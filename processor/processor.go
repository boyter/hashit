// SPDX-License-Identifier: MIT

package processor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/gosuri/uiprogress"
)

// Global Version
var Version = "1.5.0 (beta)"

// Verbose enables verbose logging output
var Verbose = false

// VeryVerbose enables verbose logging output
var VeryVerbose = false

// Debug enables debug logging output
var Debug = false

// Trace enables trace logging output which is extremely verbose
var Trace = false

// MTime enable mtime calculation and output
var MTime = false

// Progress uses ui bar to display the progress of files
var Progress = false

// Recursive to walk directories
var Recursive = false

// Do not print out results as they are processed
var NoStream = false

// Overwrite allows output to replace an existing file
var Overwrite = false

// Absolute outputs absolute paths, which is the default for hashdeep
var Absolute = false

// AllowMoved allows an audit to pass when known files have moved, so long as they are unmodified
var AllowMoved = false

// If data is being piped in using stdin
var StandardInput = false

// Should the application print all hashes it knows about
var Hashes = false

// List of hashes that we want to process
var Hash = []string{}

// Format sets the output format of the formatter
var Format = ""

// FileOutput sets the file that output should be written to
var FileOutput = ""

// AuditFile sets the file that we want to audit against similar to hashdeep
var AuditFile = ""

// progressBars renders the --progress bars
var progressBars *uiprogress.Progress

// auditor holds the parsed AuditFile
var auditor *Auditor

// sqliteCompareHashes are the hashes both requested and stored in a SQLite AuditFile, strongest first
var sqliteCompareHashes []string

// skipFiles holds the absolute paths of files which should never be processed, such as the audit
// file or output file when they are inside the directory being processed
var skipFiles = map[string]bool{}

// addSkipFile adds a file to skipFiles, along with the files SQLite creates alongside a database
func addSkipFile(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return
	}
	for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
		skipFiles[abs+suffix] = true
	}
}

// shouldSkipFile reports if the file is one we should never process
func shouldSkipFile(path string) bool {
	if len(skipFiles) == 0 {
		return false
	}
	abs, err := filepath.Abs(path)
	return err == nil && skipFiles[abs]
}

// DirFilePaths is not set via flags but by arguments following the flags for file or directory to process
var DirFilePaths = []string{}

// FileListQueueSize is the queue of files found and ready to be processed
var FileListQueueSize = 1000

// Number of bytes in a size to enable memory maps or streaming
var StreamSize int64 = 1_000_000

// FileInput indicates we have a file passed in which consists of a
var FileInput = ""

// GitIgnore set false to enable .gitignore checks
var GitIgnore = true

// GitModuleIgnore set false to enable .gitmodules checks
var GitModuleIgnore = true

// Ignore set false to enable ignore file checks
var Ignore = true

// HashIgnore set true to enable hashignore file checks
var HashIgnore = false

// SkipHidden set to false to ignore hidden folders and files
var SkipHidden = true

// PathDenyList sets the paths that should be skipped
var PathDenyList = []string{}

// Exclude is a regular expression which is used to exclude files from being processed
var Exclude = []string{}

var NoThreads = runtime.NumCPU()

