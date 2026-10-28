// SPDX-License-Identifier: MIT

package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSqlitePathCandidates(t *testing.T) {
	wd, _ := os.Getwd()

	got := sqlitePathCandidates("./a.txt", "")
	if len(got) != 2 || got[0] != "a.txt" || got[1] != filepath.Join(wd, "a.txt") {
		t.Errorf("expected cleaned and absolute path got %v", got)
	}

	// an absolute path under the invoked from directory should also be tried relative to it
	got = sqlitePathCandidates("/srv/media/sub/a.txt", "/srv/media")
	if len(got) != 2 || got[0] != "/srv/media/sub/a.txt" || got[1] != "sub/a.txt" {
		t.Errorf("expected absolute and relative to invoked from path got %v", got)
	}

	// but never one outside of it
	got = sqlitePathCandidates("/srv/other/a.txt", "/srv/media")
	if len(got) != 1 {
		t.Errorf("expected only the absolute path got %v", got)
	}
}
