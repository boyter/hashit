// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestMain runs hashit itself when asked to, so the audit tests can run it as a separate process
// without building it first, as an audit exits the process with its result
func TestMain(m *testing.M) {
	if os.Getenv("HASHIT_TEST_MAIN") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// runHashit runs hashit in dir returning its exit code and combined output
func runHashit(t *testing.T, dir string, args ...string) (int, string) {
	t.Helper()
	// the test binary may have been run by a relative path, which the directory changes
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(self, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HASHIT_TEST_MAIN=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err == nil {
		return 0, out.String()
	}
	if e, ok := err.(*exec.ExitError); ok {
		return e.ExitCode(), out.String()
	}
	t.Fatalf("unable to run hashit %v: %v", args, err)
	return 0, ""
}

// treeName is the name of the audited tree, which is also used for a directory inside it, so a
// path from the parent can name a different file through the tree itself
const treeName = "data"

var (
	scenarioDirs  = []string{"", "sub", treeName, "meta", "sub/deep", treeName + "/" + treeName}
	scenarioNames = []string{"a", "b", "README", "x.txt", "hashes.md5", "m.hd", "m.db"}
)

// auditScenario is a tree which has a manifest written for it, may be copied or moved, then is
// changed and audited. The expected result is worked out from what was changed, not from hashit
type auditScenario struct {
	format      string // hashdeep, sum or sqlite
	invokedFrom string // for hashdeep, strip or sub to remove the invoked from line or point it at sub
	absolute    bool   // manifest written with --absolute
	fromParent  bool   // manifest written from the tree's parent with the tree as the path
	location    string // where the manifest is kept: out (beside the tree), root or meta
	relocation  string // none, copy (original kept) or move
	auditFrom   string // root, parent or sub, being the directory the audit is run from
	allowNew    bool
	allowMoved  bool
	files       map[string]string // path relative to the tree to content
	changes     []string
}

func (s *auditScenario) String() string {
	return fmt.Sprintf("format=%s invokedFrom=%s absolute=%v fromParent=%v location=%s relocation=%s auditFrom=%s allowNew=%v allowMoved=%v changes=%v files=%v",
		s.format, s.invokedFrom, s.absolute, s.fromParent, s.location, s.relocation, s.auditFrom, s.allowNew, s.allowMoved, s.changes, s.files)
}

func (s *auditScenario) manifestName() string {
	switch s.format {
	case "hashdeep":
		return "m.hd"
	case "sum":
		return "hashes.md5"
	}
	return "m.db"
}

// manifestRel is the path of the manifest relative to the tree, or empty when kept beside it
func (s *auditScenario) manifestRel() string {
	switch s.location {
	case "root":
		return s.manifestName()
	case "meta":
		return "meta/" + s.manifestName()
	}
	return ""
}

func newAuditScenario(seed uint64) *auditScenario {
	r := rand.New(rand.NewSource(int64(seed)))
	pick := func(xs ...string) string { return xs[r.Intn(len(xs))] }
	s := &auditScenario{
		format:     pick("hashdeep", "sum", "sqlite"),
		absolute:   r.Intn(8) == 0,
		fromParent: r.Intn(3) == 0,
		location:   pick("out", "root", "meta"),
		relocation: pick("none", "none", "copy", "move"),
		auditFrom:  pick("root", "root", "parent", "sub"),
		allowNew:   r.Intn(2) == 0,
		allowMoved: r.Intn(3) != 0,
		files:      map[string]string{},
	}
	if s.format == "sum" && s.absolute {
		s.absolute = false
	}
	if s.format == "hashdeep" && r.Intn(3) == 0 {
		s.invokedFrom = pick("strip", "sub")
	}
	// sub always exists so the audit can be run from it
	s.files["sub/"+pick(scenarioNames...)] = "content 0\n"
	for i, n := 1, 2+r.Intn(8); i <= n; i++ {
		name := pick(scenarioNames...)
		if used := s.paths(); r.Intn(2) == 0 {
			// the same name in another directory, which another base may name
			name = filepath.Base(used[r.Intn(len(used))])
		}
		p := filepath.ToSlash(filepath.Join(pick(scenarioDirs...), name))
		if p == s.manifestRel() || s.isDir(p) || s.hasDirAt(p) {
			continue
		}
		content := fmt.Sprintf("content %d\n", i)
		if r.Intn(8) == 0 {
			content = "content 0\n" // identical files
		}
		s.files[p] = content
	}

	if r.Intn(2) == 0 {
		// one name in several directories, so another base names a different file of that name
		name := filepath.Base(pick(s.paths()...))
		for i, n := 0, 2+r.Intn(2); i < n; i++ {
			p := filepath.ToSlash(filepath.Join(pick(scenarioDirs...), name))
			if _, ok := s.files[p]; ok || p == s.manifestRel() || s.isDir(p) || s.hasDirAt(p) {
				continue
			}
			s.files[p] = fmt.Sprintf("content %s %d\n", name, i)
		}
	}

	paths := s.paths()
	// files sharing a name with another are changed more often, as which record they are is decided
	// between the directories records can be relative to
	var shared []string
	for _, p := range paths {
		for _, q := range paths {
			if p != q && filepath.Base(p) == filepath.Base(q) {
				shared = append(shared, p)
				break
			}
		}
	}
	file := func() string {
		if len(shared) > 0 && r.Intn(4) != 0 {
			return shared[r.Intn(len(shared))]
		}
		return paths[r.Intn(len(paths))]
	}
	for i, n := 0, r.Intn(4); i < n; i++ {
		f, g := file(), file()
		switch r.Intn(9) {
		case 0:
			s.changes = append(s.changes, "modify "+f)
		case 1:
			s.changes = append(s.changes, "delete "+f)
		case 2:
			s.changes = append(s.changes, "rename "+f+" "+f+".renamed")
		case 3:
			s.changes = append(s.changes, "rename "+f+" "+filepath.ToSlash(filepath.Join(pick(scenarioDirs...), filepath.Base(f)+".moved")))
		case 4:
			s.changes = append(s.changes, "swap "+f+" "+g)
		case 5:
			s.changes = append(s.changes, "copy "+f+" "+f+".copy")
		case 6:
			s.changes = append(s.changes, "copy "+f+" "+filepath.ToSlash(filepath.Join(pick(scenarioDirs...), "new")))
		case 7:
			// a backup is kept of a file before it is changed
			s.changes = append(s.changes, "copy "+f+" "+f+".orig", "modify "+f)
		case 8:
			// moved onto the path of another known file
			s.changes = append(s.changes, "rename "+f+" "+g+".old", "rename "+g+" "+f)
		}
	}
	return s
}

func (s *auditScenario) paths() []string {
	var paths []string
	for p := range s.files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// isDir reports if a file is already used as a directory
func (s *auditScenario) isDir(p string) bool {
	for f := range s.files {
		if strings.HasPrefix(f, p+"/") {
			return true
		}
	}
	return false
}

// hasDirAt reports if a directory of a path is already used as a file
func (s *auditScenario) hasDirAt(p string) bool {
	for d := filepath.Dir(p); d != "."; d = filepath.Dir(d) {
		if _, ok := s.files[filepath.ToSlash(d)]; ok {
			return true
		}
	}
	return false
}

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for p, content := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// readTree returns every file under root with its content, skipping those excluded
func readTree(t *testing.T, root string, exclude map[string]bool) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || exclude[path] {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		files[filepath.ToSlash(rel)] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// applyChange makes a change to the tree, returning false when it cannot be made, such as when
// an earlier change removed the file
func applyChange(root string, change string) bool {
	parts := strings.Fields(change)
	full := func(p string) string { return filepath.Join(root, filepath.FromSlash(p)) }
	exists := func(p string) bool { _, err := os.Lstat(full(p)); return err == nil }
	switch parts[0] {
	case "modify":
		if !exists(parts[1]) {
			return false
		}
		f, err := os.OpenFile(full(parts[1]), os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			return false
		}
		_, err = f.WriteString("modified\n")
		return f.Close() == nil && err == nil
	case "delete":
		return os.Remove(full(parts[1])) == nil
	case "rename", "copy":
		if !exists(parts[1]) || exists(parts[2]) {
			return false
		}
		if err := os.MkdirAll(filepath.Dir(full(parts[2])), 0755); err != nil {
			return false
		}
		if parts[0] == "rename" {
			return os.Rename(full(parts[1]), full(parts[2])) == nil
		}
		content, err := os.ReadFile(full(parts[1]))
		return err == nil && os.WriteFile(full(parts[2]), content, 0644) == nil
	case "swap":
		if parts[1] == parts[2] || !exists(parts[1]) || !exists(parts[2]) {
			return false
		}
		tmp := full(parts[1]) + ".swap"
		return os.Rename(full(parts[1]), tmp) == nil &&
			os.Rename(full(parts[2]), full(parts[1])) == nil &&
			os.Rename(tmp, full(parts[2])) == nil
	}
	return false
}

// expectedAudit works out whether an audit should pass from the files known when the manifest was
// written and the files on disk afterwards. Files are matched by path first, a file at the path
// of a known file with other content being modified, then files which are not where they were are
// paired with known files which are gone by content, being moved, otherwise new or missing
func expectedAudit(known map[string]string, disk map[string]string, allowNew bool, allowMoved bool) (bool, string) {
	var modified, missing, newFiles []string
	missingByContent := map[string][]string{}
	for p, content := range known {
		got, ok := disk[p]
		switch {
		case !ok:
			missingByContent[content] = append(missingByContent[content], p)
		case got != content:
			modified = append(modified, p)
		}
	}
	moved := 0
	var unknown []string
	for p := range disk {
		if _, ok := known[p]; !ok {
			unknown = append(unknown, p)
		}
	}
	sort.Strings(unknown)
	for _, p := range unknown {
		if gone := missingByContent[disk[p]]; len(gone) > 0 {
			missingByContent[disk[p]] = gone[1:]
			moved++
			continue
		}
		newFiles = append(newFiles, p)
	}
	for _, gone := range missingByContent {
		missing = append(missing, gone...)
	}
	pass := len(modified) == 0 && len(missing) == 0 && (allowNew || len(newFiles) == 0) && (allowMoved || moved == 0)
	return pass, fmt.Sprintf("modified=%v missing=%v new=%v moved=%d", modified, missing, newFiles, moved)
}

// ambiguous reports if which known file a file on disk is cannot be told from paths and content
// alone, being identical files or files of the same name in different directories, where a file
// may be matched through another directory than the one the audit is matched through
func ambiguous(known map[string]string, disk map[string]string) bool {
	contents := map[string]int{}
	names := map[string]map[string]bool{}
	for _, files := range []map[string]string{known, disk} {
		for p, content := range files {
			contents[content]++
			name := filepath.Base(p)
			if names[name] == nil {
				names[name] = map[string]bool{}
			}
			names[name][p] = true
		}
	}
	for content, n := range contents {
		if n > 2 || n == 2 && !sameFile(known, disk, content) {
			return true
		}
	}
	for _, paths := range names {
		if len(paths) > 1 {
			return true
		}
	}
	return false
}

// sameFile reports if content appears once in each of known and disk, so is not a duplicate
func sameFile(known map[string]string, disk map[string]string, content string) bool {
	count := func(files map[string]string) int {
		n := 0
		for _, c := range files {
			if c == content {
				n++
			}
		}
		return n
	}
	return count(known) == 1 && count(disk) == 1
}

// runAuditScenario writes the manifest, relocates and changes the tree, audits it and returns what
// was wrong with the result, if anything. A pass is always checked, as an audit must never pass when
// a known file was modified or is missing, while a failure is only checked when which known file
// each file is is unambiguous, unless strict
func runAuditScenario(t *testing.T, s *auditScenario, strict bool) string {
	if !s.resolvable() {
		return ""
	}
	w := t.TempDir()
	if real, err := filepath.EvalSymlinks(w); err == nil {
		w = real
	}
	tree := filepath.Join(w, treeName)
	writeTree(t, tree, s.files)

	manifest := filepath.Join(w, "out", s.manifestName())
	if s.location != "out" {
		manifest = filepath.Join(tree, filepath.FromSlash(s.manifestRel()))
	}
	if err := os.MkdirAll(filepath.Dir(manifest), 0755); err != nil {
		t.Fatal(err)
	}
	args := []string{"-f", s.format, "-o", manifest}
	if s.format == "sum" {
		args = append(args, "-c", "md5")
	}
	if s.absolute {
		args = append(args, "--absolute")
	}
	writeDir, path := tree, "."
	if s.fromParent {
		writeDir, path = w, treeName
	}
	if rc, out := runHashit(t, writeDir, append(args, path)...); rc != 0 {
		t.Fatalf("writing manifest failed rc=%d: %s\n%v", rc, out, s)
	}
	if s.invokedFrom != "" {
		editInvokedFrom(t, manifest, s.invokedFrom, filepath.Join(tree, "sub"))
	}

	// the tree keeps its name so manifests written from its parent still name its files
	audited := tree
	switch s.relocation {
	case "copy", "move":
		audited = filepath.Join(w, s.relocation, treeName)
		if err := os.MkdirAll(filepath.Dir(audited), 0755); err != nil {
			t.Fatal(err)
		}
		if s.relocation == "move" {
			if err := os.Rename(tree, audited); err != nil {
				t.Fatal(err)
			}
		} else if err := copyDir(tree, audited); err != nil {
			t.Fatal(err)
		}
	}
	if s.location != "out" {
		manifest = filepath.Join(audited, filepath.FromSlash(s.manifestRel()))
	}

	exclude := map[string]bool{manifest: true}
	for _, suffix := range []string{"-wal", "-shm", "-journal"} {
		exclude[manifest+suffix] = true
	}
	known := readTree(t, audited, exclude)

	var applied []string
	for _, change := range s.changes {
		if applyChange(audited, change) {
			applied = append(applied, change)
		}
	}
	disk := readTree(t, audited, exclude)

	auditDir, auditPath := audited, "."
	switch s.auditFrom {
	case "parent":
		auditDir, auditPath = filepath.Dir(audited), treeName
	case "sub":
		auditDir, auditPath = filepath.Join(audited, "sub"), ".."
	}
	auditArgs := []string{"-v", "-a", manifest, auditPath}
	if s.allowNew {
		auditArgs = append(auditArgs, "--allow-new")
	}
	if s.allowMoved {
		auditArgs = append(auditArgs, "--allow-moved")
	}
	rc, out := runHashit(t, auditDir, auditArgs...)
	want, why := expectedAudit(known, disk, s.allowNew, s.allowMoved)

	problem := ""
	switch {
	case rc == 2:
		problem = "audit error"
	case rc == 0 && !want:
		problem = "audit passed but should fail"
	case rc == 1 && want && (strict || !ambiguous(known, disk)):
		problem = "audit failed but should pass"
	default:
		return ""
	}
	return fmt.Sprintf("%s\n%v\napplied=%v expected %s\n%s", problem, s, applied, why, out)
}

// resolvable reports if a directory hashit matches through names the audited files: the directory
// the manifest was written from, the manifest's directory or the directory the audit is run from.
// Otherwise nothing says where the records are relative to, so what an audit should find is unknown
func (s *auditScenario) resolvable() bool {
	if s.absolute {
		return s.relocation == "none"
	}
	// the directory the records are relative to, root or parent of the audited tree
	target := "root"
	if s.fromParent {
		target = "parent"
	}
	// the directory the manifest was written from only names the audited tree if it was not moved
	if s.format != "sum" && s.invokedFrom == "" && s.relocation == "none" {
		return true
	}
	return s.location == "root" && target == "root" || s.auditFrom == target
}

// editInvokedFrom removes the invoked from line of a hashdeep file, as some hashdeep versions and
// hand edited files have none, or points it at another directory, as when it is stale
func editInvokedFrom(t *testing.T, manifest string, edit string, other string) {
	t.Helper()
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.SplitAfter(string(content), "\n") {
		if strings.HasPrefix(line, "## Invoked from: ") {
			if edit == "strip" {
				continue
			}
			line = "## Invoked from: " + other + "\n"
		}
		lines = append(lines, line)
	}
	if err := os.WriteFile(manifest, []byte(strings.Join(lines, "")), 0644); err != nil {
		t.Fatal(err)
	}
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, 0644)
	})
}

// FuzzAudit audits random trees which were changed after their manifest was written, in every
// format, comparing the result with what was changed. Run with go test -fuzz FuzzAudit
func FuzzAudit(f *testing.F) {
	if runtime.GOOS == "windows" {
		f.Skip("scenarios use / separated paths and moving open directories")
	}
	for _, seed := range auditSeeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		if problem := runAuditScenario(t, newAuditScenario(seed), false); problem != "" {
			t.Error(problem)
		}
	})
}

