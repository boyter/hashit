// SPDX-License-Identifier: MIT

package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/boyter/gocodewalker"
)

// isExcludedDirectory reports if a directory name is excluded by --exclude-dir or --not-match,
// for symlinks to directories which the walker sees as files. Ignore file rules such as
// .gitignore are not applied to them
func isExcludedDirectory(name string) bool {
	for _, d := range PathDenyList {
		if strings.TrimRight(d, "/") == name {
			return true
		}
	}
	for _, exclude := range Exclude {
		if regexp.MustCompile(exclude).MatchString(name) {
			return true
		}
	}
	return false
}

// isAncestorOrSelf reports if dir is the same as or contains path, both being real absolute paths
func isAncestorOrSelf(dir string, path string) bool {
	if dir == path {
		return true
	}
	return strings.HasPrefix(path, strings.TrimSuffix(dir, string(filepath.Separator))+string(filepath.Separator))
}

// isSymlinkCycle reports if following the symlink at location to target would re-enter a
// directory on the path to it, which is how hashdeep detects cycles. Each directory on the path is
// resolved, as earlier symlinks on the path may lead anywhere. This only runs for symlinks to
// directories so the extra lookups are rare
func isSymlinkCycle(location string, target string) bool {
	dir := filepath.Dir(location)
	for {
		if real := realPath(dir); real != "" && isAncestorOrSelf(target, real) {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}

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
			printFileError(fmt.Sprintf("unable to walk %s: %s", toWalk, err.Error()))
		}
	}()

	for f := range fileListQueue {
		// the walker does not follow symlinks, so a symlink to a directory arrives as a file.
		// Follow it as hashdeep does, unless it leads back into a directory already being walked
		if fi, err := os.Lstat(f.Location); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Stat(f.Location); err == nil && target.IsDir() {
				// the walker sees the symlink as a file, so apply directory exclusions ourselves
				if isExcludedDirectory(f.Filename) {
					continue
				}
				real := realPath(f.Location)
				if real == "" {
					printFileError(fmt.Sprintf("unable to resolve symlink %s", prefix+f.Location))
				} else if isSymlinkCycle(f.Location, real) {
					printWarn(fmt.Sprintf("Skipping symlink %s as it creates a cycle", prefix+f.Location))
				} else {
					walkDirectoryWithIgnore(f.Location, prefix, output)
				}
				continue
			}
		}

		output <- prefix + f.Location
	}
}
