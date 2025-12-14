// SPDX-License-Identifier: MIT

package processor

import (
	"fmt"
	"github.com/boyter/gocodewalker"
	"os"
	"path/filepath"
	"regexp"
)

func walkDirectory(toWalk string, output chan string) {

	walkErr := filepath.WalkDir(toWalk, func(root string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			output <- root
		}

		return nil
	})

	if walkErr != nil {
		if Verbose {
			printVerbose(fmt.Sprintf("error walking: %s", toWalk))
		}
	}
}

func walkDirectoryWithIgnore(toWalk string, output chan string) {
	fileListQueue := make(chan *gocodewalker.File, 1000)
	//fileWalker := gocodewalker.NewParallelFileWalker([]string{toWalk}, fileListQueue)
	fileWalker := gocodewalker.NewFileWalker(toWalk, fileListQueue)

	// The user flags are to enable processing, while gocodewalker is to disable
	// so we need to invert the values.
	fileWalker.IgnoreGitIgnore = !GitIgnore
	fileWalker.IgnoreIgnoreFile = !Ignore
	fileWalker.IgnoreGitModules = !GitModuleIgnore
	fileWalker.IncludeHidden = true
	fileWalker.ExcludeDirectory = PathDenyList

	if HashIgnore {
		fileWalker.CustomIgnore = []string{".hashignore"}
	}

	// handle the errors by printing them out and then ignore
	errorHandler := func(err error) bool {
		printError(err.Error())
		return true
	}
	fileWalker.SetErrorHandler(errorHandler)

	for _, exclude := range Exclude {
		regexpResult, err := regexp.Compile(exclude)
		if err == nil {
			fileWalker.ExcludeFilenameRegex = append(fileWalker.ExcludeFilenameRegex, regexpResult)
			fileWalker.ExcludeDirectoryRegex = append(fileWalker.ExcludeDirectoryRegex, regexpResult)
		} else {
			printError(err.Error())
		}
	}

	go func() {
		err := fileWalker.Start()
		if err != nil {
			printError(err.Error())
		}
	}()

	for f := range fileListQueue {
		output <- f.Location
	}
}
