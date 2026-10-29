// SPDX-License-Identifier: MIT

package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/boyter/hashit/assets"
	"github.com/boyter/hashit/processor/database"
	_ "modernc.org/sqlite"
)

// Get the time as standard UTC/Zulu format
func getFormattedTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Prints a message to stderr if flag to enable warning output is set
func printVerbose(msg string) {
	if Verbose {
		_, _ = fmt.Fprintf(os.Stderr, "VERBOSE %s: %s\n", getFormattedTime(), msg)
	}
}

// Prints a message to stderr if flag to enable debug output is set, so it does not mix with output
func printDebug(msg string) {
	if Debug {
		_, _ = fmt.Fprintf(os.Stderr, "DEBUG %s: %s\n", getFormattedTime(), msg)
	}
}

// checkedWriter records the first error writing to it, so output which could not be written,
// such as to a full disk, is reported rather than silently lost
type checkedWriter struct {
	w   io.Writer
	mu  sync.Mutex
	err error
}

func (c *checkedWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if err != nil {
		c.mu.Lock()
		if c.err == nil {
			c.err = err
		}
		c.mu.Unlock()
	}
	return n, err
}

// stdout is used for all results written to standard output
var stdout = &checkedWriter{w: os.Stdout}

// Exit codes, so scripts can tell a failed audit apart from hashit being unable to do its job
const (
	ExitOK          = 0 // everything processed, and any audit passed
	ExitAuditFailed = 1 // everything processed, but the audit failed
	ExitError       = 2 // an error was printed, such as an invalid option or unreadable file
)

// errorOccurred is set by every printError, so the exit code reflects any error regardless of
// whether processing stopped or continued
var errorOccurred atomic.Bool

// Used for failures, whether processing stops or continues and later exits with ExitError
func printError(msg string) {
	errorOccurred.Store(true)
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("ERROR %s: %s", getFormattedTime(), msg))
}

// PrintError is printError for use outside the package
func PrintError(msg string) {
	printError(msg)
}

// exitCode returns the code to exit with once processing is done
func exitCode(valid bool) int {
	if errorOccurred.Load() {
		return ExitError
	}
	if !valid {
		return ExitAuditFailed
	}
	return ExitOK
}

// Used for problems which do not stop processing, such as a flag being ignored
func printWarn(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("WARN %s: %s", getFormattedTime(), msg))
}

// fileErrors is set when any file could not be processed, so that we exit non zero at the end
// in the same way md5sum and similar tools do, rather than stopping at the first error
var fileErrors atomic.Bool

// failedFiles holds the paths of files which could not be read, so an audit can report them as
// unreadable rather than missing
var failedFiles sync.Map

// failedDirs holds the absolute paths of directories which could not be read
var failedDirs sync.Map

// isUnderFailedDir reports if the absolute path is inside a directory which could not be read
func isUnderFailedDir(abs string) bool {
	found := false
	failedDirs.Range(func(key, _ any) bool {
		if isAncestorOrSelf(key.(string), abs) {
			found = true
			return false
		}
		return true
	})
	return found
}

// fileFailed reports a file which could not be read and records it for the audit and exit code
func fileFailed(path string, msg string) {
	failedFiles.Store(path, true)
	printFileError(msg)
}

// printFileError reports a file which could not be processed and records it for the exit code
func printFileError(msg string) {
	fileErrors.Store(true)
	printError(msg)
}

// Prints a message to stderr if flag to enable trace output is set, so it does not mix with output
func printTrace(msg string) {
	if Trace {
		_, _ = fmt.Fprintf(os.Stderr, "TRACE %s: %s\n", getFormattedTime(), msg)
	}
}

// Returns the current time as a millisecond timestamp
func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Returns the current time as a nanosecond timestamp as some things
// are far too fast to measure using nanoseconds
func makeTimestampNano() int64 {
	return time.Now().UnixNano()
}

func fileSummarize(input chan Result) (string, bool) {
	if AuditFile != "" {
		if isSQLiteDB(AuditFile) {
			return doSqliteAudit(input)
		}
		return doAudit(input)
	}

	switch {
	case strings.ToLower(Format) == "json":
		return toJSON(input), true
	case strings.ToLower(Format) == "hashdeep":
		return toHashDeep(input), true
	case strings.ToLower(Format) == "sum": // Similar to md5sum sha1sum output format
		return toSum(input), true
	case strings.ToLower(Format) == "hashonly":
		return toHashOnly(input)
	case strings.ToLower(Format) == "sqlite":
		return toSqlite(input)
	}

	return toText(input)
}

