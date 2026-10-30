// SPDX-License-Identifier: MIT

package processor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// AuditRecord represents a single file entry from an audit file, which can be
// either hashdeep or sum (md5sum, sha256sum etc...) format
type AuditRecord struct {
	Filename string
	Size     int64             // -1 when the audit file does not record size
	Hashes   map[string]string // hashit hash name to lowercase hex value
	Order    int               // position in the audit file, used to report in file order
	Matched  bool
	Conflict bool // the audit file lists different values of a hash for the file, so it cannot match
}

// Auditor parses and holds an audit file which can then be used for audit purposes
// by providing methods to look up records by either path or hash
type Auditor struct {
	records     []*AuditRecord
	pathLookup  map[string]*AuditRecord   // every path key a record can be found by
	hashLookup  map[string][]*AuditRecord // "hashname:value" to records, used to find moved files
	hashNames   []string                  // hashes present in the audit file which we are able to calculate
	bases       []string                  // absolute directories relative paths may be relative to, in priority order
	primaryBase string                    // directory relative paths were created relative to, empty for the working directory
	cwd         string                    // working directory, looked up once as it is needed for every path
	duplicates  int                       // paths listed more than once, reported once parsing is done
	sumFile     bool                      // a sum file, where every line for a path is checked
}

type FileStatus int

const (
	FileMatched FileStatus = iota
	FileModified
	FileNew
	FileMoved // both of the below come from parsing the audit
	FileMissing
)

// hashdeepColumns maps hashdeep column names to hashit hash names. Hashdeep also supports
// tiger and whirlpool which we cannot calculate, so those columns are ignored
var hashdeepColumns = map[string]string{
	"md5":    HashNames.MD5,
	"sha1":   HashNames.SHA1,
	"sha256": HashNames.SHA256,
}

// bsdTagNames maps the algorithm names used by the BSD style output of md5sum --tag
// shasum, b2sum and others to hashit hash names
var bsdTagNames = map[string]string{
	"CRC32":       HashNames.CRC32,
	"XXH64":       HashNames.XxHash64,
	"MD4":         HashNames.MD4,
	"MD5":         HashNames.MD5,
	"SHA1":        HashNames.SHA1,
	"SHA224":      HashNames.SHA224,
	"SHA256":      HashNames.SHA256,
	"SHA384":      HashNames.SHA384,
	"SHA512":      HashNames.SHA512,
	"BLAKE2b-256": HashNames.Blake2b256,
	"BLAKE2b":     HashNames.Blake2b512,
	"BLAKE2b-512": HashNames.Blake2b512,
	"BLAKE3":      HashNames.Blake3,
	"ED2K":        HashNames.Ed2k,
	"SHA3-224":    HashNames.Sha3224,
	"SHA3-256":    HashNames.Sha3256,
	"SHA3-384":    HashNames.Sha3384,
	"SHA3-512":    HashNames.Sha3512,
}

// sumExtensions maps checksum file extensions to the hashes they usually contain, which lets us
// tell apart hashes of the same length such as md4 and md5. Where tools can write more than one
// length, such as b2sum -l 256, the hash matching the length of the value is used
var sumExtensions = map[string][]string{
	".crc32":    {HashNames.CRC32},
	".md4":      {HashNames.MD4},
	".md5":      {HashNames.MD5},
	".sha1":     {HashNames.SHA1},
	".sha224":   {HashNames.SHA224},
	".sha256":   {HashNames.SHA256},
	".sha384":   {HashNames.SHA384},
	".sha512":   {HashNames.SHA512},
	".blake2b":  {HashNames.Blake2b512, HashNames.Blake2b256},
	".b2":       {HashNames.Blake2b512, HashNames.Blake2b256},
	".blake3":   {HashNames.Blake3},
	".sha3-224": {HashNames.Sha3224},
	".sha3-256": {HashNames.Sha3256},
	".sha3-384": {HashNames.Sha3384},
	".sha3-512": {HashNames.Sha3512},
	".b3":       {HashNames.Blake3},
	".ed2k":     {HashNames.Ed2k},
}

