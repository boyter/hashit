

 Dual-licensed under MIT or the [UNLICENSE](http://unlicense.org).


Examples of SUM's

http://releases.ubuntu.com/16.04/MD5SUMS
http://releases.ubuntu.com/16.04/SHA1SUMS

time go run main.go --hash all --debug --no-mmap /c/Users/bboyter/Downloads/tmp

https://linhost.info/2010/05/using-hashdeep-to-ensure-data-integrity/


https://github.com/jessek/hashdeep/issues/4
https://github.com/jessek/hashdeep/issues/358

hashdeep -r vendor > audit.txt
hashdeep -r -a -k audit.txt vendor