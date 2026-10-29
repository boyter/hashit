// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/boyter/hashit/processor"
	"github.com/spf13/cobra"
)

var verboseCount int

func main() {
	//f, _ := os.Create("hashit.pprof")
	//_ = pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use:     "hashit",
		Short:   "hashit [FILE or DIRECTORY]",
		Long:    "Hash It!\nVersion " + processor.Version + "\nBen Boyter <ben@boyter.org>",
		Version: processor.Version,
		Run: func(cmd *cobra.Command, args []string) {
			var filePaths []string
			for _, arg := range args {
				// an existing path is used as is, so files with [ ] * or ? in their name can be
				// supplied, otherwise expand it ourselves as shells such as cmd.exe do not
				if _, err := os.Lstat(arg); err == nil {
					filePaths = append(filePaths, arg)
					continue
				}

				matches, err := filepath.Glob(arg)
				if err != nil || len(matches) == 0 {
					// If there's an error or no matches, treat the argument as a literal path
					filePaths = append(filePaths, arg)
				} else {
					filePaths = append(filePaths, matches...)
				}
			}

			// -vv is counted as -v twice, which hashdeep users will expect to mean --vv
			processor.Verbose = verboseCount >= 1
			if verboseCount >= 2 {
				processor.VeryVerbose = true
			}

			processor.HashSupplied = cmd.Flags().Changed("hash")
			processor.FormatSupplied = cmd.Flags().Changed("format")
			processor.DirFilePaths = filePaths
			processor.Process()
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringSliceVarP(
		&processor.Hash,
		"hash",
		"c",
		[]string{"md5", "sha1", "sha256", "sha512"},
		"hashes to be run for each file (set to 'all' for all possible hashes); hashdeep and sum audits use the hashes in the audit file",
	)
	flags.StringVarP(
		&processor.Format,
		"format",
		"f",
		"text",
		"set output format [text, json, sum, hashdeep, hashonly, sqlite]",
	)
	flags.StringVarP(
		&processor.AuditFile,
		"audit",
		"a",
		"",
		"audit against supplied file; accepts hashdeep, sum (md5sum, sha256sum etc...) or hashit sqlite format",
	)
	flags.BoolVarP(
		&processor.Recursive,
		"recursive",
		"r",
		false,
		"traverse directories when multiple paths are supplied; a single directory or no path is always traversed",
	)
	flags.BoolVar(
		&processor.Hashes,
		"hashes",
		false,
		"list all supported hashes",
	)
	flags.StringVarP(
		&processor.FileOutput,
		"output",
		"o",
		"",
		"output filename (default stdout, or hashit.db for the sqlite format)",
	)
	flags.BoolVar(
		&processor.AllowMoved,
		"allow-moved",
		false,
		"allow an audit to pass when known files have moved or been renamed but are unmodified",
	)
	flags.BoolVar(
		&processor.Absolute,
		"absolute",
		false,
		"output absolute paths, which is what hashdeep does without -l",
	)
	flags.BoolVar(
		&processor.Overwrite,
		"overwrite",
		false,
		"allow output to replace an existing regular file; a symlink which cannot be followed is always refused",
	)
	flags.BoolVar(
		&processor.NoStream,
		"no-stream",
		false,
		"do not print results as files are processed for text, sum and hashonly formats; other formats and --output never stream",
	)
	flags.Int64Var(
		&processor.StreamSize,
		"stream-size",
		1_000_000,
		"files larger than this size in bytes are read in chunks rather than all at once",
	)
	flags.CountVarP(
		&verboseCount,
		"verbose",
		"v",
		"verbose audit output, listing every modified, moved, new and missing file; -vv is the same as --vv",
	)
	flags.BoolVar(
		&processor.VeryVerbose,
		"vv",
		false,
		"very verbose audit output, as --verbose and also listing every matched file",
	)
	flags.BoolVarP(
		&processor.Progress,
		"progress",
		"p",
		false,
		"display progress of files as they are processed",
	)
	flags.BoolVar(
		&processor.Debug,
		"debug",
		false,
		"enable debug output",
	)
	flags.BoolVar(
		&processor.Trace,
		"trace",
		false,
		"enable trace output",
	)
	flags.IntVar(
		&processor.NoThreads,
		"threads",
		runtime.NumCPU(),
		"number of threads processing files, by default the number of CPU cores",
	)
	flags.BoolVar(
		&processor.MTime,
		"mtime",
		false,
		"enable mtime output",
	)
	flags.StringVarP(
		&processor.FileInput,
		"input",
		"i",
		"",
		"input file of newline separated file locations to process",
	)

	flags.BoolVar(
		&processor.GitIgnore,
		"gitignore",
		false,
		"enable .gitignore file logic",
	)
	flags.BoolVar(
		&processor.GitModuleIgnore,
		"gitmodule",
		false,
		"enable .gitmodules file logic",
	)

	flags.BoolVar(
		&processor.SkipHidden,
		"skip-hidden",
		false, // Default to false, meaning hidden files are PROCESSED by default
		"skip hidden files and directories when walking directories",
	)
	flags.StringSliceVar(
		&processor.PathDenyList,
		"exclude-dir",
		[]string{},
		"directories to exclude when walking directories, by name such as node_modules or the end of a path such as vendor/cache",
	)
	flags.BoolVar(
		&processor.Ignore,
		"ignore",
		false,
		"enable .ignore file logic",
	)
	flags.BoolVar(
		&processor.HashIgnore,
		"hashignore",
		false,
		"enable .hashignore file logic",
	)
	flags.StringArrayVarP(
		&processor.Exclude,
		"not-match",
		`M`,
		[]string{},
		"ignore files and directories whose name matches regular expression when walking directories",
	)

	// an invalid flag or argument is an error the same as any other, and is printed the same way
	rootCmd.SilenceErrors = true
	if err := rootCmd.Execute(); err != nil {
		processor.PrintError(err.Error())
		os.Exit(processor.ExitError)
	}
}
