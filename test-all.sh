#!/bin/bash

echo "Running go generate..."
go generate

echo "Running sqlc generate..."
sqlc generate

echo "Running unit tests..."
go test -shuffle on ./... || exit

echo "Building application..."
go build -ldflags="-s -w" || exit

GREEN='\033[1;32m'
RED='\033[0;31m'
NC='\033[0m'

if ./hashit --not-a-real-option > /dev/null ; then
    echo -e "${RED}================================================="
    echo -e "FAILED Invalid option should produce error code "
    echo -e "======================================================="
    exit
else
    echo -e "${GREEN}PASSED invalid option test"
fi

if ./hashit > /dev/null ; then
    echo -e "${GREEN}PASSED no directory specified test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should run correctly with no directory specified"
    echo -e "======================================================="
    exit
fi

if ./hashit * > /dev/null ; then
    echo -e "${GREEN}PASSED all files test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should run correctly with all files specified"
    echo -e "======================================================="
    exit
fi

if ./hashit "*.go" | grep -q -i 'main.go'; then
    echo -e "${GREEN}PASSED globbing test for *.go"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Globbing test for *.go"
    echo -e "======================================================="
    exit
fi

if ./hashit --debug --trace --verbose -f text --hash md5 --no-stream --stream-size 10 -r main.go > /dev/null ; then
    echo -e "${GREEN}PASSED multiple options test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should run correctly with multiple options"
    echo -e "======================================================="
    exit
fi

if ./hashit processor > /dev/null ; then
    echo -e "${GREEN}PASSED directory specified test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should run correctly with directory specified"
    echo -e "======================================================="
    exit
fi

if ./hashit main.go -c md5 | grep -q -i 'md5'; then
    echo -e "${GREEN}PASSED short hash test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to work with short hash"
    echo -e "======================================================="
    exit
fi

for i in 'crc32' 'xxHash64' 'md4' 'md5' 'sha1' 'sha256' 'sha512'
do
    if ./hashit main.go --hash $i | grep -q -i $i; then
        echo -e "${GREEN}PASSED hash test $i"
    else
        echo -e "${RED}======================================================="
        echo -e "FAILED Should be able to work with hash $i"
        echo -e "======================================================="
        exit
    fi
done

if ./hashit main.go --hash blake2b256 | grep -q -i 'blake2b-256'; then
    echo -e "${GREEN}PASSED hash test blake2b256"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to work with hash blake2b256"
    echo -e "======================================================="
    exit
fi

if ./hashit main.go --hash blake2b512 | grep -q -i 'blake2b-512'; then
    echo -e "${GREEN}PASSED hash test blake2b512"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to work with hash blake2b512"
    echo -e "======================================================="
    exit
fi

if ./hashit main.go --hash sha3224 | grep -q -i 'sha3-224'; then
    echo -e "${GREEN}PASSED hash test sha3224"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to work with hash sha3224"
    echo -e "======================================================="
    exit
fi

if ./hashit main.go --hash sha3256 | grep -q -i 'sha3-256'; then
    echo -e "${GREEN}PASSED hash test sha3256"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to work with hash sha3256"
    echo -e "======================================================="
    exit
fi

if ./hashit main.go --hash sha3384 | grep -q -i 'sha3-384'; then
    echo -e "${GREEN}PASSED hash test sha3384"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to work with hash sha3384"
    echo -e "======================================================="
    exit
fi

if ./hashit main.go --hash sha3512 | grep -q -i 'sha3-512'; then
    echo -e "${GREEN}PASSED hash test sha3512"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to work with hash sha3512"
    echo -e "======================================================="
    exit
fi

if ./hashit main.go --hashes | grep -q -i 'md5'; then
    echo -e "${GREEN}PASSED hashes test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to display hashes"
    echo -e "======================================================="
    exit
fi