// sumLengths is the hash assumed for a given hex length when the file extension does not tell us
var sumLengths = map[int]string{
	8:   HashNames.CRC32,
	16:  HashNames.XxHash64,
	32:  HashNames.MD5,
	40:  HashNames.SHA1,
	56:  HashNames.SHA224,
	64:  HashNames.SHA256,
	96:  HashNames.SHA384,
	128: HashNames.SHA512,
}

// hashLengths is the hex length of every hash we can calculate
var hashLengths = map[string]int{
	HashNames.CRC32:      8,
	HashNames.XxHash64:   16,
	HashNames.MD4:        32,
	HashNames.MD5:        32,
	HashNames.SHA1:       40,
	HashNames.SHA224:     56,
	HashNames.SHA256:     64,
	HashNames.SHA384:     96,
	HashNames.SHA512:     128,
	HashNames.Blake2b256: 64,
	HashNames.Blake2b512: 128,
	HashNames.Blake3:     64,
	HashNames.Sha3224:    56,
	HashNames.Sha3256:    64,
	HashNames.Sha3384:    96,
	HashNames.Sha3512:    128,
	HashNames.Ed2k:       32,
}

// hashPriority is the order hashes are compared and indexed in, strongest first
var hashPriority = []string{
	HashNames.SHA512, HashNames.Sha3512, HashNames.Blake2b512, HashNames.SHA384, HashNames.Sha3384,
	HashNames.SHA256, HashNames.Sha3256, HashNames.Blake2b256, HashNames.Blake3,
	HashNames.SHA224, HashNames.Sha3224, HashNames.SHA1, HashNames.MD5, HashNames.MD4, HashNames.Ed2k,
	HashNames.XxHash64, HashNames.CRC32,
}

// NewAuditor parses the supplied audit file content. The filename is used to determine
// where relative paths in a sum file are relative to and what hash a sum file contains
func NewAuditor(input string, auditFilename string) (*Auditor, error) {
	hdl := Auditor{
		pathLookup: map[string]*AuditRecord{},
		hashLookup: map[string][]*AuditRecord{},
	}
	hdl.cwd = workingDir()

	var err error
	if isHashdeepFormat(input) {
		err = hdl.parseHashdeepFile(input)
	} else {
		hdl.sumFile = true
		err = hdl.parseSumFile(input, auditFilename)
	}

	// Relative paths are matched against the working directory, and also against the directory
	// containing the audit file, as tools such as RapidCRC do for sum files. For hashdeep files
	// this lets a manifest stored at the root of a set of files be audited after they have been
	// copied elsewhere, where the invoked from directory no longer applies
	hdl.bases = auditBases(hdl.primaryBase, auditFilename, hdl.cwd)
	hdl.removeDuplicates()
	if hdl.duplicates != 0 {
		printWarn(fmt.Sprintf("audit file lists the same path more than once, only the first is used for each, duplicates: %d", hdl.duplicates))
	}
	if conflicts := hdl.conflicts(); conflicts != 0 {
		printWarn(fmt.Sprintf("audit file lists different hashes for the same path, which are reported as modified, paths: %d", conflicts))
	}
	if err != nil {
		return nil, err
	}

	// an audit file of an empty directory has no records, and passes when audited against one
	hdl.buildLookups()
	if len(hdl.records) != 0 && len(hdl.hashNames) == 0 {
		return nil, errors.New("audit file contains no hashes that hashit supports")
	}

	return &hdl, nil
}

func isHashdeepFormat(input string) bool {
	for _, line := range strings.SplitN(input, "\n", 3) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		return strings.HasPrefix(line, "%%%% HASHDEEP")
	}
	return false
}

// HashNames returns the hashes that need to be calculated in order to audit
func (hdl *Auditor) HashNames() []string {
	return hdl.hashNames
}

func (hdl *Auditor) Count() int {
	return len(hdl.records)
}

// Records returns every record in the order it appeared in the audit file
func (hdl *Auditor) Records() []*AuditRecord {
	return hdl.records
}

// Lookup finds the record for a file on disk using its absolute path, returning nil if not found
func (hdl *Auditor) Lookup(file string) *AuditRecord {
	return hdl.pathLookup[absFrom(hdl.cwd, file)]
}

