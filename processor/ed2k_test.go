// SPDX-License-Identifier: MIT

package processor

import (
	"encoding/hex"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"go.felesatra.moe/hash/ed2k"
	"golang.org/x/crypto/md4"
)

// ed2kReference is a direct implementation of https://wiki.anidb.net/Ed2k-hash for comparison.
// Data up to and including one chunk is the md4 of the data, otherwise it is the md4 of the
// md4 of each chunk. This is the "correct" variant which does not append an empty chunk when
// the size is an exact multiple of the chunk size
func ed2kReference(data []byte) string {
	if len(data) <= ed2k.ChunkSize {
		h := md4.New()
		h.Write(data)
		return hex.EncodeToString(h.Sum(nil))
	}

	var hashlist []byte
	for len(data) > 0 {
		n := min(len(data), ed2k.ChunkSize)
		h := md4.New()
		h.Write(data[:n])
		hashlist = h.Sum(hashlist)
		data = data[n:]
	}
	h := md4.New()
	h.Write(hashlist)
	return hex.EncodeToString(h.Sum(nil))
}

func TestEd2kMatchesReference(t *testing.T) {
	t.Cleanup(resetState)
	Hash = []string{HashNames.Ed2k}

	// sizes either side of the chunk boundaries, as files larger than a chunk which were not an
	// exact multiple of it were hashed incorrectly, see https://github.com/boyter/hashit/issues/25
	sizes := []int{
		0, 1, 4096,
		ed2k.ChunkSize - 1, ed2k.ChunkSize, ed2k.ChunkSize + 1,
		10_000_000,
		2*ed2k.ChunkSize - 1, 2 * ed2k.ChunkSize, 2*ed2k.ChunkSize + 5,
	}

	rng := rand.New(rand.NewSource(25))
	dir := t.TempDir()

	for _, size := range sizes {
		data := make([]byte, size)
		rng.Read(data)
		want := ed2kReference(data)

		h := ed2k.New()
		_, _ = h.Write(data)
		if got := hex.EncodeToString(h.Sum(nil)); got != want {
			t.Errorf("size %d: library expected %s got %s", size, want, got)
		}

		// Sum must not change the hash state, so calling it twice gives the same result
		if got := hex.EncodeToString(h.Sum(nil)); got != want {
			t.Errorf("size %d: second Sum expected %s got %s", size, want, got)
		}

		res, _ := processReadFile("file", &data)
		if res.Ed2k != want {
			t.Errorf("size %d: processReadFile expected %s got %s", size, want, res.Ed2k)
		}

		file := filepath.Join(dir, "file.bin")
		if err := os.WriteFile(file, data, 0600); err != nil {
			t.Fatal(err)
		}
		res, err := processScanner(file, size, nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.Ed2k != want {
			t.Errorf("size %d: processScanner expected %s got %s", size, want, res.Ed2k)
		}
	}
}

func TestEd2kKnownValues(t *testing.T) {
	// values confirmed against rhash. Sizes which are an exact multiple of the chunk size are not
	// included as rhash implements the variant which appends an empty chunk for those
	cases := map[int]string{
		0:          "31d6cfe0d16ae931b73c59d7e0c089c0",
		10_000_000: "fae3c97b2cf23ff30e100af8c624f832",
	}
	for size, want := range cases {
		h := ed2k.New()
		_, _ = h.Write(make([]byte, size))
		if got := hex.EncodeToString(h.Sum(nil)); got != want {
			t.Errorf("size %d: expected %s got %s", size, want, got)
		}
	}
}
