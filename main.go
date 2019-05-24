package main

import (
	"github.com/boyter/hashit/processor"
	"github.com/spf13/cobra"
	"os"
)

//go:generate go run scripts/include.go
func main() {
	//f, _ := os.Create("hashit.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use:     "hashit",
		Short:   "hashit [FILE or DIRECTORY]",
		Long:    "Hash It!\nBen Boyter <ben@boyter.org> + Contributors",
		Version: processor.Version,
		Run: func(cmd *cobra.Command, args []string) {
			processor.DirFilePaths = args
			processor.Process()
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringSliceVar(
		&processor.Hash,
		"hash",
		[]string{"md5", "sha1", "sha256", "sha512"},
		"hashes to be run for each file (set to 'all' for all possible hashes)",
	)
	flags.StringVarP(
		&processor.Format,
		"format",
		"f",
		"text",
		"set output format [text, json, hashdeep]",
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
		&processor.NoMmap,
		"no-mmap",
		false,
		"never use memory maps",
	)
	flags.BoolVar(
		&processor.NoStream,
		"no-stream",
		false,
		"do not stream out results as processed",
	)
	flags.BoolVarP(
		&processor.Verbose,
		"verbose",
		"v",
		false,
		"verbose output",
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
