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

		fi, err := file.Stat()

		if err != nil {
			printError(fmt.Sprintf("Unable to get file info for file %s with error %s", res, err.Error()))
			continue
		}
		fsize := fi.Size()
		file.Close()

		// Greater than 1 million bytes
		if fsize > 1000000 {
			// If Windows ignore memory maps and stream the file off disk
			//if runtime.GOOS == "windows" {
			//	// Should be done like memory map
			//	scanner := bufio.NewScanner(file)
			//
			//	for scanner.Scan() {
			//		scanner.Bytes()
			//	}
			//} else {
			if Debug {
				printDebug(fmt.Sprintf("%s size = %d using memory map", res, fsize))
			}

			// Memory map the file and process
			processMemoryMap(res)
			//}

		} else {
			processReadFile(res)
		}
	}
}

// For files over a certain size it is faster to process them using
// memory mapped files which this method does
// NB this does not play well with Windows as it will never
// be able to unmap the file "error unmapping: FlushFileBuffers: Access is denied."
func processMemoryMap(filename string) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		printError(fmt.Sprintf("opening file %s: %s", file, err.Error()))
		return
	}

	mmap, err := mmapgo.Map(file, mmapgo.RDONLY, 0)
	if err != nil {
		printError(fmt.Sprintf("mapping file %s: %s", filename, err.Error()))
		return
	}

	// Create channels for each hash
	md5_d := md5.New()
	sha1_d := sha1.New()
	sha256_d := sha256.New()
	sha512_d := sha512.New()
	blake2b_256_d := blake2b.New256()

	md5c := make(chan []byte, 10)
	sha1c := make(chan []byte, 10)
	sha256c := make(chan []byte, 10)
	sha512c := make(chan []byte, 10)
	blake2b_256_c := make(chan []byte, 10)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for b := range md5c {
			md5_d.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range sha1c {
			sha1_d.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range sha256c {
			sha256_d.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range sha512c {
			sha512_d.Write(b)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range blake2b_256_c {
			blake2b_256_d.Write(b)
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
		blake2b_256_c <- mmap[i:end]
	}
	close(md5c)
	close(sha1c)
	close(sha256c)
	close(sha512c)
	close(blake2b_256_c)

	wg.Wait()

	fmt.Println(filename)
	fmt.Println("        MD5 " + hex.EncodeToString(md5_d.Sum(nil)))
	fmt.Println("       SHA1 " + hex.EncodeToString(sha1_d.Sum(nil)))
	fmt.Println("     SHA256 " + hex.EncodeToString(sha256_d.Sum(nil)))
	fmt.Println("     SHA512 " + hex.EncodeToString(sha512_d.Sum(nil)))
	fmt.Println("Blake2b 256 " + hex.EncodeToString(blake2b_256_d.Sum(nil)))
	fmt.Println("")

	if err := mmap.Unmap(); err != nil {
		printError(fmt.Sprintf("unmapping file %s: %s", filename, err.Error()))
	}
}

// For files under a certain size its faster to just read them into memory in one
// chunk and then process them which this method does
// NB there is little point in multi-processing at this level, it would be
// better done on the input channel if required
func processReadFile(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		printError(fmt.Sprintf("Unable to read file %s into memory with error %s", filename, err.Error()))
		return
	}

	md5_digest := md5.New()
	md5_digest.Write(content)
	sha1_digest := sha1.New()
	sha1_digest.Write(content)
	sha256_digest := sha256.New()
	sha256_digest.Write(content)
	sha512_digest := sha512.New()
	sha512_digest.Write(content)
	blake2bs_256_digest := blake2b.New256()
	blake2bs_256_digest.Write(content)

	fmt.Println(filename)
	fmt.Println("        MD5 " + hex.EncodeToString(md5_digest.Sum(nil)))
	fmt.Println("      SHA-1 " + hex.EncodeToString(sha1_digest.Sum(nil)))
	fmt.Println("    SHA-256 " + hex.EncodeToString(sha256_digest.Sum(nil)))
	fmt.Println("    SHA-512 " + hex.EncodeToString(sha512_digest.Sum(nil)))
	fmt.Println("Blake2b-256 " + hex.EncodeToString(blake2bs_256_digest.Sum(nil)))
	fmt.Println("")
}