// LookupPrimary finds the record for a file on disk, and reports whether the file matched the
// record's own path rather than through another directory such as the audit file's
func (hdl *Auditor) LookupPrimary(file string) (*AuditRecord, bool) {
	r := hdl.Lookup(file)
	if r == nil {
		return nil, false
	}
	return r, hdl.primaryKey(r) == absFrom(hdl.cwd, file)
}

// primaryKey is the absolute path a record refers to using the directory paths were created
// relative to, being the invoked from directory for hashdeep files or the working directory
func (hdl *Auditor) primaryKey(r *AuditRecord) string {
	return absFrom(hdl.bases[0], r.Filename)
}

// escapesDir reports if a relative path leads outside the directory it is relative to
func escapesDir(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Find checks a file on disk against the audit records by path and marks the record as seen
func (hdl *Auditor) Find(res Result) (*AuditRecord, FileStatus) {
	r := hdl.Lookup(res.File)
	// a record already claimed by another file on disk, such as when two paths resolve to the
	// same record, is not counted twice, and the second file is treated as new
	if r == nil || r.Matched {
		return nil, FileNew
	}

	r.Matched = true
	if recordMatches(r, res) {
		return r, FileMatched
	}
	return r, FileModified
}

// FindMoved looks for a record which has not been matched by path but whose size and hashes
// match the supplied result, marking it as matched and returning it if found
func (hdl *Auditor) FindMoved(res Result) *AuditRecord {
	for _, name := range hdl.hashNames {
		v := res.hashValue(name)
		if v == "" {
			continue
		}
		for _, r := range hdl.hashLookup[name+":"+v] {
			if !r.Matched && recordMatches(r, res) {
				r.Matched = true
				return r
			}
		}
	}
	return nil
}

func (hdl *Auditor) GetUnmatched() []*AuditRecord {
	var unmatched []*AuditRecord
	for _, r := range hdl.records {
		if !r.Matched {
			unmatched = append(unmatched, r)
		}
	}
	return unmatched
}

// recordMatches compares every hash the record has, as well as size if known. A record must
// share at least one hash with the result to be considered a match
func recordMatches(r *AuditRecord, res Result) bool {
	if r.Conflict {
		return false
	}
	if r.Size >= 0 && r.Size != res.Bytes {
		return false
	}

	compared := false
	for name, v := range r.Hashes {
		rv := res.hashValue(name)
		if rv == "" {
			continue
		}
		if rv != v {
			return false
		}
		compared = true
	}
	return compared
}

// absFrom returns the absolute path p refers to, a relative path being resolved against the
// absolute directory base. Paths are only ever compared absolute, as a bare relative path such as
// x only means the same file when run from the same directory. This is done without asking the
// operating system for the working directory, as it is needed for every record and file
func absFrom(base string, p string) string {
	if filepath.IsAbs(p) {
		return osClean(p)
	}
	// a Windows path such as D:x is relative to the working directory of that drive
	if vol := filepath.VolumeName(p); vol != "" {
		if strings.EqualFold(vol, filepath.VolumeName(base)) {
			return filepath.Join(base, p[len(vol):])
		}
		if abs, err := filepath.Abs(p); err == nil {
			return abs
		}
		return filepath.Clean(p)
	}
	// a Windows path such as \x is relative to the root of the current drive
	if filepath.VolumeName(p) == "" && len(p) != 0 && os.IsPathSeparator(p[0]) {
		return filepath.Join(filepath.VolumeName(base), p)
	}
	if base == "" {
		return osClean(p)
	}
	return osClean(base + string(filepath.Separator) + p)
}

// dotDotResolved caches the directory the operating system resolves each path ending in .. to
var dotDotResolved sync.Map

// osClean cleans a path, except that a path containing .. has everything up to its last .. resolved
// by the operating system. After a symlink .. leads to the parent of the directory the link points
// to, not the directory containing the link, so cleaning a/link/.. to a/ would name a different
// directory to the one the operating system reads. A path whose .. the operating system cannot
// resolve is returned as written, so it never names an existing file. Windows resolves .. by
// cleaning the path too
func osClean(path string) string {
	clean := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return clean
	}
	end := lastDotDot(path)
	if end == -1 {
		return clean
	}

	head := path[:end]
	resolved, ok := dotDotResolved.Load(head)
	if !ok {
		resolved, _ = dotDotResolved.LoadOrStore(head, resolveDotDot(head))
	}
	if resolved.(string) == "" {
		// the OS cannot open it, so it is kept as written rather than cleaned into another path
		return path
	}
	return filepath.Join(resolved.(string), path[end:])
}

