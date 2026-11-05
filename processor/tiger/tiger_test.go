package tiger

import (
	"encoding/hex"
	"strings"
	"testing"
)

// test vectors from the Tiger reference, which hashdeep and rhash agree with
func TestTiger(t *testing.T) {
	cases := map[string]string{
		"":      "3293ac630c13f0245f92bbb1766e16167a4e58492dde73f3",
		"abc":   "2aab1484e8c158f2bfb8c5ff41b57a525129131c957b5f93",
		"Tiger": "dd00230799f5009fec6debc838bb6a27df2b9d6f110c7937",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-": "f71c8583902afb879edfe610f82c0d4786a3a534504486b5",
		strings.Repeat("a", 1000000):                                       "6db0e2729cbead93d715c6a7d36302e9b3cee0d2bc314b41",
	}
	for input, want := range cases {
		d := New()
		// written in pieces, so input split across blocks is hashed the same
		for i := 0; i < len(input); i += 37 {
			_, _ = d.Write([]byte(input[i:min(i+37, len(input))]))
		}
		if got := hex.EncodeToString(d.Sum(nil)); got != want {
			t.Errorf("tiger of %.20q: expected %s got %s", input, want, got)
		}
	}
}
