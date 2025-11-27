// SPDX-License-Identifier: MIT

package processor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/gosuri/uiprogress"
)

// Global Version
var Version = "1.5.0 (beta)"

// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// Trace enables trace logging output which is extremely verbose
var Trace = false

// MTime enable mtime calculation and output
var MTime = false

// Progress uses ui bar to display the progress of files
var Progress = false

// Recursive to walk directories
var Recursive = false

// Do not print out results as they are processed
var NoStream = false

// If data is being piped in using stdin
var StandardInput = false

// Should the application print all hashes it knows about
var Hashes = false

// List of hashes that we want to process
var Hash = []string{}

// Format sets the output format of the formatter
var Format = ""

// FileOutput sets the file that output should be written to
var FileOutput = ""

// AuditFile sets the file that we want to audit against similar to hashdeep
var AuditFile = ""

// DirFilePaths is not set via flags but by arguments following the flags for file or directory to process
var DirFilePaths = []string{}

// FileListQueueSize is the queue of files found and ready to be processed
var FileListQueueSize = 1000

// Number of bytes in a size to enable memory maps or streaming
var StreamSize int64 = 1_000_000

// If set will enable the internal file audit logic to kick in
var FileAudit = false

// FileInput indicates we have a file passed in which consists of a
var FileInput = ""

var NoThreads = runtime.NumCPU()

// String mapping for hash names
var HashNames = Result{
	CRC32:      "crc32",
	XxHash64:   "xxhash64",
	MD4:        "md4",
	MD5:        "md5",
	SHA1:       "sha1",
	SHA256:     "sha256",
	SHA512:     "sha512",
	Blake2b256: "blake2b256",
	Blake2b512: "blake2b512",
	Blake3:     "blake3",
	Sha3224:    "sha3224",
	Sha3256:    "sha3256",
	Sha3384:    "sha3384",
	Sha3512:    "sha3512",
	Ed2k:       "ed2k",
}

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	// Display the supported hashes then bail out
	if Hashes {
		printHashes()
		return
	}

	// Check if we are accepting data from stdin
	if len(DirFilePaths) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			StandardInput = true
		}
	}

	// If nothing was supplied as an argument to run against, assume run against everything in the
	// current directory recursively
	if len(DirFilePaths) == 0 {
		DirFilePaths = append(DirFilePaths, ".")
	}

	// If a single argument is supplied, enable recursive as if it's a file no problem
	// but if its a directory the user probably wants to hash everything in that directory
	if len(DirFilePaths) == 1 {
		Recursive = true
	}

	// Clean up hashes by setting all input to lowercase
	Hash = formatHashInput()

	// Results ready to be printed
	fileSummaryQueue := make(chan Result, FileListQueueSize)

	if StandardInput {
		go processStandardInput(fileSummaryQueue)
	} else {
		// Files ready to be read from disk
		fileListQueue := make(chan string, FileListQueueSize)

		if FileInput == "" {
			// Spawn routine to start finding files on disk
			go func() {
				// Check if the paths or files added exist and inform the user if they don't
				for _, f := range DirFilePaths {
					fp := filepath.Clean(f)
					fi, err := os.Stat(fp)

					// If there is an error which is usually does not exist then exit non zero
					if err != nil {
						printError(fmt.Sprintf("file or directory issue: %s %s", fp, err.Error()))
						os.Exit(1)
					} else {
						if fi.IsDir() {
							if Recursive {
								walkDirectory(fp, fileListQueue)
							}
						} else {
							fileListQueue <- fp
						}
					}

				}
				close(fileListQueue)
			}()
		} else {
			// Open the file
			go func() {
				file, err := os.Open(FileInput)
				if err != nil {
					printError(fmt.Sprintf("failed to open input file: %s, %s", FileInput, err.Error()))
					os.Exit(1)
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)

				// Read the file line by line
				for scanner.Scan() {
					line := scanner.Text()
					fileListQueue <- line
				}
				close(fileListQueue)

				// Check for errors during scanning
				if err := scanner.Err(); err != nil {
					printError(fmt.Sprintf("error reading input file: %s, %s", FileInput, err.Error()))
					os.Exit(1)
				}
			}()
		}

		if Progress {
			uiprogress.Start() // start rendering of progress bars
		}

		var wg sync.WaitGroup
		for i := 0; i < NoThreads; i++ {
			wg.Add(1)
			go func() {
				fileProcessorWorker(fileListQueue, fileSummaryQueue)
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			close(fileSummaryQueue)
		}()
	}

	result, valid := fileSummarize(fileSummaryQueue)

	if FileOutput == "" {
		fmt.Print(result)
		if !valid {
			os.Exit(1)
		}
	} else {
		// we don't write out sqlite
		if strings.ToLower(Format) != "sqlite" {
			_ = os.WriteFile(FileOutput, []byte(result), 0600)
		}
		fmt.Println("results written to " + FileOutput)
	}
}

// ToLower all of the input hashes so we can match them easily
func formatHashInput() []string {
	h := []string{}
	for _, x := range Hash {
		h = append(h, strings.ToLower(x))
	}
	return h
}

// Check if a hash was supplied to the input so we know if we should calculate it
func hasHash(hash string) bool {
	for _, x := range Hash {
		if x == "all" {
			return true
		}

		if x == hash {
			return true
		}
	}

	return false
}
