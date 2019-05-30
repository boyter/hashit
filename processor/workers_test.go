package processor

import (
	"testing"
)

func TestProcessReadFile(t *testing.T) {
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

///

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
