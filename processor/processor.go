package processor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Global Version
var Version = "0.1.0"

// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// Trace enables trace logging output which is extremely verbose
var Trace = false

// Recursive to walk directories
var Recursive = false

// If set to true disables the use of memory maps
var NoMmap = false

// Do not print out results as they are processed
var NoStream = false

// Should the application print all hashes it knows about
var Hashes = false

// List of hashes that we want to process
var Hash = []string{}

// Format sets the output format of the formatter
var Format = ""

// FileOutput sets the file that output should be written to
var FileOutput = ""

// DirFilePaths is not set via flags but by arguments following the flags for file or directory to process
var DirFilePaths = []string{}
var isDir = false

// FileListQueueSize is the queue of files found and ready to be processed
var FileListQueueSize = 1000

// Number of bytes in a size to enable memory maps or streaming
var StreamSize int64 = 1000000

// String mapping for hash names
var HashNames = Result{
	MD4:        "md4",
	MD5:        "md5",
	SHA1:       "sha1",
	SHA256:     "sha256",
	SHA512:     "sha512",
	Blake2b256: "blake2b256",
	Blake2b512: "blake2b512",
	Sha3224:    "sha3224",
	Sha3256:    "sha3256",
	Sha3384:    "sha3384",
	Sha3512:    "sha3512",
}

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	// Display the supported hashes then bail out
	if Hashes {
		fmt.Println(fmt.Sprintf("        MD4 (%s)", HashNames.MD4))
		fmt.Println(fmt.Sprintf("        MD5 (%s)", HashNames.MD5))
		fmt.Println(fmt.Sprintf("       SHA1 (%s)", HashNames.SHA1))
		fmt.Println(fmt.Sprintf("     SHA256 (%s)", HashNames.SHA256))
		fmt.Println(fmt.Sprintf("     SHA512 (%s)", HashNames.SHA512))
		fmt.Println(fmt.Sprintf("Blake2b-256 (%s)", HashNames.Blake2b256))
		fmt.Println(fmt.Sprintf("Blake2b-512 (%s)", HashNames.Blake2b512))
		fmt.Println(fmt.Sprintf("   SHA3-224 (%s)", HashNames.Sha3224))
		fmt.Println(fmt.Sprintf("   SHA3-256 (%s)", HashNames.Sha3256))
		fmt.Println(fmt.Sprintf("   SHA3-384 (%s)", HashNames.Sha3384))
		fmt.Println(fmt.Sprintf("   SHA3-512 (%s)", HashNames.Sha3512))
		return
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

	// Clean up hashes by setting all to lower
	h := []string{}
	for _, x := range Hash {
		h = append(h, strings.ToLower(x))
	}
	Hash = h

	fileListQueue := make(chan string, FileListQueueSize)    // Files ready to be read from disk
	fileSummaryQueue := make(chan Result, FileListQueueSize) // Results ready to be printed

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

	go fileProcessorWorker(fileListQueue, fileSummaryQueue)
	result := fileSummarize(fileSummaryQueue)

	if FileOutput == "" {
		fmt.Print(result)
	} else {
		_ = ioutil.WriteFile(FileOutput, []byte(result), 0600)
		fmt.Println("results written to " + FileOutput)
	}
}

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
