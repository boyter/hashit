package main

import (
	"github.com/boyter/hashit/processor"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

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
			processor.DirFilePaths = args
			processor.Process()
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringSliceVarP(
		&processor.Hash,
		"hash",
		"c",
		[]string{"md5", "sha1", "sha256", "sha512"},
		"hashes to be run for each file (set to 'all' for all possible hashes)",
	)
	flags.StringVarP(
		&processor.Format,
		"format",
		"f",
		"text",
		"set output format [text, json, sum, hashdeep, hashonly]",
	)
	flags.BoolVarP(
		&processor.Recursive,
		"recursive",
		"r",
		false,
		"recursive subdirectories are traversed",
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
		"output filename (default stdout)",
	)
	flags.BoolVar(
		&processor.NoStream,
		"no-stream",
		false,
		"do not stream out results as processed",
	)
	flags.Int64Var(
		&processor.StreamSize,
		"stream-size",
		1000000,
		"min size of file in bytes where stream processing starts",
	)
	flags.BoolVarP(
		&processor.Verbose,
		"verbose",
		"v",
		false,
		"verbose output",
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
		"input file of newline seperated file locations to process",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