if echo "hello" | ./hashit --hash xxhash64 | grep -q -i 'e4c191d091bd8853'; then
    echo -e "${GREEN}PASSED stdin xxhash64 test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to process xxhash64 stdin"
    echo -e "======================================================="
    exit
fi

if echo "hello" | ./hashit --hash md5 | grep -q -i 'b1946ac92492d2347c6235b4d2611184'; then
    echo -e "${GREEN}PASSED stdin md5 test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to process md5 stdin"
    echo -e "======================================================="
    exit
fi

if echo "hello" | ./hashit --hash sha1 | grep -q -i 'f572d396fae9206628714fb2ce00f72e94f2258f'; then
    echo -e "${GREEN}PASSED stdin sha1 test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to process sha1 stdin"
    echo -e "======================================================="
    exit
fi

if echo "hello" | ./hashit --hash sha256 | grep -q -i '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03'; then
    echo -e "${GREEN}PASSED stdin sha256 test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to process sha256 stdin"
    echo -e "======================================================="
    exit
fi

if echo "hello" | ./hashit --hash ed2k | grep -q -i '63481c78ae04c201fa01ea9d2b1db56d'; then
    echo -e "${GREEN}PASSED stdin ed2k test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to process ed2k stdin"
    echo -e "======================================================="
    exit
fi

if ./hashit --mtime ./hashit | grep -q -i 'Mtime'; then
    echo -e "${GREEN}PASSED mtime default"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to deal with mtime default"
    echo -e "======================================================="
    exit
fi

if ./hashit --mtime -f hashdeep ./hashit | grep -q -i 'mtime'; then
    echo -e "${GREEN}PASSED mtime hashdeep"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to deal with mtime hashdeep"
    echo -e "======================================================="
    exit
fi

if echo "hello" | ./hashit --mtime 2>&1 | grep -q -i 'ERROR'; then
    echo -e "${GREEN}PASSED stdin mtime test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to deal with mtime stdin"
    echo -e "======================================================="
    exit
fi

a=$(./hashit --no-stream * | sort | md5sum)
b=$(./hashit * | sort | md5sum)
if [ "$a" == "$b" ]; then
    echo -e "${GREEN}PASSED stream output test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED stream output test"
    echo -e "================================================="
    exit
fi

a=$(./hashit --format sum --hash md5 main.go)
b=$(md5sum main.go)
if [ "$a" == "$b" ]; then
    echo -e "${GREEN}PASSED sum md5 format test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sum md5 format test"
    echo -e "================================================="
    exit
fi

a=$(./hashit --format sum --hash sha1 main.go)
b=$(sha1sum main.go)
if [ "$a" == "$b" ]; then
    echo -e "${GREEN}PASSED sum sha1 format test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sum sha1 format test"
    echo -e "================================================="
    exit
fi

a=$(./hashit --format sum --hash sha256 main.go)
b=$(sha256sum main.go)
if [ "$a" == "$b" ]; then
    echo -e "${GREEN}PASSED sum sha256 format test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sum sha256 format test"
    echo -e "================================================="
    exit
fi

a=$(./hashit --format sum --hash sha512 main.go)
b=$(sha512sum main.go)
if [ "$a" == "$b" ]; then
    echo -e "${GREEN}PASSED sum sha512 format test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sum sha512 format test"
    echo -e "================================================="
    exit
fi

for i in '' '--stream-size 0'
do
    if ./hashit $i LICENSE | grep -q -i '227f999ca03b135a1b4d69bde84afb16'; then
        echo -e "${GREEN}PASSED stream $i hash test"
    else
        echo -e "${RED}======================================================="
        echo -e "FAILED $i test"
        echo -e "======================================================="
        exit
    fi
done

a=$(./hashit --format sum --hash all ./LICENSE)
b=$(./hashit --format sum --hash all --stream-size 1 ./LICENSE)
if [ "$a" == "$b" ]; then
    echo -e "${GREEN}PASSED small scanner test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED small scanner test"
    echo -e "================================================="
    exit
fi

