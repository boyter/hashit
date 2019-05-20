package processor

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	mmapgo "github.com/edsrzf/mmap-go"
	"github.com/minio/blake2b-simd"
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
		if fi.Size() > 1 {
			// If Windows ignore memory maps and stream the file off disk
			if runtime.GOOS == "windows" {
			} else {
				// Memory map the file and process
				processMemoryMap(res)
			}

		} else {
			// Suck the file into memory and process
			content, err := ioutil.ReadFile(res)
			if err != nil {
				printError(fmt.Sprintf("Unable to read file %s into memory with error %s", res, err.Error()))
				continue
			}

			processContent(content, res)
		}
	}
}

func processMemoryMap(filename string) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)

	if err != nil {
		panic(err.Error())
	}

	mmap, err := mmapgo.Map(file, mmapgo.RDONLY, 0)

	if err != nil {
		fmt.Println("error mapping:", err)
	}

	//

	// Create channels for each hash
	md5_digest := md5.New()
	sha1_digest := sha1.New()
	sha256_digest := sha256.New()
	sha512_digest := sha512.New()
	blake2b_digest := blake2b.New256()

	md5c := make(chan []byte, 10)
	sha1c := make(chan []byte, 10)
	sha256c := make(chan []byte, 10)
	sha512c := make(chan []byte, 10)
	blake2bc := make(chan []byte, 10)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for b := range md5c {
			md5_digest.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range sha1c {
			sha1_digest.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range sha256c {
			sha256_digest.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range sha512c {
			sha512_digest.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range blake2bc {
			blake2b_digest.Write(b)
		}
		wg.Done()
	}()

	total := len(mmap)

	for i := 0; i < total; i += 1000000 {
		end := i + 1000000
		if end > total {
			end = total
		}

		md5c <- mmap[i:end]
		sha1c <- mmap[i:end]
		sha256c <- mmap[i:end]
		sha512c <- mmap[i:end]
		blake2bc <- mmap[i:end]
	}
	close(md5c)
	close(sha1c)
	close(sha256c)
	close(sha512c)
	close(blake2bc)

	wg.Wait()

	md5_string := hex.EncodeToString(md5_digest.Sum(nil))
	sha1_string := hex.EncodeToString(sha1_digest.Sum(nil))
	sha256_string := hex.EncodeToString(sha256_digest.Sum(nil))
	sha512_string := hex.EncodeToString(sha512_digest.Sum(nil))
	blake2bc_string := hex.EncodeToString(blake2b_digest.Sum(nil))

	fmt.Println(filename)
	fmt.Println("    MD5 " + md5_string)
	fmt.Println("   SHA1 " + sha1_string)
	fmt.Println(" SHA256 " + sha256_string)
	fmt.Println(" SHA512 " + sha512_string)
	fmt.Println("Blake2b " + blake2bc_string)
	fmt.Println("")

	if err := mmap.Unmap(); err != nil {
		fmt.Println("error unmapping:", err)
	}
}

func Mmap() {
	file, err := os.OpenFile("main.go", os.O_RDONLY, 0644)

	if err != nil {
		panic(err.Error())
	}

	mmap, err := mmapgo.Map(file, mmapgo.RDONLY, 0)

	fmt.Println("Length", len(mmap))

	count := 0
	for _, currentByte := range mmap {
		if currentByte == '\n' {
			count++
		}
	}

	fmt.Println("Newlines", count)

	if err != nil {
		fmt.Println("error mapping:", err)
	}

	if err := mmap.Unmap(); err != nil {
		fmt.Println("error unmapping:", err)
	}
}

func processContent(content []byte, filename string) {
	var wg sync.WaitGroup
	md5_string := ""
	sha1_string := ""
	sha256_string := ""
	sha512_string := ""
	wg.Add(1)
	go func(c []byte) {
		md5_digest := md5.New()
		md5_digest.Write(c)
		md5_string = hex.EncodeToString(md5_digest.Sum(nil))
		wg.Done()
	}(content)
	wg.Add(1)
	go func(c []byte) {
		sha1_digest := sha1.New()
		sha1_digest.Write(c)
		sha1_string = hex.EncodeToString(sha1_digest.Sum(nil))
		wg.Done()
	}(content)
	wg.Add(1)
	go func(c []byte) {
		sha256_digest := sha256.New()
		sha256_digest.Write(c)
		sha256_string = hex.EncodeToString(sha256_digest.Sum(nil))
		wg.Done()
	}(content)
	wg.Add(1)
	go func(c []byte) {
		sha512_digest := sha512.New()
		sha512_digest.Write(c)
		sha512_string = hex.EncodeToString(sha512_digest.Sum(nil))
		wg.Done()
	}(content)
	wg.Wait()
	fmt.Println(filename)
	fmt.Println("   MD5 " + md5_string)
	fmt.Println("  SHA1 " + sha1_string)
	fmt.Println("SHA256 " + sha256_string)
	fmt.Println("SHA512 " + sha512_string)
	fmt.Println("")
}
