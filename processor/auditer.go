package processor

import (
	"fmt"
	"os"
)

func Auditer() {
	// open the audit file
	file, err := os.ReadFile(AuditFile)
	if err != nil {
		printError(err.Error())
		return
	}

	hashdeepFile, err := parseHashdeepFile(string(file))
	if err != nil {
		printError(err.Error())
		return
	}

	fmt.Println(hashdeepFile)
}
