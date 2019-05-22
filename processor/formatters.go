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
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("ERROR %s: %s", getFormattedTime(), msg))
}

// Prints a message to stdout if flag to enable trace output is set
func printTrace(msg string) {
	if Trace {
		fmt.Println(fmt.Sprintf("TRACE %s: %s", getFormattedTime(), msg))
	}
}

// Returns the current time as a millisecond timestamp
func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Returns the current time as a nanosecond timestamp as some things
// are far too fast to measure using nanoseconds
func makeTimestampNano() int64 {
	return time.Now().UnixNano()
}

func fileSummarize(input chan Result) string {
	switch {
	case strings.ToLower(Format) == "json":
		return toJSON(input)
	case strings.ToLower(Format) == "hashdeep":
		return toHashDeep(input)
	}

	return toText(input)
}

func toText(input chan Result) string {
	var str strings.Builder

	for res := range input {
		str.WriteString(fmt.Sprintf("%s (%d bytes)\n", res.File, res.Bytes))
		if hasHash(s_md5) {
			str.WriteString("        MD5 " + res.MD5 + "\n")
		}
		if hasHash(s_sha1) {
			str.WriteString("       SHA1 " + res.SHA1 + "\n")
		}
		if hasHash(s_sha256) {
			str.WriteString("     SHA256 " + res.SHA256 + "\n")
		}
		if hasHash(s_sha512) {
			str.WriteString("     SHA512 " + res.SHA512 + "\n")
		}
		if hasHash(s_blake2b256) {
			str.WriteString("Blake2b-256 " + res.Blake2b256 + "\n")
		}
		if hasHash(s_blake2b512) {
			str.WriteString("Blake2b-512 " + res.Blake2b512 + "\n")
		}

		if NoStream == false {
			fmt.Println(str.String())
			str.Reset()
		}
	}

	return str.String()
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

	// TODO you can turn on/off hashes in hashdeep EG hashdeep -c sha1,sha256 processor/*
	// TODO which is not currently supported below

	pwd, err := os.Getwd()
	if err != nil {
		printError(fmt.Sprintf("unable to determine working directory: %s", err.Error()))
		pwd = ""
	}


	str.WriteString("%%%% HASHIT-" + Version + "\n")
	str.WriteString("%%%% size,md5,sha256,filename\n")
	str.WriteString(fmt.Sprintf("## Invoked from: %s\n", pwd))
	str.WriteString(fmt.Sprintf("## $ %s\n", strings.Join(os.Args, " ")))
	str.WriteString("##\n")

	for res := range input {
		str.WriteString(fmt.Sprintf("%d,%s,%s,%s\n", res.Bytes, res.MD5, res.SHA256, res.File))
	}

	return str.String()
}
