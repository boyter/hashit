package main

import (
	"github.com/boyter/hashit/processor"
	"github.com/spf13/cobra"
	"os"
)

var verbose = false

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

	flags.BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"verbose output",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