// String mapping for hash names
var HashNames = Result{
	CRC32:      "crc32",
	XxHash64:   "xxhash64",
	MD4:        "md4",
	MD5:        "md5",
	SHA1:       "sha1",
	SHA224:     "sha224",
	SHA256:     "sha256",
	SHA384:     "sha384",
	SHA512:     "sha512",
	Blake2b256: "blake2b256",
	Blake2b512: "blake2b512",
	Blake3:     "blake3",
	Sha3224:    "sha3224",
	Sha3256:    "sha3256",
	Sha3384:    "sha3384",
	Sha3512:    "sha3512",
	Ed2k:       "ed2k",
}

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	// Display the supported hashes then bail out
	if Hashes {
		printHashes()
		return
	}

	// Check if we are accepting data from stdin, which is never the case when reading the files
	// to process from an input file or auditing, as both are about files on disk
	if len(DirFilePaths) == 0 && FileInput == "" && AuditFile == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			StandardInput = true
		}
	}

	// If nothing was supplied as an argument to run against, assume run against everything in the
	// current directory recursively
	if FileInput != "" && len(DirFilePaths) != 0 {
		printWarn("paths supplied as arguments are ignored when using --input")
	}

	defaultPath := false
	if len(DirFilePaths) == 0 {
		DirFilePaths = append(DirFilePaths, ".")
		defaultPath = true
	}

	// If a single argument is supplied, enable recursive as if it's a file no problem
	// but if its a directory the user probably wants to hash everything in that directory.
	// When auditing directories are always processed, otherwise every file in a skipped
	// directory would be reported as missing
	if len(DirFilePaths) == 1 || AuditFile != "" {
		Recursive = true
	}

	// Clean up hashes by setting all input to lowercase
	Hash = formatHashInput()

	// --vv is a more verbose -v, so it should include everything -v outputs
	if VeryVerbose {
		Verbose = true
	}

	if err := validateOptions(); err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	if AllowMoved && AuditFile == "" {
		printWarn("--allow-moved only applies when auditing with --audit, ignoring flag")
	}

	// Where audit file is set we load it up front so we only calculate the hashes it contains
	if AuditFile != "" {
		if isSQLiteDB(AuditFile) {
			// we calculate the hashes requested and compare those the database also has
			stored, err := sqliteAuditHashes(AuditFile)
			if err != nil {
				printError(fmt.Sprintf("unable to read audit database %s: %s", AuditFile, err.Error()))
				os.Exit(1)
			}
			sqliteCompareHashes = nil
			for _, h := range stored {
				if hasHash(h) {
					sqliteCompareHashes = append(sqliteCompareHashes, h)
				}
			}
			if len(sqliteCompareHashes) == 0 {
				printWarn(fmt.Sprintf("none of the requested hashes %s are in audit database which has %s, every file will be reported as modified", strings.Join(Hash, ","), strings.Join(stored, ",")))
			}
		} else {
			file, err := os.ReadFile(AuditFile)
			if err != nil {
				printError(fmt.Sprintf("unable to read audit file: %s", err.Error()))
				os.Exit(1)
			}

			auditor, err = NewAuditor(string(file), AuditFile)
			if err != nil {
				printError(fmt.Sprintf("unable to parse audit file %s: %s", AuditFile, err.Error()))
				os.Exit(1)
			}
			// the audit file cannot contain its own hash, so never process or expect it
			auditor.Exclude(AuditFile)
			Hash = auditor.HashNames()
		}
		addSkipFile(AuditFile)
	}

	// The sqlite format writes to the database as we go, so default the name up front
	sqliteOutput := strings.ToLower(Format) == "sqlite" && AuditFile == ""
	if sqliteOutput && FileOutput == "" {
		FileOutput = "hashit.db"
	}

	// Check before doing any work that we are not going to destroy an existing file
	if FileOutput != "" {
		if err := prepareOutput(FileOutput, sqliteOutput); err != nil {
			printError(err.Error())
			os.Exit(1)
		}
	}

	// The output file is being written while we process, so its hash is meaningless
	if FileOutput != "" {
		addSkipFile(FileOutput)
	}

	// If format is set to hashdeep
	if strings.ToLower(Format) == "hashdeep" {
		// if hashdeep, we need to ensure md5 is set or sha256, or both, and if neither are set both
		if !hasHash("md5") && !hasHash("sha256") {
			printWarn("hashdeep format without supported hash - setting md5 and sha256")
			Hash = []string{"md5", "sha256"}
		}
	}

	// Results ready to be printed
	fileSummaryQueue := make(chan Result, FileListQueueSize)

	if StandardInput {
		if MTime {
			printWarn("cannot use --mtime option with standard input, ignoring flag")
		}
		go processStandardInput(fileSummaryQueue)
	} else {
		// Files ready to be read from disk
		fileListQueue := make(chan string, FileListQueueSize)

		if FileInput == "" {
			// Spawn routine to start finding files on disk
			go func() {
				// Check if the paths or files added exist and inform the user if they don't
				for _, f := range DirFilePaths {
					fp := filepath.Clean(f)
					fi, err := os.Stat(fp)

					// keep a leading ./ the user supplied as hashdeep and md5sum do, so that our
					// output matches theirs when run the same way
					prefix := ""
					if Absolute {
						if abs, err := filepath.Abs(fp); err == nil {
							fp = abs
						}
					} else if !defaultPath {
						prefix = dotPrefix(f, fp)
					}

					// A path which does not exist is reported and the rest still processed, then we
					// exit non zero at the end, in the same way as a file which cannot be read
					if err != nil {
						printFileError(fmt.Sprintf("file or directory issue: %s %s", fp, err.Error()))
						continue
					}

					if !fi.IsDir() {
						fileListQueue <- prefix + fp
						continue
					}

					if Recursive {
						walkDirectoryWithIgnore(fp, prefix, fileListQueue)
					} else {
						printWarn(fmt.Sprintf("Skipping directory %s, use -r to process directories when supplying multiple paths", prefix+fp))
					}
				}
				close(fileListQueue)
			}()
		} else {
			// Open the file
			go func() {
				file, err := os.Open(FileInput)
				if err != nil {
					printError(fmt.Sprintf("failed to open input file: %s, %s", FileInput, err.Error()))
					os.Exit(1)
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)

				// Read the file line by line
				for scanner.Scan() {
					// skip blank lines, and allow lists written with windows line endings
					line := strings.TrimRight(scanner.Text(), "\r")
					if line == "" {
						continue
					}
					if Absolute {
						if abs, err := filepath.Abs(line); err == nil {
							line = abs
						}
					}
					fileListQueue <- line
				}
				close(fileListQueue)

				// Check for errors during scanning
				if err := scanner.Err(); err != nil {
					printError(fmt.Sprintf("error reading input file: %s, %s", FileInput, err.Error()))
					os.Exit(1)
				}
			}()
		}

		if Progress {
			// render to stderr so the bars do not end up in redirected output
			progressBars = uiprogress.New()
			progressBars.SetOut(os.Stderr)
			progressBars.Start()
		}

		var wg sync.WaitGroup
		for i := 0; i < NoThreads; i++ {
			wg.Add(1)
			go func() {
				fileProcessorWorker(fileListQueue, fileSummaryQueue)
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			close(fileSummaryQueue)
		}()
	}

	result, valid := fileSummarize(fileSummaryQueue)

	// any file which could not be processed means the output is incomplete
	if fileErrors.Load() {
		valid = false
	}

	if FileOutput == "" {
		fmt.Print(result)
		if !valid {
			os.Exit(1)
		}
	} else {
		// we don't write out sqlite as it is written as we go
		if !sqliteOutput {
			if err := os.WriteFile(FileOutput, []byte(result), 0600); err != nil {
				printError(fmt.Sprintf("unable to write output file %s: %s", FileOutput, err.Error()))
				os.Exit(1)
			}
		}
		fmt.Println("results written to " + FileOutput)
		if !valid {
			os.Exit(1)
		}
	}
}

