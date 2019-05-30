package processor

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/minio/blake2b-simd"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/sha3"
	"io"
	"log"
	"os"
	"sync"
)

func fileProcessorWorker(input chan string, output chan Result) {
	for res := range input {
		if Debug {
			printDebug(fmt.Sprintf("processing %s", res))
		}

		// Open the file and determine if we should read it from disk or memory map
		// based on how large it is reported as being
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

		if fsize > StreamSize {
			if Debug {
				printDebug(fmt.Sprintf("%s bytes=%d using scanner", res, fsize))
			}

			fileStartTime := makeTimestampMilli()
			r, err := processScanner(res)
			if Trace {
				printTrace(fmt.Sprintf("milliseconds processMemoryMap: %s: %d", res, makeTimestampMilli()-fileStartTime))
			}

			if err == nil {
				r.File = res
				r.Bytes = fsize
				output <- r
			}

		} else {
			if Debug {
				printDebug(fmt.Sprintf("%s bytes=%d using read file", res, fsize))
			}

			fileStartTime := makeTimestampNano()

			var n int64 = bytes.MinRead
			if size := fsize + bytes.MinRead; size > n {
				n = size
			}
			content, _ := readAll(file, n)

			var r Result

			// For larger files if we have more than one hash try parallel
			if fsize > 200000 && len(Hash) >= 1 && !hasHash("all") {
				r, err = processReadFileParallel(res, &content)
			} else {
				r, err = processReadFile(res, &content)
			}

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processReadFileParallel: %s: %d", res, makeTimestampNano()-fileStartTime))
			}

			if err == nil {
				r.File = res
				r.Bytes = fsize
				output <- r
			}
		}
		_ = file.Close()
	}
}