// resolveDotDot returns the directory a path ending in .. names. It is cleaned as text when that
// names the same directory the OS opens, so paths through a symlinked directory such as the
// working directory are kept as written, and otherwise is where the OS resolves it to. An empty
// string is returned when the OS cannot open it
func resolveDotDot(head string) string {
	fi, err := os.Stat(head)
	if err != nil {
		return ""
	}
	clean := filepath.Clean(head)
	if cfi, err := os.Stat(clean); err == nil && os.SameFile(fi, cfi) {
		return clean
	}
	if real, err := filepath.EvalSymlinks(head); err == nil {
		return physicalAbs(real)
	}
	return ""
}

// lastDotDot returns the index of the end of the last .. component in the path, or -1
func lastDotDot(path string) int {
	if !strings.Contains(path, "..") {
		return -1
	}
	end := -1
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || os.IsPathSeparator(path[i]) {
			if path[start:i] == ".." {
				end = i
			}
			start = i + 1
		}
	}
	return end
}

// workingDir returns the working directory, or an empty string when it cannot be found such as
// when it has been deleted, in which case relative paths are compared as written
func workingDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

// auditBases returns the absolute directories relative paths in an audit file may be relative to,
// in priority order: the directory they were created relative to, then the directory containing
// the audit file, both where it really is and where the path supplied for it is when that is a
// symlink, then the working directory
func auditBases(primary string, auditFile string, cwd string) []string {
	var bases []string
	add := func(dir string) {
		dir = absFrom(cwd, dir)
		if !contains(bases, dir) {
			bases = append(bases, dir)
		}
	}
	if primary != "" {
		add(primary)
	} else {
		add(cwd)
	}
	if real := realPath(auditFile); real != "" {
		add(filepath.Dir(real))
	}
	add(filepath.Dir(auditFile))
	add(cwd)
	return bases
}

// Exclude removes any record which refers to the supplied path, which is used to remove the
// audit file itself as it can never match its own hash
func (hdl *Auditor) Exclude(path string) {
	target := auditFileTarget{abs: absFrom(hdl.cwd, path), real: realPath(path)}

	var records []*AuditRecord
	for _, r := range hdl.records {
		if !hdl.isOwnEntry(r, target) {
			records = append(records, r)
		}
	}
	if len(records) == len(hdl.records) {
		return
	}

	hdl.records = nil
	hdl.pathLookup = map[string]*AuditRecord{}
	hdl.hashLookup = map[string][]*AuditRecord{}
	hdl.hashNames = nil
	for _, r := range records {
		hdl.addRecord(r)
	}
	hdl.buildLookups()
}

// recordKeys returns every absolute path a record may refer to, in the priority order of bases
func (hdl *Auditor) recordKeys(r *AuditRecord) []string {
	var keys []string
	for _, base := range hdl.bases {
		if k := absFrom(base, r.Filename); !contains(keys, k) {
			keys = append(keys, k)
		}
	}
	return keys
}

// isOwnEntry reports if the record is the audit file listing itself, which happens when it is
// written by a shell redirect into the directory being hashed. The record must resolve to the
// audit file using the directory paths are relative to, being the invoked from directory for
// hashdeep files or the working directory otherwise. A record which only resolves to it some
// other way, such as through the audit file's own directory after a tree has been copied, must
// also describe an empty file, as the audit file was when it was hashed. Otherwise it may be a
// different file which happens to share the audit file's name, and must still be audited
func (hdl *Auditor) isOwnEntry(r *AuditRecord, target auditFileTarget) bool {
	if target.matches([]string{hdl.primaryKey(r)}) {
		return true
	}
	return target.matches(hdl.recordKeys(r)) && isEmptyRecord(r)
}

// auditFileTarget is the audit file being excluded, by absolute and real path, as it may be
// supplied through a symlinked path which differs from the path its records were written with
type auditFileTarget struct {
	abs  string
	real string
}

