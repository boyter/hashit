package processor

import "time"

// Holds the result after processing the hashes for the file
type Result struct {
	File       string
	CRC32      string `json:",omitempty"`
	XxHash64   string `json:",omitempty"`
	MD4        string `json:",omitempty"`
	MD5        string `json:",omitempty"`
	SHA1       string `json:",omitempty"`
	SHA256     string `json:",omitempty"`
	SHA512     string `json:",omitempty"`
	Blake2b256 string `json:",omitempty"`
	Blake2b512 string `json:",omitempty"`
	Blake3     string `json:",omitempty"`
	Sha3224    string `json:",omitempty"`
	Sha3256    string `json:",omitempty"`
	Sha3384    string `json:",omitempty"`
	Sha3512    string `json:",omitempty"`
	Bytes      int64
	MTime      *time.Time `json:",omitzero"`
}
