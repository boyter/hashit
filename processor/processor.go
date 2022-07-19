package processor

import (
	"fmt"
	"github.com/gosuri/uiprogress"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Global Version
var Version = "0.1.0"

// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// Trace enables trace logging output which is extremely verbose
var Trace = false

// Progress uses ui bar to display the progress of files
var Progress = false

// ProgressLarge uses ui bar to display the progress of files
var ProgressLarge = false

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
var isDir = false

// FileListQueueSize is the queue of files found and ready to be processed
var FileListQueueSize = 1000

// Number of bytes in a size to enable memory maps or streaming
var StreamSize int64 = 1_000_000

// If set will enable the internal file audit logic to kick in
var FileAudit = false

// String mapping for hash names
var HashNames = Result{
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
}

// Raw hashDatabase loaded
var hashDatabase = map[string]Result{}

// Hash to name lookup
var hashLookup = map[string]string{}

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	// Display the supported hashes then bail out
	if Hashes {
		printHashes()
		return
	}

	// Check if we are accepting data from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		StandardInput = true
	}

	// If nothing was supplied as an argument to run against assume run against everything in the
	// current directory recursively
	if len(DirFilePaths) == 0 {
		DirFilePaths = append(DirFilePaths, ".")
	}

	// If a single argument is supplied enable recursive as if its a file no problem
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
							isDir = true
							walkDirectory(fp, fileListQueue)
						}
					} else {
						fileListQueue <- fp
					}
				}

			}
			close(fileListQueue)
		}()

		if Progress || ProgressLarge {
			uiprogress.Start() // start rendering of progress bars
		}

		var wg sync.WaitGroup
		for i := 0; i < runtime.NumCPU(); i++ {
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
		_ = ioutil.WriteFile(FileOutput, []byte(result), 0600)
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
