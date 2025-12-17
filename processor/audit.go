// SPDX-License-Identifier: MIT

package processor

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/boyter/hashit/processor/database"
)

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

// doSqliteAudit performs an audit against a SQLite database, comparing all available hashes.
func doSqliteAudit(input chan Result) (string, bool) {
	db, err := connectSqliteDb(AuditFile)
	if err != nil {
		printError(fmt.Sprintf("failed to open audit database: %s", err.Error()))
		return "", false
	}
	defer db.Close()

	queries := database.New(db)

	// For detecting missing files, load all known paths into a map.
	// This could be memory intensive for very large databases but is the simplest approach.

	rows, err := db.Query("SELECT filepath FROM file_hashes")
	if err != nil {
		printError(fmt.Sprintf("failed to query file paths from audit database: %s", err.Error()))
		return "", false
	}
	defer rows.Close()

	expectedFiles := make(map[string]bool)
	for rows.Next() {
		var filepath string
		if err := rows.Scan(&filepath); err != nil {
			printError(fmt.Sprintf("failed to scan filepath from audit database: %s", err.Error()))
			continue
		}
		expectedFiles[filepath] = true
	}
	knownFileCount := len(expectedFiles)

	examinedCount := 0
	matched := 0
	filesModified := 0
	newFiles := 0
	status := Passed

	for res := range input {
		examinedCount++
		// Remove from expected files map as we have now seen it
		delete(expectedFiles, res.File)

		dbRecord, err := queries.FileHashByFilePath(context.Background(), res.File)
		if err != nil {
			// If the record is not found, it's a new file.
			if strings.Contains(err.Error(), "no rows in result set") {
				newFiles++
				status = Failed
				if Verbose {
					fmt.Printf("%v: File new\n", res.File)
				}
			} else {
				printError(fmt.Sprintf("failed to query audit database for file %s: %s", res.File, err.Error()))
			}
			continue
		}

		// Record found, now perform "paranoid" multi-hash comparison
		modified := false

		// Helper function for comparison to reduce boilerplate
		checkHash := func(hashName, resHash string, dbHash sql.NullString) bool {
			// If the hash exists in the DB, and the new hash was calculated, compare them.
			if dbHash.Valid && resHash != "" && resHash != dbHash.String {
				if Verbose {
					fmt.Printf("%v: File modified (hash mismatch: %s)\n", res.File, hashName)
				}
				return true // Indicates modification
			}
			return false
		}

		if checkHash("crc32", res.CRC32, dbRecord.Crc32) {
			modified = true
		}
		if !modified && checkHash("xxhash64", res.XxHash64, dbRecord.Xxhash64) {
			modified = true
		}
		if !modified && checkHash("md4", res.MD4, dbRecord.Md4) {
			modified = true
		}
		if !modified && checkHash("md5", res.MD5, dbRecord.Md5) {
			modified = true
		}
		if !modified && checkHash("sha1", res.SHA1, dbRecord.Sha1) {
			modified = true
		}
		if !modified && checkHash("sha256", res.SHA256, dbRecord.Sha256) {
			modified = true
		}
		if !modified && checkHash("sha512", res.SHA512, dbRecord.Sha512) {
			modified = true
		}
		if !modified && checkHash("blake2b256", res.Blake2b256, dbRecord.Blake2b256) {
			modified = true
		}
		if !modified && checkHash("blake2b512", res.Blake2b512, dbRecord.Blake2b512) {
			modified = true
		}
		if !modified && checkHash("blake3", res.Blake3, dbRecord.Blake3) {
			modified = true
		}
		if !modified && checkHash("sha3-224", res.Sha3224, dbRecord.Sha3224) {
			modified = true
		}
		if !modified && checkHash("sha3-256", res.Sha3256, dbRecord.Sha3256) {
			modified = true
		}
		if !modified && checkHash("sha3-384", res.Sha3384, dbRecord.Sha3384) {
			modified = true
		}
		if !modified && checkHash("sha3-512", res.Sha3512, dbRecord.Sha3512) {
			modified = true
		}
		if !modified && checkHash("ed2k", res.Ed2k, dbRecord.Ed2k) {
			modified = true
		}

		// Also check file size
		if !modified && res.Bytes != dbRecord.Size {
			modified = true
			if Verbose {
				fmt.Printf("%v: File modified (size mismatch: got %d, expected %d)\n", res.File, res.Bytes, dbRecord.Size)
			}
		}

		if modified {
			filesModified++
			status = Failed
		} else {
			matched++
			if VeryVerbose {
				fmt.Printf("%v: Ok\n", res.File)
			}
		}
	}

	filesMissing := len(expectedFiles)
	if filesMissing > 0 {
		status = Failed
		if Verbose {
			for f := range expectedFiles {
				fmt.Printf("%v: File expected but not found\n", f)
			}
		}
	}

	// Note: Moved file detection is not implemented in this version for simplicity.

	return fmt.Sprintf(`hashit: SQLite Audit %s
       Files examined: %d
Known files expecting: %d
        Files matched: %d
       Files modified: %d
      New files found: %d
        Files missing: %d`+"\n", status, examinedCount, knownFileCount, matched, filesModified, newFiles, filesMissing), status == Passed
}
