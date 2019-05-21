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
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
)

func fileProcessorWorker(input chan string, output chan Result) {
	for res := range input {

		if Debug {
			printDebug(fmt.Sprintf("processing %s", res))
		}

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
		_ = file.Close()

		if fsize > StreamSize {
			// If Windows always ignore memory maps and stream the file off disk
			if runtime.GOOS == "windows" || NoMmap == true {

				if Debug {
					printDebug(fmt.Sprintf("%s bytes=%d using scanner", res, fsize))
				}

				// TODO should return a struct with the values we have
				processScanner(res)
			} else {

				if Debug {
					printDebug(fmt.Sprintf("%s bytes=%d using memory map", res, fsize))
				}

				r, err := processMemoryMap(res)
				if err == nil {
					r.File = res
					r.Bytes = fsize
					output <- r
				}
			}

		} else {
			if Debug {
				printDebug(fmt.Sprintf("%s bytes=%d using read file", res, fsize))
			}

			r, err := processReadFile(res)
			if err == nil {
				r.File = res
				r.Bytes = fsize
				output <- r
			}
		}
	}
	close(output)
}

// TODO compare this to memory maps
// Random tests indicate that mmap is faster when not in power save mode
func processScanner(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		printError(fmt.Sprintf("opening file %s: %s", file, err.Error()))
		return
	}
	defer file.Close()

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

	if hasHash("md5") {
		wg.Add(1)
		go func() {
			for b := range md5c {
				md5_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("sha1") {
		wg.Add(1)
		go func() {
			for b := range sha1c {
				sha1_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("sha256") {
		wg.Add(1)
		go func() {
			for b := range sha256c {
				sha256_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("sha512") {
		wg.Add(1)
		go func() {
			for b := range sha512c {
				sha512_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("blake2b256") {
		wg.Add(1)
		go func() {
			for b := range blake2b_256_c {
				blake2b_256_d.Write(b)
			}
			wg.Done()
		}()
	}


	data := make([]byte, 8192) // 8192 appears to be optimal
	for {
		data = data[:cap(data)]
		n, err := file.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}

			printError(fmt.Sprintf("reading file %s: %s", filename, err.Error()))
			return
		}

		data = data[:n]

		if hasHash("md5") {
			md5c <- data
		}
		if hasHash("sha1") {
			sha1c <-data
		}
		if hasHash("sha256") {
			sha256c <-data
		}
		if hasHash("sha512") {
			sha512c <- data
		}
		if hasHash("blake2b256") {
			blake2b_256_c <- data
		}
	}

	close(md5c)
	close(sha1c)
	close(sha256c)
	close(sha512c)
	close(blake2b_256_c)

	wg.Wait()

	fmt.Println(filename)
	if hasHash("md5") {
		fmt.Println("        MD5 " + hex.EncodeToString(md5_d.Sum(nil)))
	}
	if hasHash("sha1") {
		fmt.Println("       SHA1 " + hex.EncodeToString(sha1_d.Sum(nil)))
	}
	if hasHash("sha256") {
		fmt.Println("     SHA256 " + hex.EncodeToString(sha256_d.Sum(nil)))
	}
	if hasHash("sha512") {
		fmt.Println("     SHA512 " + hex.EncodeToString(sha512_d.Sum(nil)))
	}
	if hasHash("blake2b256") {
		fmt.Println("Blake2b 256 " + hex.EncodeToString(blake2b_256_d.Sum(nil)))
	}
	fmt.Println("")
}

// For files over a certain size it is faster to process them using
// memory mapped files which this method does
// NB this does not play well with Windows as it will never
// be able to unmap the file "error unmapping: FlushFileBuffers: Access is denied."
func processMemoryMap(filename string) (Result, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		printError(fmt.Sprintf("opening file %s: %s", file, err.Error()))
		return Result{}, err
	}

	mmap, err := mmapgo.Map(file, mmapgo.RDONLY, 0)
	if err != nil {
		printError(fmt.Sprintf("mapping file %s: %s", filename, err.Error()))
		return Result{}, err
	}

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

	if hasHash("md5") {
		wg.Add(1)
		go func() {
			for b := range md5c {
				md5_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("sha1") {
		wg.Add(1)
		go func() {
			for b := range sha1c {
				sha1_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("sha256") {
		wg.Add(1)
		go func() {
			for b := range sha256c {
				sha256_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("sha512") {
		wg.Add(1)
		go func() {
			for b := range sha512c {
				sha512_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash("blake2b256") {
		wg.Add(1)
		go func() {
			for b := range blake2b_256_c {
				blake2b_256_d.Write(b)
			}
			wg.Done()
		}()
	}

	total := len(mmap)

	for i := 0; i < total; i += 1000000 {
		end := i + 1000000
		if end > total {
			end = total
		}

		if hasHash("md5") {
			md5c <- mmap[i:end]
		}
		if hasHash("sha1") {
			sha1c <- mmap[i:end]
		}
		if hasHash("sha256") {
			sha256c <- mmap[i:end]
		}
		if hasHash("sha512") {
			sha512c <- mmap[i:end]
		}
		if hasHash("blake2b256") {
			blake2b_256_c <- mmap[i:end]
		}
	}

	close(md5c)
	close(sha1c)
	close(sha256c)
	close(sha512c)
	close(blake2b_256_c)

	wg.Wait()

	if err := mmap.Unmap(); err != nil {
		printError(fmt.Sprintf("unmapping file %s: %s", filename, err.Error()))
	}

	return Result{
		File: filename,
		Bytes: int64(total),
		MD5: hex.EncodeToString(md5_d.Sum(nil)),
		SHA1: hex.EncodeToString(sha1_d.Sum(nil)),
		SHA256: hex.EncodeToString(sha256_d.Sum(nil)),
		SHA512: hex.EncodeToString(sha512_d.Sum(nil)),
		Blake2b: hex.EncodeToString(blake2b_256_d.Sum(nil)),
	}, nil
}

// For files under a certain size its faster to just read them into memory in one
// chunk and then process them which this method does
// NB there is little point in multi-processing at this level, it would be
// better done on the input channel if required
func processReadFile(filename string) (Result, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		printError(fmt.Sprintf("Unable to read file %s into memory with error %s", filename, err.Error()))
		return Result{}, err
	}

	result := Result{}

	if hasHash("md5") {
		md5_digest := md5.New()
		md5_digest.Write(content)
		result.MD5 = hex.EncodeToString(md5_digest.Sum(nil))
	}

	if hasHash("sha1") {
		sha1_digest := sha1.New()
		sha1_digest.Write(content)
		result.SHA1 = hex.EncodeToString(sha1_digest.Sum(nil))
	}

	if hasHash("sha256") {
		sha256_digest := sha256.New()
		sha256_digest.Write(content)
		result.SHA256 = hex.EncodeToString(sha256_digest.Sum(nil))
	}

	if hasHash("sha512") {
		sha512_digest := sha512.New()
		sha512_digest.Write(content)
		result.SHA512 = hex.EncodeToString(sha512_digest.Sum(nil))
	}

	if hasHash("blake2b256") {
		blake2bs_256_digest := blake2b.New256()
		blake2bs_256_digest.Write(content)
		result.Blake2b = hex.EncodeToString(blake2bs_256_digest.Sum(nil))
	}

	return result, nil
}
