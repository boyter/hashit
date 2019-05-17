package processor

import (
	"fmt"
	"github.com/karrick/godirwalk"
)

func walkDirectory(toWalk string, output chan string) {
	_ = godirwalk.Walk(toWalk, &godirwalk.Options{
		// Unsorted is meant to make the walk faster and we need to sort after processing anyway
		Unsorted: true,
		Callback: func(root string, info *godirwalk.Dirent) error {
			if !info.IsDir() {
				output <- info.Name()
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
}
