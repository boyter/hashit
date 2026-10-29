// SPDX-License-Identifier: MIT

package processor

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// AuditRecord represents a single file entry from an audit file, which can be
// either hashdeep or sum (md5sum, sha256sum etc...) format
type AuditRecord struct {
	Filename string
	Size     int64             // -1 when the audit file does not record size
	Hashes   map[string]string // hashit hash name to lowercase hex value
	Order    int               // position in the audit file, used to report in file order
	Matched  bool
}

// Auditor parses and holds an audit file which can then be used for audit purposes
// by providing methods to look up records by either path or hash
type Auditor struct {
	records     []*AuditRecord
	pathLookup  map[string]*AuditRecord   // every path key a record can be found by
	hashLookup  map[string][]*AuditRecord // "hashname:value" to records, used to find moved files
	hashNames   []string                  // hashes present in the audit file which we are able to calculate
	baseDirs    []string                  // directories relative paths in the audit file may be relative to
	primaryBase string                    // directory relative paths were created relative to, empty for the working directory
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

	var err error
	if isHashdeepFormat(input) {
		err = hdl.parseHashdeepFile(input)
	} else {
		err = hdl.parseSumFile(input, auditFilename)
	}

	// Relative paths are matched against the working directory, and also against the directory
	// containing the audit file, as tools such as RapidCRC do for sum files. For hashdeep files
	// this lets a manifest stored at the root of a set of files be audited after they have been
	// copied elsewhere, where the invoked from directory no longer applies
	if abs, err := filepath.Abs(filepath.Dir(auditFilename)); err == nil {
		hdl.baseDirs = append(hdl.baseDirs, abs)
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

// Lookup finds the record for a file on disk using its path, returning nil if not found
func (hdl *Auditor) Lookup(file string) *AuditRecord {
	for _, k := range pathKeys(file, "") {
		if r, ok := hdl.pathLookup[k]; ok {
			return r
		}
	}
	return nil
}

// Find checks a file on disk against the audit records by path and marks the record as seen
func (hdl *Auditor) Find(res Result) (*AuditRecord, FileStatus) {
	r := hdl.Lookup(res.File)
	if r == nil {
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

// pathKeys returns the keys a path can be looked up by, being the cleaned path and the absolute
// path. Relative paths are made absolute against base, or the working directory when base is empty
func pathKeys(p string, base string) []string {
	clean := filepath.Clean(p)
	keys := []string{clean}

	var abs string
	if filepath.IsAbs(clean) {
		return keys
	}
	if base != "" {
		abs = filepath.Join(base, clean)
		if !filepath.IsAbs(abs) {
			abs, _ = filepath.Abs(abs)
		}
	} else {
		abs, _ = filepath.Abs(clean)
	}
	if abs != "" && abs != clean {
		keys = append(keys, abs)
	}
	return keys
}

// Exclude removes any record which refers to the supplied path, which is used to remove the
// audit file itself as it can never match its own hash
func (hdl *Auditor) Exclude(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return
	}

	var records []*AuditRecord
	for _, r := range hdl.records {
		if !hdl.isOwnEntry(r, abs) {
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

// recordKeys returns every path key a record can be found by, resolving it against each base directory
func (hdl *Auditor) recordKeys(r *AuditRecord) []string {
	if len(hdl.baseDirs) == 0 {
		return pathKeys(r.Filename, "")
	}

	var keys []string
	for _, base := range hdl.baseDirs {
		for _, k := range pathKeys(r.Filename, base) {
			if !contains(keys, k) {
				keys = append(keys, k)
			}
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
func (hdl *Auditor) isOwnEntry(r *AuditRecord, abs string) bool {
	if contains(pathKeys(r.Filename, hdl.primaryBase), abs) {
		return true
	}
	return contains(hdl.recordKeys(r), abs) && isEmptyRecord(r)
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
	seen := map[string]bool{}
	for _, r := range hdl.records {
		for _, k := range hdl.recordKeys(r) {
			// first record wins where the same file appears more than once
			if _, ok := hdl.pathLookup[k]; !ok {
				hdl.pathLookup[k] = r
			}
		}
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

	var header []string
	filenameIdx := -1
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		// filenames can start or end with spaces so only remove line endings
		line := strings.TrimRight(scanner.Text(), "\r")

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
			if dir, ok := strings.CutPrefix(line, "## Invoked from: "); ok && len(hdl.baseDirs) == 0 {
				hdl.baseDirs = append(hdl.baseDirs, strings.TrimSpace(dir))
				hdl.primaryBase = strings.TrimSpace(dir)
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
		r.Hashes[name] = value
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if improper != 0 {
		// a file such as text or json output has no valid lines, and is not an audit file at all
		if len(hdl.records) == 0 {
			return fmt.Errorf("audit file has %d lines and none are in a supported format", improper)
		}
		printWarn(fmt.Sprintf("audit file has %d improperly formatted lines which were ignored", improper))
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
