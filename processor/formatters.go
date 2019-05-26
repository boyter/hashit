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
		//case strings.ToLower(Format) == "clean":
		//	return toClean(input)
	case strings.ToLower(Format) == "sum":
		return toSum(input)
	}
	// TODO add output format that is just the hash for single files
	// TODO add output format that confirms to md5sum

	return toText(input)
}

// Mimics how md5sum sha1sum etc... work
func toSum(input chan Result) string {
	var str strings.Builder

	first := true

	for res := range input {
		if first == false {
			str.WriteString("\n")
		} else {
			first = false
		}

		if hasHash(HashNames.MD4) {
			str.WriteString(res.MD4 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.MD5) {
			str.WriteString(res.MD5 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.SHA1) {
			str.WriteString(res.SHA1 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.SHA256) {
			str.WriteString(res.SHA256 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.SHA512) {
			str.WriteString(res.SHA512 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Blake2b256) {
			str.WriteString(res.Blake2b256 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Blake2b512) {
			str.WriteString(res.Blake2b512 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3224) {
			str.WriteString(res.Sha3224 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3256) {
			str.WriteString(res.Sha3256 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3384) {
			str.WriteString(res.Sha3384 + "  " + res.File + "\n")
		}
		if hasHash(HashNames.Sha3512) {
			str.WriteString(res.Sha3512 + "  " + res.File + "\n")
		}

		if NoStream == false {
			fmt.Print(str.String())
			str.Reset()
		}
	}

	return str.String()
}

func toText(input chan Result) string {
	var str strings.Builder

	first := true

	for res := range input {
		if first == false {
			str.WriteString("\n")
		} else {
			first = false
		}

		str.WriteString(fmt.Sprintf("%s (%d bytes)\n", res.File, res.Bytes))

		if hasHash(HashNames.MD4) {
			str.WriteString("        MD4 " + res.MD4 + "\n")
		}
		if hasHash(HashNames.MD5) {
			str.WriteString("        MD5 " + res.MD5 + "\n")
		}
		if hasHash(HashNames.SHA1) {
			str.WriteString("       SHA1 " + res.SHA1 + "\n")
		}
		if hasHash(HashNames.SHA256) {
			str.WriteString("     SHA256 " + res.SHA256 + "\n")
		}
		if hasHash(HashNames.SHA512) {
			str.WriteString("     SHA512 " + res.SHA512 + "\n")
		}
		if hasHash(HashNames.Blake2b256) {
			str.WriteString("Blake2b-256 " + res.Blake2b256 + "\n")
		}
		if hasHash(HashNames.Blake2b512) {
			str.WriteString("Blake2b-512 " + res.Blake2b512 + "\n")
		}
		if hasHash(HashNames.Sha3224) {
			str.WriteString("   SHA3-224 " + res.Sha3224 + "\n")
		}
		if hasHash(HashNames.Sha3256) {
			str.WriteString("   SHA3-256 " + res.Sha3256 + "\n")
		}
		if hasHash(HashNames.Sha3384) {
			str.WriteString("   SHA3-384 " + res.Sha3384 + "\n")
		}
		if hasHash(HashNames.Sha3512) {
			str.WriteString("   SHA3-512 " + res.Sha3512 + "\n")
		}

		if NoStream == false {
			fmt.Print(str.String())
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

func printHashes() {
	fmt.Println(fmt.Sprintf("        MD4 (%s)", HashNames.MD4))
	fmt.Println(fmt.Sprintf("        MD5 (%s)", HashNames.MD5))
	fmt.Println(fmt.Sprintf("       SHA1 (%s)", HashNames.SHA1))
	fmt.Println(fmt.Sprintf("     SHA256 (%s)", HashNames.SHA256))
	fmt.Println(fmt.Sprintf("     SHA512 (%s)", HashNames.SHA512))
	fmt.Println(fmt.Sprintf("Blake2b-256 (%s)", HashNames.Blake2b256))
	fmt.Println(fmt.Sprintf("Blake2b-512 (%s)", HashNames.Blake2b512))
	fmt.Println(fmt.Sprintf("   SHA3-224 (%s)", HashNames.Sha3224))
	fmt.Println(fmt.Sprintf("   SHA3-256 (%s)", HashNames.Sha3256))
	fmt.Println(fmt.Sprintf("   SHA3-384 (%s)", HashNames.Sha3384))
	fmt.Println(fmt.Sprintf("   SHA3-512 (%s)", HashNames.Sha3512))
}