const (
	Passed = "passed"
	Failed = "failed"
)

// isDuplicateFile reports if the file has already been seen, recording it if not
func isDuplicateFile(seen map[string]bool, file string) bool {
	key, err := filepath.Abs(file)
	if err != nil {
		key = filepath.Clean(file)
	}
	if seen[key] {
		return true
	}
	seen[key] = true
	return false
}

// auditFailed determines if an audit failed. Same as hashdeep, the audit only passes when everything
// is exactly where it was, unless moved files have been allowed
func auditFailed(modified, moved, newFiles, missing int) bool {
	// a file which could not be read or a path which does not exist means the audit is incomplete
	if fileErrors.Load() {
		return true
	}
	if moved > 0 && !AllowMoved {
		return true
	}
	return modified > 0 || newFiles > 0 || missing > 0
}

func doAudit(input chan Result) (string, bool) {
	examinedCount := 0
	matched := 0
	filesModified := 0
	moved := 0

	// Files are processed in parallel so arrive in any order. We hold the outcome for each
	// record and report at the end in the order of the audit file, followed by new files
	type outcome struct {
		status FileStatus
		file   string // file on disk, which differs from the record for moved files
	}
	outcomes := map[*AuditRecord]outcome{}
	var newFileCandidates []Result

	// Two files on disk can resolve to the same record, so decide which claims it once every file
	// has arrived, so the result does not depend on the order files were processed in. A file
	// matching the record's own path wins over one matching through another directory, then the
	// first path in sort order
	type claim struct {
		res     Result
		primary bool
	}
	claims := map[*AuditRecord]claim{}

	seen := map[string]bool{}
	for res := range input {
		// overlapping paths such as . and ./sub supply the same file twice
		if isDuplicateFile(seen, res.File) {
			continue
		}
		examinedCount++

		r, primary := auditor.LookupPrimary(res.File)
		if r == nil {
			newFileCandidates = append(newFileCandidates, res)
			continue
		}

		existing, ok := claims[r]
		if !ok {
			claims[r] = claim{res: res, primary: primary}
			continue
		}
		if primary != existing.primary && primary || primary == existing.primary && res.File < existing.res.File {
			claims[r] = claim{res: res, primary: primary}
			newFileCandidates = append(newFileCandidates, existing.res)
		} else {
			newFileCandidates = append(newFileCandidates, res)
		}
	}

	for r, c := range claims {
		r.Matched = true
		if recordMatches(r, c.res) {
			matched++
			outcomes[r] = outcome{status: FileMatched, file: c.res.File}
		} else {
			filesModified++
			outcomes[r] = outcome{status: FileModified, file: c.res.File}
		}
	}

	// With every file on disk accounted for, anything new may be a known file which was
	// moved or renamed, which we identify by matching size and hashes against unmatched records
	var newFiles []Result
	// in path order, so which of several identical files is reported as moved is always the same
	sort.Slice(newFileCandidates, func(i, j int) bool { return newFileCandidates[i].File < newFileCandidates[j].File })
	for _, res := range newFileCandidates {
		if r := auditor.FindMoved(res); r != nil {
			moved++
			outcomes[r] = outcome{status: FileMoved, file: res.File}
		} else {
			newFiles = append(newFiles, res)
		}
	}

	// files which exist but could not be read are reported as such rather than as not found
	unreadable := map[*AuditRecord]bool{}
	var unreadableNew []string
	failedFiles.Range(func(key, _ any) bool {
		// a record claimed by another file is reported for that file, so the unreadable file is
		// listed on its own, as it is when it has no record
		if r := auditor.Lookup(key.(string)); r != nil {
			if _, claimed := outcomes[r]; !claimed {
				unreadable[r] = true
				return true
			}
		}
		unreadableNew = append(unreadableNew, key.(string))
		return true
	})
	isUnreadable := func(r *AuditRecord) bool {
		if unreadable[r] {
			return true
		}
		for _, k := range auditor.recordKeys(r) {
			if filepath.IsAbs(k) && isUnderFailedDir(k) {
				return true
			}
		}
		return false
	}

	filesMissing := 0
	for _, r := range auditor.Records() {
		o, ok := outcomes[r]
		if !ok {
			filesMissing++
			if Verbose {
				if isUnreadable(r) {
					fmt.Fprintf(stdout, "%v: File could not be read\n", r.Filename)
				} else {
					fmt.Fprintf(stdout, "%v: File expected but not found\n", r.Filename)
				}
			}
			continue
		}

		switch o.status {
		case FileMatched:
			if VeryVerbose {
				fmt.Fprintf(stdout, "%v: Ok\n", o.file)
			}
		case FileModified:
			if Verbose {
				fmt.Fprintf(stdout, "%v: File modified\n", o.file)
			}
		case FileMoved:
			if Verbose {
				fmt.Fprintf(stdout, "%v -> %v: File moved\n", r.Filename, o.file)
			}
		}
	}

	if Verbose {
		for _, res := range sortedPaths(newFiles) {
			fmt.Fprintf(stdout, "%v: File new\n", res.File)
		}
		// files on disk which are not in the audit file and could not be read
		sort.Strings(unreadableNew)
		for _, f := range unreadableNew {
			fmt.Fprintf(stdout, "%v: File could not be read\n", f)
		}
	}

	status := Passed
	if auditFailed(filesModified, moved, len(newFiles), filesMissing) {
		status = Failed
	}

	// the below is output based on what we get from hashdeep
	// verbose (not very verbose)
	return fmt.Sprintf(`hashit: Audit %s
       Files examined: %d
Known files expecting: %d
        Files matched: %d
       Files modified: %d
          Files moved: %d
      New files found: %d
        Files missing: %d`+"\n", status, examinedCount, auditor.Count(), matched, filesModified, moved, len(newFiles), filesMissing), status == Passed
}

