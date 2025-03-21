package processor

import (
	"bufio"
	"encoding/csv"
	"strings"
)

// auditRecord represents a single entry from the hashdeep output
// but is designed to work with other formats in the future
type auditRecord struct {
	Size     string
	MD5      string
	SHA256   string
	Filename string
}

// Auditor parses and holds a audit file which can then be used for audit purposes
// by providing methods to look up values by either name or hash and optimised for all
type Auditor struct {
	fileLookup map[string]auditRecord   // filename optimised lookup
	md5Lookup  map[string][]auditRecord // md5 optimised lookup
}

// TODO modify to accept multiple types
func NewAuditor(input string) (*Auditor, error) {
	hdl := Auditor{}

	file, err := hdl.parseHashdeepFile(input)
	if err != nil {
		return nil, err
	}

	hdl.fileLookup = file
	hdl.md5Lookup = map[string][]auditRecord{}

	for _, v := range file {
		hdl.md5Lookup[v.MD5] = append(hdl.md5Lookup[v.MD5], v)
	}

	return &hdl, nil
}

type FileStatus int

const (
	FileMatched FileStatus = iota
	FileModified
	FileMoved
	FileUnknown
)

func (hdl *Auditor) Count() int {
	return len(hdl.fileLookup)
}

func (hdl *Auditor) Find(file, md5, sha256 string) FileStatus {
	// check if filename exists
	r, ok := hdl.fileLookup[file]
	if ok {
		// ok file exists, check if the hash's match
		if r.MD5 == md5 && r.SHA256 == sha256 {
			return FileMatched
		}

		// hash does not match so file has changed
		return FileModified
	}

	matches, ok := hdl.md5Lookup[md5]
	if ok {
		for _, m := range matches {
			if m.MD5 == md5 && r.SHA256 == sha256 {
				return FileMoved
			}
		}
	}

	return FileUnknown
}

// parseHashdeepFile accepts a hashdeep format in and builds the internal
// audit processor on it
func (hdl *Auditor) parseHashdeepFile(input string) (map[string]auditRecord, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var header []string
	auditLookup := map[string]auditRecord{}
	csvStarted := false

	// Read the string line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle header lines starting with %
		if strings.HasPrefix(line, "%") {
			// Extract the header (e.g., "size,md5,sha256,filename")
			if strings.Contains(line, "size") {
				header = strings.Split(strings.TrimPrefix(line, "%%%% "), ",")
				csvStarted = true
			}
			continue
		}

		// Skip comment lines starting with #
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Process CSV data lines once header is found
		if csvStarted {
			reader := csv.NewReader(strings.NewReader(line))
			// Allow for potential spaces after commas
			reader.TrimLeadingSpace = true
			record, err := reader.Read()
			if err != nil {
				return nil, err
			}

			// Map record to auditRecord based on header
			var fh auditRecord
			for i, field := range header {
				if i >= len(record) {
					break
				}
				switch field {
				case "size":
					fh.Size = record[i]
				case "md5":
					fh.MD5 = record[i]
				case "sha256":
					fh.SHA256 = record[i]
				case "filename":
					fh.Filename = record[i]
				}
			}
			auditLookup[fh.Filename] = fh
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return auditLookup, nil
}
