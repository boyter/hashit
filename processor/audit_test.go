// SPDX-License-Identifier: MIT

package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSqlitePathCandidates(t *testing.T) {
	wd, _ := os.Getwd()

	expect := func(got []string, want ...string) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("expected %v got %v", want, got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("expected %v got %v", want, got)
			}
		}
	}

	// databases may store relative paths with or without ./
	expect(sqlitePathCandidates("./a.txt", ""), "./a.txt", "a.txt", filepath.Join(wd, "a.txt"))
	expect(sqlitePathCandidates("a.txt", ""), "a.txt", "./a.txt", filepath.Join(wd, "a.txt"))

	// an absolute path under the invoked from directory should also be tried relative to it
	expect(sqlitePathCandidates("/srv/media/sub/a.txt", "/srv/media"), "/srv/media/sub/a.txt", "sub/a.txt", "./sub/a.txt")

	// but never one outside of it
	expect(sqlitePathCandidates("/srv/other/a.txt", "/srv/media"), "/srv/other/a.txt")
}

func TestDotPrefix(t *testing.T) {
	cases := map[string]string{
		".":       "./",
		"./":      "./",
		"./d":     "./",
		"./d/":    "./",
		"d":       "",
		"/abs":    "",
		"./../d":  "",
		"../d":    "",
		".hidden": "",
	}
	for supplied, want := range cases {
		if got := dotPrefix(supplied, filepath.Clean(supplied)); got != want {
			t.Errorf("%q: expected %q got %q", supplied, want, got)
		}
	}
}