// Mimics how md5sum sha1sum etc... work
func toSum(input chan Result) string {
	var str strings.Builder

	for res := range input {
		if hasHash(HashNames.CRC32) {
			str.WriteString(sumLine(res.CRC32, res.File))
		}
		if hasHash(HashNames.XxHash64) {
			str.WriteString(sumLine(res.XxHash64, res.File))
		}
		if hasHash(HashNames.MD4) {
			str.WriteString(sumLine(res.MD4, res.File))
		}
		if hasHash(HashNames.MD5) {
			str.WriteString(sumLine(res.MD5, res.File))
		}
		if hasHash(HashNames.SHA1) {
			str.WriteString(sumLine(res.SHA1, res.File))
		}
		if hasHash(HashNames.SHA224) {
			str.WriteString(sumLine(res.SHA224, res.File))
		}
		if hasHash(HashNames.SHA256) {
			str.WriteString(sumLine(res.SHA256, res.File))
		}
		if hasHash(HashNames.SHA384) {
			str.WriteString(sumLine(res.SHA384, res.File))
		}
		if hasHash(HashNames.SHA512) {
			str.WriteString(sumLine(res.SHA512, res.File))
		}
		if hasHash(HashNames.Blake2b256) {
			str.WriteString(sumLine(res.Blake2b256, res.File))
		}
		if hasHash(HashNames.Blake2b512) {
			str.WriteString(sumLine(res.Blake2b512, res.File))
		}
		if hasHash(HashNames.Blake3) {
			str.WriteString(sumLine(res.Blake3, res.File))
		}
		if hasHash(HashNames.Sha3224) {
			str.WriteString(sumLine(res.Sha3224, res.File))
		}
		if hasHash(HashNames.Sha3256) {
			str.WriteString(sumLine(res.Sha3256, res.File))
		}
		if hasHash(HashNames.Sha3384) {
			str.WriteString(sumLine(res.Sha3384, res.File))
		}
		if hasHash(HashNames.Sha3512) {
			str.WriteString(sumLine(res.Sha3512, res.File))
		}
		if hasHash(HashNames.Ed2k) {
			str.WriteString(sumLine(res.Ed2k, res.File))
		}

		// written as a comment so md5sum -c and similar tools ignore it rather than warn
		if MTime && res.MTime != nil {
			str.WriteString(fmt.Sprintf("# mtime %s  %s\n", res.MTime.Format("2006-01-02 15:04:05"), sumEscaper.Replace(res.File)))
		}

		if !NoStream && FileOutput == "" {
			fmt.Fprint(stdout, str.String())
			str.Reset()
		}
	}

	return str.String()
}

var sumEscaper = strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r")

// sumLine formats a single line the same way md5sum does, which escapes filenames
// containing a backslash or newline and marks the line with a leading backslash
func sumLine(hash string, file string) string {
	if strings.ContainsAny(file, "\\\n\r") {
		return "\\" + hash + "  " + sumEscaper.Replace(file) + "\n"
	}
	return hash + "  " + file + "\n"
}

