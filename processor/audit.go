// SPDX-License-Identifier: MIT

package processor

import (
	"context"
	"fmt"
	"os"
	"reflect"
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

		// Use reflection to compare all hash fields
		resVal := reflect.ValueOf(res)
		dbVal := reflect.ValueOf(dbRecord)

		for i := 0; i < resVal.NumField(); i++ {
			fieldName := resVal.Type().Field(i).Name
			// Skip non-hash fields
			if fieldName == "File" || fieldName == "Bytes" || fieldName == "MTime" || fieldName == "Err" {
				continue
			}

			resHash := resVal.Field(i).String()
			// Only compare if the new hash was actually calculated
			if resHash == "" {
				continue
			}

			dbField := dbVal.FieldByName(fieldName)
			if dbField.IsValid() {
				// The DB field is a sql.NullString, so we need to access its String and Valid properties
				dbHash := dbField.FieldByName("String").String()
				dbValid := dbField.FieldByName("Valid").Bool()

				if dbValid && resHash != dbHash {
					modified = true
					if Verbose {
						fmt.Printf("%v: File modified (hash mismatch: %s)\n", res.File, fieldName)
					}
					break // No need to check other hashes
				}
			}
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