// matches reports if any of the keys is the audit file. Keys are only resolved when they share
// the name of the audit file or of the path supplied for it, so large audit files do not resolve
// every record
func (t auditFileTarget) matches(keys []string) bool {
	for _, k := range keys {
		if k == t.abs || (t.real != "" && k == t.real) {
			return true
		}
		name := filepath.Base(k)
		if t.real != "" && (name == filepath.Base(t.abs) || name == filepath.Base(t.real)) && realPath(k) == t.real {
			return true
		}
	}
	return false
}

// removeDuplicates drops records for a path already listed, keeping the first, as otherwise the
// duplicate is never matched and is reported as missing although the file exists
func (hdl *Auditor) removeDuplicates() {
	keyOf := hdl.primaryKey
	if hdl.sumFile && len(hdl.bases) > 1 {
		keyOf = hdl.sumDuplicateKeys()
	}

	seen := map[string]*AuditRecord{}
	var records []*AuditRecord
	for _, r := range hdl.records {
		// compare the resolved path, so the same file listed by relative and absolute path is a
		// duplicate too
		key := keyOf(r)
		if first, ok := seen[key]; ok {
			// md5sum -c checks every line, so the hashes of a sum file are merged
			if hdl.sumFile {
				for name, value := range r.Hashes {
					mergeSumHash(first, name, value)
				}
			}
			continue
		}
		seen[key] = r
		records = append(records, r)
	}
	if dupes := len(hdl.records) - len(records); dupes != 0 {
		if !hdl.sumFile {
			hdl.duplicates += dupes
		}
		hdl.records = nil
		for _, r := range records {
			hdl.addRecord(r)
		}
	}
}

// sumDuplicateKeys returns the path each record of a sum file is compared to. A sum file has no
// directory its paths were created relative to, so two lines can name the same file through the
// sum file's directory but not through the working directory. Lines naming the same path through
// any directory are compared by the first of their paths which exists, as that is the file they
// are checked against, and the rest by their path through the working directory
func (hdl *Auditor) sumDuplicateKeys() func(*AuditRecord) string {
	// the records naming each path through each directory, so most records need no lookup on disk
	counts := make([]map[string]int, len(hdl.bases))
	for i, base := range hdl.bases {
		counts[i] = make(map[string]int, len(hdl.records))
		for _, r := range hdl.records {
			counts[i][absFrom(base, r.Filename)]++
		}
	}
	return func(r *AuditRecord) string {
		shared := false
		for i, base := range hdl.bases {
			if counts[i][absFrom(base, r.Filename)] > 1 {
				shared = true
				break
			}
		}
		if shared {
			for _, k := range hdl.recordKeys(r) {
				if _, err := os.Lstat(k); err == nil {
					return k
				}
			}
		}
		return hdl.primaryKey(r)
	}
}

// isEmptyRecord reports if the record describes an empty file, by size or when the size is not
// recorded, as in sum files, by every hash being that of empty content
func isEmptyRecord(r *AuditRecord) bool {
	if r.Size >= 0 {
		return r.Size == 0
	}

	// calculate the hash of empty content for every hash, restoring the requested hashes after
	requested := Hash
	Hash = []string{"all"}
	empty, _ := processReadFile("", &[]byte{})
	Hash = requested

	for name, v := range r.Hashes {
		if empty.hashValue(name) != v {
			return false
		}
	}
	return len(r.Hashes) != 0
}

func (hdl *Auditor) addRecord(r *AuditRecord) {
	r.Order = len(hdl.records)
	hdl.records = append(hdl.records, r)
}

func (hdl *Auditor) buildLookups() {
	// Every record's resolution against the directory paths were created relative to takes priority
	// over any other resolution, then against the next directory, so a record resolved through the
	// audit file's directory never takes the path of a different record. For each directory the
	// first record wins
	for _, base := range hdl.bases {
		for _, r := range hdl.records {
			if k := absFrom(base, r.Filename); hdl.pathLookup[k] == nil {
				hdl.pathLookup[k] = r
			}
		}
	}

	seen := map[string]bool{}
	for _, r := range hdl.records {
		for name, v := range r.Hashes {
			seen[name] = true
			hdl.hashLookup[name+":"+v] = append(hdl.hashLookup[name+":"+v], r)
		}
	}

	for _, name := range hashPriority {
		if seen[name] {
			hdl.hashNames = append(hdl.hashNames, name)
		}
	}
}