func toHashOnly(input chan Result) (string, bool) {
	var str strings.Builder

	for res := range input {
		if hasHash(HashNames.CRC32) {
			str.WriteString(res.CRC32 + "\n")
		}
		if hasHash(HashNames.XxHash64) {
			str.WriteString(res.XxHash64 + "\n")
		}
		if hasHash(HashNames.MD4) {
			str.WriteString(res.MD4 + "\n")
		}
		if hasHash(HashNames.MD5) {
			str.WriteString(res.MD5 + "\n")
		}
		if hasHash(HashNames.SHA1) {
			str.WriteString(res.SHA1 + "\n")
		}
		if hasHash(HashNames.SHA224) {
			str.WriteString(res.SHA224 + "\n")
		}
		if hasHash(HashNames.SHA256) {
			str.WriteString(res.SHA256 + "\n")
		}
		if hasHash(HashNames.SHA384) {
			str.WriteString(res.SHA384 + "\n")
		}
		if hasHash(HashNames.SHA512) {
			str.WriteString(res.SHA512 + "\n")
		}
		if hasHash(HashNames.Blake2b256) {
			str.WriteString(res.Blake2b256 + "\n")
		}
		if hasHash(HashNames.Blake2b512) {
			str.WriteString(res.Blake2b512 + "\n")
		}
		if hasHash(HashNames.Blake3) {
			str.WriteString(res.Blake3 + "\n")
		}
		if hasHash(HashNames.Sha3224) {
			str.WriteString(res.Sha3224 + "\n")
		}
		if hasHash(HashNames.Sha3256) {
			str.WriteString(res.Sha3256 + "\n")
		}
		if hasHash(HashNames.Sha3384) {
			str.WriteString(res.Sha3384 + "\n")
		}
		if hasHash(HashNames.Sha3512) {
			str.WriteString(res.Sha3512 + "\n")
		}
		if hasHash(HashNames.Ed2k) {
			str.WriteString(res.Ed2k + "\n")
		}

		if MTime && res.MTime != nil {
			str.WriteString(fmt.Sprintf("%s\n", res.MTime.Format("2006-01-02 15:04:05")))
		}

		if !NoStream && FileOutput == "" {
			fmt.Fprint(stdout, str.String())
			str.Reset()
		}
	}

	return str.String(), true
}

func toText(input chan Result) (string, bool) {
	var str strings.Builder
	first := true

	for res := range input {
		if !first {
			str.WriteString("\n")
		} else {
			first = false
		}

		str.WriteString(fmt.Sprintf("%s (%d bytes)\n", res.File, res.Bytes))

		if hasHash(HashNames.CRC32) {
			str.WriteString("      CRC32 " + res.CRC32 + "\n")
		}
		if hasHash(HashNames.XxHash64) {
			str.WriteString("   xxHash64 " + res.XxHash64 + "\n")
		}
		if hasHash(HashNames.MD4) {
			str.WriteString("        MD4 " + res.MD4 + "\n")
		}
		if hasHash(HashNames.MD5) {
			str.WriteString("        MD5 " + res.MD5 + "\n")
		}
		if hasHash(HashNames.SHA1) {
			str.WriteString("       SHA1 " + res.SHA1 + "\n")
		}
		if hasHash(HashNames.SHA224) {
			str.WriteString("     SHA224 " + res.SHA224 + "\n")
		}
		if hasHash(HashNames.SHA256) {
			str.WriteString("     SHA256 " + res.SHA256 + "\n")
		}
		if hasHash(HashNames.SHA384) {
			str.WriteString("     SHA384 " + res.SHA384 + "\n")
		}
		if hasHash(HashNames.SHA512) {
			str.WriteString("     SHA512 " + res.SHA512 + "\n")
		}
		if hasHash(HashNames.Blake2b256) {
			str.WriteString("Blake2b-256 " + res.Blake2b256 + "\n")
		}
		if hasHash(HashNames.Blake2b512) {
			str.WriteString("Blake2b-512 " + res.Blake2b512 + "\n")
		}
		if hasHash(HashNames.Blake3) {
			str.WriteString("     Blake3 " + res.Blake3 + "\n")
		}
		if hasHash(HashNames.Sha3224) {
			str.WriteString("   SHA3-224 " + res.Sha3224 + "\n")
		}
		if hasHash(HashNames.Sha3256) {
			str.WriteString("   SHA3-256 " + res.Sha3256 + "\n")
		}
		if hasHash(HashNames.Sha3384) {
			str.WriteString("   SHA3-384 " + res.Sha3384 + "\n")
		}
		if hasHash(HashNames.Sha3512) {
			str.WriteString("   SHA3-512 " + res.Sha3512 + "\n")
		}
		if hasHash(HashNames.Ed2k) {
			str.WriteString("       ed2k " + res.Ed2k + "\n")
		}

		if MTime && res.MTime != nil {
			str.WriteString(fmt.Sprintf("      MTime %s\n", res.MTime.Format("2006-01-02 15:04:05")))
		}

		if !NoStream && FileOutput == "" {
			fmt.Fprint(stdout, str.String())
			str.Reset()
		}
	}

	return str.String(), true
}

