package processor

import (
	"os"
	"path/filepath"
	"strings"
)

// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// If set to true disables the use of memory maps
var NoMmap = false

// List of hashes that we want to process
var Hashes = []string{}

// DirFilePaths is not set via flags but by arguments following the flags for file or directory to process
var DirFilePaths = []string{}

// FileListQueueSize is the queue of files found and ready to be processed
var FileListQueueSize = 1000

// Number of bytes in a size to enable memory maps or streaming
var StreamSize int64 = 1000000

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	// Clean up any invalid arguments before setting everything up
	if len(DirFilePaths) == 0 {
		DirFilePaths = append(DirFilePaths, ".")
	}

	// Clean up hashes by setting all to lower
	cleanhash := []string{}
	for _, x := range Hashes {
		cleanhash = append(cleanhash, strings.ToLower(x))
	}
	Hashes = cleanhash

	// Check if the paths or files added exist and exit if not
	for _, f := range DirFilePaths {
		fpath := filepath.Clean(f)

		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			printError("file or directory does not exist: " + fpath)
			os.Exit(1)
		}
	}

	fileListQueue := make(chan string, FileListQueueSize) // Files ready to be read from disk
	fileSummaryQueue := make(chan Result, FileListQueueSize)

	// Spawn routine to start finding files on disk
	go func() {
		for _, f := range DirFilePaths {
			walkDirectory(f, fileListQueue)
		}
		close(fileListQueue)
	}()

	// TODO multi-process this
	fileProcessorWorker(fileListQueue)
	fileSummarize(fileSummaryQueue)

	// TODO formatter here
}


func hasHash(hash string) bool {
	for _, x := range Hashes {
		if x == "all" {
			return true
		}

		if x == hash {
			return true
		}
	}

	return false
}