package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Get the time as standard UTC/Zulu format
func getFormattedTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Prints a message to stdout if flag to enable warning output is set
func printVerbose(msg string) {
	if Verbose {
		fmt.Println(fmt.Sprintf("VERBOSE %s: %s", getFormattedTime(), msg))
	}
}

// Prints a message to stdout if flag to enable debug output is set
func printDebug(msg string) {
	if Debug {
		fmt.Println(fmt.Sprintf("DEBUG %s: %s", getFormattedTime(), msg))
	}
}

// Used when explicitly for os.exit output when crashing out
func printError(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, "ERROR %s: %s", getFormattedTime(), msg)
}

func fileSummarize(input chan Result) string {
	switch {
	case strings.ToLower(Format) == "json":
		return toJSON(input)
	case strings.ToLower(Format) == "csv":
	//return toCSV(input)
	case strings.ToLower(Format) == "hashdeep":
			return toHashDeep(input)
	}

	for res := range input {
		fmt.Println(res.File)
		fmt.Println("   MD5 " + res.MD5)
		fmt.Println("  SHA1 " + res.SHA1)
		fmt.Println("SHA512 " + res.SHA512)
		fmt.Println("")
	}

	return ""
}

func toJSON(input chan Result) string {
	results := []Result{}
	for res := range input {
		results = append(results, res)
	}

	jsonString, _ := json.Marshal(results)
	return string(jsonString)
}

func toHashDeep(input chan Result) string {
	var str strings.Builder

	str.WriteString("%%%% HASHIT-" + Version + "\n")
	str.WriteString("%%%% size,md5,sha256,filename\n")
	str.WriteString("## Invoked from: NEEDS TO GO HERE\n")
	str.WriteString("## $ hashdeep NEEDS TO GO HERE\n")
	str.WriteString("##\n")

	for res := range input {
		str.WriteString(fmt.Sprintf("%d,%s,%s,%s\n", res.Bytes, res.MD5, res.SHA256, res.File))
	}

	return str.String()
}