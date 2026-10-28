// SPDX-License-Identifier: MIT

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
	Ed2k       string `json:"ed2k,omitempty"`
	Bytes      int64
	MTime      *time.Time `json:",omitzero"`
}

// hashValue returns the value of the hash with the supplied name, or empty if not calculated
func (r Result) hashValue(name string) string {
	switch name {
	case HashNames.CRC32:
		return r.CRC32
	case HashNames.XxHash64:
		return r.XxHash64
	case HashNames.MD4:
		return r.MD4
	case HashNames.MD5:
		return r.MD5
	case HashNames.SHA1:
		return r.SHA1
	case HashNames.SHA256:
		return r.SHA256
	case HashNames.SHA512:
		return r.SHA512
	case HashNames.Blake2b256:
		return r.Blake2b256
	case HashNames.Blake2b512:
		return r.Blake2b512
	case HashNames.Blake3:
		return r.Blake3
	case HashNames.Sha3224:
		return r.Sha3224
	case HashNames.Sha3256:
		return r.Sha3256
	case HashNames.Sha3384:
		return r.Sha3384
	case HashNames.Sha3512:
		return r.Sha3512
	case HashNames.Ed2k:
		return r.Ed2k
	}
	return ""
}
