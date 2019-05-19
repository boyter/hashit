package processor

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
)

func fileProcessorWorker(input chan string) {
	for res := range input {
		// Open the file and determine if we should read it from disk or memory map
		file, err := os.OpenFile(res, os.O_RDONLY, 0644)

		if err != nil {
			printError(fmt.Sprintf("Unable to process file %s with error %s", res, err.Error()))
			continue
		}

		defer file.Close()
		fi, err := file.Stat()

		if err != nil {
			printError(fmt.Sprintf("Unable to get file info for file %s with error %s", res, err.Error()))
			continue
		}

		// Greater than 1 million bytes
		if fi.Size() > 1000000 {
			// If Windows ignore memory maps and stream the file off disk
			if runtime.GOOS == "windows" {
			} else {

			}

		} else {
			// Suck the file into memory and process
			content, err := ioutil.ReadFile(res)
			if err != nil {
				printError(fmt.Sprintf("Unable to read file %s into memory with error %s", res, err.Error()))
				continue
			}

			var wg sync.WaitGroup

			wg.Add(1)
			md5_string := ""
			go func() {
				md5_digest := md5.New()
				md5_digest.Write(content)
				md5_string = hex.EncodeToString(md5_digest.Sum(nil))
				wg.Done()
			}()


			////sha1_digest := sha1.New()
			////sha256_digest := sha256.New()
			////sha512_digest := sha512.New()

			fmt.Println()
		}
	}
}