func toJSON(input chan Result) string {
	results := []Result{}
	for res := range input {
		results = append(results, res)
	}

	jsonString, _ := json.Marshal(results)
	return string(jsonString) + "\n"
}

func toHashDeep(input chan Result) string {
	var str strings.Builder

	pwd, err := os.Getwd()
	if err != nil {
		printWarn(fmt.Sprintf("unable to determine working directory: %s", err.Error()))
		pwd = ""
	}

	str.WriteString("%%%% HASHDEEP-1.0\n")
	if !contains(Hash, "sha256") && !contains(Hash, "all") {
		str.WriteString("%%%% size,md5,filename")
	} else if contains(Hash, "sha256") && !contains(Hash, "all") && !contains(Hash, "md5") {
		str.WriteString("%%%% size,sha256,filename")
	} else {
		str.WriteString("%%%% size,md5,sha256,filename")
	}

	str.WriteString("\n")

	str.WriteString(fmt.Sprintf("## Invoked from: %s\n", pwd))
	str.WriteString(fmt.Sprintf("## $ %s\n", strings.Join(os.Args, " ")))
	str.WriteString("##\n")

	for res := range input {
		// Bytes first, always the same.
		str.WriteString(fmt.Sprintf("%d,", res.Bytes))

		// Follows the same logic as the headers.
		if !contains(Hash, "sha256") && !contains(Hash, "all") {
			str.WriteString(fmt.Sprintf("%s,", res.MD5))
		} else if contains(Hash, "sha256") && !contains(Hash, "all") && !contains(Hash, "md5") {
			str.WriteString(fmt.Sprintf("%s,", res.SHA256))
		} else {
			str.WriteString(fmt.Sprintf("%s,%s,", res.MD5, res.SHA256))
		}

		// Finish with filename and newline.
		str.WriteString(res.File)
		str.WriteString("\n")

		// hashdeep rejects files with columns it does not know, and a column after the filename
		// is ambiguous when the filename contains a comma, so mtime is written as a comment line
		// after the file which hashdeep ignores
		if MTime && res.MTime != nil {
			str.WriteString(fmt.Sprintf("## mtime %s\n", res.MTime.Format("2006-01-02 15:04:05")))
		}
	}

	return str.String()
}