// parseHashdeepFile accepts a hashdeep format in and builds the internal
// audit processor on it
func (hdl *Auditor) parseHashdeepFile(input string) error {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	scanner.Split(scanLinesKeepCR)

	var header []string
	filenameIdx := -1
	lineNo := 0

	// a filename can end with \r, so one is only removed when the file has windows line endings,
	// which the HASHDEEP line starting each section shows
	crlf := false

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if strings.HasPrefix(line, "%%%% HASHDEEP") {
			crlf = strings.HasSuffix(line, "\r")
		}
		// filenames can start or end with spaces so only remove line endings
		if crlf {
			line = strings.TrimSuffix(line, "\r")
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle header lines starting with %, which can appear more than once
		// when hashdeep files have been concatenated together
		if strings.HasPrefix(line, "%%%%") {
			h := strings.TrimSpace(strings.TrimPrefix(line, "%%%%"))
			if strings.HasPrefix(h, "size") {
				header = strings.Split(h, ",")
				filenameIdx = -1
				for i, f := range header {
					header[i] = strings.TrimSpace(f)
					if header[i] == "filename" {
						filenameIdx = i
					}
				}
				if filenameIdx == -1 {
					return fmt.Errorf("hashdeep header on line %d has no filename column", lineNo)
				}
			}
			continue
		}

		// Skip comment lines starting with #, but record where hashdeep was invoked from
		// as relative paths in the file are relative to it
		if strings.HasPrefix(line, "#") {
			if dir, ok := strings.CutPrefix(line, "## Invoked from: "); ok && hdl.primaryBase == "" {
				// a directory name can start or end with spaces
				hdl.primaryBase = dir
			}
			continue
		}

		if header == nil {
			return fmt.Errorf("hashdeep line %d found before header", lineNo)
		}

		fields, err := splitHashdeepLine(line, len(header), filenameIdx)
		if err != nil {
			return fmt.Errorf("hashdeep line %d: %w", lineNo, err)
		}

		r := &AuditRecord{Size: -1, Hashes: map[string]string{}}
		for i, column := range header {
			switch column {
			case "size":
				size, err := strconv.ParseInt(fields[i], 10, 64)
				if err != nil {
					return fmt.Errorf("hashdeep line %d: invalid size %q", lineNo, fields[i])
				}
				r.Size = size
			case "filename":
				r.Filename = fields[i]
			default:
				if name, ok := hashdeepColumns[column]; ok {
					r.Hashes[name] = strings.ToLower(fields[i])
				}
			}
		}

		hdl.addRecord(r)
	}

	return scanner.Err()
}

// splitHashdeepLine splits a hashdeep line into its columns. Hashdeep does not quote or escape
// filenames, so a filename can contain commas or quotes. Every other column cannot, so we take
// the columns before the filename from the left and any after it (such as mtime) from the right
func splitHashdeepLine(line string, columns int, filenameIdx int) ([]string, error) {
	fields := make([]string, columns)
	rest := line

	for i := 0; i < filenameIdx; i++ {
		idx := strings.Index(rest, ",")
		if idx == -1 {
			return nil, fmt.Errorf("expected %d columns", columns)
		}
		fields[i] = rest[:idx]
		rest = rest[idx+1:]
	}

	for i := columns - 1; i > filenameIdx; i-- {
		idx := strings.LastIndex(rest, ",")
		if idx == -1 {
			return nil, fmt.Errorf("expected %d columns", columns)
		}
		fields[i] = rest[idx+1:]
		rest = rest[:idx]
	}

	if rest == "" {
		return nil, errors.New("empty filename")
	}
	fields[filenameIdx] = rest
	return fields, nil
}

