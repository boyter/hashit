// SPDX-License-Identifier: MIT

package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAbsFrom(t *testing.T) {
	cases := []struct{ base, p, want string }{
		{"/srv/media", "a.txt", "/srv/media/a.txt"},
		{"/srv/media", "./sub/../a.txt", "/srv/media/a.txt"},
		{"/srv/media", "../other/a.txt", "/srv/other/a.txt"},
		{"/srv/media", "/abs/./a.txt", "/abs/a.txt"},
	}
	for _, c := range cases {
		if got := absFrom(c.base, c.p); got != filepath.FromSlash(c.want) {
			t.Errorf("%q %q: expected %q got %q", c.base, c.p, c.want, got)
		}
	}
}

func TestAuditBases(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	link := filepath.Join(dir, "link")
	if err := os.Mkdir(real, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(link, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(real, "audit.txt"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(real, "audit.txt"), filepath.Join(link, "m")); err != nil {
		t.Skip("symlinks not supported")
	}
	realDir := realPath(real)

	// the invoked from directory first, then where the audit file really is, then where the
	// path supplied for it is, then the working directory, without repeats
	got := auditBases("/inv", filepath.Join(link, "m"), "/cwd")
	want := []string{filepath.FromSlash("/inv"), realDir, link, filepath.FromSlash("/cwd")}
	if len(got) != len(want) {
		t.Fatalf("expected %v got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("expected %v got %v", want, got)
		}
	}

	// without an invoked from directory the working directory is first
	got = auditBases("", filepath.Join(real, "audit.txt"), "/cwd")
	if len(got) != 2 || got[0] != filepath.FromSlash("/cwd") || got[1] != realDir {
		t.Errorf("expected [/cwd %s] got %v", realDir, got)
	}
}

func TestOutputDir(t *testing.T) {
	cases := map[string]string{
		"new.txt":       ".",
		"dl/../new.txt": "dl/..",
		"/new.txt":      "/",
		"/a/b/new.txt":  "/a/b",
		"./new.txt":     ".",
		"a//new.txt":    "a/",
	}
	for p, want := range cases {
		if got := outputDir(filepath.FromSlash(p)); got != filepath.FromSlash(want) {
			t.Errorf("%q: expected %q got %q", p, want, got)
		}
	}
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

func TestAuditFailed(t *testing.T) {
	t.Cleanup(func() { AllowMoved = false })

	cases := []struct {
		allowMoved                         bool
		modified, moved, newFiles, missing int
		want                               bool
	}{
		{false, 0, 0, 0, 0, false},
		{false, 0, 1, 0, 0, true},
		{true, 0, 1, 0, 0, false},
		{true, 1, 1, 0, 0, true},
		{true, 0, 1, 1, 0, true},
		{true, 0, 1, 0, 1, true},
		{true, 0, 0, 0, 0, false},
	}

	for _, c := range cases {
		AllowMoved = c.allowMoved
		if got := auditFailed(c.modified, c.moved, c.newFiles, c.missing); got != c.want {
			t.Errorf("%+v: expected %v got %v", c, c.want, got)
		}
	}
}

func TestSqliteFileURI(t *testing.T) {
	cases := map[string]string{
		"/tmp/x.db":     "file:///tmp/x.db?mode=ro",
		"/tmp/x?y#z.db": "file:///tmp/x%3Fy%23z.db?mode=ro",
		"C:/data/x.db":  "file:///C:/data/x.db?mode=ro",
	}
	for abs, want := range cases {
		if got := sqliteFileURI(abs, "mode=ro"); got != want {
			t.Errorf("%s: expected %s got %s", abs, want, got)
		}
	}
}

func TestEscapesDir(t *testing.T) {
	for rel, want := range map[string]bool{"..": true, "../a": true, "..notes": false, "..cfg/x": false, "a/..b": false} {
		if got := escapesDir(rel); got != want {
			t.Errorf("%q: expected %v got %v", rel, want, got)
		}
	}
}