var validFormats = []string{"text", "json", "sum", "hashdeep", "hashonly", "sqlite"}

// validateOptions checks the requested format and hashes are ones we support, rather than
// silently falling back to text output or skipping hashes we do not know
func validateOptions() error {
	if !contains(validFormats, strings.ToLower(Format)) {
		return fmt.Errorf("unknown format %q, supported formats are %s", Format, strings.Join(validFormats, ", "))
	}

	for _, h := range Hash {
		if h != "all" && !contains(hashPriority, h) {
			return fmt.Errorf("unknown hash %q, use --hashes to list supported hashes", h)
		}
	}

	if NoThreads < 1 {
		return fmt.Errorf("--threads must be at least 1, got %d", NoThreads)
	}

	for _, exclude := range Exclude {
		if _, err := regexp.Compile(exclude); err != nil {
			return fmt.Errorf("invalid --not-match regular expression %q: %s", exclude, err.Error())
		}
	}

	// check the output can be created before doing any work, rather than failing at the end
	if FileOutput != "" {
		if fi, err := os.Stat(filepath.Dir(FileOutput)); err != nil || !fi.IsDir() {
			return fmt.Errorf("unable to write output file %s as its directory does not exist", FileOutput)
		}
	}
	return nil
}

// prepareOutput refuses to write over an existing file unless overwrite has been allowed. An
// existing sqlite database is removed, as writing into it would merge with the previous results
func prepareOutput(path string, sqlite bool) error {
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to check output file %s: %s", path, err.Error())
	}

	if !Overwrite {
		return fmt.Errorf("output file %s already exists, use --overwrite to replace it", path)
	}

	if !sqlite {
		return nil
	}

	// guard against a mistyped output name destroying something which is not a database
	if fi, err := os.Stat(path); err == nil && fi.Size() != 0 && !isSQLiteDB(path) {
		return fmt.Errorf("output file %s is not a SQLite database, refusing to replace it", path)
	}

	for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
		if err := os.Remove(path + suffix); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to remove existing output file %s: %s", path+suffix, err.Error())
		}
	}
	return nil
}

// dotPrefix returns "./" when the supplied path started with it, or was the current directory,
// and the cleaned path does not already start with it. Paths which are absolute or leave the
// current directory such as "./../a" are returned as cleaned without a prefix
func dotPrefix(supplied string, cleaned string) string {
	sep := string(filepath.Separator)
	if supplied != "." && !strings.HasPrefix(supplied, "./") && !strings.HasPrefix(supplied, "."+sep) {
		return ""
	}
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+sep) {
		return ""
	}
	return "." + sep
}

// ToLower all of the input hashes so we can match them easily
func formatHashInput() []string {
	h := []string{}
	for _, x := range Hash {
		h = append(h, strings.ToLower(x))
	}
	return h
}

// Check if a hash was supplied to the input so we know if we should calculate it
func hasHash(hash string) bool {
	for _, x := range Hash {
		if x == "all" {
			return true
		}

		if x == hash {
			return true
		}
	}

	return false
}
