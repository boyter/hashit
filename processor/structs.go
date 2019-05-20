package processor

// Holds the result after processing the hashes for the file
type Result struct {
	File    string
	MD5     string
	SHA1    string
	SHA256  string
	SHA512  string
	Blake2b string
	Bytes   int64
}