// TestAuditSweep runs HASHIT_AUDIT_SWEEP random scenarios from HASHIT_AUDIT_SEED, which finds more
// than go test -fuzz as the fuzzer mutates around a few seeds. Audits known to be wrong are found too
func TestAuditSweep(t *testing.T) {
	n, _ := strconv.ParseUint(os.Getenv("HASHIT_AUDIT_SWEEP"), 10, 64)
	if n == 0 {
		t.Skip("set HASHIT_AUDIT_SWEEP to the number of scenarios")
	}
	start, _ := strconv.ParseUint(os.Getenv("HASHIT_AUDIT_SEED"), 10, 64)
	for seed := start; seed < start+n; seed++ {
		if problem := runAuditScenario(t, newAuditScenario(seed), false); problem != "" {
			t.Errorf("seed %d: %s", seed, problem)
		}
	}
}

// auditSeeds are run by go test without -fuzz
var auditSeeds = func() []uint64 {
	var seeds []uint64
	for i := uint64(1); i <= 60; i++ {
		seeds = append(seeds, i)
	}
	return seeds
}()

// TestAuditMatrix audits a tree with no repeated names or content in every format, manifest
// location, relocation and directory the audit is run from, unchanged and after each kind of change
func TestAuditMatrix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("scenarios use / separated paths and moving open directories")
	}
	files := map[string]string{"a": "a\n", "sub/b": "b\n", "meta/c": "c\n", "sub/deep/d": "d\n"}
	changes := [][]string{
		nil,
		{"modify a"},
		{"modify sub/deep/d"},
		{"delete sub/b"},
		{"rename a a.renamed"},
		{"rename sub/b meta/b"},
		{"swap a sub/b"},
		{"copy a sub/a.copy"},
		{"copy sub/b sub/b.orig", "modify sub/b"},
		{"rename sub/b a.old", "rename a sub/b"},
	}
	for _, format := range []string{"hashdeep", "sum", "sqlite"} {
		for _, location := range []string{"out", "root", "meta"} {
			for _, relocation := range []string{"none", "copy", "move"} {
				for _, auditFrom := range []string{"root", "parent", "sub"} {
					for _, fromParent := range []bool{false, true} {
						for i, change := range changes {
							for flags := 0; flags < 4; flags++ {
								s := &auditScenario{format: format, location: location, relocation: relocation,
									auditFrom: auditFrom, fromParent: fromParent, files: files, changes: change,
									allowNew: flags&1 != 0, allowMoved: flags&2 != 0}
								if !s.resolvable() {
									continue
								}
								name := fmt.Sprintf("%s/%s/%s/%s/parent=%v/change%d/flags%d", format, location, relocation, auditFrom, fromParent, i, flags)
								t.Run(name, func(t *testing.T) {
									t.Parallel()
									if problem := runAuditScenario(t, s, true); problem != "" {
										t.Error(problem)
									}
								})
							}
						}
					}
				}
			}
		}
	}
}