func toSqlite(input chan Result) (string, bool) {
	// if not file output specified we need to do it ourselves
	if FileOutput == "" {
		FileOutput = "hashit.db"
	}

	// handle sql conversions where null
	toSqlNull := func(input string) sql.NullString {
		if input == "" {
			return sql.NullString{
				Valid: false,
			}
		}
		return sql.NullString{
			Valid:  true,
			String: input,
		}
	}

	db, err := connectSqliteDb(outputTemp)
	if err != nil {
		printError(fmt.Sprintf("problem creating database %s: %s", FileOutput, err.Error()))
		return "", false
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	db.SetMaxOpenConns(1) // we are writing, so set the number of writes to 1

	queries := database.New(db)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		printError(fmt.Sprintf("problem with tx %s", FileOutput))
		return "", false
	}
	withTx := queries.WithTx(tx)

	// relative paths are relative to where we were run from, which the audit needs to resolve them
	if pwd, err := os.Getwd(); err == nil {
		if err := withTx.MetadataSet(context.Background(), database.MetadataSetParams{Key: metadataInvokedFrom, Value: pwd}); err != nil {
			printError(err.Error())
			return "", false
		}
	}

	// record which hashes were calculated, so an audit does not have to guess from the rows
	var calculated []string
	for _, name := range hashPriority {
		if hasHash(name) {
			calculated = append(calculated, name)
		}
	}
	if err := withTx.MetadataSet(context.Background(), database.MetadataSetParams{Key: metadataHashes, Value: strings.Join(calculated, ",")}); err != nil {
		printError(err.Error())
		return "", false
	}

	count := 0
	for res := range input {

		mtime := ""
		if res.MTime != nil {
			mtime = res.MTime.Format(time.RFC3339Nano)
		}

		_, err = withTx.FileHashInsertReplace(context.Background(), database.FileHashInsertReplaceParams{
			Filepath:   res.File,
			Crc32:      toSqlNull(res.CRC32),
			Xxhash64:   toSqlNull(res.XxHash64),
			Md4:        toSqlNull(res.MD4),
			Md5:        toSqlNull(res.MD5),
			Sha1:       toSqlNull(res.SHA1),
			Sha224:     toSqlNull(res.SHA224),
			Sha256:     toSqlNull(res.SHA256),
			Sha384:     toSqlNull(res.SHA384),
			Sha512:     toSqlNull(res.SHA512),
			Blake2b256: toSqlNull(res.Blake2b256),
			Blake2b512: toSqlNull(res.Blake2b512),
			Blake3:     toSqlNull(res.Blake3),
			Sha3224:    toSqlNull(res.Sha3224),
			Sha3256:    toSqlNull(res.Sha3256),
			Sha3384:    toSqlNull(res.Sha3384),
			Sha3512:    toSqlNull(res.Sha3512),
			Ed2k:       toSqlNull(res.Ed2k),
			Size:       res.Bytes,
			Mtime:      toSqlNull(mtime),
		})
		if err != nil {
			printError(err.Error())
			return "", false
		}

		if count >= 1000 {
			count = 0

			err := tx.Commit()
			if err != nil {
				printError(err.Error())
				return "", false
			}

			tx, err = db.BeginTx(context.Background(), nil)
			if err != nil {
				printError(err.Error())
				return "", false
			}
			withTx = queries.WithTx(tx)
		}
		count++
	}

	// its possible this was already commited so ignore
	err = tx.Commit()
	if err != nil {
		printError(err.Error())
	}

	// ensure we merge the WAL into a single file, and leave the database out of WAL mode so that
	// opening it read-only to audit does not create -wal and -shm files next to it
	_, err = db.Exec("PRAGMA wal_checkpoint(TRUNCATE); PRAGMA journal_mode = DELETE;")
	if err != nil {
		printError(err.Error())
	}

	return "", true
}

func printHashes() {
	fmt.Fprintf(stdout, "      CRC32 (%s)\n", HashNames.CRC32)
	fmt.Fprintf(stdout, "   xxHash64 (%s)\n", HashNames.XxHash64)
	fmt.Fprintf(stdout, "        MD4 (%s)\n", HashNames.MD4)
	fmt.Fprintf(stdout, "        MD5 (%s)\n", HashNames.MD5)
	fmt.Fprintf(stdout, "       SHA1 (%s)\n", HashNames.SHA1)
	fmt.Fprintf(stdout, "     SHA224 (%s)\n", HashNames.SHA224)
	fmt.Fprintf(stdout, "     SHA256 (%s)\n", HashNames.SHA256)
	fmt.Fprintf(stdout, "     SHA384 (%s)\n", HashNames.SHA384)
	fmt.Fprintf(stdout, "     SHA512 (%s)\n", HashNames.SHA512)
	fmt.Fprintf(stdout, "Blake2b-256 (%s)\n", HashNames.Blake2b256)
	fmt.Fprintf(stdout, "Blake2b-512 (%s)\n", HashNames.Blake2b512)
	fmt.Fprintf(stdout, "     Blake3 (%s)\n", HashNames.Blake3)
	fmt.Fprintf(stdout, "   SHA3-224 (%s)\n", HashNames.Sha3224)
	fmt.Fprintf(stdout, "   SHA3-256 (%s)\n", HashNames.Sha3256)
	fmt.Fprintf(stdout, "   SHA3-384 (%s)\n", HashNames.Sha3384)
	fmt.Fprintf(stdout, "   SHA3-512 (%s)\n", HashNames.Sha3512)
	fmt.Fprintf(stdout, "       ed2k (%s)\n", HashNames.Ed2k)
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}

	return false
}

func connectSqliteDb(pathName string) (*sql.DB, error) {
	// build a file URI so a path containing ? or # is not taken as the start of parameters
	abs, err := filepath.Abs(pathName)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", sqliteFileURI(abs, "_pragma=busy_timeout(5000)"))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`pragma journal_mode = wal;
pragma synchronous = normal;
pragma temp_store = memory;
pragma mmap_size = 268435456;
pragma foreign_keys = on;`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = db.Exec(assets.Migrations)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
