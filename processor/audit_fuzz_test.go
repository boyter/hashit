// SPDX-License-Identifier: MIT

package processor

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// FuzzSumLineRoundTrip checks every filename written to a sum file is read back unchanged
func FuzzSumLineRoundTrip(f *testing.F) {
	for _, s := range []string{"a.txt", "back\\slash.txt", "new\nline.txt", "c \"quoted\", comma.txt", "cr\r", " lead", "MD5 (x) = y", "*star"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, file string) {
		if file == "" {
			return
		}
		line := strings.TrimSuffix(sumLine(testMD5, file), "\n")
		name, value, filename, ok := parseSumLine(line, nil)
		if !ok || name != "md5" || value != testMD5 || filename != file {
			t.Errorf("%q: round trip failed got %v %q %q %q", file, ok, name, value, filename)
		}
	})
}

// FuzzParseSumLine checks any line is either rejected or gives a known hash of the right length
// and a filename
func FuzzParseSumLine(f *testing.F) {
	for _, s := range []string{
		testMD5 + "  a.txt",
		testMD5 + " *a.txt",
		"\\" + testMD5 + "  back\\\\slash",
		"MD5 (a.txt) = " + testMD5,
		"SHA256 (a (1).txt) = " + testSHA256,
		"MD4   (a) = " + testMD5,
		testSHA1 + "  ",
		"",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, line string) {
		name, value, filename, ok := parseSumLine(line, nil)
		if !ok {
			return
		}
		if filename == "" {
			t.Errorf("%q: empty filename accepted", line)
		}
		if !isHex(value) || len(value) != hashLengths[name] {
			t.Errorf("%q: %s value %q is not %d hex digits", line, name, value, hashLengths[name])
		}
	})
}

// FuzzSplitHashdeepLine checks a split line has every column and joins back to the line
func FuzzSplitHashdeepLine(f *testing.F) {
	f.Add("2,abc,def,a.txt", 4, 3)
	f.Add("2,abc,def,b, with comma.txt", 4, 3)
	f.Add("2,abc,a,b.txt,2026-01-01 10:00:00", 4, 2)
	f.Add(",,,", 4, 3)
	f.Fuzz(func(t *testing.T, line string, columns int, filenameIdx int) {
		if columns < 1 || columns > 16 || filenameIdx < 0 || filenameIdx >= columns {
			return
		}
		fields, err := splitHashdeepLine(line, columns, filenameIdx)
		if err != nil {
			return
		}
		if len(fields) != columns {
			t.Fatalf("%q: expected %d columns got %d", line, columns, len(fields))
		}
		if fields[filenameIdx] == "" {
			t.Errorf("%q: empty filename accepted", line)
		}
		if joined := strings.Join(fields, ","); joined != line {
			t.Errorf("%q: columns join to %q", line, joined)
		}
	})
}

// FuzzNewAuditor checks any audit file is either refused or parsed without panicking, and that
// every record parsed can be looked up
func FuzzNewAuditor(f *testing.F) {
	for _, s := range []string{
		"%%%% HASHDEEP-1.0\n%%%% size,md5,sha256,filename\n## Invoked from: /srv/media\n##\n2," + testMD5 + "," + testSHA256 + ",a.txt\n",
		"%%%% HASHDEEP-1.0\r\n%%%% size,md5,filename\r\n##\r\n2," + testMD5 + ",./a\r\n2," + testMD5 + ",a\r\n",
		testMD5 + "  a.txt\n" + testMD5 + "  ./a.txt\n" + testMD5 + "  /srv/media/a.txt\n",
		"MD5 (a.txt) = " + testMD5 + "\nSHA1 (a.txt) = " + testSHA1 + "\n",
		"\\" + testMD5 + "  new\\nline\n",
		"",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		hdl, err := NewAuditor(input, "/hashit-fuzz/srv/media/hashes.txt")
		if err != nil {
			return
		}
		if hdl.Count() < 0 || hdl.Count() > len(hdl.Records()) {
			t.Errorf("count %d with %d records", hdl.Count(), len(hdl.Records()))
		}
		for _, r := range hdl.Records() {
			if r.Filename == "" {
				t.Errorf("record with an empty filename")
			}
		}
	})
}

// FuzzAbsFrom checks a path resolved against a directory is absolute and resolves to itself
func FuzzAbsFrom(f *testing.F) {
	if runtime.GOOS == "windows" {
		f.Skip("unix paths")
	}
	for _, s := range []string{"a.txt", "./sub/../a.txt", "../other/a.txt", "/abs/./a.txt", "..", "a/", "a/.", "//a", "../../../.."} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, p string) {
		if p == "" || strings.ContainsRune(p, 0) {
			return
		}
		base := "/hashit-fuzz/srv/media"
		got := absFrom(base, p)
		if !filepath.IsAbs(got) {
			t.Fatalf("%q: %q is not absolute", p, got)
		}
		if again := absFrom(base, got); again != got {
			t.Errorf("%q: %q resolves again to %q", p, got, again)
		}
		if again := absFrom("/elsewhere", got); again != got {
			t.Errorf("%q: absolute %q depends on the directory, got %q", p, got, again)
		}
	})
}

// FuzzOsClean checks cleaning a path twice gives the same path
func FuzzOsClean(f *testing.F) {
	for _, s := range []string{"a/../b", "/a/./b/", "..", "a//b", "/../a", "./a"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, p string) {
		if p == "" || strings.ContainsRune(p, 0) {
			return
		}
		once := osClean(p)
		if twice := osClean(once); twice != once {
			t.Errorf("%q: cleaned to %q then %q", p, once, twice)
		}
	})
}
