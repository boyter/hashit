// SPDX-License-Identifier: MIT

package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/boyter/hashit/assets"
	"github.com/boyter/hashit/processor/database"
	_ "modernc.org/sqlite"
)

// Get the time as standard UTC/Zulu format
func getFormattedTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Prints a message to stdout if flag to enable warning output is set
func printVerbose(msg string) {
	if Verbose {
		fmt.Printf("VERBOSE %s: %s\n", getFormattedTime(), msg)
	}
}

// Prints a message to stdout if flag to enable debug output is set
func printDebug(msg string) {
	if Debug {
		fmt.Printf("DEBUG %s: %s\n", getFormattedTime(), msg)
	}
}

// Used when explicitly for os.exit output when crashing out
func printError(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("ERROR %s: %s", getFormattedTime(), msg))
}

// Prints a message to stdout if flag to enable trace output is set
func printTrace(msg string) {
	if Trace {
		fmt.Println(fmt.Sprintf("TRACE %s: %s", getFormattedTime(), msg))
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

func doAudit(input chan Result) (string, bool) {
	// open the audit file
	file, err := os.ReadFile(AuditFile)
	if err != nil {
		printError(err.Error())
		return "", false
	}

	// parse the file into the lookup
	auditLookup, err := parseHashdeepFile(string(file))
	if err != nil {
		printError(err.Error())
		return "", false
	}

	matched := 0
	//partialMatch := 0
	//moved := 0
	//newFiles := 0
	//missingFile := 0

	for res := range input {
		_, ok := auditLookup[res.File]
		//fmt.Println(res.File, ok)

		if ok {
			matched++
		}
	}

	return fmt.Sprintf(`
hashdeep: Audit failed
   Input files examined: 0
  Known files expecting: 0
          Files matched: %d
Files partially matched: 0
            Files moved: 8
        New files found: 3
  Known files not found: 4`, matched), true

	// verbose (not very verybose)
	// output looks like the below
	//

	return "", true
}

// Mimics how md5sum sha1sum etc... work
func toSum(input chan Result) string {
	var str strings.Builder

	first := true

	for res := range input {
		if !first {
			str.WriteString("\n")
		} else {
			first = false
		}

		if hasHash(HashNames.CRC32) {
			str.WriteString(res.CRC32 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.XxHash64) {
			str.WriteString(res.XxHash64 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.MD4) {
			str.WriteString(res.MD4 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.MD5) {
			str.WriteString(res.MD5 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.SHA1) {
			str.WriteString(res.SHA1 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.SHA256) {
			str.WriteString(res.SHA256 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.SHA512) {
			str.WriteString(res.SHA512 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Blake2b256) {
			str.WriteString(res.Blake2b256 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Blake2b512) {
			str.WriteString(res.Blake2b512 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Blake3) {
			str.WriteString(res.Blake3 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3224) {
			str.WriteString(res.Sha3224 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3256) {
			str.WriteString(res.Sha3256 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3384) {
			str.WriteString(res.Sha3384 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3512) {
			str.WriteString(res.Sha3512 + "  " + res.File + "\n")
		}

		if !NoStream && FileOutput == "" {
			fmt.Print(str.String())
			str.Reset()
		}
	}

	return str.String()
}

func toHashOnly(input chan Result) (string, bool) {
	var str strings.Builder
	valid := true

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
		if hasHash(HashNames.SHA256) {
			str.WriteString(res.SHA256 + "\n")
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

		if !NoStream && FileOutput == "" {
			fmt.Print(str.String())
			str.Reset()
		}
	}

	return str.String(), valid
}

func toText(input chan Result) (string, bool) {
	var str strings.Builder
	valid := true
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
		if hasHash(HashNames.SHA256) {
			str.WriteString("     SHA256 " + res.SHA256 + "\n")
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

		if !NoStream && FileOutput == "" {
			fmt.Print(str.String())
			str.Reset()
		}
	}

	return str.String(), valid
}

func toJSON(input chan Result) string {
	results := []Result{}
	for res := range input {
		results = append(results, res)
	}

	jsonString, _ := json.Marshal(results)
	return string(jsonString)
}

func toHashDeep(input chan Result) string {
	var str strings.Builder

	pwd, err := os.Getwd()
	if err != nil {
		printError(fmt.Sprintf("unable to determine working directory: %s", err.Error()))
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
		str.WriteString(fmt.Sprintf("%s\n", res.File))
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

	db, err := connectSqliteDb(FileOutput)
	if err != nil {
		printError(fmt.Sprintf("problem connecting to db %s", FileOutput))
		return "", false
	}
	defer db.Close()
	db.SetMaxOpenConns(1) // we are writing, so set the number of writes to 1

	queries := database.New(db)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		printError(fmt.Sprintf("problem with tx %s", FileOutput))
		return "", false
	}
	withTx := queries.WithTx(tx)

	count := 0
	for res := range input {
		_, err = withTx.FileHashInsertReplace(context.Background(), database.FileHashInsertReplaceParams{
			Filepath:   res.File,
			Crc32:      toSqlNull(res.CRC32),
			Xxhash64:   toSqlNull(res.XxHash64),
			Md4:        toSqlNull(res.MD4),
			Md5:        toSqlNull(res.MD5),
			Sha1:       toSqlNull(res.SHA1),
			Sha256:     toSqlNull(res.SHA256),
			Sha512:     toSqlNull(res.SHA512),
			Blake2b256: toSqlNull(res.Blake2b256),
			Blake2b512: toSqlNull(res.Blake2b512),
			Blake3:     toSqlNull(res.Blake3),
			Sha3224:    toSqlNull(res.Sha3224),
			Sha3256:    toSqlNull(res.Sha3256),
			Sha3384:    toSqlNull(res.Sha3384),
			Sha3512:    toSqlNull(res.Sha3512),
			Size:       res.Bytes,
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

	// ensure we merge the WAL into a single file
	_, err = db.Exec("PRAGMA wal_checkpoint(FULL)")
	if err != nil {
		printError(err.Error())
	}

	return "", true
}

func printHashes() {
	fmt.Printf("      CRC32 (%s)\n", HashNames.CRC32)
	fmt.Printf("   xxHash64 (%s)\n", HashNames.XxHash64)
	fmt.Printf("        MD4 (%s)\n", HashNames.MD4)
	fmt.Printf("        MD5 (%s)\n", HashNames.MD5)
	fmt.Printf("       SHA1 (%s)\n", HashNames.SHA1)
	fmt.Printf("     SHA256 (%s)\n", HashNames.SHA256)
	fmt.Printf("     SHA512 (%s)\n", HashNames.SHA512)
	fmt.Printf("Blake2b-256 (%s)\n", HashNames.Blake2b256)
	fmt.Printf("Blake2b-512 (%s)\n", HashNames.Blake2b512)
	fmt.Printf("     Blake3 (%s)\n", HashNames.Blake3)
	fmt.Printf("   SHA3-224 (%s)\n", HashNames.Sha3224)
	fmt.Printf("   SHA3-256 (%s)\n", HashNames.Sha3256)
	fmt.Printf("   SHA3-384 (%s)\n", HashNames.Sha3384)
	fmt.Printf("   SHA3-512 (%s)\n", HashNames.Sha3512)
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
	db, err := sql.Open("sqlite", fmt.Sprintf("%s?_busy_timeout=5000", pathName))
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
		printError("pragma issue " + err.Error())
	}

	_, err = db.Exec(assets.Migrations)
	if err != nil {
		_ = db.Close()
		printError("migrations issue " + err.Error())
	}

	return db, err
}
