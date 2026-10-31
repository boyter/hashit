// SPDX-License-Identifier: MIT

package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/boyter/hashit/processor/database"
)

// metadataInvokedFrom is the metadata key holding the directory hashit was run from when
// creating the database, which relative paths in it are relative to
const metadataInvokedFrom = "invoked_from"

// metadataHashes is the metadata key holding the comma separated hashes stored in the database
const metadataHashes = "hashes"

// sqliteHeader is the start of every SQLite database
const sqliteHeader = "SQLite format 3\x00"

// isSQLiteDB checks if a file is a SQLite database by checking the header. It must only be used on
// regular files, as reading the header consumes it from a pipe
func isSQLiteDB(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, len(sqliteHeader))
	if _, err = io.ReadFull(file, header); err != nil {
		return false
	}

	return string(header) == sqliteHeader
}

// connectSqliteDbReadOnly opens the database without running migrations or changing journal
// mode, as an audit should never modify the thing it is auditing. A single connection is used
// as the temporary tables used during the audit only exist for the connection that created them
func connectSqliteDbReadOnly(pathName string) (*sql.DB, error) {
	abs, err := osAbs(pathName)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", sqliteFileURI(abs, "mode=ro"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	// temporary tables can be as large as the audit, so keep them on disk rather than memory
	if _, err = db.Exec("pragma temp_store = file; pragma busy_timeout = 5000;"); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// sqliteFileURI builds a file URI for SQLite, so a path containing ? or # is not taken as the
// start of parameters. Windows paths such as C:/x.db need a leading slash, as otherwise the drive
// would be read as the host
func sqliteFileURI(abs string, query string) string {
	path := filepath.ToSlash(abs)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := url.URL{Scheme: "file", Path: path, RawQuery: query}
	return u.String()
}

// fileHashToRecord converts a database row into an AuditRecord so it can be compared
// in the same way as hashdeep and sum audit files
func fileHashToRecord(f database.FileHash) *AuditRecord {
	r := &AuditRecord{Filename: f.Filepath, Size: f.Size, Hashes: map[string]string{}}
	add := func(name string, v sql.NullString) {
		if v.Valid && v.String != "" {
			r.Hashes[name] = strings.ToLower(v.String)
		}
	}
	add(HashNames.CRC32, f.Crc32)
	add(HashNames.XxHash64, f.Xxhash64)
	add(HashNames.MD4, f.Md4)
	add(HashNames.MD5, f.Md5)
	add(HashNames.SHA1, f.Sha1)
	add(HashNames.SHA224, f.Sha224)
	add(HashNames.SHA256, f.Sha256)
	add(HashNames.SHA384, f.Sha384)
	add(HashNames.SHA512, f.Sha512)
	add(HashNames.Blake2b256, f.Blake2b256)
	add(HashNames.Blake2b512, f.Blake2b512)
	add(HashNames.Blake3, f.Blake3)
	add(HashNames.Sha3224, f.Sha3224)
	add(HashNames.Sha3256, f.Sha3256)
	add(HashNames.Sha3384, f.Sha3384)
	add(HashNames.Sha3512, f.Sha3512)
	add(HashNames.Ed2k, f.Ed2k)
	return r
}

// sqliteAuditHashes returns the hashes stored in the audit database, strongest first. Every row
// is written by the same run, so the first row is representative
func sqliteAuditHashes(pathName string) ([]string, error) {
	db, err := connectSqliteDbReadOnly(pathName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var hasTable int
	if err := db.QueryRow("select count(*) from sqlite_master where type = 'table' and name = 'file_hashes'").Scan(&hasTable); err != nil {
		return nil, err
	}
	if hasTable == 0 {
		return nil, errors.New("audit database is not a hashit database, it has no file_hashes table")
	}

	// databases from development builds before sha224 and sha384 were added lack those columns
	var hasColumn int
	if err := db.QueryRow("select count(*) from pragma_table_info('file_hashes') where name = 'sha224'").Scan(&hasColumn); err != nil {
		return nil, err
	}
	if hasColumn == 0 {
		return nil, errors.New("audit database was created by an older development version of hashit, regenerate it with --format sqlite --overwrite")
	}

	// each path is looked up as a single row and rows are told apart by rowid, so a table made
	// some other way could have a path listed twice with only one audited, or a rowid column
	// hiding the real one
	var keyed, shadowed, withoutRowid int
	if err := db.QueryRow(`select
    (select count(*) = 1 and sum(name = 'filepath') = 1 from pragma_table_info('file_hashes') where pk > 0),
    (select count(*) from pragma_table_info('file_hashes') where lower(name) in ('rowid', 'oid', '_rowid_')),
    (select count(*) from pragma_table_list where name = 'file_hashes' and wr = 1)`).Scan(&keyed, &shadowed, &withoutRowid); err != nil {
		return nil, err
	}
	if keyed != 1 || shadowed != 0 || withoutRowid != 0 {
		return nil, errors.New("audit database was not created by hashit, its file_hashes table must have filepath as its primary key and a rowid")
	}

	// a database of an empty directory has no records, and passes when audited against one
	// databases record which hashes they hold, but those from earlier development builds do not,
	// in which case the first row is used
	if stored, err := database.New(db).MetadataGet(context.Background(), metadataHashes); err == nil && stored != "" {
		var hashes []string
		for _, name := range hashPriority {
			if contains(strings.Split(stored, ","), name) {
				hashes = append(hashes, name)
			}
		}
		if len(hashes) != 0 {
			return hashes, nil
		}
	}

	first, err := database.New(db).FileHashFirst(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	r := fileHashToRecord(first)
	var hashes []string
	for _, name := range hashPriority {
		if _, ok := r.Hashes[name]; ok {
			hashes = append(hashes, name)
		}
	}
	if len(hashes) == 0 {
		return nil, errors.New("audit database contains no hashes")
	}
	return hashes, nil
}

// buildSqlitePathIndex fills temp.paths with the absolute path every record may refer to, in the
// same priority order as other audit files: the directory hashit was invoked from, then the
// database's directory, then the working directory. A path takes the record from the most
// preferred directory, then the first record. Records written as hashit writes them, such as
// sub/a or ./sub/a, are looked up directly by sqliteLookup, so only other spellings such as
// sub//a are indexed. It is held in a temporary table rather than memory so databases larger
// than memory can be audited
func buildSqlitePathIndex(ctx context.Context, db *sql.DB, bases []string) error {
	if _, err := db.ExecContext(ctx, `create temp table paths (
    abs text primary key,
    filepath text not null,
    level integer not null,
    rid integer not null
)`); err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	insert, err := tx.PrepareContext(ctx, `insert into temp.paths (abs, filepath, level, rid) values (?, ?, ?, ?)
on conflict (abs) do update set filepath = excluded.filepath, level = excluded.level, rid = excluded.rid where excluded.level < paths.level`)
	if err != nil {
		return err
	}
	defer insert.Close()

	rows, err := tx.QueryContext(ctx, "select rowid, filepath from file_hashes order by rowid")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rid int64
		var filePath string
		if err := rows.Scan(&rid, &filePath); err != nil {
			return err
		}
		direct := isDirectSpelling(filePath)
		if direct && len(workingDirAncestors()) == 0 {
			continue
		}
		var keys []string
		for level, base := range bases {
			k := absFrom(base, filePath)
			if contains(keys, k) {
				continue
			}
			keys = append(keys, k)
			// when the working directory was reached through a symlink a path can be compared
			// as another spelling, which sqliteLookup does not try, so only those are indexed
			if direct && k == absPath(base, filePath) {
				continue
			}
			if _, err := insert.ExecContext(ctx, k, filePath, level, rid); err != nil {
				return err
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_ = rows.Close()
	return tx.Commit()
}

// isDirectSpelling reports whether sqliteLookup finds a record by its path without the index,
// being a clean absolute path, or a clean relative path without .. optionally starting with ./
func isDirectSpelling(p string) bool {
	if filepath.IsAbs(p) {
		return filepath.Clean(p) == p
	}
	// Windows paths such as \a and D:a are relative to a drive, so are left to the index
	if p == "" || filepath.VolumeName(p) != "" || os.IsPathSeparator(p[0]) {
		return false
	}
	// sqliteLookup only tries the path relative to each directory, which never goes up and back
	// down again such as ../sub/a, so any path with .. is left to the index
	p = strings.TrimPrefix(p, "."+string(filepath.Separator))
	if p == "" || os.IsPathSeparator(p[0]) || p == ".." || strings.HasPrefix(p, ".."+string(filepath.Separator)) {
		return false
	}
	return filepath.Clean(p) == p
}

// sqliteLookup finds the record for the absolute path of a file on disk, returning its path in
// the database and the priority of the directory it matched through, or false if none matches
func sqliteLookup(ctx context.Context, tx *sql.Tx, bases []string, abs string) (string, int, bool, error) {
	for level, base := range bases {
		var candidates []any
		if level == 0 {
			candidates = append(candidates, abs)
		}
		// the path relative to base is only used when it leads back to the same file, as
		// relative paths are compared exactly as absFrom resolves them
		if rel, err := filepath.Rel(base, abs); err == nil && rel != "." && absFrom(base, rel) == abs {
			candidates = append(candidates, rel, "."+string(filepath.Separator)+rel)
		}
		if len(candidates) == 0 {
			candidates = append(candidates, nil)
		}
		in := strings.Repeat("?, ", len(candidates)-1) + "?"

		var filePath string
		err := tx.QueryRowContext(ctx, `select filepath from (
    select rowid as rid, filepath from file_hashes where filepath in (`+in+`)
    union all
    select rid, filepath from temp.paths where abs = ? and level = ?
) order by rid limit 1`, append(candidates, abs, level)...).Scan(&filePath)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return "", 0, false, err
		}
		return filePath, level, true, nil
	}
	return "", 0, false, nil
}

// claimSqliteMatches decides which file on disk claims each record once the directory the audit
// is matched through is chosen, see chooseLevel, recording it in temp.seen. Files matching through
// another directory and files losing a record to another are returned so they are treated as new.
// Two files on disk can resolve to the same record, so which claims it does not depend on the
// order files were processed in. A file matching through a directory earlier in priority order
// wins, being the record's own path before another directory, then the first path in sort order
func claimSqliteMatches(ctx context.Context, tx *sql.Tx, chosen int) (int, int, []Result, error) {
	if _, err := tx.ExecContext(ctx, `create temp table seen (
    filepath text primary key,
    status integer not null,
    disk_file text not null unique
)`); err != nil {
		return 0, 0, nil, err
	}
	if _, err := tx.ExecContext(ctx, `insert into temp.seen (filepath, status, disk_file)
select filepath, status, disk_file from (
    select filepath, status, disk_file, row_number() over (partition by filepath order by level, disk_file) as n
    from temp.matches where independent = 1 or ? = -1 or level = ?
) where n = 1`, chosen, chosen); err != nil {
		return 0, 0, nil, err
	}

	matched, modified := 0, 0
	rows, err := tx.QueryContext(ctx, "select status, count(*) from temp.seen group by status")
	if err != nil {
		return 0, 0, nil, err
	}
	for rows.Next() {
		var status, n int
		if err := rows.Scan(&status, &n); err != nil {
			_ = rows.Close()
			return 0, 0, nil, err
		}
		if FileStatus(status) == FileMatched {
			matched += n
		} else {
			modified += n
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, 0, nil, err
	}
	_ = rows.Close()

	rows, err = tx.QueryContext(ctx, "select disk_file, result from temp.matches where disk_file not in (select disk_file from temp.seen)")
	if err != nil {
		return 0, 0, nil, err
	}
	defer rows.Close()
	var unclaimed []Result
	for rows.Next() {
		var diskFile, result string
		if err := rows.Scan(&diskFile, &result); err != nil {
			return 0, 0, nil, err
		}
		var res Result
		if err := json.Unmarshal([]byte(result), &res); err != nil {
			return 0, 0, nil, err
		}
		// JSON replaces bytes which are not UTF-8, so the name is taken from disk_file
		res.File = diskFile
		unclaimed = append(unclaimed, res)
	}
	return matched, modified, unclaimed, rows.Err()
}

// markSqliteDuplicates records in temp.duplicates the rows no file on disk was checked against
// which name the same path through the directory the audit is matched through as another row with
// the same values, such as from hashing . and ./sub together, as otherwise the row is never looked
// up and is reported as missing although the file exists. Of such rows the earliest is kept when
// none was checked. A row with other values, or naming another path through that directory such as
// an absolute path into a tree which was copied elsewhere, is still reported. When no directory was
// chosen rows naming the same path through any directory are the same, see judgingBases
func markSqliteDuplicates(ctx context.Context, tx *sql.Tx, bases []string) (int, error) {
	if _, err := tx.ExecContext(ctx, "create temp table duplicates (rid integer primary key)"); err != nil {
		return 0, err
	}
	var unseen bool
	if err := tx.QueryRowContext(ctx, "select exists (select 1 from file_hashes where filepath not in (select filepath from temp.seen))").Scan(&unseen); err != nil || !unseen {
		return 0, err
	}
	// only rows whose name and values are listed more than once can be duplicates, so when many
	// files are missing they are not all resolved. Values are compared as fileHashToRecord reads
	// them, and the name is what follows the last separator. A row this misses is still reported
	path := `replace(filepath, '\', '/')`
	values := []string{"substr(" + path + ", length(rtrim(" + path + ", replace(" + path + ", '/', ''))) + 1)"}
	for _, column := range []string{"crc32", "xxhash64", "md4", "md5", "sha1", "sha224", "sha256", "sha384", "sha512",
		"blake2b_256", "blake2b_512", "blake3", "sha3_224", "sha3_256", "sha3_384", "sha3_512", "ed2k"} {
		values = append(values, "lower(ifnull("+column+", ''))")
	}
	key := "size || '|' || " + strings.Join(values, " || '|' || ")
	if _, err := tx.ExecContext(ctx, "create temp table repeated (rid integer primary key, key text not null)"); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `insert into temp.repeated select f.rowid, `+key+` from file_hashes f
where `+key+` in (select `+key+` from file_hashes group by 1 having count(*) > 1)`); err != nil {
		return 0, err
	}

	// rows with the same values are read together, so only the paths of one set are held at a time
	rows, err := tx.QueryContext(ctx, `select r.rid, r.key, f.filepath, f.filepath in (select filepath from temp.seen)
from temp.repeated r join file_hashes f on f.rowid = r.rid order by r.key, r.rid`)
	if err != nil {
		return 0, err
	}
	type row struct {
		rid     int64
		checked bool
		keys    []string
	}
	var duplicates []int64
	var set []row
	// rows of a set naming the same path through any of the directories are grouped together
	settle := func() {
		group := make([]int, len(set))
		for i := range group {
			group[i] = i
		}
		find := func(i int) int {
			for group[i] != i {
				group[i] = group[group[i]]
				i = group[i]
			}
			return i
		}
		owner := map[string]int{}
		for i, r := range set {
			for _, k := range r.keys {
				if j, ok := owner[k]; ok {
					a, b := find(i), find(j)
					group[max(a, b)] = min(a, b)
				} else {
					owner[k] = i
				}
			}
		}
		checked := map[int]bool{}
		for i, r := range set {
			if r.checked {
				checked[find(i)] = true
			}
		}
		// rows are in rowid order, so the earliest of a group none was checked against is its root
		for i, r := range set {
			if g := find(i); !r.checked && (checked[g] || g != i) {
				duplicates = append(duplicates, r.rid)
			}
		}
		set = set[:0]
	}
	current := ""
	for rows.Next() {
		var r row
		var values, filePath string
		if err := rows.Scan(&r.rid, &values, &filePath, &r.checked); err != nil {
			_ = rows.Close()
			return 0, err
		}
		if values != current {
			settle()
			current = values
		}
		for _, base := range bases {
			r.keys = append(r.keys, absFrom(base, filePath))
		}
		set = append(set, r)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, err
	}
	_ = rows.Close()
	settle()

	for _, rid := range duplicates {
		if _, err := tx.ExecContext(ctx, "insert into temp.duplicates (rid) values (?)", rid); err != nil {
			return 0, err
		}
	}
	return len(duplicates), nil
}

// doSqliteAudit performs an audit against a SQLite database, comparing all available hashes.
// To scale to databases larger than available memory, which files have been seen is tracked
// in a temporary table rather than in memory, and the final report streams over the database
func doSqliteAudit(input chan Result) (string, bool) {
	db, err := connectSqliteDbReadOnly(AuditFile)
	if err != nil {
		printError(fmt.Sprintf("failed to open audit database: %s", err.Error()))
		return "", false
	}
	defer db.Close()

	ctx := context.Background()
	queries := database.New(db)

	cwd := workingDir()

	// databases created before metadata was added do not have this, in which case paths are
	// relative to the working directory
	invokedFrom, _ := queries.MetadataGet(ctx, metadataInvokedFrom)
	bases := auditBases(invokedFrom, AuditFile, cwd)

	if err := buildSqlitePathIndex(ctx, db, bases); err != nil {
		printError(fmt.Sprintf("failed to index audit database: %s", err.Error()))
		return "", false
	}

	// every file matching a record is held until the directory the audit is matched through is
	// chosen, see chooseLevel, with its result so it can be treated as new
	_, err = db.Exec(`create temp table matches (
    disk_file text primary key,
    filepath text not null,
    status integer not null,
    level integer not null,
    independent integer not null,
    result text not null
)`)
	if err != nil {
		printError(fmt.Sprintf("failed to create temporary audit table: %s", err.Error()))
		return "", false
	}

	// wrap inserts in a transaction as otherwise every insert is its own sync to disk
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		printError(fmt.Sprintf("failed to begin audit transaction: %s", err.Error()))
		return "", false
	}
	insertMatch, err := tx.PrepareContext(ctx, "insert or ignore into temp.matches (disk_file, filepath, status, level, independent, result) values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		printError(fmt.Sprintf("failed to prepare audit statement: %s", err.Error()))
		return "", false
	}
	txQueries := queries.WithTx(tx)

	examinedCount := 0

	// new files are usually few, so are held in memory indexed by hash to identify moves
	var newFileCandidates []Result

	// lookupRecord finds the record for a file on disk by its absolute path, with the priority of
	// the directory it matched through and whether its path names the file through every directory
	lookupRecord := func(file string) (*AuditRecord, int, bool, error) {
		abs := absFrom(cwd, file)
		filePath, level, ok, err := sqliteLookup(ctx, tx, bases, abs)
		if err != nil || !ok {
			return nil, 0, false, err
		}
		row, err := txQueries.FileHashByFilePath(ctx, filePath)
		if err != nil {
			return nil, 0, false, err
		}
		// an absolute path names the same file through every directory, see baseLevel
		return fileHashToRecord(row), level, filepath.IsAbs(filePath), nil
	}
	votes := map[int]int{}

	seenFiles := map[string]bool{}
	for res := range input {
		// overlapping paths such as . and ./sub supply the same file twice
		if isDuplicateFile(seenFiles, res.File) {
			continue
		}
		examinedCount++

		record, level, independent, err := lookupRecord(res.File)
		if err != nil {
			printError(fmt.Sprintf("failed to query audit database for file %s: %s", res.File, err.Error()))
			continue
		}
		if record == nil {
			newFileCandidates = append(newFileCandidates, res)
			continue
		}

		status := FileModified
		if recordMatches(record, res) {
			status = FileMatched
		}
		b, err := json.Marshal(res)
		if err != nil {
			printError(fmt.Sprintf("failed to record audit result for file %s: %s", res.File, err.Error()))
			continue
		}
		if _, err := insertMatch.ExecContext(ctx, res.File, record.Filename, int(status), level, independent, string(b)); err != nil {
			printError(fmt.Sprintf("failed to record audit result for file %s: %s", res.File, err.Error()))
			continue
		}
		if !independent {
			votes[level]++
		}
	}
	_ = insertMatch.Close()

	chosen := chooseLevel(votes)
	matched, filesModified, unclaimed, err := claimSqliteMatches(ctx, tx, chosen)
	if err != nil {
		_ = tx.Rollback()
		printError(fmt.Sprintf("failed to read audit results: %s", err.Error()))
		return "", false
	}
	newFileCandidates = append(newFileCandidates, unclaimed...)

	// files which exist but could not be read are reported as such rather than as not found. A
	// record claimed by another file is reported for that file, so the unreadable file is listed
	// on its own, as it is when it has no record or its record was moved to another file
	unreadable := map[string]string{}
	unreadableAt := map[string]int{} // unreadable file to the directory its record was found through
	var unreadableNew []string
	var failedList []string
	failedFiles.Range(func(key, _ any) bool {
		failedList = append(failedList, key.(string))
		return true
	})
	sort.Strings(failedList)
	for _, f := range failedList {
		abs := absFrom(cwd, f)
		filePath, level, ok, err := sqliteLookup(ctx, tx, bases, abs)
		if err == nil && ok {
			unreadableAt[abs] = level
			var claimed int
			err = tx.QueryRowContext(ctx, "select count(*) from temp.seen where filepath = ?", filePath).Scan(&claimed)
			ok = err == nil && claimed == 0
		}
		if ok {
			unreadable[filePath] = f
		} else {
			unreadableNew = append(unreadableNew, f)
		}
	}
	isUnreadable := func(filename string) bool {
		if _, ok := unreadable[filename]; ok {
			return true
		}
		// another spelling of an unreadable file, which the lookup only returns one record for,
		// through the directory its record was found through, so a path through another directory
		// such as the working directory is not taken as another record's unreadable file
		for level, base := range bases {
			if at, ok := unreadableAt[absFrom(base, filename)]; ok && at == level {
				return true
			}
		}
		for _, base := range bases {
			if isUnderFailedDir(absFrom(base, filename)) {
				return true
			}
		}
		return false
	}

	duplicates, err := markSqliteDuplicates(ctx, tx, judgingBases(bases, chosen))
	if err != nil {
		_ = tx.Rollback()
		printError(fmt.Sprintf("failed to check audit database for duplicates: %s", err.Error()))
		return "", false
	}
	if duplicates != 0 {
		printWarn(fmt.Sprintf("audit database lists the same path more than once, only the first is used for each, duplicates: %d", duplicates))
	}
	if err := tx.Commit(); err != nil {
		printError(fmt.Sprintf("failed to commit audit results: %s", err.Error()))
		return "", false
	}

	// index new files by the strongest hash we calculated which every record also has
	hashes := sqliteCompareHashes
	moveIndex := map[string][]int{}
	// in path order, so which of several identical files is reported as moved is always the same
	sort.Slice(newFileCandidates, func(i, j int) bool { return newFileCandidates[i].File < newFileCandidates[j].File })
	if len(hashes) != 0 {
		for i, res := range newFileCandidates {
			if v := res.hashValue(hashes[0]); v != "" {
				moveIndex[v] = append(moveIndex[v], i)
			}
		}
	}
	movedTo := make([]bool, len(newFileCandidates))

	// stream every record in path order joined against what we saw, reporting as we go
	rows, err := db.QueryContext(ctx, `select f.filepath, f.crc32, f.xxhash64, f.md4, f.md5, f.sha1, f.sha224, f.sha256, f.sha384, f.sha512,
    f.blake2b_256, f.blake2b_512, f.blake3, f.sha3_224, f.sha3_256, f.sha3_384, f.sha3_512, f.ed2k, f.size, f.mtime,
    s.status, s.disk_file
from file_hashes f left join temp.seen s on s.filepath = f.filepath
where f.rowid not in (select rid from temp.duplicates)
order by f.filepath`)
	if err != nil {
		printError(fmt.Sprintf("failed to query audit database: %s", err.Error()))
		return "", false
	}
	defer rows.Close()

	knownFileCount := 0
	moved := 0
	filesMissing := 0

	for rows.Next() {
		var f database.FileHash
		var status sql.NullInt64
		var diskFile sql.NullString
		err := rows.Scan(&f.Filepath, &f.Crc32, &f.Xxhash64, &f.Md4, &f.Md5, &f.Sha1, &f.Sha224, &f.Sha256, &f.Sha384, &f.Sha512,
			&f.Blake2b256, &f.Blake2b512, &f.Blake3, &f.Sha3224, &f.Sha3256, &f.Sha3384, &f.Sha3512, &f.Ed2k,
			&f.Size, &f.Mtime, &status, &diskFile)
		if err != nil {
			printError(fmt.Sprintf("failed to read audit database: %s", err.Error()))
			return "", false
		}
		knownFileCount++

		if status.Valid {
			switch FileStatus(status.Int64) {
			case FileMatched:
				if VeryVerbose {
					fmt.Fprintf(stdout, "%v: Ok\n", diskFile.String)
				}
			case FileModified:
				if Verbose {
					fmt.Fprintf(stdout, "%v: File modified\n", diskFile.String)
				}
			}
			continue
		}

		// not seen on disk, so either moved or missing
		record := fileHashToRecord(f)
		foundMove := false
		if len(hashes) != 0 {
			for _, i := range moveIndex[record.Hashes[hashes[0]]] {
				if !movedTo[i] && recordMatches(record, newFileCandidates[i]) {
					movedTo[i] = true
					foundMove = true
					moved++
					if Verbose {
						fmt.Fprintf(stdout, "%v -> %v: File moved\n", record.Filename, newFileCandidates[i].File)
					}
					if f, ok := unreadable[record.Filename]; ok {
						unreadableNew = append(unreadableNew, f)
					}
					break
				}
			}
		}

		if !foundMove {
			filesMissing++
			if Verbose {
				if isUnreadable(record.Filename) {
					fmt.Fprintf(stdout, "%v: File could not be read\n", record.Filename)
				} else {
					fmt.Fprintf(stdout, "%v: File expected but not found\n", record.Filename)
				}
			}
		}
	}
	if err := rows.Err(); err != nil {
		printError(fmt.Sprintf("failed to read audit database: %s", err.Error()))
		return "", false
	}

	var newFiles []Result
	for i, res := range newFileCandidates {
		if !movedTo[i] {
			newFiles = append(newFiles, res)
		}
	}

	if Verbose {
		for _, res := range sortedPaths(newFiles) {
			fmt.Fprintf(stdout, "%v: File new\n", res.File)
		}
		// files on disk which are not in the database and could not be read
		sort.Strings(unreadableNew)
		for _, f := range unreadableNew {
			fmt.Fprintf(stdout, "%v: File could not be read\n", f)
		}
	}

	status := Passed
	if auditFailed(filesModified, moved, len(newFiles), filesMissing) {
		status = Failed
	}

	return fmt.Sprintf(`hashit: SQLite Audit %s
       Files examined: %d
Known files expecting: %d
        Files matched: %d
       Files modified: %d
          Files moved: %d
      New files found: %d
        Files missing: %d`+"\n", status, examinedCount, knownFileCount, matched, filesModified, moved, len(newFiles), filesMissing), status == Passed
}
