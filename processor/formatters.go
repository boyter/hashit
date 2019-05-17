package processor

import (
	"fmt"
	"os"
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

// Used when explicitly for os.exit output when crashing out
func printError(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "ERROR %s: %s", getFormattedTime(), msg)
}
