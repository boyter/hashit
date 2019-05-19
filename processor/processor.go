package processor

import (
	"os"
	"path/filepath"
	"runtime"
)

// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// DirFilePaths is not set via flags but by arguments following the flags for file or directory to process
var DirFilePaths = []string{}

// FileListQueueSize is the queue of files found and ready to be processed
var FileListQueueSize = runtime.NumCPU() * 100

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	// Clean up any invalid arguments before setting everything up
	if len(DirFilePaths) == 0 {
		DirFilePaths = append(DirFilePaths, ".")
	}

	// Check if the paths or files added exist and exit if not
	for _, f := range DirFilePaths {
		fpath := filepath.Clean(f)

		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			printError("file or directory does not exist: " + fpath)
			os.Exit(1)
		}
	}

	fileListQueue := make(chan string, FileListQueueSize) // Files ready to be read from disk

	// Spawn routine to start finding files on disk
	go func() {
		for _, f := range DirFilePaths {
			walkDirectory(f, fileListQueue)
		}
		close(fileListQueue)
	}()

	fileProcessorWorker(fileListQueue)
}
