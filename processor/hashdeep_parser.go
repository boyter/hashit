package processor

import (
	"bufio"
	"encoding/csv"
	"strings"
)

// HashdeepAuditRecord represents a single entry from the hashdeep output
type HashdeepAuditRecord struct {
	Size     string
	MD5      string
	SHA256   string // Optional, only present if in header
	Filename string
}

func parseHashdeepFile(input string) ([]HashdeepAuditRecord, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var header []string
	var hashes []HashdeepAuditRecord
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

			// Map record to HashdeepAuditRecord based on header
			var fh HashdeepAuditRecord
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
			hashes = append(hashes, fh)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hashes, nil
}