echo ""
echo "Running SQLite audit tests..."

# Test 1: Creation and passing audit
mkdir -p /tmp/hashit_sqlite
echo "test file 1" > /tmp/hashit_sqlite/file1.txt
./hashit --hash all --format sqlite --output /tmp/hashit_sqlite/audit.db /tmp/hashit_sqlite/
AUDIT_OUTPUT=$(./hashit --audit /tmp/hashit_sqlite/audit.db /tmp/hashit_sqlite/)
AUDIT_EXIT_CODE=$?
if [ $AUDIT_EXIT_CODE -eq 0 ] && echo "$AUDIT_OUTPUT" | grep -q -i 'Audit passed'; then
    echo -e "${GREEN}PASSED SQLite audit basic test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED SQLite audit basic test (Exit Code: $AUDIT_EXIT_CODE)"
    echo -e "Output was:"
    echo -e "$AUDIT_OUTPUT"
    echo -e "======================================================="
    exit
fi

# Test 2: Modified file
echo "modified content" > /tmp/hashit_sqlite/file1.txt
AUDIT_OUTPUT=$(./hashit --audit /tmp/hashit_sqlite/audit.db /tmp/hashit_sqlite/)
AUDIT_EXIT_CODE=$?
if [ $AUDIT_EXIT_CODE -ne 0 ] && echo "$AUDIT_OUTPUT" | grep -q -i 'Files modified: 1'; then
    echo -e "${GREEN}PASSED SQLite audit modified file test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED SQLite audit modified file test (Exit Code: $AUDIT_EXIT_CODE)"
    echo -e "Output was:"
    echo -e "$AUDIT_OUTPUT"
    echo -e "======================================================="
    exit
fi

# Test 3: New file
# NB we expect 1 modified and 1 new file here because of the previous test
echo "new file" > /tmp/hashit_sqlite/file2.txt
AUDIT_OUTPUT=$(./hashit --audit /tmp/hashit_sqlite/audit.db /tmp/hashit_sqlite/)
AUDIT_EXIT_CODE=$?
if [ $AUDIT_EXIT_CODE -ne 0 ] && echo "$AUDIT_OUTPUT" | grep -q -i 'New files found: 1'; then
    echo -e "${GREEN}PASSED SQLite audit new file test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED SQLite audit new file test (Exit Code: $AUDIT_EXIT_CODE)"
    echo -e "Output was:"
    echo -e "$AUDIT_OUTPUT"
    echo -e "======================================================="
    exit
fi

# Test 4: Missing file
mkdir -p /tmp/hashit_sqlite_missing
echo "file1" > /tmp/hashit_sqlite_missing/file1.txt
echo "file2" > /tmp/hashit_sqlite_missing/file2.txt
./hashit --hash all --format sqlite --output /tmp/hashit_sqlite_missing/audit.db /tmp/hashit_sqlite_missing/
rm /tmp/hashit_sqlite_missing/file2.txt
AUDIT_OUTPUT=$(./hashit --audit /tmp/hashit_sqlite_missing/audit.db /tmp/hashit_sqlite_missing/)
AUDIT_EXIT_CODE=$?
if [ $AUDIT_EXIT_CODE -ne 0 ] && echo "$AUDIT_OUTPUT" | grep -q -i 'Files missing: 1'; then
    echo -e "${GREEN}PASSED SQLite audit missing file test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED SQLite audit missing file test (Exit Code: $AUDIT_EXIT_CODE)"
    echo -e "Output was:"
    echo -e "$AUDIT_OUTPUT"
    echo -e "======================================================="
    exit
fi

echo -e "${NC}Cleaning up..."
rm ./hashit
rm -f ./audit.txt
rm -rf /tmp/hashit
rm -rf /tmp/hashit_sqlite
rm -rf /tmp/hashit_sqlite_missing

echo -e "${GREEN}================================================="
echo -e "ALL TESTS PASSED"
echo -e "================================================="