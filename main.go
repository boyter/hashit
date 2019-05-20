package main

import (
	"github.com/boyter/hashit/processor"
	"github.com/spf13/cobra"
	"os"
)

//go:generate go run scripts/include.go
func main() {
	rootCmd := &cobra.Command{
		Use:     "hashit",
		Short:   "hashit [FILE or DIRECTORY]",
		Long:    "Hash It!\nBen Boyter <ben@boyter.org> + Contributors",
		Version: "0.1.0",
		Run: func(cmd *cobra.Command, args []string) {
			processor.DirFilePaths = args
			processor.Process()
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringSliceVar(
		&processor.Hashes,
		"hashes",
		[]string{"md5", "sha1", "sha512"},
		"hashes to be run for each file (set to all for all possible hashes)",
	)
	flags.BoolVar(
		&processor.NoMmap,
		"no-mmap",
		false,
		"never use memory maps",
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
		"enabled debug output",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
