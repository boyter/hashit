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

If you are comfortable using Go and have >= 1.25 installed the usual `go get -u github.com/boyter/hashit/` will install for you.

Binaries will be on [Releases](https://github.com/boyter/hashit/releases) for Windows, GNU/Linux and macOS for both i386 and x86_64 bit machines once it hits version 1.0.0.

If you would like to assist with getting `hashit` added into apt/homebrew/chocolatey/etc... please submit a PR or at least raise an issue with instructions.

### Development

You need to have Go installed. Minimum version is Go 1.25 https://go.dev/

Install the following tools, either via the indicated command or what is suggested on site

- sqlc brew install sqlc https://sqlc.dev/

sqlc is used for the audit functionality as well as the SQLite output format. If you never change this functionality
it may not be required, however _never_ edit the `./processor/db/` files directly.

### Pitch

Why use `hashit`?

 - It is very fast
 - You can get multiple hashes "for free" on any CPU with multiple cores
 - Works very well across multiple platforms without slowdown (Windows, Linux, macOS)
 - Supports many hashes `hashit --hashes` CRC32, xxHash64, MD4, MD5, SHA1, SHA256, SHA512, Blake2b-256, Blake2b-512, Blake3, SHA3-224, SHA3-256, SHA3-384, SHA3-512
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
  -a, --audit string            audit against supplied file; audit file must be in hashdeep output format
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
  -i, --input string            input file of newline seperated file locations to process
      --mtime                   enable mtime output
      --no-stream               do not stream out results as processed
  -M, --not-match stringArray   ignore files and directories matching regular expression
  -o, --output string           output filename (default stdout)
  -p, --progress                display progress of files as they are processed
  -r, --recursive               recursive subdirectories are traversed
      --stream-size int         min size of file in bytes where stream processing starts (default 1000000)
      --threads int             number of threads processing files, by default the number of CPU cores (default 8)
      --trace                   enable trace output
  -v, --verbose                 verbose output
      --version                 version for hashit
      --vv                      very verbose output
```

Output should look something like the below for operations on this repository

```
$ hashit README.md
README.md (6616 bytes)
        MD5 17cad37b7b873eed74e15cafd7855f6d
       SHA1 8e4256ef60302acf72e9c25bea5a41e195cf14ec
     SHA256 5ff9cd136f827570acc98be42d5a76a9409c83e3b9300ff2466de48c71760ca0
     SHA512 e4d108219def5c36089f7cd45d4531548d52a41cc5992004c78e816fbe64d63a21bc7ea1303d7a31bd693bf4f5435c916fbbe4e9d3e1fd0b1982a9734b4ec739

$ hashit --hash all README.md
README.md (9460 bytes)
      CRC32 8f023940
   xxHash64 8bc915722510645e
        MD4 179e0501d57b741e822964a85a4f1923
        MD5 2a8517703f0ddf7be45d0d51b84f1e49
       SHA1 886521f2d1f8e4b461a364d3cad826636831eff5
     SHA256 cb664826dd6982d3d0350f4453b36bd0168a996a20b8e0e719f0e224cf761808
     SHA512 4894b771b8b7412a12ecb8bd6e9b2a5d1f7c540d0c293fddd6231fec18c13af7f1333f9cf621c1ce7677a363947f28a3ab706868e3ca2e93f2597b798e7a25d1
Blake2b-256 5c8dd15df397217d742ca94ebd3f4d6b7c4bf74fc94b61e1ad14bd0b41882960
Blake2b-512 c4e22434524549d882a475076092586e7543bbba5ff401761cecab8b016c9b877035e4b6a3c59ba8d7e95147a1f13bd54aa0ec300f833ddf4a87e7bd09982efc
     Blake3 2b3e1660f329f570182e8f448e516d1ce545984e5a2d00046cd52689847e2894
   SHA3-224 9bceeaeed4d1ead3af3098429cbe3f1a729b40feebf3b9d576ca1210
   SHA3-256 ee226be159cbaee0a49d94e3a2a2bb9fa75d59a039e9149ae16f419a2d41e84d
   SHA3-384 f9a1c8860b40fa58bb2f113bbae3ab7bfe25c8f5bfa4312dc4e09c75f678faf8005f7cbeac8f01f5adbd00ea480e694d
   SHA3-512 d7e5d8b4abd2bdf2c9628e1698e08577c30a4626f303d3e3e0a8d3385068418adace23460c6303d2537ee35dd6963cb5c9a04fdee5bf6e307e2240835872775f
       ed2k 179e0501d57b741e822964a85a4f1923

$ hashit scripts
scripts/include.go (1835 bytes)
        MD5 73d7180f48af0b44e4ca7ae650335ac1
       SHA1 dcd8b23288c1604c7b541c7626ae569bec9001b6
     SHA256 929750fcace21c4a261d19824f691eae8691958989e8c1870b0934cfa9493462
     SHA512 b37ac5a309f9006b740fb0933fe5c4569923cab0fe822c1e2fbf0fbd2a15e9787681ec509ca9f7ea13d921a82257ecc3a32e2dfa18cc6892ea82978befe2629c
```

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

# Now, audit with hashdeep
$ hashdeep -r -a -k audit.txt processor
hashdeep: Audit passed
          Files matched: 8
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
       Files examined: 8
Known files expecting: 8
        Files matched: 8
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

`hashit` can also detect moved files, distinguishing them from new or missing files.

#### Key Differences from `hashdeep`

While `hashit` aims for compatibility, there are some minor differences in the command-line interface:

*   **Audit Flag:** `hashdeep` uses two flags to start an audit (`-a -k <file>`), whereas `hashit` uses a single flag that takes the audit file as its argument (`-a <file>`).
*   **Displaying Failed Hashes:** `hashdeep` has an `-X` flag to display the new hashes of modified files. `hashit` does not currently have an equivalent for this feature.

Note that you don't have to specify the directory you want to run against. Running `hashit` will assume you want to run against the current directory.

If you supply a single argument to `hashit` and its a file it will process it. If you supply a single argument and it is a directory it will recurse that directory.

If you supply multiple arguments which consist of files and directories then directories will be skipped. This allows wild card's to be used and just process the files in a directory.

```
$ hashit *
Gopkg.toml (736 bytes)
        MD5 3e88135ebf43e8199ce0c19c8bebd925
       SHA1 c3324f86a84c7a3ee941fcf94d221877aa25e799
     SHA256 712667bfe27e245e0f34b2d64ad6b0d21580c058a01ce8374e5b6fa408a97f66
     SHA512 d897a428487e31143b1b1ba9b3fa237bc9ed3aac7ad15ecd9507b7f58eaf5f1059385ebfcbeed4505e2cbf63113f70a0252acbf9f738f25fc79e563e7c8b32b6

Gopkg.lock (1514 bytes)
        MD5 d19eb48da41406376a361a1c4a97e496
       SHA1 977b14e5ed765137c0e146789ed37c52a8cd51ee
     SHA256 5b76da627bd28a7bc7459b3281e3582616b21c4694302ae00b8888689effe35c
     SHA512 95b9450cf19a3ef5c287841b3606ac7a1ad3066308a638fb8be908d649360a656ae41291e657e6f4250c87b0e18e05d73fca111fede9ed496eb600afdb245a0d

README.md (7094 bytes)
        MD5 df97688a1cb29d4b06d630df034b6e16
       SHA1 d1c94c30c97f8fac75fe485e8352da813c4d4bb7
     SHA256 78ce6718977135593fbd9867d6c2a68f91aa6d2846fe7bf08cbe9924790517b9
     SHA512 759a3ef175466639fdc5bfbd1a8a47087410148405b15a3ac4bc0c88e677207f7c50f62cee947dabbbcf739c1031f3ff1ecdfd1688d83f566c75068b11a1f680

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
