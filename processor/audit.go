// SPDX-License-Identifier: MIT

package processor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// isSQLiteDB checks if a file is a SQLite database by checking the header.
func isSQLiteDB(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 16)
	_, err = file.Read(header)
	if err != nil {
		return false
	}

	return string(header) == "SQLite format 3\x00"
}

// connectSqliteDbReadOnly opens the database without running migrations or changing journal
// mode, as an audit should never modify the thing it is auditing. A single connection is used
// as the temporary tables used during the audit only exist for the connection that created them
func connectSqliteDbReadOnly(pathName string) (*sql.DB, error) {
	abs, err := filepath.Abs(pathName)
	if err != nil {
		return nil, err
	}

	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs), RawQuery: "mode=ro"}
	db, err := sql.Open("sqlite", u.String())
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

	// a database of an empty directory has no records, and passes when audited against one
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

// sqlitePathCandidates returns the paths a file on disk may be stored under in the database,
// in order of preference. Relative paths are stored relative to where hashit was invoked from
func sqlitePathCandidates(file string, invokedFrom string) []string {
	var candidates []string
	seen := map[string]bool{}
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			candidates = append(candidates, p)
		}
	}

	// relative paths may have been stored with or without a leading ./ depending on how
	// hashit was run, so try both
	addRelative := func(p string) {
		add(p)
		if !filepath.IsAbs(p) && !strings.HasPrefix(p, "..") {
			add("." + string(filepath.Separator) + p)
		}
	}

	clean := filepath.Clean(file)
	add(file)
	addRelative(clean)

	abs, err := filepath.Abs(clean)
	if err != nil {
		return candidates
	}

	if invokedFrom != "" {
		if rel, err := filepath.Rel(invokedFrom, abs); err == nil && !strings.HasPrefix(rel, "..") {
			addRelative(rel)
		}
	}
	add(abs)
	return candidates
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

	// databases created before metadata was added do not have this, in which case only
	// paths relative to the current directory or absolute paths will match
	invokedFrom, _ := queries.MetadataGet(ctx, metadataInvokedFrom)

	_, err = db.Exec(`create temp table seen (
    filepath text primary key,
    status integer not null,
    disk_file text not null
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
	insertSeen, err := tx.PrepareContext(ctx, "insert or ignore into temp.seen (filepath, status, disk_file) values (?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		printError(fmt.Sprintf("failed to prepare audit statement: %s", err.Error()))
		return "", false
	}
	txQueries := queries.WithTx(tx)

	examinedCount := 0
	matched := 0
	filesModified := 0

	// new files are usually few, so are held in memory indexed by hash to identify moves
	var newFileCandidates []Result

	seenFiles := map[string]bool{}
	for res := range input {
		// overlapping paths such as . and ./sub supply the same file twice
		if isDuplicateFile(seenFiles, res.File) {
			continue
		}
		examinedCount++

		var record *AuditRecord
		for _, p := range sqlitePathCandidates(res.File, invokedFrom) {
			row, err := txQueries.FileHashByFilePath(ctx, p)
			if err == nil {
				record = fileHashToRecord(row)
				break
			}
			if !errors.Is(err, sql.ErrNoRows) {
				printError(fmt.Sprintf("failed to query audit database for file %s: %s", res.File, err.Error()))
			}
		}

		if record == nil {
			newFileCandidates = append(newFileCandidates, res)
			continue
		}

		status := FileModified
		if recordMatches(record, res) {
			status = FileMatched
			matched++
		} else {
			filesModified++
		}

		if _, err := insertSeen.ExecContext(ctx, record.Filename, int(status), res.File); err != nil {
			printError(fmt.Sprintf("failed to record audit result for file %s: %s", res.File, err.Error()))
		}
	}

	_ = insertSeen.Close()
	if err := tx.Commit(); err != nil {
		printError(fmt.Sprintf("failed to commit audit results: %s", err.Error()))
		return "", false
	}

	// index new files by the strongest hash we calculated which every record also has
	hashes := sqliteCompareHashes
	moveIndex := map[string][]int{}
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
order by f.filepath`)
	if err != nil {
		printError(fmt.Sprintf("failed to query audit database: %s", err.Error()))
		return "", false
	}
	defer rows.Close()

	// files which exist but could not be read are reported as such rather than as not found
	unreadable := map[string]bool{}
	var failedList []string
	failedFiles.Range(func(key, _ any) bool {
		failedList = append(failedList, key.(string))
		for _, p := range sqlitePathCandidates(key.(string), invokedFrom) {
			unreadable[p] = true
		}
		return true
	})
	reportedUnreadable := map[string]bool{}
	isUnreadable := func(filename string) bool {
		if unreadable[filename] {
			reportedUnreadable[filename] = true
			return true
		}
		abs := filename
		if !filepath.IsAbs(abs) && invokedFrom != "" {
			abs = filepath.Join(invokedFrom, abs)
		}
		if a, err := filepath.Abs(abs); err == nil && isUnderFailedDir(a) {
			return true
		}
		return false
	}

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
		sort.Strings(failedList)
		for _, f := range failedList {
			known := false
			for _, p := range sqlitePathCandidates(f, invokedFrom) {
				if reportedUnreadable[p] {
					known = true
				}
			}
			if !known {
				fmt.Fprintf(stdout, "%v: File could not be read\n", f)
			}
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
