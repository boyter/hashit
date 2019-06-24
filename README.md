Hash It!
--------

A hash tool which can work like hashdeep or md5sum, sha1sum, etc... When you want to find the hash or hashes of a file quickly, cross platform using a single command.

Yes the name is *very intentional* and similar to what say when I realise I need to get multiple hashes of a large file. This tool attempts to solve that pain.


[![Build Status](https://travis-ci.org/boyter/hashit.svg?branch=master)](https://travis-ci.org/boyter/hashit)
[![Scc Count Badge](https://sloc.xyz/github/boyter/hashit/)](https://github.com/boyter/hashit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/hashit)](https://goreportcard.com/report/github.com/boyter/hashit)

Dual-licensed under MIT or the [UNLICENSE](http://unlicense.org).

Other similar projects,

 - [hashdeep](https://github.com/jessek/hashdeep) originally called md5deep
 
### Install

If you are comfortable using Go and have >= 1.10 installed the usual `go get -u github.com/boyter/hashit/` will install for you.

Binaries will be on [Releases](https://github.com/boyter/hashit/releases) for Windows, GNU/Linux and macOS for both i386 and x86_64 bit machines once it hits version 1.0.0.

If you would like to assist with getting `hashit` added into apt/homebrew/chocolatey/etc... please submit a PR or at least raise an issue with instructions.


### Pitch

Why use `hashit`?

 - It is very fast
 - You can get multiple hashes "for free" on any CPU with multiple cores
 - Works very well across multiple platforms without slowdown (Windows, Linux, macOS)


### Usage

Command line usage of `hashit` is designed to be as simple as possible.
Full details can be found in `hashit --help` or `hashit -h`.

```
$ hashit -h
Hash It!
Ben Boyter <ben@boyter.org>

Usage:
  hashit [flags]

Flags:
  -a, --audit string      audit mode, validates argument against known hashes
      --debug             enable debug output
  -x, --file-audit        enable file audit logic where files will be checked against internal list
  -f, --format string     set output format [text, json, sum, hashdeep] (default "text")
  -c, --hash strings      hashes to be run for each file (set to 'all' for all possible hashes) (default [md5,sha1,sha256,sha512])
      --hashes            list all supported hashes
  -h, --help              help for hashit
      --no-stream         do not stream out results as processed
  -o, --output string     output filename (default stdout)
  -r, --recursive         recursive subdirectories are traversed
      --stream-size int   min size of file in bytes where stream processing starts (default 1000000)
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
README.md (4193 bytes)
        MD4 ae9c394c03361d2d714fc57eb8c3e7d5
        MD5 99fd7b856204d9b2cd6d141c1b7db29b
       SHA1 c99ff93266bfcf3f351d684f8a33e3329513afc3
     SHA256 38c9f6413588d15820c3550e5a192f0c34713459b478ff7d96e4fc883097a4b1
     SHA512 90592400b4d741ad7017230a9c8a98a3ad8eacf609a53284eb316884e557bbf71832c6dac6437aa051cb2b9d34e3849c3954946b5cab42c121c44d114b3481c3
Blake2b-256 547e63f9c2c3a4d91e1d8ca003f1455468d2b55ff88f6a53ba0c6b6cce154c62
Blake2b-512 a36027916066416aa6ac20917b41a12115e0b7e62ccebcfb3c28913c4d08bb713e6917529304b192ae45626a528c2730975d0d432fbb50f2e652a8146a34af12
   SHA3-224 5e8cb68cbcfeaccce743ab391fbf266a6a1e4799049fabccd1388a42
   SHA3-256 4975f9781b142fa433d65232cf86839c1e36ae57437f60a99281739820b66437
   SHA3-384 5a69081e37557f15f9f6a438d250c5e1db96dff94f320c6f25fffda68b611b8df36abf616a25466ecaa432402bcada61
   SHA3-512 ee7fa0b2e7329f383f7bc71881daee50b9a85daec5be70059f2a072ee0101275224911bf8d135fd60a361ad6941a37de849dd0904284f873ed026887593ded12     

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

If you supply multiple arguments which consist of files and directories then directories will be skipped. This allows wild card to be used and just process the files in a directory easily.

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


#### Misc stuff below

Examples of SUM's

http://releases.ubuntu.com/16.04/MD5SUMS
http://releases.ubuntu.com/16.04/SHA1SUMS

https://linhost.info/2010/05/using-hashdeep-to-ensure-data-integrity/


https://github.com/jessek/hashdeep/issues/4
https://github.com/jessek/hashdeep/issues/358

hashdeep -r vendor > audit.txt
hashdeep -r -a -k audit.txt vendor