// scanLinesKeepCR splits lines like bufio.ScanLines, but leaves any \r before the newline in place
func scanLinesKeepCR(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// conflicts counts the records with different values listed for one of their hashes
func (hdl *Auditor) conflicts() int {
	n := 0
	for _, r := range hdl.records {
		if r.Conflict {
			n++
		}
	}
	return n
}

// mergeSumHash adds a hash from a sum file line to the record for its file, a different value
// for a hash it already has meaning the file can never match
func mergeSumHash(r *AuditRecord, name string, value string) {
	if existing, ok := r.Hashes[name]; ok && existing != value {
		r.Conflict = true
		return
	}
	r.Hashes[name] = value
}

// parseSumFile parses the output of md5sum, sha256sum and similar tools, including hashit's
// sum format. Both the GNU "hash  filename" and BSD "MD5 (filename) = hash" styles are accepted
func (hdl *Auditor) parseSumFile(input string, auditFilename string) error {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	extHashes := sumExtensions[strings.ToLower(filepath.Ext(auditFilename))]
	byFile := map[string]*AuditRecord{}
	improper := 0

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// hashit's sum format with --mtime before 1.5.0 wrote a "2006-01-02 15:04:05  filename" line
		// per file, it now writes them as comments which are skipped above
		if sumMtimeLine.MatchString(line) {
			continue
		}

		name, value, filename, ok := parseSumLine(line, extHashes)
		if !ok {
			improper++
			continue
		}

		// hashit's sum format writes one line per hash, so merge them into a single record
		r, exists := byFile[filename]
		if !exists {
			r = &AuditRecord{Filename: filename, Size: -1, Hashes: map[string]string{}}
			byFile[filename] = r
			hdl.addRecord(r)
		}
		// md5sum -c checks every line, so the same path listed again with a different value of the
		// same hash can never match
		mergeSumHash(r, name, value)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if improper != 0 {
		// a file such as text or json output has no valid lines, and is not an audit file at all
		if len(hdl.records) == 0 {
			return fmt.Errorf("audit file has no lines in a supported format, %d were not recognised", improper)
		}
		printWarn(fmt.Sprintf("audit file has improperly formatted lines which were ignored: %d", improper))
	}

	return nil
}

var sumMtimeLine = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}  `)

// parseSumLine parses a single line from a sum file, returning the hash name, value and filename
func parseSumLine(line string, extHashes []string) (string, string, string, bool) {
	// md5sum escapes filenames containing a backslash or newline and marks the line with a leading \
	escaped := strings.HasPrefix(line, "\\")
	if escaped {
		line = line[1:]
	}

	var name, value, filename string

	// GNU style "hash  filename" for text mode or "hash *filename" for binary mode
	if idx := strings.Index(line, " "); idx > 0 && idx+2 <= len(line) && isHex(strings.ToLower(line[:idx])) {
		value = strings.ToLower(line[:idx])
		mode := line[idx+1]
		if mode != ' ' && mode != '*' {
			return "", "", "", false
		}
		filename = line[idx+2:]

		for _, h := range extHashes {
			if hashLengths[h] == len(value) {
				name = h
				break
			}
		}
		if name == "" {
			n, ok := sumLengths[len(value)]
			if !ok {
				return "", "", "", false
			}
			name = n
		}
	} else {
		// BSD style "MD5 (filename) = hash"
		open := strings.Index(line, " (")
		closeIdx := strings.LastIndex(line, ") = ")
		if open <= 0 || closeIdx < open {
			return "", "", "", false
		}
		// rhash pads the name to align the output, such as "MD4   (file) = hash"
		n, ok := bsdTagNames[strings.TrimSpace(line[:open])]
		if !ok {
			return "", "", "", false
		}
		name = n
		filename = line[open+2 : closeIdx]
		value = strings.ToLower(strings.TrimSpace(line[closeIdx+4:]))
		if !isHex(value) || len(value) != hashLengths[name] {
			return "", "", "", false
		}
	}

	if escaped {
		filename = unescapeSumFilename(filename)
	}

	if filename == "" {
		return "", "", "", false
	}

	return name, value, filename, true
}

func unescapeSumFilename(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '\\':
				sb.WriteByte('\\')
				i++
				continue
			case 'n':
				sb.WriteByte('\n')
				i++
				continue
			case 'r':
				sb.WriteByte('\r')
				i++
				continue
			}
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// sortedPaths is a helper to report new files in a stable order
func sortedPaths(results []Result) []Result {
	sort.Slice(results, func(i, j int) bool {
		return results[i].File < results[j].File
	})
	return results
}