// TODO compare this to memory maps
// Random tests indicate that mmap is faster when not in power save mode
func processScanner(filename string) (Result, error) {
	file, err := os.Open(filename)
	if err != nil {
		printError(fmt.Sprintf("opening file %s: %s", filename, err.Error()))
		return Result{}, err
	}
	defer file.Close()

	md4_d := md4.New()
	md5_d := md5.New()
	sha1_d := sha1.New()
	sha256_d := sha256.New()
	sha512_d := sha512.New()
	blake2b_256_d := blake2b.New256()
	blake2b_512_d := blake2b.New512()
	sha3_224_d := sha3.New224()
	sha3_256_d := sha3.New256()
	sha3_384_d := sha3.New384()
	sha3_512_d := sha3.New512()

	md4c := make(chan []byte, 10)
	md5c := make(chan []byte, 10)
	sha1c := make(chan []byte, 10)
	sha256c := make(chan []byte, 10)
	sha512c := make(chan []byte, 10)
	blake2b_256_c := make(chan []byte, 10)
	blake2b_512_c := make(chan []byte, 10)
	sha3_224_c := make(chan []byte, 10)
	sha3_256_c := make(chan []byte, 10)
	sha3_384_c := make(chan []byte, 10)
	sha3_512_c := make(chan []byte, 10)

	var wg sync.WaitGroup

	if hasHash(HashNames.MD4) {
		wg.Add(1)
		go func() {
			for b := range md4c {
				md4_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.MD5) {
		wg.Add(1)
		go func() {
			for b := range md5c {
				md5_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA1) {
		wg.Add(1)
		go func() {
			for b := range sha1c {
				sha1_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA256) {
		wg.Add(1)
		go func() {
			for b := range sha256c {
				sha256_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA512) {
		wg.Add(1)
		go func() {
			for b := range sha512c {
				sha512_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Blake2b256) {
		wg.Add(1)
		go func() {
			for b := range blake2b_256_c {
				blake2b_256_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Blake2b512) {
		wg.Add(1)
		go func() {
			for b := range blake2b_512_c {
				blake2b_512_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Sha3224) {
		wg.Add(1)
		go func() {
			for b := range sha3_224_c {
				sha3_224_d.Write(b)
			}
			wg.Done()
		}()
	}
	if hasHash(HashNames.Sha3256) {
		wg.Add(1)
		go func() {
			for b := range sha3_256_c {
				sha3_256_d.Write(b)
			}
			wg.Done()
		}()
	}
	if hasHash(HashNames.Sha3384) {
		wg.Add(1)
		go func() {
			for b := range sha3_384_c {
				sha3_384_d.Write(b)
			}
			wg.Done()
		}()
	}
	if hasHash(HashNames.Sha3512) {
		wg.Add(1)
		go func() {
			for b := range sha3_512_c {
				sha3_512_d.Write(b)
			}
			wg.Done()
		}()
	}

	data := make([]byte, 4194304)
	for {
		n, err := file.Read(data)
		if err != nil && err != io.EOF {
			printError(fmt.Sprintf("reading file %s: %s", filename, err.Error()))
			return Result{}, err
		}

		// Need to make a copy here as it can be modified before
		// the goroutine processes it in the channel
		tmp := make([]byte, len(data))
		copy(tmp, data)

		if hasHash(HashNames.MD4) {
			md4c <- tmp[:n]
		}
		if hasHash(HashNames.MD5) {
			md5c <- tmp[:n]
		}
		if hasHash(HashNames.SHA1) {
			sha1c <- tmp[:n]
		}
		if hasHash(HashNames.SHA256) {
			sha256c <- tmp[:n]
		}
		if hasHash(HashNames.SHA512) {
			sha512c <- tmp[:n]
		}
		if hasHash(HashNames.Blake2b256) {
			blake2b_256_c <- tmp[:n]
		}
		if hasHash(HashNames.Blake2b512) {
			blake2b_512_c <- tmp[:n]
		}
		if hasHash(HashNames.Sha3224) {
			sha3_224_c <- tmp[:n]
		}
		if hasHash(HashNames.Sha3256) {
			sha3_256_c <- tmp[:n]
		}
		if hasHash(HashNames.Sha3384) {
			sha3_384_c <- tmp[:n]
		}
		if hasHash(HashNames.Sha3512) {
			sha3_512_c <- tmp[:n]
		}

		if err == io.EOF {
			break
		}
	}

	close(md4c)
	close(md5c)
	close(sha1c)
	close(sha256c)
	close(sha512c)
	close(blake2b_256_c)
	close(blake2b_512_c)
	close(sha3_224_c)
	close(sha3_256_c)
	close(sha3_384_c)
	close(sha3_512_c)

	wg.Wait()

	return Result{
		File:       filename,
		Bytes:      0,
		MD4:        hex.EncodeToString(md4_d.Sum(nil)),
		MD5:        hex.EncodeToString(md5_d.Sum(nil)),
		SHA1:       hex.EncodeToString(sha1_d.Sum(nil)),
		SHA256:     hex.EncodeToString(sha256_d.Sum(nil)),
		SHA512:     hex.EncodeToString(sha512_d.Sum(nil)),
		Blake2b256: hex.EncodeToString(blake2b_256_d.Sum(nil)),
		Blake2b512: hex.EncodeToString(blake2b_512_d.Sum(nil)),
		Sha3224:    hex.EncodeToString(sha3_224_d.Sum(nil)),
		Sha3256:    hex.EncodeToString(sha3_256_d.Sum(nil)),
		Sha3384:    hex.EncodeToString(sha3_384_d.Sum(nil)),
		Sha3512:    hex.EncodeToString(sha3_512_d.Sum(nil)),
	}, nil
}

func processStandardInput(output chan Result) {
	total, nChunks := int64(0), int64(0)
	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 0, 4*1024)

	md4_d := md4.New()
	md5_d := md5.New()
	sha1_d := sha1.New()
	sha256_d := sha256.New()
	sha512_d := sha512.New()
	blake2b_256_d := blake2b.New256()
	blake2b_512_d := blake2b.New512()
	sha3_224_d := sha3.New224()
	sha3_256_d := sha3.New256()
	sha3_384_d := sha3.New384()
	sha3_512_d := sha3.New512()

	md4c := make(chan []byte, 10)
	md5c := make(chan []byte, 10)
	sha1c := make(chan []byte, 10)
	sha256c := make(chan []byte, 10)
	sha512c := make(chan []byte, 10)
	blake2b_256_c := make(chan []byte, 10)
	blake2b_512_c := make(chan []byte, 10)
	sha3_224_c := make(chan []byte, 10)
	sha3_256_c := make(chan []byte, 10)
	sha3_384_c := make(chan []byte, 10)
	sha3_512_c := make(chan []byte, 10)

	var wg sync.WaitGroup

	if hasHash(HashNames.MD4) {
		wg.Add(1)
		go func() {
			for b := range md4c {
				md4_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.MD5) {
		wg.Add(1)
		go func() {
			for b := range md5c {
				md5_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA1) {
		wg.Add(1)
		go func() {
			for b := range sha1c {
				sha1_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA256) {
		wg.Add(1)
		go func() {
			for b := range sha256c {
				sha256_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA512) {
		wg.Add(1)
		go func() {
			for b := range sha512c {
				sha512_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Blake2b256) {
		wg.Add(1)
		go func() {
			for b := range blake2b_256_c {
				blake2b_256_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Blake2b512) {
		wg.Add(1)
		go func() {
			for b := range blake2b_512_c {
				blake2b_512_d.Write(b)
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Sha3224) {
		wg.Add(1)
		go func() {
			for b := range sha3_224_c {
				sha3_224_d.Write(b)
			}
			wg.Done()
		}()
	}
	if hasHash(HashNames.Sha3256) {
		wg.Add(1)
		go func() {
			for b := range sha3_256_c {
				sha3_256_d.Write(b)
			}
			wg.Done()
		}()
	}
	if hasHash(HashNames.Sha3384) {
		wg.Add(1)
		go func() {
			for b := range sha3_384_c {
				sha3_384_d.Write(b)
			}
			wg.Done()
		}()
	}
	if hasHash(HashNames.Sha3512) {
		wg.Add(1)
		go func() {
			for b := range sha3_512_c {
				sha3_512_d.Write(b)
			}
			wg.Done()
		}()
	}

	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]

		if n == 0 {
			if err == nil {
				continue
			}

			if err == io.EOF {
				break
			}

			log.Fatal(err)
		}

		nChunks++
		total += int64(len(buf))

		if hasHash(HashNames.MD4) {
			md4c <- buf
		}
		if hasHash(HashNames.MD5) {
			md5c <- buf
		}
		if hasHash(HashNames.SHA1) {
			sha1c <- buf
		}
		if hasHash(HashNames.SHA256) {
			sha256c <- buf
		}
		if hasHash(HashNames.SHA512) {
			sha512c <- buf
		}
		if hasHash(HashNames.Blake2b256) {
			blake2b_256_c <- buf
		}
		if hasHash(HashNames.Blake2b512) {
			blake2b_512_c <- buf
		}
		if hasHash(HashNames.Sha3224) {
			sha3_224_c <- buf
		}
		if hasHash(HashNames.Sha3256) {
			sha3_256_c <- buf
		}
		if hasHash(HashNames.Sha3384) {
			sha3_384_c <- buf
		}
		if hasHash(HashNames.Sha3512) {
			sha3_512_c <- buf
		}

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}

	close(md4c)
	close(md5c)
	close(sha1c)
	close(sha256c)
	close(sha512c)
	close(blake2b_256_c)
	close(blake2b_512_c)
	close(sha3_224_c)
	close(sha3_256_c)
	close(sha3_384_c)
	close(sha3_512_c)

	wg.Wait()

	output <- Result{
		File:       "stdin",
		Bytes:      total,
		MD4:        hex.EncodeToString(md4_d.Sum(nil)),
		MD5:        hex.EncodeToString(md5_d.Sum(nil)),
		SHA1:       hex.EncodeToString(sha1_d.Sum(nil)),
		SHA256:     hex.EncodeToString(sha256_d.Sum(nil)),
		SHA512:     hex.EncodeToString(sha512_d.Sum(nil)),
		Blake2b256: hex.EncodeToString(blake2b_256_d.Sum(nil)),
		Blake2b512: hex.EncodeToString(blake2b_512_d.Sum(nil)),
		Sha3224:    hex.EncodeToString(sha3_224_d.Sum(nil)),
		Sha3256:    hex.EncodeToString(sha3_256_d.Sum(nil)),
		Sha3384:    hex.EncodeToString(sha3_384_d.Sum(nil)),
		Sha3512:    hex.EncodeToString(sha3_512_d.Sum(nil)),
	}

	close(output)
}

// For files under a certain size its faster to just read them into memory in one
// chunk and then process them which this method does
// NB there is little point in multi-processing at this level, it would be
// better done on the input channel if required
func processReadFileParallel(filename string, content *[]byte) (Result, error) {
	startTime := makeTimestampNano()

	var wg sync.WaitGroup
	result := Result{}

	if hasHash(HashNames.MD4) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := md4.New()
			d.Write(*content)
			result.MD4 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing md4: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.MD5) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := md5.New()
			d.Write(*content)
			result.MD5 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing md5: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA1) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := sha1.New()
			d.Write(*content)
			result.SHA1 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing sha1: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA256) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := sha256.New()
			d.Write(*content)
			result.SHA256 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing sha256: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.SHA512) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := sha512.New()
			d.Write(*content)
			result.SHA512 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing sha512: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Blake2b256) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := blake2b.New256()
			d.Write(*content)
			result.Blake2b256 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing blake2b-256: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Blake2b512) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := blake2b.New512()
			d.Write(*content)
			result.Blake2b512 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing blake2b-512: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Sha3224) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := sha3.New224()
			d.Write(*content)
			result.Sha3224 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing sha3-224: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Sha3256) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := sha3.New256()
			d.Write(*content)
			result.Sha3256 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing sha3-256: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Sha3384) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := sha3.New384()
			d.Write(*content)
			result.Sha3384 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing sha3-384: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	if hasHash(HashNames.Sha3512) {
		wg.Add(1)
		go func() {
			startTime = makeTimestampNano()
			d := sha3.New512()
			d.Write(*content)
			result.Sha3512 = hex.EncodeToString(d.Sum(nil))

			if Trace {
				printTrace(fmt.Sprintf("nanoseconds processing sha3-512: %s: %d", filename, makeTimestampNano()-startTime))
			}
			wg.Done()
		}()
	}

	wg.Wait()
	return result, nil
}

func processReadFile(filename string, content *[]byte) (Result, error) {
	startTime := makeTimestampNano()

	result := Result{}

	if hasHash(HashNames.MD4) {
		startTime = makeTimestampNano()
		d := md4.New()
		d.Write(*content)
		result.MD4 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing md4: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.MD5) {
		startTime = makeTimestampNano()
		d := md5.New()
		d.Write(*content)
		result.MD5 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing md5: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.SHA1) {
		startTime = makeTimestampNano()
		d := sha1.New()
		d.Write(*content)
		result.SHA1 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing sha1: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.SHA256) {
		startTime = makeTimestampNano()
		d := sha256.New()
		d.Write(*content)
		result.SHA256 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing sha256: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.SHA512) {
		startTime = makeTimestampNano()
		d := sha512.New()
		d.Write(*content)
		result.SHA512 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing sha512: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.Blake2b256) {
		startTime = makeTimestampNano()
		d := blake2b.New256()
		d.Write(*content)
		result.Blake2b256 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing blake2b-256: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.Blake2b512) {
		startTime = makeTimestampNano()
		d := blake2b.New512()
		d.Write(*content)
		result.Blake2b512 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing blake2b-512: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.Sha3224) {
		startTime = makeTimestampNano()
		d := sha3.New224()
		d.Write(*content)
		result.Sha3224 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing sha3-224: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.Sha3256) {
		startTime = makeTimestampNano()
		d := sha3.New256()
		d.Write(*content)
		result.Sha3256 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing sha3-256: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.Sha3384) {
		startTime = makeTimestampNano()
		d := sha3.New384()
		d.Write(*content)
		result.Sha3384 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing sha3-384: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	if hasHash(HashNames.Sha3512) {
		startTime = makeTimestampNano()
		d := sha3.New512()
		d.Write(*content)
		result.Sha3512 = hex.EncodeToString(d.Sum(nil))

		if Trace {
			printTrace(fmt.Sprintf("nanoseconds processing sha3-512: %s: %d", filename, makeTimestampNano()-startTime))
		}
	}

	return result, nil
}

// Copied from Go io/ioutil
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	var buf bytes.Buffer
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
