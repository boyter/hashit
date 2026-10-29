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
	abs, err := filepath.Abs(pathName)
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
// preferred directory, then the first record. It is held in a temporary table rather than memory
// so databases larger than memory can be audited
func buildSqlitePathIndex(ctx context.Context, db *sql.DB, bases []string) error {
	if _, err := db.ExecContext(ctx, `create temp table paths (
    abs text primary key,
    filepath text not null,
    level integer not null
)`); err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	insert, err := tx.PrepareContext(ctx, `insert into temp.paths (abs, filepath, level) values (?, ?, ?)
on conflict (abs) do update set filepath = excluded.filepath, level = excluded.level where excluded.level < paths.level`)
	if err != nil {
		return err
	}
	defer insert.Close()

	rows, err := tx.QueryContext(ctx, "select filepath from file_hashes order by rowid")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return err
		}
		var keys []string
		for level, base := range bases {
			k := absFrom(base, filePath)
			if contains(keys, k) {
				continue
			}
			keys = append(keys, k)
			if _, err := insert.ExecContext(ctx, k, filePath, level); err != nil {
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

	cwd, err := os.Getwd()
	if err != nil {
		printError(fmt.Sprintf("unable to determine the working directory: %s", err.Error()))
		return "", false
	}

	// databases created before metadata was added do not have this, in which case paths are
	// relative to the working directory
	invokedFrom, _ := queries.MetadataGet(ctx, metadataInvokedFrom)
	bases := auditBases(invokedFrom, AuditFile, cwd)

	if err := buildSqlitePathIndex(ctx, db, bases); err != nil {
		printError(fmt.Sprintf("failed to index audit database: %s", err.Error()))
		return "", false
	}

	// result holds the file on disk for a record matched through a directory other than the one
	// hashit was invoked from, as a file matching the record's own path may later take it over
	_, err = db.Exec(`create temp table seen (
    filepath text primary key,
    status integer not null,
    disk_file text not null,
    is_primary integer not null,
    result text
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
	insertSeen, err := tx.PrepareContext(ctx, "insert or ignore into temp.seen (filepath, status, disk_file, is_primary, result) values (?, ?, ?, ?, ?)")
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

	// lookupRecord finds the record for a file on disk by its absolute path, and whether it matched
	// through the directory hashit was invoked from
	lookupRecord := func(file string) (*AuditRecord, bool, error) {
		var filePath string
		var level int
		err := tx.QueryRowContext(ctx, "select filepath, level from temp.paths where abs = ?", absFrom(cwd, file)).Scan(&filePath, &level)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		row, err := txQueries.FileHashByFilePath(ctx, filePath)
		if err != nil {
			return nil, false, err
		}
		return fileHashToRecord(row), level == 0, nil
	}

	seenFiles := map[string]bool{}
	for res := range input {
		// overlapping paths such as . and ./sub supply the same file twice
		if isDuplicateFile(seenFiles, res.File) {
			continue
		}
		examinedCount++

		record, primary, err := lookupRecord(res.File)
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

		var stored sql.NullString
		if !primary {
			b, err := json.Marshal(res)
			if err != nil {
				printError(fmt.Sprintf("failed to record audit result for file %s: %s", res.File, err.Error()))
				continue
			}
			stored = sql.NullString{String: string(b), Valid: true}
		}

		result, err := insertSeen.ExecContext(ctx, record.Filename, int(status), res.File, primary, stored)
		if err != nil {
			printError(fmt.Sprintf("failed to record audit result for file %s: %s", res.File, err.Error()))
			continue
		}

		// Two files on disk can resolve to the same record, so which claims it must not depend on
		// the order files were processed in. A file matching the record's own path wins over one
		// matching through another directory, then the first path in sort order. The other file
		// is treated as new
		if n, _ := result.RowsAffected(); n == 0 {
			var oldStatus int
			var oldFile string
			var oldPrimary bool
			var oldResult sql.NullString
			err := tx.QueryRowContext(ctx, "select status, disk_file, is_primary, result from temp.seen where filepath = ?", record.Filename).Scan(&oldStatus, &oldFile, &oldPrimary, &oldResult)
			if err != nil {
				printError(fmt.Sprintf("failed to read audit result for file %s: %s", res.File, err.Error()))
				continue
			}

			var old Result
			wins := oldResult.Valid && (primary && !oldPrimary || primary == oldPrimary && res.File < oldFile)
			if wins {
				if err := json.Unmarshal([]byte(oldResult.String), &old); err != nil {
					wins = false
				}
			}
			if !wins {
				newFileCandidates = append(newFileCandidates, res)
				continue
			}

			if _, err := tx.ExecContext(ctx, "update temp.seen set status = ?, disk_file = ?, is_primary = ?, result = ? where filepath = ?", int(status), res.File, primary, stored, record.Filename); err != nil {
				printError(fmt.Sprintf("failed to record audit result for file %s: %s", res.File, err.Error()))
				continue
			}
			if FileStatus(oldStatus) == FileMatched {
				matched--
			} else {
				filesModified--
			}
			newFileCandidates = append(newFileCandidates, old)
		}

		if status == FileMatched {
			matched++
		} else {
			filesModified++
		}
	}

	// files which exist but could not be read are reported as such rather than as not found. A
	// record claimed by another file is reported for that file, so the unreadable file is listed
	// on its own, as it is when it has no record
	unreadable := map[string]bool{}
	var unreadableNew []string
	var failedList []string
	failedFiles.Range(func(key, _ any) bool {
		failedList = append(failedList, key.(string))
		return true
	})
	sort.Strings(failedList)
	for _, f := range failedList {
		var filePath string
		err := tx.QueryRowContext(ctx, "select p.filepath from temp.paths p left join temp.seen s on s.filepath = p.filepath where p.abs = ? and s.filepath is null", absFrom(cwd, f)).Scan(&filePath)
		if err == nil {
			unreadable[filePath] = true
		} else {
			unreadableNew = append(unreadableNew, f)
		}
	}
	isUnreadable := func(filename string) bool {
		if unreadable[filename] {
			return true
		}
		for _, base := range bases {
			if isUnderFailedDir(absFrom(base, filename)) {
				return true
			}
		}
		return false
	}

	_ = insertSeen.Close()
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
