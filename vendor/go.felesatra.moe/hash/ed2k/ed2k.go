// Copyright (C) 2021  Allen Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ed2k implements the eD2k hash.
//
// See https://en.wikipedia.org/wiki/Ed2k_URI_scheme.
package ed2k

import (
	"hash"

	"golang.org/x/crypto/md4"
)

// The blocksize of eD2k in bytes.
const BlockSize = md4.BlockSize

// The size of an eD2k checksum in bytes.
const Size = md4.Size

// The chunk size of eD2k in bytes.
const ChunkSize = 9728000

// A Hash is an implementation of hash.Hash for eD2k.
type Hash struct {
	// Total bytes written so far.
	written  int64
	hashlist []byte
	subhash  hash.Hash
}

// New returns a new Hash computing the eD2k checksum.
func New() *Hash {
	return &Hash{
		subhash: md4.New(),
	}
}

// Write adds more data to the running hash.
// It never returns an error.
func (h *Hash) Write(p []byte) (int, error) {
	total := len(p)
	for len(p) > 0 {
		p2 := h.limitNextChunk(p)
		n, _ := h.subhash.Write(p2)
		h.written += int64(n)
		p = p[n:]
		if h.written%ChunkSize == 0 {
			h.hashlist = h.subhash.Sum(h.hashlist)
			h.subhash.Reset()
		}
	}
	return total, nil
}

// limitNextChunk limits the input slice to at most the next chunk to write.
func (h *Hash) limitNextChunk(b []byte) []byte {
	remainder := ChunkSize - int(h.written%ChunkSize)
	if len(b) > remainder {
		return b[:remainder]
	}
	return b
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (h *Hash) Sum(b []byte) []byte {
	if h.written == 0 {
		return h.subhash.Sum(b)
	}

	// If there's an uncommitted chunk, hash it and add to the hashlist
	if h.written%ChunkSize != 0 {
		h.hashlist = h.subhash.Sum(h.hashlist)
	}

	// If the total written data is less than or equal to one chunk,
	// the hashlist already contains the final hash (or the hash of the single chunk).
	if h.written <= ChunkSize {
		return append(b, h.hashlist...)
	}

	// For multiple chunks, hash the hashlist itself.
	h2 := md4.New()
	h2.Write(h.hashlist)
	return h2.Sum(b)
}

// Reset resets the Hash to its initial state.
func (h *Hash) Reset() {
	h.written = 0
	h.hashlist = h.hashlist[:0]
}

// Size returns the number of bytes Sum will return.
func (h *Hash) Size() int {
	return Size
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (h *Hash) BlockSize() int {
	return BlockSize
}
