// SPDX-License-Identifier: MIT

package processor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/boyter/gocodewalker"
)

// isExcludedDirectory reports if a directory is excluded by --exclude-dir or --not-match, for
// symlinks to directories which the walker sees as files. As with the walker, --exclude-dir
// matches the end of the path, such as node_modules or vendor/cache, and --not-match the name
func isExcludedDirectory(location string, name string) bool {
	path := "/" + strings.Trim(filepath.ToSlash(location), "/")
	for _, d := range PathDenyList {
		d = strings.Trim(filepath.ToSlash(d), "/")
		if d != "" && strings.HasSuffix(path, "/"+d) {
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
	return physicalAbs(real)
}

// physicalAbs returns the absolute path of a path with its symlinks already resolved, such as
// ../../x. filepath.Abs joins it to the working directory as the shell named it, which when that
// was reached through a symlink is not where the operating system resolves .. to, so the working
// directory as the operating system has it is used instead
func physicalAbs(resolved string) string {
	if filepath.IsAbs(resolved) {
		return filepath.Clean(resolved)
	}
	if cwd, err := syscall.Getwd(); err == nil {
		return filepath.Join(cwd, resolved)
	}
	abs, err := filepath.Abs(resolved)
	if err != nil {
		return ""
	}
	return abs
}

// osAbs returns the absolute path the operating system resolves path to, for opening a file by
// an absolute path, such as a SQLite database
func osAbs(path string) (string, error) {
	if real := realPath(path); real != "" {
		return real, nil
	}
	return filepath.Abs(path)
}

// walkDirectoryWithIgnore walks the directory sending every file found to output, with
// prefix added to the start of each path
func walkDirectoryWithIgnore(toWalk string, prefix string, output chan fileJob) {
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
		// record directories which could not be read, so an audit can report their files as
		// unreadable rather than missing
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			if abs := absFrom(workingDir(), pathErr.Path); filepath.IsAbs(abs) {
				failedDirs.Store(abs, true)
			}
		}
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
				if isExcludedDirectory(f.Location, f.Filename) {
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

		output <- fileJob{path: prefix + f.Location}
	}
}
