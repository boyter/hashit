Hash It!
--------

A hash tool which can work like hashdeep or md5sum, sha1sum, etc... When you want to find the hash or hashes of a file quickly, cross platform using a single command.

Yes the name is *very intentional* and similar to what say when I realise I need to get multiple hashes of a large file. This tool attempts to solve that pain.


[![Go](https://github.com/boyter/hashit/actions/workflows/go.yml/badge.svg)](https://github.com/boyter/hashit/actions/workflows/go.yml)
[![Scc Count Badge](https://sloc.xyz/github/boyter/hashit/)](https://github.com/boyter/hashit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/hashit)](https://goreportcard.com/report/github.com/boyter/hashit)

Dual-licensed under MIT or the [UNLICENSE](http://unlicense.org).

Other similar projects,

 - [hashdeep](https://github.com/jessek/hashdeep) originally called md5deep

### Support

Using `hashit` commercially? If you want priority support for `hashit` you can purchase a years worth https://app.gumroad.com/products/mcpni which entitles you to priority direct email support from the developer.
 
### Install

If you are comfortable using Go and have >= 1.16 installed the usual `go get -u github.com/boyter/hashit/` will install for you.

Binaries will be on [Releases](https://github.com/boyter/hashit/releases) for Windows, GNU/Linux and macOS for both i386 and x86_64 bit machines once it hits version 1.0.0.

If you would like to assist with getting `hashit` added into apt/homebrew/chocolatey/etc... please submit a PR or at least raise an issue with instructions.

### Development

You need to have Go installed. Minimum version is Go 1.24 https://go.dev/

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
$ hashit -h
Hash It!
Version 1.2.0
Ben Boyter <ben@boyter.org>

Usage:
  hashit [flags]

Flags:
      --debug             enable debug output
  -f, --format string     set output format [text, json, sum, hashdeep, hashonly] (default "text")
  -c, --hash strings      hashes to be run for each file (set to 'all' for all possible hashes) (default [md5,sha1,sha256,sha512])
      --hashes            list all supported hashes
  -h, --help              help for hashit
      --no-stream         do not stream out results as processed
  -o, --output string     output filename (default stdout)
  -p, --progress          display progress of files as they are processed
  -r, --recursive         recursive subdirectories are traversed
      --stream-size int   min size of file in bytes where stream processing starts (default 1000000)
      --threads int       number of threads processing files, by default the number of CPU cores (default 8)
      --trace             enable trace output
  -v, --verbose           verbose output
      --version           version for hashit
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
README.md (8964 bytes)
      CRC32 a6f59c84
   xxHash64 d3b87287c02996de
        MD4 f106887bd63a7b7c5269a039c501b6b6
        MD5 22ed507c2a5fc7b822c8080340d799ab
       SHA1 2e74fa58da1f19b66b5826cf73d0c28a6c48aed6
     SHA256 c84475732426797adfc3b1f14717d1aeeee976ca19ddd596f2ea5e59c932a062
     SHA512 4f77e1d43cfcdb04c1ad99b5d53087b960ea7301da21f76de42cfefe77eb6f0bb3d2781db6c99f3f99d161e76c68d5a74a990b81e4e908a7104aef7d512efd31
Blake2b-256 59627f6afdb0bcd783a9fe1f77763ff87c5768135e839fd9f16b440828f2a8e6
Blake2b-512 9d00cd557c20b15ee85fdcc292d53604f4783d79dde6d3aada67391423633f28a1e5279519260b85171558bafc4b3785adc1854b88b004e402e49962426d86b4
     Blake3 dcef77d73bc12c8d931b253835eae13854d18b8b16c784c089929815592a4506
   SHA3-224 3912468ee2c0dfe3a4bf5f2bf4c2c36b684a465a73445d9651cf5e4d
   SHA3-256 ca7bcbd1ad389727a2502cfd5b785968ebff7453551eb44ffd1e879f47fc00dc
   SHA3-384 b97839411747e2174ed9d4dbf415e7945160746962f3208a4fd05ded8013c569da66db2fa8e8001c438896e160a153c7
   SHA3-512 be6ac69cbb38fd2dc2646e7ec3e04878c0bc8824fe196da63ac577fae35a026251ef59618e9584bc2fbb274424a6a5251aea0ab64148469eb1b704a6c836627e


$ hashit scripts
scripts/include.go (1835 bytes)
        MD5 73d7180f48af0b44e4ca7ae650335ac1
       SHA1 dcd8b23288c1604c7b541c7626ae569bec9001b6
     SHA256 929750fcace21c4a261d19824f691eae8691958989e8c1870b0934cfa9493462
     SHA512 b37ac5a309f9006b740fb0933fe5c4569923cab0fe822c1e2fbf0fbd2a15e9787681ec509ca9f7ea13d921a82257ecc3a32e2dfa18cc6892ea82978befe2629c
```

hashit can produce `hashdeep` compatible audit files and its ability to do the audit is coming,

```
$ hashit --format hashdeep processor
%%%% HASHDEEP-1.0
%%%% size,md5,sha256,filename
## Invoked from: /home/bboyter/Go/src/github.com/boyter/hashit
## $ hashit --format hashdeep processor
##
12093,4bf92524edf098a6f5ada7b9ce6ae933,54be4e99f2635c00c6eb769a0342c2c040eac9b4f10627233e6dea8b9b20981b,processor/constants.go
18786,5b0971442f17ae00b7ad6087855d5089,b1074780eee33b1c7d548b1b94c6743691dcbc5c7d475d685c9ca77a8b7905ba,processor/workers.go
5856,58043d636928a4c0e7e6a04e69d385de,12f0e925a67d10da9327f11976c9156ba158458874d5d6fde632c27e27dead67,processor/processor.go
758,76697adfb4c818d816d3092c04fdeb46,af61af65db73a2aec2d2bea66468d9e7c44bc92bade2561754b426484a7f235b,processor/file.go
8883,8865704b023f417ffa2d6d347f5d2164,56a04cd56fb30b4ccb7b1344fbee119607b514eac57c99222dbe1319020adb5a,processor/formatters.go
444,22158f610b8a262ca8ae68424ad96ad9,47f026a06d7ced7ecbbdaf199926678cbc003b7a387eb9bbee78a2a0340297bf,processor/structs.go
2840,ce29ce9a95713628e1d8e43a51027ac1,7dcc785a34ce95c4e741e92177f221e6d05d9c1663481f35c54286fc6645934f,processor/workers_test.go
```

The output of the above can be run through hashdeep for verification,

```
$ hashit --format hashdeep processor > audit.txt && hashdeep -l -r -a -v -k audit.txt processor
hashdeep: Audit passed
          Files matched: 7
Files partially matched: 0
            Files moved: 0
        New files found: 0
  Known files not found: 0
```

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
