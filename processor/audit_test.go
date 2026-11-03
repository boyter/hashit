// SPDX-License-Identifier: MIT

package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAbsFrom(t *testing.T) {
	dir := realPath(t.TempDir())
	if err := os.MkdirAll(filepath.Join(dir, "srv", "media", "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	srv := filepath.Join(dir, "srv")
	cases := []struct{ base, p, want string }{
		{srv + "/media", "a.txt", srv + "/media/a.txt"},
		{srv + "/media", "./sub/../a.txt", srv + "/media/a.txt"},
		{srv + "/media", "../other/a.txt", srv + "/other/a.txt"},
		{srv + "/media", "/abs/./a.txt", "/abs/a.txt"},
		// a directory which no longer exists, such as the one an audit file was created in
		{srv + "/gone", "../media/a.txt", srv + "/media/a.txt"},
	}
	for _, c := range cases {
		if runtime.GOOS == "windows" && c.want[0] == '/' {
			continue // takes the base's drive, see TestAbsFromWindows
		}
		want := filepath.FromSlash(c.want)
		if runtime.GOOS == "windows" {
			want = filepath.Clean(want)
		}
		if got := absFrom(filepath.FromSlash(c.base), filepath.FromSlash(c.p)); got != want {
			t.Errorf("%q %q: expected %q got %q", c.base, c.p, want, got)
		}
	}
}

func TestAbsFromWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows paths")
	}
	cases := []struct{ base, p, want string }{
		{`D:\srv`, `D:a.txt`, `D:\srv\a.txt`},
		{`D:\srv`, `d:a.txt`, `D:\srv\a.txt`},
		{`D:\srv`, `\a.txt`, `D:\a.txt`},
		{`D:\srv`, `E:\a.txt`, `E:\a.txt`},
	}
	for _, c := range cases {
		if got := absFrom(c.base, c.p); got != c.want {
			t.Errorf("%q %q: expected %q got %q", c.base, c.p, c.want, got)
		}
	}
	// another drive is relative to that drive's working directory, so is absolute on it
	if got := absFrom(`D:\srv`, `E:a.txt`); !filepath.IsAbs(got) || filepath.VolumeName(got) != "E:" {
		t.Errorf("expected an absolute path on E: got %q", got)
	}
}

func TestOutputDirWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows paths")
	}
	cases := map[string]string{
		`C:new.txt`:    `C:.`,
		`C:\new.txt`:   `C:\`,
		`C:\a\new.txt`: `C:\a`,
		`C:a\new.txt`:  `C:a`,
	}
	for p, want := range cases {
		if got := outputDir(p); got != want {
			t.Errorf("%q: expected %q got %q", p, want, got)
		}
	}
}

func TestAuditBases(t *testing.T) {
	// resolved, as the temporary directory may be reached through a symlink, such as /var on macOS
	dir := realPath(t.TempDir())
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
		if got := dotPrefix(supplied, filepath.Clean(supplied)); got != filepath.FromSlash(want) {
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
		if got := escapesDir(filepath.FromSlash(rel)); got != want {
			t.Errorf("%q: expected %v got %v", rel, want, got)
		}
	}
}

func TestIsDirectSpelling(t *testing.T) {
	cases := map[string]bool{
		"a":       true,
		"sub/a":   true,
		"./sub/a": true,
		"/abs/a":  runtime.GOOS != "windows", // \abs\a has no drive on windows
		"sub//a":  false,
		"sub/./a": false,
		"x/../a":  false,
		"/abs//a": false,
		"./":      false,
		"":        false,
		"sub/a/":  false,
		"././sub": false,
		"../a":    false,
		"./../a":  false,
		"..":      false,
		".//a":    false,
		"..a":     true,
	}
	for p, want := range cases {
		if got := isDirectSpelling(filepath.FromSlash(p)); got != want {
			t.Errorf("%q: expected %v got %v", p, want, got)
		}
	}
}

func TestOsClean(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows resolves .. by cleaning the path")
	}
	dir := realPath(t.TempDir())
	if err := os.MkdirAll(filepath.Join(dir, "real", "d", "a..b"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "real", "d"), filepath.Join(dir, "link")); err != nil {
		t.Skip("symlinks not supported")
	}
	if err := os.Symlink(dir, filepath.Join(dir, "self")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "real", "file"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "locked"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "real", "d"), filepath.Join(dir, "locked", "lk")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(dir, "locked"), 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, "locked"), 0755) })
	if err := os.Symlink(filepath.Join(dir, "gone"), filepath.Join(dir, "dangling")); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"link/../x":     "real/x",
		"link/../d/a":   "real/d/a",
		"real/../a":     "a",
		"real/./d//a":   "real/d/a",
		"link/../../a":  "a",
		"real/d/../..":  "",
		"..notes/a":     "..notes/a",
		"link/a..b/../": "link",
		// a symlink which cleaning names the same directory as the OS reads is kept as written
		"self/real/../a": "self/a",
		// a directory which no longer exists cannot have been a symlink, so is cleaned as text
		"gone/../a":            "a",
		"gone/x/../../a":       "a",
		"real/gone/../../a":    "a",
		"link/gone/../../a":    "real/a",
		"real/d/gone/../../..": "",
		// .. after a symlink or file which cannot be followed is kept as written
		"dangling/../a":       "dangling/../a",
		"dangling/x/../../a":  "dangling/x/../../a",
		"real/file/../a":      "real/file/../a",
		"real/file/x/../../a": "real/file/x/../../a",
		// a directory which cannot be searched may hold a symlink
		"locked/lk/../a": "locked/lk/../a",
	}
	for p, want := range cases {
		path := filepath.Join(dir, "") + "/" + p
		expected := filepath.Join(dir, want)
		if strings.Contains(want, "..") {
			expected = path
		}
		if got := osClean(path); got != expected {
			t.Errorf("%q: expected %q got %q", p, expected, got)
		}
	}
}

func TestJoinText(t *testing.T) {
	cases := []struct{ dir, rest, want string }{
		{"/a/b", "/c", "/a/b/c"},
		{"/a/b", "c", "/a/b/c"},
		{"/", "/c", "/c"},
		{"/", "c", "/c"},
		{"/a", "", "/a"},
		{"/a", "/x/../y", "/a/x/../y"},
	}
	for _, c := range cases {
		if got := joinText(filepath.FromSlash(c.dir), filepath.FromSlash(c.rest)); got != filepath.FromSlash(c.want) {
			t.Errorf("%q %q: expected %q got %q", c.dir, c.rest, c.want, got)
		}
	}
}

func TestHasCollateClause(t *testing.T) {
	cases := []struct {
		sql  string
		want bool
	}{
		{"create table file_hashes (filepath text primary key, md5 text)", false},
		{"create table file_hashes (filepath text collate nocase primary key)", true},
		{"CREATE TABLE file_hashes (filepath TEXT COLLATE NOCASE PRIMARY KEY)", true},
		{"create table file_hashes (filepath text primary key, x text collate\tnocase)", true},
		{"create table file_hashes (filepath text primary key, collateral text)", false},
		{"create table file_hashes (filepath text primary key, \"collate\" text)", false},
		{"create table file_hashes (filepath text primary key, [collate] text, `collate` text)", false},
		{"create table file_hashes (filepath text primary key, note text default 'see collate list')", false},
		{"create table file_hashes (filepath text primary key, note text default 'it''s collate')", false},
		{"create table file_hashes (filepath text primary key -- collate nocase\n)", false},
		{"create table file_hashes (filepath text primary key /* collate */, x text collate binary)", true},
		{"create table file_hashes (filepath text primary key, \"a\"\"collate\" text) ", false},
	}
	for _, c := range cases {
		if got := hasCollateClause(c.sql); got != c.want {
			t.Errorf("%q: expected %v got %v", c.sql, c.want, got)
		}
	}
}

func TestJudgeTied(t *testing.T) {
	bases := []string{"/b0", "/b1", "/b2"}
	cases := []struct {
		name     string
		chosen   int
		tied     []int
		fails    map[int]bool
		reported int
		calls    int
	}{
		{name: "passing two way tie judges each once", chosen: 0, tied: []int{0, 1}, reported: 0, calls: 2},
		{name: "passing three way tie judges each once", chosen: 1, tied: []int{2, 0, 1}, reported: 1, calls: 3},
		{name: "chosen failing is reported", chosen: 0, tied: []int{0, 1}, fails: map[int]bool{0: true, 1: true}, reported: 0, calls: 3},
		{name: "chosen failing alone is reported", chosen: 0, tied: []int{0, 1}, fails: map[int]bool{0: true}, reported: 0, calls: 2},
		{name: "other failing is reported", chosen: 0, tied: []int{0, 1, 2}, fails: map[int]bool{2: true}, reported: 2, calls: 4},
		{name: "first other failing is reported", chosen: 2, tied: []int{0, 1, 2}, fails: map[int]bool{0: true, 1: true}, reported: 0, calls: 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var calls []int
			var last int
			summary, passed := judgeTied("m", bases, c.chosen, c.tied, func(level int) (string, bool) {
				calls = append(calls, level)
				last = level
				return fmt.Sprintf("judged %d", level), !c.fails[level]
			})
			if len(calls) != c.calls {
				t.Errorf("expected %d judgements got %v", c.calls, calls)
			}
			if last != c.reported || summary != fmt.Sprintf("judged %d", c.reported) {
				t.Errorf("expected %d reported got %d %q", c.reported, last, summary)
			}
			if passed == (len(c.fails) != 0) {
				t.Errorf("expected passed %v got %v", len(c.fails) == 0, passed)
			}
			if warningsMuted || stdout.w != os.Stdout {
				t.Errorf("output not restored")
			}
		})
	}
}