// TestAuditUserScenario is a sum file written in the directory it lists, as in
// https://github.com/boyter/hashit/issues/12, with another manifest written beside it later
func TestAuditUserScenario(t *testing.T) {
	files := map[string]string{"a": "a\n", "sub/b": "b\n"}
	for _, allowNew := range []bool{false, true} {
		s := &auditScenario{format: "sum", location: "root", relocation: "none", auditFrom: "root",
			files: files, changes: []string{"copy a hashes2.md5"}, allowNew: allowNew}
		if problem := runAuditScenario(t, s, true); problem != "" {
			t.Error(problem)
		}
	}
}

// auditRegressions are audits which were wrong, found by reviews and the sweep, see TestAuditSweep
var auditRegressions = []struct {
	name   string
	strict bool
	s      *auditScenario
}{
	{
		// the manifest's directory names meta/README as record ./README, so the record which names
		// it through the directory the audit is matched through is never looked up
		name: "file moved onto a known path of a name repeated below the manifest passes in a copy",
		s: &auditScenario{format: "hashdeep", location: "meta", relocation: "copy", auditFrom: "root", allowMoved: true,
			files:   map[string]string{"README": "top\n", "meta/README": "meta\n", "w": "w\n", "sub/x": "x\n"},
			changes: []string{"rename meta/README meta/README.old", "rename w meta/README"}},
	},
	{
		name: "file moved onto a known path of a name repeated below the manifest passes in a copy, sqlite",
		s: &auditScenario{format: "sqlite", location: "meta", relocation: "copy", auditFrom: "root", allowMoved: true,
			files:   map[string]string{"README": "top\n", "meta/README": "meta\n", "w": "w\n", "sub/x": "x\n"},
			changes: []string{"rename meta/README meta/README.old", "rename w meta/README"}},
	},
	{
		name: "file moved onto a known path of a name repeated below the manifest passes in a moved tree",
		s: &auditScenario{format: "hashdeep", location: "meta", relocation: "move", auditFrom: "root", allowMoved: true,
			files:   map[string]string{"README": "top\n", "meta/README": "meta\n", "w": "w\n", "sub/x": "x\n"},
			changes: []string{"rename meta/README meta/README.old", "rename w meta/README"}},
	},
	{
		name:   "unchanged copy of a tree with a name repeated below the manifest fails",
		strict: true,
		s: &auditScenario{format: "hashdeep", location: "meta", relocation: "copy", auditFrom: "root",
			files: map[string]string{"README": "top\n", "meta/README": "meta\n", "w": "w\n", "sub/x": "x\n"}},
	},
	{
		name:   "unchanged copy of a tree with a name repeated below the manifest fails, sqlite",
		strict: true,
		s: &auditScenario{format: "sqlite", location: "meta", relocation: "copy", auditFrom: "root",
			files: map[string]string{"README": "top\n", "meta/README": "meta\n", "w": "w\n", "sub/x": "x\n"}},
	},
	{
		name:   "unchanged tree audited from a subdirectory without invoked from fails",
		strict: true,
		s: &auditScenario{format: "hashdeep", invokedFrom: "strip", location: "root", relocation: "none", auditFrom: "sub",
			files: map[string]string{"b": "BBB\n", "sub/b": "SUBB\n", "sub/c": "C\n"}},
	},
	{
		// with no invoked from line the working directory sub names sub/b as record ./b
		name: "modified file with a backup passes audited from a subdirectory without invoked from",
		s: &auditScenario{format: "hashdeep", invokedFrom: "strip", location: "root", relocation: "none", auditFrom: "sub",
			allowNew: true, allowMoved: true,
			files:   map[string]string{"b": "BBB\n", "sub/b": "SUBB\n", "sub/c": "C\n"},
			changes: []string{"copy sub/b sub/b.orig", "modify sub/b"}},
	},
	{
		// a stale invoked from line names some files through sub, which then outvotes the
		// manifest's directory, as each file votes only for the first directory naming it
		name: "modified file with a backup passes with a stale invoked from",
		s: &auditScenario{format: "hashdeep", invokedFrom: "sub", location: "meta", relocation: "none", auditFrom: "root",
			allowNew: true, allowMoved: true,
			files: map[string]string{"data/README": "content 0\n", "data/b": "content 1\n", "data/data/README": "content README 0\n",
				"data/hashes.md5": "content 9\n", "hashes.md5": "content 0\n", "sub/README": "content 8\n",
				"sub/deep/README": "content README 1\n", "sub/hashes.md5": "content 3\n"},
			changes: []string{"copy sub/hashes.md5 sub/hashes.md5.orig", "modify sub/hashes.md5"}},
	},
	{
		// through the working directory sub, ./x.txt names sub/x.txt, which is identical to x.txt, and
		// ./deep/x.txt names sub/deep/x.txt, so they are not looked up through the sum file's directory
		name: "file moved onto a known path of a repeated name passes for a sum file audited from a subdirectory",
		s: &auditScenario{format: "sum", location: "root", relocation: "none", auditFrom: "sub", allowMoved: true,
			files: map[string]string{"data/README": "content 6\n", "data/m.db": "content 2\n", "data/m.hd": "content 4\n",
				"data/x.txt": "content 1\n", "meta/x.txt": "content 5\n", "sub/deep/x.txt": "content 3\n",
				"sub/x.txt": "content 0\n", "x.txt": "content 0\n"},
			changes: []string{"rename sub/x.txt sub/deep/x.txt.old", "rename sub/deep/x.txt sub/x.txt"}},
	},
	{
		// through the working directory ./data/hashes.md5 is the audit file, so is dropped as its
		// own entry, although through the sum file's directory it is another file
		name: "deleted file of the sum file's name below a directory of the tree's name passes",
		s: &auditScenario{format: "sum", location: "root", relocation: "none", auditFrom: "parent",
			files:   map[string]string{"x": "x\n", treeName + "/hashes.md5": "inner\n"},
			changes: []string{"delete " + treeName + "/hashes.md5"}},
	},
	{
		name:   "unchanged file of the sum file's name below a directory of the tree's name is new",
		strict: true,
		s: &auditScenario{format: "sum", location: "root", relocation: "none", auditFrom: "parent",
			files: map[string]string{"x": "x\n", treeName + "/hashes.md5": "inner\n"}},
	},
	{
		// ./README through the working directory and ./sub/README through the sum file's directory
		// are grouped as the same file before the directory is chosen, and their hashes conflict
		name:   "unchanged sum file at the root audited from a subdirectory with a repeated name fails",
		strict: true,
		s: &auditScenario{format: "sum", location: "root", relocation: "none", auditFrom: "sub",
			files: map[string]string{"README": "top\n", "sub/README": "sub\n", "sub/c": "c\n"}},
	},
	{
		// the stale invoked from sub and the working directory each name sub/m.hd, a tie which went
		// to invoked from, through which sub/m.hd matches ./m.hd and ./sub/m.hd was moved
		name: "file moved onto a known path passes with a stale invoked from naming as many files",
		s: &auditScenario{format: "hashdeep", invokedFrom: "sub", location: "meta", relocation: "none", auditFrom: "root",
			allowNew: true, allowMoved: true,
			files:   map[string]string{"m.hd": "content 2\n", "sub/m.hd": "content 0\n"},
			changes: []string{"rename sub/m.hd m.hd.old", "rename m.hd sub/m.hd"}},
	},
	{
		// the manifest's directory and the working directory each name meta/README, a tie which went
		// to the manifest's directory, through which meta/README matches ./README
		name: "file moved onto a known path below the manifest passes in a copy naming as many files, hashdeep",
		s: &auditScenario{format: "hashdeep", location: "meta", relocation: "copy", auditFrom: "root", allowMoved: true,
			files:   map[string]string{"README": "top\n", "meta/README": "meta\n"},
			changes: []string{"rename meta/README meta/README.old", "rename README meta/README"}},
	},
	{
		// the manifest's directory and the working directory each name meta/README, a tie which went
		// to the manifest's directory, through which meta/README matches ./README
		name: "file moved onto a known path below the manifest passes in a copy naming as many files, sqlite",
		s: &auditScenario{format: "sqlite", location: "meta", relocation: "copy", auditFrom: "root", allowMoved: true,
			files:   map[string]string{"README": "top\n", "meta/README": "meta\n"},
			changes: []string{"rename meta/README meta/README.old", "rename README meta/README"}},
	},
	{
		// the manifest's directory and the working directory each name meta/README, a tie which went
		// to the manifest's directory, through which meta/README matches ./README
		name: "file moved onto a known path below the manifest passes in a copy naming as many files, sum",
		s: &auditScenario{format: "sum", location: "meta", relocation: "copy", auditFrom: "root", allowMoved: true,
			files:   map[string]string{"README": "top\n", "meta/README": "meta\n"},
			changes: []string{"rename meta/README meta/README.old", "rename README meta/README"}},
	},
}

func TestAuditRegressions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("scenarios use / separated paths and moving open directories")
	}
	for _, c := range auditRegressions {
		t.Run(c.name, func(t *testing.T) {
			if problem := runAuditScenario(t, c.s, c.strict); problem != "" {
				t.Error(problem)
			}
		})
	}
}

// knownAuditBugs are audits which are wrong and not yet fixed. Each is skipped while it is still
// wrong, and fails once it is right so it can be moved to auditRegressions
var knownAuditBugs []struct {
	name   string
	strict bool
	s      *auditScenario
}

func TestAuditKnownBugs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("scenarios use / separated paths and moving open directories")
	}
	for _, bug := range knownAuditBugs {
		t.Run(bug.name, func(t *testing.T) {
			if problem := runAuditScenario(t, bug.s, bug.strict); problem != "" {
				t.Skipf("known bug: %s", problem)
			}
			t.Errorf("known bug is fixed, move it to auditRegressions")
		})
	}
}
