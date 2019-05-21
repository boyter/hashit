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
	str.WriteString("## $ hashdeep processor/file.go processor/formatters.go processor/processor.go processor/structs.go processor/workers.go\n")
	str.WriteString("##\n")

	for res := range input {
		str.WriteString(fmt.Sprintf("%d,%s,%s,%s\n", res.Bytes, res.MD5, res.SHA256, res.File))
	}

	//%%%% HASHDEEP-1.0
	//%%%% size,md5,sha256,filename
	//## Invoked from: /mnt/c/Users/bboyter/Documents/Go/src/github.com/boyter/hashit
	//## $ hashdeep processor/file.go processor/formatters.go processor/processor.go processor/structs.go processor/workers.go
	//##
	//1459,aa62c46f8afb02d3923309f997a21a32,5683ad437b2210ec63fc78db6ca3b3f8930ad1287e5ff1fd4d156ccf05c243f2,/mnt/c/Users/bboyter/Documents/Go/src/github.com/boyter/hashit/processor/formatters.go
	//2278,f9e775a4ecfa51a1c5638a767012d5fc,ae7030d6495a15c6583516a89ebf166ce7a84cf2a3fced66f25af422824ed90d,/mnt/c/Users/bboyter/Documents/Go/src/github.com/boyter/hashit/processor/processor.go
	//214,60ed00bfe019fbf3e894cd6336b4b6be,d814dcc4beb533a7a55655519531c5c0a0f525007b8f6218775330cf54ee0578,/mnt/c/Users/bboyter/Documents/Go/src/github.com/boyter/hashit/processor/structs.go
	//528,d4b44389202317f4fe541a26d22ee5ec,eb5189227751bbe489d4a0dcd477fad87bf5bcff8a333e8b410487a1c0037e3b,/mnt/c/Users/bboyter/Documents/Go/src/github.com/boyter/hashit/processor/file.go
	//7916,d305c56350944199d831c439244f67cd,06df06a6f8a379a4a391a3e496799b4b8e253b38b0973893d0a27032dff284a7,/mnt/c/Users/bboyter/Documents/Go/src/github.com/boyter/hashit/processor/workers.go

	return str.String()
}