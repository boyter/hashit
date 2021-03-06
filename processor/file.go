package processor

import (
	"fmt"
	"github.com/karrick/godirwalk"
	"strings"
)

func walkDirectory(toWalk string, output chan string) {
	err := godirwalk.Walk(toWalk, &godirwalk.Options{
		Unsorted: false, // We want the run to be deterministic
		Callback: func(root string, info *godirwalk.Dirent) error {
			if !info.IsDir() {
				output <- root
			}

			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			if Verbose {
				printVerbose(fmt.Sprintf("error walking: %s %s", osPathname, err))
			}
			return godirwalk.SkipNode
		},
	})

	// If err and we get specific error it's a file which we want to process
	if err != nil && strings.Contains(err.Error(), "cannot Walk non-directory") {
		output <- toWalk
	}
}
