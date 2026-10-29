Hash It!
--------

A hash tool which can work like hashdeep or md5sum, sha1sum, etc... When you want to find the hash or hashes of a file quickly, cross platform using a single command.

Yes the name is *very intentional* and similar to what say when I realise I need to get multiple hashes of a large file. This tool attempts to solve that pain.


[![Go](https://github.com/boyter/hashit/actions/workflows/go.yml/badge.svg)](https://github.com/boyter/hashit/actions/workflows/go.yml)
[![Scc Count Badge](https://sloc.xyz/github/boyter/hashit/)](https://github.com/boyter/hashit/)

Licensed under the MIT license.

Other similar projects,

 - [hashdeep](https://github.com/jessek/hashdeep) originally called md5deep

### Support

Using `hashit` commercially? If you want priority support for `hashit` you can purchase a years worth https://app.gumroad.com/products/mcpni which entitles you to priority direct email support from the developer.
 
### Install

If you are comfortable using Go and have >= 1.25 installed `go install github.com/boyter/hashit@latest` will install the latest tagged release for you. Features described here which are newer than that release need `@master`.

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
 - Builds on illumos and Solaris, without the `sqlite` format, auditing SQLite databases or `--progress`, as the libraries they use do not support those platforms
 - Supports many hashes `hashit --hashes` CRC32, xxHash64, MD4, MD5, SHA1, SHA224, SHA256, SHA384, SHA512, Blake2b-256, Blake2b-512, Blake3, SHA3-224, SHA3-256, SHA3-384, SHA3-512, ed2k. Select them with `--hash` using the lowercase names shown in brackets by `hashit --hashes`, such as `sha3256` or `blake2b256`, or `--hash all`. The `ed2k` hash of a file which is an exact multiple of 9,728,000 bytes uses the variant without an extra empty chunk, so differs from tools such as `rhash` for those sizes
 - Output is compatible with `hashdeep`, and audits can be verified by either tool (see [Absolute Paths](#absolute-paths))

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
      --exclude-dir strings     directories to exclude when walking directories, by name such as node_modules or the end of a path such as vendor/cache
  -f, --format string           set output format [text, json, sum, hashdeep, hashonly, sqlite] (default "text")
      --gitignore               enable .gitignore file logic
      --gitmodule               enable .gitmodules file logic
  -c, --hash strings            hashes to be run for each file (set to 'all' for all possible hashes); hashdeep and sum audits use the hashes in the audit file (default [md5,sha1,sha256,sha512])
      --hashes                  list all supported hashes
      --hashignore              enable .hashignore file logic
  -h, --help                    help for hashit
      --ignore                  enable .ignore file logic
  -i, --input string            input file of newline separated file locations to process
      --mtime                   enable mtime output
      --no-stream               do not print results as files are processed for text, sum and hashonly formats; other formats and --output never stream
  -M, --not-match stringArray   ignore files and directories whose name matches regular expression when walking directories
  -o, --output string           output filename (default stdout, or hashit.db for the sqlite format)
      --overwrite               allow output to replace an existing regular file; a symlink which cannot be followed is always refused
  -p, --progress                display progress of files as they are processed
  -r, --recursive               traverse directories when multiple paths are supplied; a single directory or no path is always traversed
      --skip-hidden             skip hidden files and directories when walking directories
      --stream-size int         files larger than this size in bytes are read in chunks rather than all at once (default 1000000)
      --threads int             number of threads processing files, by default the number of CPU cores (default 8)
      --trace                   enable trace output
  -v, --verbose count           verbose audit output, listing every modified, moved, new and missing file; -vv is the same as --vv
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
     SHA224 434524e15237261766cb0eaca93ae3692211001e641c6a4b8c428ab5
     SHA256 fb3f44f5e74b957107f89b027896250ecff74718b1fa8bf0566874e142e54351
     SHA384 2521569e2a92464bd7327682ebce99014809a1fed6f71ffbc47cee55e76076ce3d24a41d0e413768fc9cf99990dee779
     SHA512 882eb7ee56cda80778dbd44f514e68a14126a61a11a825f9c030a3425b941d6a3636561a466347178d8bf204dbedc806471d248637091b311ca37ec18a1c084b
Blake2b-256 824c7551cee5a6996c422988f658d2622890fa52e2fcf63d0aa6c01abc69e77d
Blake2b-512 ad692390b5e680f24490d5f4676f279f84e802c79b98ef36369404b88d8ced237dbde429098f6e7bb221b6ae8fff4ab1920b03fd187c3025c4f81be2a3def672
     Blake3 339416b7c6155ddd00a7da8c6fbb40b472fbc113c7ca532f1929cffffa81f892
   SHA3-224 845411d717203ef1896ab41ff4b2d468958b2e067807713bd49867ed
   SHA3-256 e334df710e2791c6df492edd61d1d17022be91044937e29d8613ddfb43f4aec6
   SHA3-384 8301eb66f20215a14a77aa6e977a9d08524caf31fa09a9fd97e7b3c4c3e9266159d9c35c661960024654ff36f5131302
   SHA3-512 1024928532bccfe9feb0b3f9fe54eda6fb05685fb3236eb09cb3765f0d1face3ea5f5dc681aaeec44134cd89691348a03cc6a2abfe1eae55d8eb274a51f97e3d
       ed2k a4961e3d704e28a40debbf1f1e4a62cd

$ mkdir -p example/sub && echo hello > example/a.txt && echo world > example/sub/b.txt
$ hashit --hash md5 example
example/a.txt (6 bytes)
        MD5 b1946ac92492d2347c6235b4d2611184

example/sub/b.txt (6 bytes)
        MD5 591785b794601e212b260e25925636fd
```

Files are processed in parallel, so the order of the output may vary between runs, even with `--threads 1`. For a stable order pipe `sum` output through `sort`.

#### Paths

- No path processes the current directory, unless data is piped or redirected into `hashit` from a file, in which case it hashes standard input. A path of `-` is not standard input. A single directory is always recursed.
- When multiple paths are supplied, directories are skipped with a warning unless `-r` is used. Audits always recurse.
- Arguments which match an existing path are used as is, otherwise they are expanded as a wildcard, as shells such as `cmd.exe` do not do this. So `hashit '*.go'` works on any platform. Unlike most shells this expansion also matches hidden files. A wildcard which matches several paths counts as several paths, so directories it matches are only recursed with `-r`.
- `--input` reads the paths to process from a file, one per line, which are used exactly as written. Wildcards are not expanded and directories are an error even with `-r`. Empty lines and Windows line endings are allowed, and paths supplied as arguments are ignored with a warning.
- Symbolic links to files and directories are followed as `hashdeep` does, skipping any which would create a cycle. Symbolic links which cannot be followed, such as broken links or loops, FIFOs, sockets and devices found while walking a directory are skipped with a warning, while one supplied directly or in an `--input` file is an error.
- Hidden files are included unless `--skip-hidden` is used. `.gitignore`, `.ignore` and `.gitmodules` files are not respected unless `--gitignore`, `--ignore` or `--gitmodule` are used. `--exclude-dir` takes a directory name such as `node_modules`, or the end of a path such as `vendor/cache`, matched at any depth. `--not-match` (`-M`) takes a regular expression matched against file and directory names and can be repeated. These only apply when walking directories, not to files supplied directly. A symbolic link to a directory is checked against them by its own name, but is then walked on its own, so ignore file rules from directories above it do not apply inside it.
- Hash names for `--hash` are not case sensitive, so `-c MD5,SHA256` works.

#### Output Formats

`--format` (`-f`) selects the output format. Results are written to standard output, or to a file with `--output` (`-o`). The output file is written to a temporary file next to it first, so the directory must be writable, being the directory of the file a symbolic link points to when `--output` is one, and an existing file is only replaced with `--overwrite`, is never replaced when it is a directory or other special file, and is kept rather than replaced when any error occurred. A new output file is still written when errors occurred, so it may not contain every file. A replaced file keeps its permission bits, but not its owner, special bits such as setuid, or extended attributes. When `--output` is a symbolic link the file it points to is replaced, and a symbolic link which cannot be followed, such as a loop or a link to a file which does not exist, is refused. The replacement is a new file, so other hard links to the old file keep the old content. If `hashit` is interrupted by `SIGINT`, `SIGTERM`, `SIGHUP`, `SIGQUIT` or `SIGPIPE` the temporary file is removed and it exits with 128 plus the signal number, such as `130` for `SIGINT`. `SIGKILL` and other signals such as `SIGUSR1` are not caught, so may leave the temporary file behind. The message `results written to` is printed to standard output. The output file, the audit file, and paths reaching them through a symbolic link, either a link to the file or a symbolically linked directory, are never included in the results. Another hard link to the output file is ordinary data and is included. Supplying one of them explicitly, such as with a wildcard like `hashit -a sums.md5 *`, skips it with a `WARN`. When nothing else was hashed an existing output file is kept rather than replaced with empty results, so `hashit -o doc.txt --overwrite doc.txt` is an error.

- `text`, the default, shown above. `--mtime` adds an `MTime` line in local time.
- `json`, an array of objects such as `[{"File":"example/a.txt","MD5":"b1946ac92492d2347c6235b4d2611184","Bytes":6}]`. The keys are `File`, `CRC32`, `XxHash64`, `MD4`, `MD5`, `SHA1`, `SHA224`, `SHA256`, `SHA384`, `SHA512`, `Blake2b256`, `Blake2b512`, `Blake3`, `Sha3224`, `Sha3256`, `Sha3384`, `Sha3512`, `ed2k` and `Bytes`. Hashes which were not calculated are omitted. `--mtime` adds `MTime` in RFC 3339 format with nanoseconds and the local time zone.
- `sum`, the same as `md5sum` and similar tools, with one line per hash per file when several hashes are selected. `--mtime` adds a `# mtime` comment line, see [Auditing with sum Files](#auditing-with-sum-files).
- `hashdeep`, see [Auditing](#auditing). `--mtime` adds a `## mtime` comment line, see [Modification Times](#modification-times).
- `hashonly`, one hash per line. With several hashes or files there is nothing to say which hash belongs to which file, so it is best used with a single hash and file. `--mtime` adds the time on its own line.
- `sqlite`, a database written to `--output`, or `hashit.db` when not supplied, see [SQLite Auditing](#sqlite-auditing-recommended-for-large-datasets).

Files larger than `--stream-size` bytes are read in chunks rather than all at once. `text`, `sum` and `hashonly` output to standard output, rather than `--output`, is printed as each file is processed unless `--no-stream` is used. `--debug`, `--trace` and `--progress` output is written to standard error.

#### Errors and Exit Codes

If any file cannot be read, such as a path which does not exist or a file which lacks permissions, an error is printed, the remaining files are still processed, and `hashit` exits with a non zero code. Messages where processing continues normally, such as a skipped directory, are labelled `WARN`.

`hashit` exits with one of the following codes, so scripts can tell a failed audit apart from `hashit` being unable to do its job:

- `0` everything was processed, and any audit passed.
- `1` everything was processed, but the audit failed.
- `2` an error was printed, such as an invalid option, a path which does not exist, a file which could not be read, or output which could not be written. This includes an audit which could not check every file.

When interrupted by a signal `hashit` exits with 128 plus the signal number, such as `130` for `SIGINT` and `141` for `SIGPIPE`, which is also the code when the program reading its output exits early, such as `hashit . | head`.

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

In this case, `hashit` would process `main.go` (and the `.hashignore` file itself) but would skip `data.log` and the entire `node_modules` directory. A `.hashignore` file in a subdirectory applies to that subdirectory, and patterns such as `!keep.log` and `/build` work as they do in `.gitignore`. A symbolically linked directory is followed but is matched as a file, so a directory pattern such as `node_modules/` does not exclude it, use `node_modules` or `--exclude-dir` instead.

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

`hashit` can also detect moved files, distinguishing them from new or missing files. As with `hashdeep`, an audit only passes when every file is exactly where it was, so moved files cause the audit to fail. To check that every known file is still present and unmodified, while allowing files to have been moved or renamed, use `--allow-moved`. New, missing or modified files still fail the audit. Files are matched by path first, so a file renamed onto the path of another known file, such as two files swapping names, is reported as modified rather than moved.

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

Use `-v` (`--verbose`) to list every file which is not matched, and `-vv` (`--vv`) to also list matched files. Results are reported in the order of the audit file, followed by any new files, before the summary:

```
./a.txt: File modified
./b.txt -> ./archive/b.txt: File moved
./c.txt: File expected but not found
./d.txt: File could not be read
./e.txt: File new
./f.txt: Ok
```

Some other points worth knowing:

- Only `hashdeep`, `sum` and `sqlite` files can be audited against. A file with no lines in a supported format, such as `text` or `json` output, is an error. `--format` is ignored when auditing, and `--hash` is ignored with a warning for `hashdeep` and `sum` files, which use the hashes they contain. `--output` writes the summary to a file.
- The audit compares the files found in the supplied paths against everything in the audit file, so auditing a subset of the files reports the rest as missing.
- A copy of a known file, where the original is still present, is reported as new. `hashdeep` reports it as moved. Both fail the audit.
- A file which cannot be read, including files in a directory which cannot be read, is reported as could not be read in `-v` output, counted as missing when it is in the audit file and not counted otherwise, and `hashit` exits with `2` rather than `1`.
- The same file supplied more than once by the same absolute path, such as through `.` and `./sub`, is only counted once. The same file reached through a symbolic link is a different path, so is counted again.
- Modification times are recorded with `--mtime` but are not compared by an audit.
- The audit file of an empty directory passes when audited against an empty directory, as does an audit file which is empty or only contains comments, which `md5sum -c` treats as an error. `hashdeep` reports an error for an audit file with no files in it.
- An audit file written by a shell redirect into the directory being hashed, such as `hashit --format hashdeep . > manifest.txt`, lists itself as an empty file. That entry is ignored when auditing, but `--output manifest.txt` avoids it entirely and is recommended, especially for the `sum` format which writes as it goes and so may record part of itself, which then fails when audited from another directory.
- Concatenated `hashdeep` files are read, but only the first `## Invoked from:` line is used, so files concatenated from different directories may report moved files. `hashdeep` also fails to audit these.

#### Modification Times

With `--mtime` the modification time of each file is written as a comment line after it, such as `## mtime 2026-01-01 10:00:00`, in local time. `hashdeep` ignores comment lines, so the file remains compatible with it, and filenames containing commas are not ambiguous.

#### Absolute Paths

`hashdeep` outputs absolute paths unless run with `-l`, whereas `hashit` outputs paths as supplied, which is the same as `hashdeep -l`. Relative paths make a manifest which can be shipped with the files and verified elsewhere. To match `hashdeep`'s default use `--absolute`.

```shell
$ hashit --absolute --format hashdeep . > ../audit.txt
$ hashdeep -r -a -k ../audit.txt .
hashdeep: Audit passed
```

#### Paths in Audit Files

Relative paths in a `hashdeep` file are resolved against the directory recorded in its `## Invoked from:` line, so an audit file created with `hashit --format hashdeep .` can be audited from anywhere using either a relative or absolute path. When the invoked from directory does not contain the files, such as after they have been copied elsewhere, paths are also resolved against the directory containing the audit file, being where the file really is and then where the path supplied for it is when that is a symbolic link, and then the current directory. Relative paths in `sum` files have no invoked from directory, so are resolved against the current directory first. Matching is case sensitive on every platform. Paths are always compared as absolute paths, so the path to audit may be supplied as relative or absolute, and `./file` and `file` are the same file. So a manifest stored at the root of a set of files, created from that directory, can be audited from anywhere after the files are copied or shipped elsewhere. Symbolic links in the path to audit are not resolved, so a manifest created through `real/` and audited through a link to it reports every file as moved, as `hashdeep` does. When two files on disk match the same record, such as when the path to audit overlaps a copy of the files, the one matching through the invoked from directory is used and the other is reported as new.

If the audit file is inside the directory being audited it is skipped, as it can never contain its own hash. `hashdeep` and `md5sum` do not do this, so when generating a file for them to verify, write it outside the directory being hashed, otherwise they will report the partially written file as changed.

#### Auditing with `sum` Files

`hashit` can also audit against the output of `md5sum`, `sha1sum`, `sha224sum`, `sha256sum`, `sha384sum`, `sha512sum`, `b2sum` and similar tools, including the BSD style output of `--tag`, as well as files generated by `hashit --format sum`. Each line must be a hash followed by its path, so SFV files, which put the path first, are not supported. Output from `hashit --format sum` contains no hash names, so when it uses a hash other than `md5`, `sha1`, `sha224`, `sha256`, `sha384` or `sha512`, give the file the matching extension listed below. Note the extensions differ from the names used with `--hash`, such as `.sha3-256` for `--hash sha3256`.

```shell
$ md5sum * > checksums.md5
$ hashit -a checksums.md5 .
```

Note that `md5sum *` does not include files in subdirectories, while `hashit` audits the whole directory, so for a directory with subdirectories generate the file with something like `find . -type f -exec md5sum {} +`.

The hash is determined by the file extension where it is one of `.md4`, `.md5`, `.sha1`, `.sha224`, `.sha256`, `.sha384`, `.sha512`, `.sha3-224`, `.sha3-256`, `.sha3-384`, `.sha3-512`, `.b2` or `.blake2b` (Blake2b-512 or Blake2b-256), `.b3` or `.blake3`, `.ed2k`, or `.crc32`, and otherwise by the length of the hash. A hash of 8 characters is assumed to be `crc32` and one of 16 `xxhash64`. Where lengths are shared the more common hash is assumed, being `md5` over `md4` and `ed2k`, `sha224` over `sha3-224`, `sha256` over `sha3-256`, `blake3` and `b2sum -l 256`, `sha384` over `sha3-384`, and `sha512` over `sha3-512` and `b2sum`. A wrong guess reports every file as modified, so name `b2sum` files `.b2` or use `--tag` output for the others. Relative paths are resolved against both the current directory and the directory containing the sum file, so run the audit from the directory the sum file was created in, or keep the sum file at the root of the files it lists.

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

To generate a SQLite database of hashes, use the `--format sqlite` flag combined with `--output`. Without `--output` the database is written to `hashit.db`. It is recommended to use `--hash all` to store all possible hashes, which enables the most robust "paranoid" audit checks when the audit is also run with `--hash all`.

```shell
# Generate a database with all hashes for the current directory
$ hashit --format sqlite --output audit.db --hash all .
results written to audit.db
```

This creates a file named `audit.db` containing a `file_hashes` table with each file's path, hashes and size, and a `metadata` table recording the directory `hashit` was run from and the hashes calculated. With `--mtime` the modification time is also stored in RFC 3339 format, otherwise it is `NULL`. Audits do not compare it. The database file is not included in its own results when it is inside the directory being hashed.

`hashit` will not write to an existing database, as that would mix the new results with an earlier audit. Pass `--overwrite` to replace it, which is only allowed when the existing file is a SQLite database or an empty file. The new database is written to a temporary file and only replaces the existing one once finished without errors, so a mistyped path does not destroy an earlier audit. When there is no existing database one is written even when errors occurred, and `hashit` exits with `2`, so it may not contain every file. Every other format also refuses to replace an existing file without `--overwrite`, though with it any regular file is replaced. A directory is never replaced.

```shell
$ hashit --format sqlite --output audit.db --hash all .
ERROR ...: output file audit.db already exists, use --overwrite to replace it
$ hashit --format sqlite --output audit.db --hash all --overwrite .
results written to audit.db
```

#### Verifying with a SQLite Audit Database

To audit a directory against a SQLite database, use the `-a` or `--audit` flag, pointing it to your database file. `hashit` will automatically detect that it's a SQLite file. Other audit files can be read from a pipe such as `/dev/stdin`, but a SQLite database must be a regular file.

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

If a file is modified, the audit will fail and report the change, and `-v` lists which files changed. The audit calculates the hashes requested with `--hash`, or the defaults of `md5`, `sha1`, `sha256` and `sha512` when not supplied, and compares each one the database also stores. To compare every stored hash use `--hash all` when auditing as well. If none of the requested hashes are in the database a warning is printed and every file is reported as modified.

```shell
# the most thorough check, calculating and comparing every hash
$ hashit --hash all -a audit.db .
```

The database is opened read-only and is never modified by an audit. If another tool has switched the database to WAL mode, SQLite creates `-wal` and `-shm` files next to it, and the audit fails when the directory is read only, so run `sqlite3 audit.db 'pragma journal_mode=delete'` first. Which files have been seen is tracked in a temporary table, which SQLite moves to a file once it outgrows its cache, so memory use does not grow with the size of the database. The table is created in SQLite's temporary directory, which can be changed by setting `SQLITE_TMPDIR` for very large audits. Files on disk which are not in the database, being new or moved files, are held in memory until the end of the audit, so auditing a directory with many new files will use more memory. The directory `hashit` was run from is stored in the database, and paths are matched the same way as `hashdeep` files, see [Paths in Audit Files](#paths-in-audit-files), so a database of relative paths can be audited from any directory using relative or absolute paths, and a database stored at the root of a set of files can be audited after they are copied elsewhere. To do this the absolute path of every record is first indexed in a temporary table.

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

As with other audits, auditing a subset of the files in the database reports the rest as missing, and `--allow-moved` passes audits where files have only moved. A database created by a development build of `hashit` before SHA-224 and SHA-384 were added is rejected with a message to regenerate it.

#### Verifying with your own tools (Python Example)

The SQLite database is a standard, open format, which means you can easily write your own scripts and tools to query it. Here is a simple Python example demonstrating how to connect to the database and verify the SHA256 hash of a single file. Paths are stored relative to how `hashit` was run, so `hashit .` stores `./README.md` while `hashit` with no arguments stores `README.md`, and the example tries the path with and without a leading `./`. Unlike `hashit -a` it does not use the stored directory `hashit` was run from, so run it from that directory.

```python
import hashlib
import os
import pathlib
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

    # hashit stores paths relative to how it was run, so "README.md" may be stored as "./README.md"
    path = os.path.normpath(file_to_check)
    candidates = [path, os.path.join(".", path)]

    conn = None
    try:
        # open read-only so a mistyped path does not create an empty database
        uri = pathlib.Path(db_path).resolve().as_uri() + "?mode=ro"
        conn = sqlite3.connect(uri, uri=True)
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
*   **Matching Modes:** `hashdeep` has matching modes, `-m` and `-x` to list files which do or do not match known hashes, and `-M` and `-X` to also display their hashes. `hashit` does not currently have an equivalent, and `-M` in `hashit` is `--not-match`, which excludes files by name.
*   **Modified Files:** `hashdeep` matches on hashes alone, so a modified file is reported as both a new file and a known file not found. `hashit` matches on path first and reports it as modified.
*   **Paths:** `hashit` outputs relative paths by default, the same as `hashdeep -l`, see [Absolute Paths](#absolute-paths). A single directory is recursed without needing `-r`.
*   **Audit File:** `hashit` skips the audit file when it is inside the directory being audited, and resolves relative paths against the directory containing it, see [Paths in Audit Files](#paths-in-audit-files).

Note that you don't have to specify the directory you want to run against. Running `hashit` will assume you want to run against the current directory, unless data is piped or redirected into it from a file, in which case it hashes standard input. Audits and `--input` never read standard input. In scripts and scheduled jobs supply the directory, for example `hashit .`, to avoid this.

If you supply a single argument to `hashit` and it's a file it will process it. If you supply a single argument and it is a directory it will recurse that directory.

If you supply multiple arguments which consist of files and directories then directories will be skipped with a warning, unless `-r` is supplied in which case they are recursed. When auditing directories are always recursed. This allows wildcards to be used and just process the files in a directory.

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
