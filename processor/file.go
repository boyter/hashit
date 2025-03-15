// SPDX-License-Identifier: MIT

package processor

import (
	"fmt"
	"os"
	"path/filepath"
)

func walkDirectory(toWalk string, output chan string) {

	walkErr := filepath.WalkDir(toWalk, func(root string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			output <- root
		}

		return nil
	})

	if walkErr != nil {
		if Verbose {
			printVerbose(fmt.Sprintf("error walking: %s", toWalk))
		}
	}
}
