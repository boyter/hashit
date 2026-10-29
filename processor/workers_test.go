// SPDX-License-Identifier: MIT

package processor

import (
	"crypto/md5"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessReadFile(t *testing.T) {
	t.Cleanup(resetState)
	Hash = append(Hash, "all")

	res, _ := processReadFileParallel("filename", &[]byte{})

	if res.MD5 != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("Expected d41d8cd98f00b204e9800998ecf8427e got %s", res.MD5)
	}

	if res.SHA1 != "da39a3ee5e6b4b0d3255bfef95601890afd80709" {
		t.Errorf("Expected da39a3ee5e6b4b0d3255bfef95601890afd80709 got %s", res.SHA1)
	}

	if res.SHA256 != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("Expected e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 got %s", res.SHA256)
	}
}

func TestEncodeIfHashEnabled(t *testing.T) {
	t.Cleanup(resetState)
	hash_md5 := md5.New()
	hash_sha256 := sha256.New()

	// Force only sha256
	Hash = []string{"sha256"}

	res_md5 := encodeIfHashEnabled(hash_md5, "md5")
	res_sha256 := encodeIfHashEnabled(hash_sha256, "sha256")

	if res_md5 != "" {
		t.Errorf("Expected empty string got %s", res_md5)
	}

	if res_sha256 == "" {
		t.Errorf("Expected digest string got %s", res_sha256)
	}
}

// Reset the global state after test.
func resetState() {
	Hash = []string{}
}

//////////////////////////////////////////////////
// Benchmarks Below
//////////////////////////////////////////////////

func BenchmarkProcessReadFile100Bytes(b *testing.B) {
	b.StopTimer()
	Hash = append(Hash, "all")

	content := ""
	for i := 0; i < 100; i++ {
		content += "1"
	}

	data := []byte(content)
	var count int64

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, _ := processReadFileParallel("filenane", &data)
		count += res.Bytes
	}
	b.Log(count)
}

func BenchmarkProcessReadFile500Bytes(b *testing.B) {
	b.StopTimer()
	Hash = append(Hash, "all")

	content := ""
	for i := 0; i < 700; i++ {
		content += "1"
	}

	data := []byte(content)
	var count int64

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, _ := processReadFileParallel("filenane", &data)
		count += res.Bytes
	}
	b.Log(count)
}

func BenchmarkProcessReadFile1000Bytes(b *testing.B) {
	b.StopTimer()
	Hash = append(Hash, "all")

	content := ""
	for i := 0; i < 1000; i++ {
		content += "1"
	}

	data := []byte(content)
	var count int64

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, _ := processReadFileParallel("filenane", &data)
		count += res.Bytes
	}
	b.Log(count)
}

func BenchmarkProcessReadFileSingle100Bytes(b *testing.B) {
	b.StopTimer()
	Hash = append(Hash, "all")

	content := ""
	for i := 0; i < 100; i++ {
		content += "1"
	}

	data := []byte(content)
	var count int64

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, _ := processReadFile("filenane", &data)
		count += res.Bytes
	}
	b.Log(count)
}

func BenchmarkProcessReadFileSingle500Bytes(b *testing.B) {
	b.StopTimer()
	Hash = append(Hash, "all")

	content := ""
	for i := 0; i < 700; i++ {
		content += "1"
	}

	data := []byte(content)
	var count int64

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, _ := processReadFile("filenane", &data)
		count += res.Bytes
	}
	b.Log(count)
}

func BenchmarkProcessReadFileSingle1000Bytes(b *testing.B) {
	b.StopTimer()
	Hash = append(Hash, "all")

	content := ""
	for i := 0; i < 1000; i++ {
		content += "1"
	}

	data := []byte(content)
	var count int64

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, _ := processReadFile("filenane", &data)
		count += res.Bytes
	}
	b.Log(count)
}

func TestAllHashPathsAgree(t *testing.T) {
	t.Cleanup(resetState)

	// files are hashed by processReadFile, processReadFileParallel or processScanner depending
	// on size, so every hash must be produced and agree across all of them
	data := make([]byte, 300_000)
	for i := range data {
		data[i] = byte(i * 7)
	}
	file := filepath.Join(t.TempDir(), "file.bin")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}

	for _, name := range hashPriority {
		Hash = []string{name}

		serial, _ := processReadFile(file, &data)
		parallel, _ := processReadFileParallel(file, &data)
		scanner, err := processScanner(file, len(data), nil)
		if err != nil {
			t.Fatal(err)
		}

		want := serial.hashValue(name)
		if want == "" {
			t.Errorf("%s: processReadFile produced no hash", name)
		}
		if got := parallel.hashValue(name); got != want {
			t.Errorf("%s: processReadFileParallel expected %q got %q", name, want, got)
		}
		if got := scanner.hashValue(name); got != want {
			t.Errorf("%s: processScanner expected %q got %q", name, want, got)
		}
	}
}
