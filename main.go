package main

import (
	"fmt"
	mmapgo "github.com/edsrzf/mmap-go"
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
			fmt.Println(args)
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

	file, err := os.OpenFile("main.go", os.O_RDONLY, 0644)

	if err != nil {
		panic(err.Error())
	}

	mmap, err := mmapgo.Map(file, mmapgo.RDONLY, 0)

	fmt.Println(len(mmap))

	count := 0
	for _, currentByte := range mmap {
		if currentByte == '\n' {
			count++
		}
	}

	fmt.Println(count)

	if err != nil {
		fmt.Println("error mapping:", err)
	}

	if err := mmap.Unmap(); err != nil {
		fmt.Println("error unmapping:", err)
	}
}
