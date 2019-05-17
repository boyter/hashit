package processor

import (
	"fmt"
	mmapgo "github.com/edsrzf/mmap-go"
	"os"
	"path/filepath"
	"runtime"
)

// Verbose enables verbose logging output
var Verbose = false

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

	// Spawn routine to start identifying files on disk
	go func() {
		for _, f := range DirFilePaths {
			walkDirectory(f, fileListQueue)
		}
		close(fileListQueue)
	}()
	fileProcessorWorker(fileListQueue)

	// If a file is small < 1 MB then read directly into memory and process
	// if it is large then mmap it and process
	//md5_digest := md5.New()
	//sha1_digest := sha1.New()
	//sha256_digest := sha256.New()
	//sha512_digest := sha512.New()
	//md5_digest.Write([]byte{})
}

func Mmap() {
	file, err := os.OpenFile("main.go", os.O_RDONLY, 0644)

	if err != nil {
		panic(err.Error())
	}

	mmap, err := mmapgo.Map(file, mmapgo.RDONLY, 0)

	fmt.Println("Length", len(mmap))

	count := 0
	for _, currentByte := range mmap {
		if currentByte == '\n' {
			count++
		}
	}

	fmt.Println("Newlines", count)

	if err != nil {
		fmt.Println("error mapping:", err)
	}

	if err := mmap.Unmap(); err != nil {
		fmt.Println("error unmapping:", err)
	}
}
