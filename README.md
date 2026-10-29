Hash It!
--------

A hash tool which can work like hashdeep or md5sum, sha1sum, etc... When you want to find the hash or hashes of a file quickly, cross platform using a single command.

Yes the name is *very intentional* and similar to what say when I realise I need to get multiple hashes of a large file. This tool attempts to solve that pain.


[![Go](https://github.com/boyter/hashit/actions/workflows/go.yml/badge.svg)](https://github.com/boyter/hashit/actions/workflows/go.yml)
[![Scc Count Badge](https://sloc.xyz/github/boyter/hashit/)](https://github.com/boyter/hashit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/hashit)](https://goreportcard.com/report/github.com/boyter/hashit)

Licensed under the MIT license.

Other similar projects,

 - [hashdeep](https://github.com/jessek/hashdeep) originally called md5deep

### Support

Using `hashit` commercially? If you want priority support for `hashit` you can purchase a years worth https://app.gumroad.com/products/mcpni which entitles you to priority direct email support from the developer.
 
### Install

If you are comfortable using Go and have >= 1.25 installed `go install github.com/boyter/hashit@latest` will install for you.

Binaries are available on [Releases](https://github.com/boyter/hashit/releases) for Windows, GNU/Linux and macOS.

If you would like to assist with getting `hashit` added into apt/homebrew/chocolatey/etc... please submit a PR or at least raise an issue with instructions.

### Development

You need to have Go installed. Minimum version is Go 1.25 https://go.dev/

Install the following tools, either via the indicated command or what is suggested on site

- sqlc brew install sqlc https://sqlc.dev/

sqlc is used for the audit functionality as well as the SQLite output format. If you never change this functionality
it may not be required, however _never_ edit the `./processor/database/` files directly. Edit `./assets/db/` and run `sqlc generate` instead.

### Pitch

Why use `hashit`?

 - It is very fast
 - You can get multiple hashes "for free" on any CPU with multiple cores
 - Works very well across multiple platforms without slowdown (Windows, Linux, macOS)
 - Supports many hashes `hashit --hashes` CRC32, xxHash64, MD4, MD5, SHA1, SHA256, SHA512, Blake2b-256, Blake2b-512, Blake3, SHA3-224, SHA3-256, SHA3-384, SHA3-512, ed2k
 - Output is compatible with `hashdeep`

### Usage

Command line usage of `hashit` is designed to be as simple as possible.
Full details can be found in `hashit --help` or `hashit -h`.

```
Hash It!
Version 1.5.0 (beta)
Ben Boyter <ben@boyter.org>

Usage:
  hashit [flags]

Flags:
      --absolute                output absolute paths, which is what hashdeep does without -l
      --allow-moved             allow an audit to pass when known files have moved or been renamed but are unmodified
  -a, --audit string            audit against supplied file; accepts hashdeep, sum (md5sum, sha256sum etc...) or hashit sqlite format
      --debug                   enable debug output
      --exclude-dir strings     directories to exclude
  -f, --format string           set output format [text, json, sum, hashdeep, hashonly, sqlite] (default "text")
      --gitignore               enable .gitignore file logic
      --gitmodule               enable .gitmodules file logic
  -c, --hash strings            hashes to be run for each file (set to 'all' for all possible hashes) (default [md5,sha1,sha256,sha512])
      --hashes                  list all supported hashes
      --hashignore              enable .hashignore file logic
  -h, --help                    help for hashit
      --ignore                  enable .ignore file logic
  -i, --input string            input file of newline separated file locations to process
      --mtime                   enable mtime output
      --no-stream               do not print results as files are processed for text, sum and hashonly formats; other formats and --output never stream
  -M, --not-match stringArray   ignore files and directories whose name matches regular expression
  -o, --output string           output filename (default stdout)
      --overwrite               allow output to replace an existing file
  -p, --progress                display progress of files as they are processed
  -r, --recursive               traverse directories when multiple paths are supplied; a single directory or no path is always traversed
      --skip-hidden             skip hidden files and directories
      --stream-size int         files larger than this size in bytes are read in chunks rather than all at once (default 1000000)
      --threads int             number of threads processing files, by default the number of CPU cores (default 8)
      --trace                   enable trace output
  -v, --verbose                 verbose audit output, listing every modified, moved, new and missing file
      --version                 version for hashit
      --vv                      very verbose audit output, as --verbose and also listing every matched file
```

Output should look something like the below for operations on this repository

```
$ hashit LICENSE
LICENSE (1067 bytes)
        MD5 227f999ca03b135a1b4d69bde84afb16
       SHA1 2c7beb1563981414f31481016816a88dc0cdba16
     SHA256 fb3f44f5e74b957107f89b027896250ecff74718b1fa8bf0566874e142e54351
     SHA512 882eb7ee56cda80778dbd44f514e68a14126a61a11a825f9c030a3425b941d6a3636561a466347178d8bf204dbedc806471d248637091b311ca37ec18a1c084b

$ hashit --hash all LICENSE
LICENSE (1067 bytes)
      CRC32 14ec1e20
   xxHash64 07fb115a53dea4ca
        MD4 a4961e3d704e28a40debbf1f1e4a62cd
        MD5 227f999ca03b135a1b4d69bde84afb16
       SHA1 2c7beb1563981414f31481016816a88dc0cdba16
     SHA256 fb3f44f5e74b957107f89b027896250ecff74718b1fa8bf0566874e142e54351
     SHA512 882eb7ee56cda80778dbd44f514e68a14126a61a11a825f9c030a3425b941d6a3636561a466347178d8bf204dbedc806471d248637091b311ca37ec18a1c084b
Blake2b-256 824c7551cee5a6996c422988f658d2622890fa52e2fcf63d0aa6c01abc69e77d
Blake2b-512 ad692390b5e680f24490d5f4676f279f84e802c79b98ef36369404b88d8ced237dbde429098f6e7bb221b6ae8fff4ab1920b03fd187c3025c4f81be2a3def672
     Blake3 339416b7c6155ddd00a7da8c6fbb40b472fbc113c7ca532f1929cffffa81f892
   SHA3-224 845411d717203ef1896ab41ff4b2d468958b2e067807713bd49867ed
   SHA3-256 e334df710e2791c6df492edd61d1d17022be91044937e29d8613ddfb43f4aec6
   SHA3-384 8301eb66f20215a14a77aa6e977a9d08524caf31fa09a9fd97e7b3c4c3e9266159d9c35c661960024654ff36f5131302
   SHA3-512 1024928532bccfe9feb0b3f9fe54eda6fb05685fb3236eb09cb3765f0d1face3ea5f5dc681aaeec44134cd89691348a03cc6a2abfe1eae55d8eb274a51f97e3d
       ed2k a4961e3d704e28a40debbf1f1e4a62cd

$ hashit --hash md5 processor/database
processor/database/db.go (601 bytes)
        MD5 f43f14400d5c84a60151efa94680442d

processor/database/models.go (663 bytes)
        MD5 d984b3ca676ceca67546d20ba8395559

processor/database/queries.sql.go (4162 bytes)
        MD5 5e44a725583f351d381d7c70e3aa21e7
```

Files are processed in parallel, so the order of the output may vary between runs.

If any file cannot be read, such as one which is missing or lacks permissions, an error is printed, the remaining files are still processed, and `hashit` exits with a non zero code.

### Ignoring Files (`.hashignore`)

When hashing a directory, you may want to exclude certain files or directories from the output, such as `node_modules`, log files, or build artifacts. `hashit` supports this through a `.hashignore` file.

The `.hashignore` file uses the **exact same syntax as a standard `.gitignore` file**. You can specify files, directories, and wildcard patterns that you want `hashit` to ignore during its scan.

By default, processing of `.hashignore` files is disabled. You must explicitly enable it using the `--hashignore` flag.

#### Example

Imagine you have the following directory structure:

```
my_project/
├── .hashignore
├── main.go
├── data.log
└── node_modules/
    └── some_package/
        └── index.js
```

And your `.hashignore` file contains:

```
# Ignore log files
*.log

# Ignore node_modules directory
node_modules/
```

To hash the project while respecting the ignore rules, you would run:

```shell
$ hashit --hashignore my_project/
```

In this case, `hashit` would process `main.go` but would skip `data.log` and the entire `node_modules` directory.

### Auditing

`hashit` provides a powerful auditing feature that is compatible with `hashdeep`. This allows you to verify the integrity of a set of files by comparing their current state against a previously generated list of known hashes.

You can use `hashit` to generate an audit file and have `hashdeep` verify it, or use `hashdeep` to generate the audit file and have `hashit` verify it.

#### Generating an Audit File

To generate an audit file compatible with `hashdeep`, use the `--format hashdeep` flag.

```shell
$ hashit --format hashdeep processor > audit.txt
```

This will create a file named `audit.txt` containing the hashes of the files in the `processor` directory.

#### Verifying with `hashdeep`

You can then use `hashdeep` to audit a directory against the file generated by `hashit`.

```shell
# First, ensure hashdeep is installed
# Then, generate the audit file with hashit
$ hashit --format hashdeep processor > audit.txt

# Now, audit with hashdeep, using -l as hashit writes relative paths
$ hashdeep -l -r -a -v -k audit.txt processor
hashdeep: Audit passed
          Files matched: 14
Files partially matched: 0
            Files moved: 0
        New files found: 0
  Known files not found: 0
```

#### Verifying with `hashit`

Similarly, you can use `hashit` to audit against a file generated by `hashdeep`. `hashit`'s `-a` flag is used to specify the audit file.

```shell
# First, ensure hashdeep is installed
# Then, generate the audit file with hashdeep
$ hashdeep -r processor > audit.txt

# Now, audit with hashit
$ hashit -a audit.txt processor
hashit: Audit passed
       Files examined: 14
Known files expecting: 14
        Files matched: 14
       Files modified: 0
          Files moved: 0
      New files found: 0
        Files missing: 0
```

#### Understanding Audit Results

`hashit`'s audit output is designed to be similar to `hashdeep`'s verbose output, providing a clear summary of what has changed.

Here's an example of a failed audit where a file was modified:

```shell
# Setup a temporary directory for the example
$ mkdir -p /tmp/hashit-audit-test
$ echo "original content" > /tmp/hashit-audit-test/file.txt

# Create an audit file
$ hashit --format hashdeep /tmp/hashit-audit-test > audit.txt

# Modify the file
$ echo "new content" >> /tmp/hashit-audit-test/file.txt

# Run the audit
$ hashit -a audit.txt /tmp/hashit-audit-test
hashit: Audit failed
       Files examined: 1
Known files expecting: 1
        Files matched: 0
       Files modified: 1
          Files moved: 0
      New files found: 0
        Files missing: 0

# Clean up
$ rm -rf /tmp/hashit-audit-test audit.txt
```

`hashit` can also detect moved files, distinguishing them from new or missing files. As with `hashdeep`, an audit only passes when every file is exactly where it was, so moved files cause the audit to fail. To check that every known file is still present and unmodified, while allowing files to have been moved or renamed, use `--allow-moved`. New, missing or modified files still fail the audit.

```shell
$ mv a.txt archive/a.txt
$ hashit --allow-moved -a audit.txt .
hashit: Audit passed
       Files examined: 8
Known files expecting: 8
        Files matched: 7
       Files modified: 0
          Files moved: 1
      New files found: 0
        Files missing: 0
```
 Use `-v` to list every file which is modified, moved, new or missing. Results are reported in the order of the audit file, followed by any new files.

#### Modification Times

With `--mtime` the modification time of each file is written as a comment line after it, such as `## mtime 2026-01-01 10:00:00`. `hashdeep` ignores comment lines, so the file remains compatible with it, and filenames containing commas are not ambiguous.

#### Absolute Paths

`hashdeep` outputs absolute paths unless run with `-l`, whereas `hashit` outputs paths as supplied, which is the same as `hashdeep -l`. Relative paths make a manifest which can be shipped with the files and verified elsewhere. To match `hashdeep`'s default use `--absolute`.

```shell
$ hashit --absolute --format hashdeep . > ../audit.txt
$ hashdeep -r -a -k ../audit.txt .
hashdeep: Audit passed
```

#### Paths in Audit Files

Relative paths in a `hashdeep` file are resolved against the directory recorded in its `## Invoked from:` line, so an audit file created with `hashit --format hashdeep .` can be audited from anywhere using either a relative or absolute path. Paths are also matched against the current directory, and `./file` and `file` are treated as the same file.

If the audit file is inside the directory being audited it is skipped, as it can never contain its own hash. `hashdeep` and `md5sum` do not do this, so when generating a file for them to verify, write it outside the directory being hashed, otherwise they will report the partially written file as changed.

#### Auditing with `sum` Files

`hashit` can also audit against the output of `md5sum`, `sha1sum`, `sha256sum`, `sha512sum` and similar tools, including the BSD style output of `--tag`, as well as files generated by `hashit --format sum`.

```shell
$ md5sum * > checksums.md5
$ hashit -a checksums.md5 .
```

Note that `md5sum *` does not include files in subdirectories, while `hashit` audits the whole directory, so for a directory with subdirectories generate the file with something like `find . -type f -exec md5sum {} +`.

The hash is determined by the file extension where it is one of `.md4`, `.md5`, `.sha1`, `.sha256`, `.sha512`, `.b2` or `.blake2b` (Blake2b-512 or Blake2b-256), `.b3` or `.blake3`, or `.crc32`, and otherwise by the length of the hash. Where lengths are shared, such as `md4` and `md5`, the more common hash is assumed. Relative paths are resolved against both the current directory and the directory containing the sum file.

Likewise the output of `hashit --format sum` can be verified by `md5sum -c` and friends, so long as only a single hash is selected:

```shell
$ hashit --hash md5 --format sum . > ../checksums.md5
$ md5sum -c ../checksums.md5
```

With `--mtime` the modification time of each file is written as a comment line such as `# mtime 2026-01-01 10:00:00  ./file`, which `md5sum -c` ignores.

### SQLite Auditing (Recommended for Large Datasets)

For auditing very large collections of files, using a text-based `hashdeep` file can become slow and memory-intensive. To address this, `hashit` supports using a SQLite database as a scalable and efficient backend for creating and verifying audits.

This method is recommended for any large-scale or repeated auditing tasks.

#### Generating a SQLite Audit Database

To generate a SQLite database of hashes, use the `--format sqlite` flag combined with `--output`. It is recommended to use `--hash all` to store all possible hashes, which enables the most robust "paranoid" audit checks when the audit is also run with `--hash all`.

```shell
# Generate a database with all hashes for the current directory
$ hashit --format sqlite --output audit.db --hash all .
results written to audit.db
```

This creates a file named `audit.db` containing a `file_hashes` table with all the file paths and their corresponding hashes.

`hashit` will not write to an existing database, as that would mix the new results with an earlier audit. Pass `--overwrite` to replace it, which is only allowed when the existing file is a SQLite database or empty. The same applies to every other format, `--output` never replaces an existing file without `--overwrite`.

```shell
$ hashit --format sqlite --output audit.db --hash all .
ERROR ...: output file audit.db already exists, use --overwrite to replace it
$ hashit --format sqlite --output audit.db --hash all --overwrite .
results written to audit.db
```

#### Verifying with a SQLite Audit Database

To audit a directory against a SQLite database, use the `-a` or `--audit` flag, pointing it to your database file. `hashit` will automatically detect that it's a SQLite file.

```shell
# Run an audit against the database we just created
$ hashit -a audit.db .
hashit: SQLite Audit passed
       Files examined: 25
Known files expecting: 25
        Files matched: 25
       Files modified: 0
          Files moved: 0
      New files found: 0
        Files missing: 0
```

If a file is modified, the audit will fail and report the change. The audit calculates the hashes requested with `--hash`, or the defaults when not supplied, and compares each one the database also stores. To compare every stored hash use `--hash all` when auditing as well. If none of the requested hashes are in the database a warning is printed and every file is reported as modified.

```shell
# the most thorough check, calculating and comparing every hash
$ hashit --hash all -a audit.db .
```

The database is opened read-only and is never modified by an audit. Which files have been seen is tracked in a temporary table on disk rather than in memory, so memory use does not grow with the size of the database. Files on disk which are not in the database, being new or moved files, are held in memory until the end of the audit, so auditing a directory with many new files will use more memory. The directory `hashit` was run from is stored in the database, so as with `hashdeep` files, a database of relative paths can be audited from any directory using relative or absolute paths.

```shell
# Example of a failed audit after modifying a file
$ echo "new content" >> README.md
$ hashit -a audit.db .
hashit: SQLite Audit failed
       Files examined: 25
Known files expecting: 25
        Files matched: 24
       Files modified: 1
          Files moved: 0
      New files found: 0
        Files missing: 0
```

#### Verifying with your own tools (Python Example)

The SQLite database is a standard, open format, which means you can easily write your own scripts and tools to query it. Here is a simple Python example demonstrating how to connect to the database and verify the SHA256 hash of a single file. Paths are stored as `hashit` was given them, so a database generated from `.` stores `./README.md`, and the example tries both forms.

```python
import hashlib
import os
import sqlite3
import sys


def sha256_of_file(path):
    """Calculates the SHA256 of a file without reading it all into memory."""
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def verify_file_hash(db_path, file_to_check):
    """Verifies the SHA256 of a file against a hashit SQLite audit database."""
    try:
        on_disk = sha256_of_file(file_to_check)
    except OSError as e:
        print(f"Error reading file {file_to_check}: {e}")
        return False

    # Paths are stored as hashit was given them, so "README.md" may be stored as "./README.md"
    candidates = [file_to_check, os.path.join(".", file_to_check)]

    conn = None
    try:
        # open read-only so a mistyped path does not create an empty database
        conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
        placeholders = ",".join("?" * len(candidates))
        record = conn.execute(
            f"SELECT filepath, sha256 FROM file_hashes WHERE filepath IN ({placeholders})",
            candidates,
        ).fetchone()
    except sqlite3.Error as e:
        print(f"Database error: {e}")
        return False
    finally:
        if conn:
            conn.close()

    if record is None:
        print(f"File not found in audit database: {file_to_check}")
        return False

    stored_path, stored = record
    if stored is None:
        print(f"{stored_path}: no SHA256 stored, generate the database with --hash sha256 or --hash all")
        return False

    ok = on_disk == stored
    print(f"{stored_path}: {'OK' if ok else 'MISMATCH'}")
    return ok


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python verify.py <path_to_audit.db> <path_to_file>")
        sys.exit(2)

    sys.exit(0 if verify_file_hash(sys.argv[1], sys.argv[2]) else 1)
```

#### Key Differences from `hashdeep`

While `hashit` aims for compatibility, there are some minor differences in the command-line interface:

*   **Audit Flag:** `hashdeep` uses two flags to start an audit (`-a -k <file>`), whereas `hashit` uses a single flag that takes the audit file as its argument (`-a <file>`).
*   **Displaying Failed Hashes:** `hashdeep` has an `-X` flag to display the new hashes of modified files. `hashit` does not currently have an equivalent for this feature.
*   **Modified Files:** `hashdeep` matches on hashes alone, so a modified file is reported as both a new file and a known file not found. `hashit` matches on path first and reports it as modified.

Note that you don't have to specify the directory you want to run against. Running `hashit` will assume you want to run against the current directory, unless data is piped into it in which case it hashes standard input.

If you supply a single argument to `hashit` and its a file it will process it. If you supply a single argument and it is a directory it will recurse that directory.

If you supply multiple arguments which consist of files and directories then directories will be skipped without a warning, unless `-r` is supplied in which case they are recursed. This allows wild card's to be used and just process the files in a directory.

```
$ hashit *
LICENSE (1067 bytes)
        MD5 227f999ca03b135a1b4d69bde84afb16
       SHA1 2c7beb1563981414f31481016816a88dc0cdba16
     SHA256 fb3f44f5e74b957107f89b027896250ecff74718b1fa8bf0566874e142e54351
     SHA512 882eb7ee56cda80778dbd44f514e68a14126a61a11a825f9c030a3425b941d6a3636561a466347178d8bf204dbedc806471d248637091b311ca37ec18a1c084b

go.mod (1247 bytes)
        MD5 da13de2c3aef119dacf5919201da1be2
       SHA1 c55eab504d96eeea9fd98b602cb632b74c769b44
     SHA256 3ca5dc09cbbed95b31fd30d32b15f773f68b44570841d0751f942f5d44032edb
     SHA512 305603841baea29bd5ca87f2d1ecf80c669247898095b512dde07634311d4ee3f4a91e1201261cb8450aae1c0d01596b4796ef98fb16a069c12f3a0a6fd30cfd

... OUTPUT SNIPPED ...
```

If you are running hashit on a slower mechanical HDD you may want to limit the number of threads which read files from
disk using `--threads 1`

```shell
$ hashit --threads 1 /mnt/slowdisk/
```

For large files you can use `-p` to see the progress of the file to get an idea of how long it might take to process.

```shell
$ hashit -p --threads 1 large.file
[==================================>---------------------------------] file: large.file
```


#### Misc stuff below

Usage of hashdeep

https://linhost.info/2010/05/using-hashdeep-to-ensure-data-integrity/

Issues to address

https://github.com/jessek/hashdeep/issues/4
https://github.com/jessek/hashdeep/issues/358

Example usage

hashdeep -r vendor > audit.txt
hashdeep -r -a -k audit.txt vendor
