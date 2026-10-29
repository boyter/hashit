// SPDX-License-Identifier: MIT

package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/boyter/gocodewalker"
)

// walkedDirectories holds the real path of every directory walked or followed through a symlink,
// so a symlink back to a directory already being walked does not loop forever. Walks happen one
// at a time from a single goroutine so it needs no locking
var walkedDirectories = map[string]bool{}

func realPath(path string) string {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return ""
	}
	abs, err := filepath.Abs(real)
	if err != nil {
		return ""
	}
	return abs
}

// walkDirectoryWithIgnore walks the directory sending every file found to output, with
// prefix added to the start of each path
func walkDirectoryWithIgnore(toWalk string, prefix string, output chan string) {
	if real := realPath(toWalk); real != "" {
		walkedDirectories[real] = true
	}

	fileListQueue := make(chan *gocodewalker.File, 1000)
	var fileWalker *gocodewalker.FileWalker
	if NoThreads != 1 {
		fileWalker = gocodewalker.NewParallelFileWalker([]string{toWalk}, fileListQueue)
	} else {
		fileWalker = gocodewalker.NewFileWalker(toWalk, fileListQueue)
	}

	// The user flags are to enable processing, while gocodewalker is to disable
	// so we need to invert the values.
	fileWalker.IgnoreGitIgnore = !GitIgnore
	fileWalker.IgnoreIgnoreFile = !Ignore
	fileWalker.IgnoreGitModules = !GitModuleIgnore
	fileWalker.IncludeHidden = !SkipHidden
	fileWalker.ExcludeDirectory = PathDenyList

	if HashIgnore {
		fileWalker.CustomIgnore = []string{".hashignore"}
	}

	// handle the errors by printing them out and then ignore
	errorHandler := func(err error) bool {
		printFileError(fmt.Sprintf("unable to walk %s: %s", toWalk, err.Error()))
		return true
	}
	fileWalker.SetErrorHandler(errorHandler)

	for _, exclude := range Exclude {
		// already checked as valid by validateOptions
		regexpResult := regexp.MustCompile(exclude)
		fileWalker.ExcludeFilenameRegex = append(fileWalker.ExcludeFilenameRegex, regexpResult)
		fileWalker.ExcludeDirectoryRegex = append(fileWalker.ExcludeDirectoryRegex, regexpResult)
	}

	go func() {
		err := fileWalker.Start()
		if err != nil {
			printError(err.Error())
		}
	}()

	for f := range fileListQueue {
		// the walker does not follow symlinks, so a symlink to a directory arrives as a file.
		// Follow it as hashdeep does, unless it leads back to a directory already walked
		if fi, err := os.Lstat(f.Location); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Stat(f.Location); err == nil && target.IsDir() {
				if real := realPath(f.Location); real != "" && walkedDirectories[real] {
					printWarn(fmt.Sprintf("Skipping symlink %s as it leads to a directory already being processed", prefix+f.Location))
				} else {
					walkDirectoryWithIgnore(f.Location, prefix, output)
				}
				continue
			}
		}

		output <- prefix + f.Location
	}
}
