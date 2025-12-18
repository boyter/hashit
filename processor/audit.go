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

	// --- Phase 1: Process all files found on disk ---
	examinedCount := 0
	matched := 0
	filesModified := 0
	newFileCandidates := []Result{}
	foundFilesOnDisk := make(map[string]bool)
	status := Passed

	for res := range input {
		examinedCount++
		foundFilesOnDisk[res.File] = true

		dbRecord, err := queries.FileHashByFilePath(context.Background(), res.File)
		if err != nil {
			// If the record is not found by path, it's a candidate for being a new or moved file.
			if strings.Contains(err.Error(), "no rows in result set") {
				newFileCandidates = append(newFileCandidates, res)
			} else {
				printError(fmt.Sprintf("failed to query audit database for file %s: %s", res.File, err.Error()))
			}
			continue
		}

		// Record found by path, so compare its hashes
		modified := false
		compared := false // Did we successfully compare at least one hash?

		compare := func(name, resHash string, dbHash sql.NullString) {
			if modified || resHash == "" || !dbHash.Valid {
				return
			}
			compared = true
			if resHash != dbHash.String {
				modified = true
				if Verbose {
					fmt.Printf("%v: File modified (hash mismatch: %s)\n", res.File, name)
				}
			}
		}

		compare("crc32", res.CRC32, dbRecord.Crc32)
		compare("xxhash64", res.XxHash64, dbRecord.Xxhash64)
		compare("md4", res.MD4, dbRecord.Md4)
		compare("md5", res.MD5, dbRecord.Md5)
		compare("sha1", res.SHA1, dbRecord.Sha1)
		compare("sha256", res.SHA256, dbRecord.Sha256)
		compare("sha512", res.SHA512, dbRecord.Sha512)
		compare("blake2b256", res.Blake2b256, dbRecord.Blake2b256)
		compare("blake2b512", res.Blake2b512, dbRecord.Blake2b512)
		compare("blake3", res.Blake3, dbRecord.Blake3)
		compare("sha3-224", res.Sha3224, dbRecord.Sha3224)
		compare("sha3-256", res.Sha3256, dbRecord.Sha3256)
		compare("sha3-384", res.Sha3384, dbRecord.Sha3384)
		compare("sha3-512", res.Sha3512, dbRecord.Sha3512)
		compare("ed2k", res.Ed2k, dbRecord.Ed2k)

		if !modified && res.Bytes != dbRecord.Size {
			modified = true
			compared = true
			if Verbose {
				fmt.Printf("%v: File modified (size mismatch: got %d, expected %d)\n", res.File, res.Bytes, dbRecord.Size)
			}
		}

		if modified {
			filesModified++
		} else if !compared {
			filesModified++
			if Verbose {
				fmt.Printf("%v: File modified (no common hashes/size to compare)\n", res.File)
			}
		} else {
			matched++
			if VeryVerbose {
				fmt.Printf("%v: Ok\n", res.File)
			}
		}
	}

	// --- Phase 2: Find Missing and Moved files ---

	// First, find all files that are in the DB but were not seen on disk.
	// We do this in pages to avoid loading everything into memory at once.
	missingFilePaths := make(map[string]bool)
	offset := int32(0)
	limit := int32(1000) // Process in blocks of 1000
	knownFileCount := 0

	for {
		dbFilePaths, err := queries.ListFilePathsPaged(context.Background(), database.ListFilePathsPagedParams{
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			printError(fmt.Sprintf("failed to query paged file paths from audit database: %s", err.Error()))
			return "", false
		}

		if len(dbFilePaths) == 0 {
			break // No more files in the database
		}
		knownFileCount += len(dbFilePaths)

		for _, dbFilePath := range dbFilePaths {
			if _, ok := foundFilesOnDisk[dbFilePath]; !ok {
				missingFilePaths[dbFilePath] = true
			}
		}
		offset += limit
	}

	// Now, reconcile new files vs missing files to find moves.
	moved := 0
	genuinelyNewFiles := []Result{}

	for _, newFile := range newFileCandidates {
		foundMove := false
		if newFile.SHA256 != "" {
			// Look for a missing file with the same SHA256 hash
			dbRecord, err := queries.FileHashBySHA256(context.Background(), sql.NullString{String: newFile.SHA256, Valid: true})
			if err == nil {
				// We found a record with the same hash. Check if its path is in our missing files list.
				if _, ok := missingFilePaths[dbRecord.Filepath]; ok {
					moved++
					foundMove = true
					if Verbose {
						fmt.Printf("%v -> %v: File moved\n", dbRecord.Filepath, newFile.File)
					}
					// Remove from missing list so it's not counted as missing
					delete(missingFilePaths, dbRecord.Filepath)
				}
			}
		}

		if !foundMove {
			genuinelyNewFiles = append(genuinelyNewFiles, newFile)
		}
	}

	if Verbose {
		for _, res := range genuinelyNewFiles {
			fmt.Printf("%v: File new\n", res.File)
		}
	}

	filesMissing := len(missingFilePaths)
	if Verbose && filesMissing > 0 {
		for f := range missingFilePaths {
			fmt.Printf("%v: File expected but not found\n", f)
		}
	}

	newFiles := len(genuinelyNewFiles)

	if filesModified > 0 || newFiles > 0 || filesMissing > 0 || moved > 0 {
		status = Failed
	}

	return fmt.Sprintf(`hashit: SQLite Audit %s
       Files examined: %d
Known files expecting: %d
        Files matched: %d
       Files modified: %d
          Files moved: %d
      New files found: %d
        Files missing: %d`+"\n", status, examinedCount, knownFileCount, matched, filesModified, moved, newFiles, filesMissing), status == Passed
}
