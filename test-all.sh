#!/bin/bash

echo "Running go generate..."
go generate

echo "Running sqlc generate..."
sqlc generate

echo "Running unit tests..."
go test -shuffle on ./... || exit

echo "Building application..."
go build -ldflags="-s -w" || exit

# illumos has no SQLite driver or progress bar support, so is built without them
echo "Building for illumos..."
GOOS=illumos GOARCH=amd64 go build -o /dev/null . || exit

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

if echo "hello" | ./hashit --mtime 2>&1 | grep -q -i 'WARN'; then
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
if ./hashit --format hashdeep processor > audit.txt && hashdeep -l -r -a -k audit.txt processor | grep -q -i 'Audit passed'; then
    echo -e "${GREEN}PASSED relative hashdeep audit test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to create relative hashdeep audit"
    echo -e "======================================================="
    exit
fi

if ./hashit --format hashdeep vendor > audit.txt && hashdeep -l -r -a -k audit.txt vendor | grep -q -i 'Audit passed'; then
    echo -e "${GREEN}PASSED large relative hashdeep audit test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to create large relative hashdeep audit"
    echo -e "======================================================="
    exit
fi

mkdir -p /tmp/hashit/
echo "hello" > /tmp/hashit/file
if ./hashit --format hashdeep /tmp/hashit/ > audit.txt && hashdeep -r -a -k audit.txt /tmp/hashit/ | grep -q -i 'Audit passed'; then
    echo -e "${GREEN}PASSED full hashdeep audit test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to create full hashdeep audit"
    echo -e "======================================================="
    exit
fi

echo "Running SQLite audit tests..."

# Test 1: Creation and passing audit
rm -rf /tmp/hashit_sqlite
mkdir -p /tmp/hashit_sqlite
echo "test file 1" > /tmp/hashit_sqlite/file1.txt
rm -f /tmp/audit_sqlite.db
./hashit --hash all --format sqlite --output /tmp/audit_sqlite.db /tmp/hashit_sqlite/
AUDIT_OUTPUT=$(./hashit --audit /tmp/audit_sqlite.db /tmp/hashit_sqlite/)
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
AUDIT_OUTPUT=$(./hashit --audit /tmp/audit_sqlite.db /tmp/hashit_sqlite/)
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
AUDIT_OUTPUT=$(./hashit --audit /tmp/audit_sqlite.db /tmp/hashit_sqlite/)
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

mkdir -p /tmp/hashit/
echo "hello" > /tmp/hashit/file
if hashdeep -r /tmp/hashit/ > audit.txt && ./hashit -a audit.txt /tmp/hashit/ | grep -q -i 'Audit passed'; then
    echo -e "${GREEN}PASSED full hashdeep audit reverse test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED Should be able to create full hashdeep audit reverse"
    echo -e "======================================================="
    exit
fi

# Test 4: Missing file
rm -rf /tmp/hashit_sqlite_missing
mkdir -p /tmp/hashit_sqlite_missing
echo "file1" > /tmp/hashit_sqlite_missing/file1.txt
echo "file2" > /tmp/hashit_sqlite_missing/file2.txt
rm -f /tmp/audit_sqlite_missing.db
./hashit --hash all --format sqlite --output /tmp/audit_sqlite_missing.db /tmp/hashit_sqlite_missing/
rm /tmp/hashit_sqlite_missing/file2.txt
AUDIT_OUTPUT=$(./hashit --audit /tmp/audit_sqlite_missing.db /tmp/hashit_sqlite_missing/)
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

echo ""
echo "Running SQLite mismatched hash audit tests..."

# Setup for mismatched hash tests
rm -rf /tmp/hashit_mismatch
mkdir -p /tmp/hashit_mismatch
echo "test file 1" > /tmp/hashit_mismatch/file1.txt

# Test 5: Audit fails because no common hashes can be compared
rm -f /tmp/audit_mismatch.db
./hashit --hash md5 --format sqlite --output /tmp/audit_mismatch.db /tmp/hashit_mismatch/
AUDIT_OUTPUT=$(./hashit --hash sha1 --audit /tmp/audit_mismatch.db /tmp/hashit_mismatch/)
AUDIT_EXIT_CODE=$?
if [ $AUDIT_EXIT_CODE -ne 0 ] && echo "$AUDIT_OUTPUT" | grep -q -i 'Files modified: 1'; then
    echo -e "${GREEN}PASSED SQLite audit no common hashes test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED SQLite audit no common hashes test (Exit Code: $AUDIT_EXIT_CODE)"
    echo -e "Output was:"
    echo -e "$AUDIT_OUTPUT"
    echo -e "======================================================="
    exit
fi

# Test 6: Audit passes with a common hash subset
./hashit --hash md5,sha1 --format sqlite --overwrite --output /tmp/audit_mismatch.db /tmp/hashit_mismatch/
AUDIT_OUTPUT=$(./hashit --hash sha1 --audit /tmp/audit_mismatch.db /tmp/hashit_mismatch/)
AUDIT_EXIT_CODE=$?
if [ $AUDIT_EXIT_CODE -eq 0 ] && echo "$AUDIT_OUTPUT" | grep -q -i 'Audit passed'; then
    echo -e "${GREEN}PASSED SQLite audit common hash subset test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED SQLite audit common hash subset test (Exit Code: $AUDIT_EXIT_CODE)"
    echo -e "Output was:"
    echo -e "$AUDIT_OUTPUT"
    echo -e "======================================================="
    exit
fi

# Test 7: Audit fails with a common hash subset after modification
echo "modified content" > /tmp/hashit_mismatch/file1.txt
AUDIT_OUTPUT=$(./hashit --hash sha1 --audit /tmp/audit_mismatch.db /tmp/hashit_mismatch/)
AUDIT_EXIT_CODE=$?
if [ $AUDIT_EXIT_CODE -ne 0 ] && echo "$AUDIT_OUTPUT" | grep -q -i 'Files modified: 1'; then
    echo -e "${GREEN}PASSED SQLite audit modified common hash subset test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED SQLite audit modified common hash subset test (Exit Code: $AUDIT_EXIT_CODE)"
    echo -e "Output was:"
    echo -e "$AUDIT_OUTPUT"
    echo -e "======================================================="
    exit
fi

# Audit edge cases, see https://github.com/boyter/hashit/issues/12
AUDITDIR=$(mktemp -d)
mkdir -p "$AUDITDIR/media/sub"
echo a > "$AUDITDIR/media/a.txt"
echo b > "$AUDITDIR/media/b, with comma.txt"
echo c > "$AUDITDIR/media/c \"quoted\".txt"
echo d > "$AUDITDIR/media/sub/ünïcode.txt"
echo e > "$AUDITDIR/media/back\\slash.txt"
HASHIT=$(pwd)/hashit

audit_test() {
    # $1 description, $2 command which must pass
    if (eval "$2") | grep -q -i 'Audit passed'; then
        echo -e "${GREEN}PASSED $1"
    else
        echo -e "${RED}======================================================="
        echo -e "FAILED $1"
        echo -e "======================================================="
        exit
    fi
}

audit_test "audit file inside audited directory test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -f hashdeep -o hashes.txt . > /dev/null && '$HASHIT' -a hashes.txt ."
audit_test "hashdeep can audit hashit output generated from . test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -f hashdeep . > ../dot.txt && hashdeep -l -r -a -k ../dot.txt ."
audit_test "hashdeep can audit hashit output generated from ./sub test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -f hashdeep ./sub > ../dotsub.txt && hashdeep -l -r -a -k ../dotsub.txt ./sub"
audit_test "audit with absolute path against relative hashdeep file test" \
    "cd / && '$HASHIT' -a '$AUDITDIR/media/hashes.txt' '$AUDITDIR/media/'"
audit_test "audit md5 only hashdeep file test" \
    "cd '$AUDITDIR/media' && hashdeep -c md5 -r -l . > ../md5.txt && '$HASHIT' -a ../md5.txt ."
audit_test "audit hashdeep relative file from parent directory test" \
    "cd '$AUDITDIR/media' && hashdeep -r -l . > ../hd.txt && cd .. && '$HASHIT' -a hd.txt media"
audit_test "audit md5sum file test" \
    "cd '$AUDITDIR/media' && find . -type f -print0 | xargs -0 md5sum > ../sums.md5 && '$HASHIT' -a ../sums.md5 ."
audit_test "audit sha256sum bsd tag file test" \
    "cd '$AUDITDIR/media' && find . -type f -print0 | xargs -0 sha256sum --tag > ../sums.sha256 && '$HASHIT' -a ../sums.sha256 ."
audit_test "audit hashit sum file test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -c md5,sha1,sha256 -f sum -o ../sums.txt . > /dev/null && '$HASHIT' -a ../sums.txt ."

if (cd "$AUDITDIR/media" && "$HASHIT" -c md5 -f sum -o ../check.md5 . > /dev/null && md5sum -c --quiet ../check.md5); then
    echo -e "${GREEN}PASSED md5sum can verify sum format test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED md5sum can verify sum format test"
    echo -e "======================================================="
    exit
fi

MTIME_OUTPUT=$(cd "$AUDITDIR/media" && "$HASHIT" -c md5 --mtime -f sum -o ../mtime.md5 . > /dev/null && md5sum -c --quiet ../mtime.md5 2>&1)
if [ $? -eq 0 ] && grep -q '^# mtime ' "$AUDITDIR/mtime.md5" && ! echo "$MTIME_OUTPUT" | grep -q -i 'improperly'; then
    echo -e "${GREEN}PASSED md5sum can verify sum format with mtime test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED md5sum can verify sum format with mtime test"
    echo -e "$MTIME_OUTPUT"
    echo -e "======================================================="
    exit
fi

echo "keep me" > "$AUDITDIR/existing.txt"
if ! "$HASHIT" -o "$AUDITDIR/existing.txt" "$AUDITDIR/media" > /dev/null 2>&1 && grep -q "keep me" "$AUDITDIR/existing.txt"; then
    echo -e "${GREEN}PASSED refuses to overwrite existing output file test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED refuses to overwrite existing output file test"
    echo -e "======================================================="
    exit
fi

if "$HASHIT" --overwrite -o "$AUDITDIR/existing.txt" "$AUDITDIR/media" > /dev/null && ! grep -q "keep me" "$AUDITDIR/existing.txt"; then
    echo -e "${GREEN}PASSED overwrite replaces existing output file test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED overwrite replaces existing output file test"
    echo -e "======================================================="
    exit
fi

echo changed >> "$AUDITDIR/media/a.txt"
if (cd "$AUDITDIR/media" && "$HASHIT" -a ../sums.md5 . | grep -q 'Files modified: 1'); then
    echo -e "${GREEN}PASSED audit detects modified file test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED audit detects modified file test"
    echo -e "======================================================="
    exit
fi
audit_test "sqlite audit file inside audited directory test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -f sqlite -o audit.db . > /dev/null && '$HASHIT' -a audit.db ."
if (cd "$AUDITDIR/media" && ! "$HASHIT" -f sqlite -o audit.db . > /dev/null 2>&1); then
    echo -e "${GREEN}PASSED sqlite refuses to overwrite existing database test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sqlite refuses to overwrite existing database test"
    echo -e "======================================================="
    exit
fi

echo "not a database" > "$AUDITDIR/notadb.txt"
if ! "$HASHIT" -f sqlite --overwrite -o "$AUDITDIR/notadb.txt" "$AUDITDIR/media" > /dev/null 2>&1 && grep -q "not a database" "$AUDITDIR/notadb.txt"; then
    echo -e "${GREEN}PASSED sqlite overwrite refuses to replace non database test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sqlite overwrite refuses to replace non database test"
    echo -e "======================================================="
    exit
fi

audit_test "sqlite overwrite replaces existing database test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -f sqlite --overwrite -o audit.db . > /dev/null && '$HASHIT' -a audit.db ."
audit_test "sqlite audit with absolute path against relative database test" \
    "cd / && '$HASHIT' -a '$AUDITDIR/media/audit.db' '$AUDITDIR/media/'"
audit_test "sqlite audit from parent directory test" \
    "cd '$AUDITDIR' && '$HASHIT' -a media/audit.db media"
if [ ! -e "$AUDITDIR/media/audit.db-wal" ] && [ ! -e "$AUDITDIR/media/audit.db-shm" ]; then
    echo -e "${GREEN}PASSED sqlite audit leaves no files behind test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sqlite audit leaves no files behind test"
    echo -e "======================================================="
    exit
fi

mv "$AUDITDIR/media/sub/ünïcode.txt" "$AUDITDIR/media/moved.txt"
if (cd "$AUDITDIR/media" && "$HASHIT" -a audit.db . | grep -q 'Files moved: 1'); then
    echo -e "${GREEN}PASSED sqlite audit detects moved file test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED sqlite audit detects moved file test"
    echo -e "======================================================="
    exit
fi
if (cd "$AUDITDIR/media" && ! "$HASHIT" -a ../sums.md5 -o ../failed-audit.txt . > /dev/null) && grep -q 'Audit failed' "$AUDITDIR/failed-audit.txt"; then
    echo -e "${GREEN}PASSED failed audit to output file exits non zero test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED failed audit to output file exits non zero test"
    echo -e "======================================================="
    exit
fi
# https://github.com/boyter/hashit/issues/19
audit_test "hashdeep can audit hashit --absolute output test" \
    "cd '$AUDITDIR/media' && '$HASHIT' --absolute -f hashdeep . > ../absolute.txt && hashdeep -r -a -k ../absolute.txt . < /dev/null"

# --allow-moved passes an audit where files moved but fails for anything else
mkdir -p "$AUDITDIR/moved/sub"
echo 1 > "$AUDITDIR/moved/a"
echo 2 > "$AUDITDIR/moved/b"
(cd "$AUDITDIR/moved" && "$HASHIT" -f hashdeep . > ../moved.txt && "$HASHIT" -f sqlite -o ../moved.db . > /dev/null)
mv "$AUDITDIR/moved/a" "$AUDITDIR/moved/sub/renamed"
for AUDITFILE in moved.txt moved.db; do
    if (cd "$AUDITDIR/moved" && ! "$HASHIT" -a "../$AUDITFILE" . > /dev/null && "$HASHIT" --allow-moved -a "../$AUDITFILE" . | grep -q 'Files moved: 1'); then
        echo -e "${GREEN}PASSED allow moved $AUDITFILE test"
    else
        echo -e "${RED}======================================================="
        echo -e "FAILED allow moved $AUDITFILE test"
        echo -e "======================================================="
        exit
    fi
done
echo 3 > "$AUDITDIR/moved/new"
for AUDITFILE in moved.txt moved.db; do
    if (cd "$AUDITDIR/moved" && ! "$HASHIT" --allow-moved -a "../$AUDITFILE" . > /dev/null); then
        echo -e "${GREEN}PASSED allow moved still fails new file $AUDITFILE test"
    else
        echo -e "${RED}======================================================="
        echo -e "FAILED allow moved still fails new file $AUDITFILE test"
        echo -e "======================================================="
        exit
    fi
done

check() {
    # $1 description, $2 command which must succeed
    if (eval "$2"); then
        echo -e "${GREEN}PASSED $1"
    else
        echo -e "${RED}======================================================="
        echo -e "FAILED $1"
        echo -e "======================================================="
        exit
    fi
}

printf '%s\n' "$AUDITDIR/media" "$AUDITDIR/does-not-exist" "$AUDITDIR/media/a.txt" > "$AUDITDIR/bad-input.txt"
check "input file with directory and missing file exits non zero test" \
    "! '$HASHIT' -c md5 -f sum -i '$AUDITDIR/bad-input.txt' > '$AUDITDIR/bad-input.out' 2>/dev/null && grep -q 'a.txt' '$AUDITDIR/bad-input.out' && ! grep -q 'd41d8cd98f00b204e9800998ecf8427e' '$AUDITDIR/bad-input.out'"
check "unknown hash exits non zero test" "! '$HASHIT' -c notahash '$AUDITDIR/media' > /dev/null 2>&1"
check "unknown format exits non zero test" "! '$HASHIT' -f notaformat '$AUDITDIR/media' > /dev/null 2>&1"
check "debug output goes to stderr test" "[ \"\$('$HASHIT' --debug --trace -c md5 -f sum '$AUDITDIR/media/a.txt' 2>/dev/null | wc -l)\" -eq 1 ]"
check "very verbose audit includes modified files test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -f hashdeep . > ../vv.txt && echo vv >> a.txt && '$HASHIT' --vv -a ../vv.txt . | grep -q 'File modified'"
check "b2sum 256 bit audit test" \
    "cd '$AUDITDIR/media' && find . -type f -print0 | xargs -0 b2sum -l 256 > ../sums.b2 && '$HASHIT' -a ../sums.b2 . | grep -q 'Audit passed'"

# symlinks, special files and multiple paths
mkdir -p "$AUDITDIR/links/sub"
echo 1 > "$AUDITDIR/links/a"
echo 2 > "$AUDITDIR/links/sub/b"
ln -s sub "$AUDITDIR/links/lnk"
check "symlinked directory followed as hashdeep does test" \
    "cd '$AUDITDIR/links' && [ \"\$(timeout 20 '$HASHIT' -c md5 -f sum . 2>/dev/null | cut -d' ' -f3 | sort | tr '\n' ' ')\" == \"\$(timeout 20 hashdeep -l -r -c md5 . < /dev/null 2>/dev/null | tail -n +6 | cut -d, -f3 | sort | tr '\n' ' ')\" ]"
ln -s .. "$AUDITDIR/links/sub/loop"
mkfifo "$AUDITDIR/links/fifo"
check "symlink loop and fifo do not hang test" "timeout 20 '$HASHIT' -c md5 -f sum '$AUDITDIR/links' > /dev/null 2>&1"
rm "$AUDITDIR/links/fifo" "$AUDITDIR/links/sub/loop"
check "missing path argument still processes other paths test" \
    "! '$HASHIT' -c md5 -f sum '$AUDITDIR/links/a' '$AUDITDIR/nope' > '$AUDITDIR/missing.out' 2>/dev/null && grep -q 'links/a' '$AUDITDIR/missing.out'"
check "skipped directory warns test" "'$HASHIT' -c md5 -f sum '$AUDITDIR/links/a' '$AUDITDIR/links/sub' 2>&1 >/dev/null | grep -q 'Skipping directory'"
check "audit with multiple paths processes directories test" \
    "cd '$AUDITDIR/links' && '$HASHIT' -f hashdeep . > ../links.txt && '$HASHIT' -a ../links.txt ./a ./sub ./lnk | grep -q 'Audit passed'"
check "progress bars go to stderr test" "[ \"\$('$HASHIT' -p -c md5 -f sum '$AUDITDIR/links/a' 2>/dev/null | wc -l)\" -eq 1 ]"
check "zero threads is an error test" "! '$HASHIT' --threads 0 '$AUDITDIR/links' > /dev/null 2>&1"
printf '%s\n\n' "$AUDITDIR/links/a" > "$AUDITDIR/blank-lines.txt"
check "blank lines in input file are ignored test" "'$HASHIT' -c md5 -f sum -i '$AUDITDIR/blank-lines.txt' > /dev/null 2>&1"

# symlink cycles through another symlink are detected as hashdeep does
mkdir -p "$AUDITDIR/cross/A" "$AUDITDIR/cross/B"
echo a > "$AUDITDIR/cross/A/fa"
echo b > "$AUDITDIR/cross/B/fb"
ln -s ../B "$AUDITDIR/cross/A/l1"
ln -s ../A "$AUDITDIR/cross/B/l2"
check "symlink cycles match hashdeep test" \
    "cd '$AUDITDIR/cross' && [ \"\$(timeout 20 '$HASHIT' -c md5 -f sum . 2>/dev/null | cut -d' ' -f3 | sort | tr '\n' ' ')\" == \"\$(timeout 20 hashdeep -l -r -c md5 . < /dev/null 2>/dev/null | tail -n +6 | cut -d, -f3 | sort | tr '\n' ' ')\" ]"
ln -s nowhere "$AUDITDIR/cross/A/broken"
check "broken symlink is skipped test" "'$HASHIT' -c md5 -f sum '$AUDITDIR/cross' > /dev/null 2>&1"

mkdir -p "$AUDITDIR/outdir"
check "overwrite refuses to replace a directory test" "! '$HASHIT' -f sqlite --overwrite -o '$AUDITDIR/outdir' '$AUDITDIR/cross' > /dev/null 2>&1 && [ -d '$AUDITDIR/outdir' ]"
mkdir -p "$AUDITDIR/emptydir"
check "audit of an empty directory passes test" \
    "'$HASHIT' -f sqlite -o '$AUDITDIR/empty.db' '$AUDITDIR/emptydir' > /dev/null && '$HASHIT' -a '$AUDITDIR/empty.db' '$AUDITDIR/emptydir' | grep -q 'Audit passed' && '$HASHIT' -f hashdeep '$AUDITDIR/emptydir' > '$AUDITDIR/empty.txt' && '$HASHIT' -a '$AUDITDIR/empty.txt' '$AUDITDIR/emptydir' | grep -q 'Audit passed'"
echo literal > "$AUDITDIR/a[1].txt"
echo other > "$AUDITDIR/a1.txt"
check "filename with glob characters is used literally test" "'$HASHIT' -c md5 -f sum '$AUDITDIR/a[1].txt' | grep -q 'a\[1\].txt'"
check "-vv is the same as --vv test" "cd '$AUDITDIR/cross' && '$HASHIT' -f hashdeep . > ../cross.txt && '$HASHIT' -vv -a ../cross.txt . | grep -q ': Ok'"
check "audit with a missing path fails test" "cd '$AUDITDIR/cross' && ! '$HASHIT' -a ../cross.txt . nope 2>/dev/null | grep -q 'Audit passed'"

# an audit file outside the tree which shares its name with a file in the tree is not confused with it
mkdir -p "$AUDITDIR/samename/d"
echo 1 > "$AUDITDIR/samename/d/a.txt"
(cd "$AUDITDIR/samename/d" && "$HASHIT" -f hashdeep . > ../a.txt && md5sum ./a.txt > ../b.txt && cp ../b.txt ../a.md5)
echo changed >> "$AUDITDIR/samename/d/a.txt"
check "audit file sharing a name with an audited file test" \
    "cd '$AUDITDIR/samename/d' && '$HASHIT' -a ../a.txt . | grep -q 'Files modified: 1' && '$HASHIT' -a ../a.md5 . | grep -q 'Files modified: 1'"

# exit codes are 0 for success, 1 for a failed audit and 2 for an error
mkdir -p "$AUDITDIR/exitcodes"
echo 1 > "$AUDITDIR/exitcodes/a"
"$HASHIT" -f hashdeep "$AUDITDIR/exitcodes" > "$AUDITDIR/exitcodes.txt"
check "exit code 0 for a passed audit test" "'$HASHIT' -a '$AUDITDIR/exitcodes.txt' '$AUDITDIR/exitcodes' > /dev/null; [ \$? -eq 0 ]"
echo changed >> "$AUDITDIR/exitcodes/a"
check "exit code 1 for a failed audit test" "'$HASHIT' -a '$AUDITDIR/exitcodes.txt' '$AUDITDIR/exitcodes' > /dev/null; [ \$? -eq 1 ]"
check "exit code 2 for an audit which could not check every file test" "'$HASHIT' -a '$AUDITDIR/exitcodes.txt' '$AUDITDIR/exitcodes' '$AUDITDIR/nope' > /dev/null 2>&1; [ \$? -eq 2 ]"
check "exit code 2 for a missing path test" "'$HASHIT' '$AUDITDIR/nope' > /dev/null 2>&1; [ \$? -eq 2 ]"
check "exit code 2 for an invalid option test" "'$HASHIT' --not-a-real-option > /dev/null 2>&1; [ \$? -eq 2 ]"

# a failed run never destroys an existing output file
mkdir -p "$AUDITDIR/keep"
echo 1 > "$AUDITDIR/keep/a"
"$HASHIT" -f sqlite -o "$AUDITDIR/keep.db" "$AUDITDIR/keep" > /dev/null
check "overwrite keeps existing output when errors occur test" \
    "! '$HASHIT' -f sqlite --overwrite -o '$AUDITDIR/keep.db' '$AUDITDIR/nope' > /dev/null 2>&1 && '$HASHIT' -a '$AUDITDIR/keep.db' '$AUDITDIR/keep' | grep -q 'Files matched: 1'"
check "output same as audit file is refused test" \
    "! '$HASHIT' -a '$AUDITDIR/keep.db' -o '$AUDITDIR/keep.db' --overwrite '$AUDITDIR/keep' > /dev/null 2>&1 && head -c 15 '$AUDITDIR/keep.db' | grep -q 'SQLite format 3'"
check "no temporary output files left behind test" "! ls -a '$AUDITDIR' | grep -q '\.tmp\$'"
if [ -w /dev/full ]; then
    check "failed write to standard output exits 2 test" "'$HASHIT' -c md5 '$AUDITDIR/keep' > /dev/full 2>/dev/null; [ \$? -eq 2 ]"
fi
"$HASHIT" "$AUDITDIR/keep" > "$AUDITDIR/keep-text.txt"
check "text output as an audit file is an error test" "'$HASHIT' -a '$AUDITDIR/keep-text.txt' '$AUDITDIR/keep' > /dev/null 2>&1; [ \$? -eq 2 ]"
mkdir -p "$AUDITDIR/overlap/s"
echo 1 > "$AUDITDIR/overlap/a"
echo 2 > "$AUDITDIR/overlap/s/b"
check "overlapping audit paths are counted once test" \
    "cd '$AUDITDIR/overlap' && '$HASHIT' -f hashdeep . > ../overlap.txt && '$HASHIT' -a ../overlap.txt . s | grep -q 'Files examined: 2'"
mkdir -p "$AUDITDIR/samename2/d"
echo real > "$AUDITDIR/samename2/d/audit.txt"
echo keep > "$AUDITDIR/samename2/d/keep"
(cd "$AUDITDIR/samename2/d" && "$HASHIT" -f hashdeep -o ../audit.txt . > /dev/null && rm audit.txt)
check "deleted file sharing the audit file name is missing test" \
    "cd '$AUDITDIR/samename2/d' && '$HASHIT' -a ../audit.txt . | grep -q 'Files missing: 1'"
mkdir -p "$AUDITDIR/redirect/orig"
echo 1 > "$AUDITDIR/redirect/orig/x"
(cd "$AUDITDIR/redirect/orig" && "$HASHIT" -f hashdeep . > manifest.txt)
cp -r "$AUDITDIR/redirect/orig" "$AUDITDIR/redirect/copy"
check "copied manifest written by redirect while original exists test" \
    "cd / && '$HASHIT' -a '$AUDITDIR/redirect/copy/manifest.txt' '$AUDITDIR/redirect/copy' | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/pnpm/real_nm"
echo x > "$AUDITDIR/pnpm/real_nm/m.js"
ln -s real_nm "$AUDITDIR/pnpm/node_modules"
check "exclude-dir applies to symlinked directories test" \
    "! '$HASHIT' -c md5 -f sum --exclude-dir node_modules '$AUDITDIR/pnpm' 2>/dev/null | grep -q 'node_modules/'"

# output replacement keeps permissions and symlinks, and cleans up when interrupted
mkdir -p "$AUDITDIR/perm/d" "$AUDITDIR/perm/keep"
echo 1 > "$AUDITDIR/perm/d/a"
(umask 0022 && "$HASHIT" -o "$AUDITDIR/perm/new.txt" "$AUDITDIR/perm/d" > /dev/null)
check "new output file honours umask test" "[ \"\$(stat -c '%a' '$AUDITDIR/perm/new.txt')\" == '644' ]"
echo old > "$AUDITDIR/perm/p.txt" && chmod 640 "$AUDITDIR/perm/p.txt"
"$HASHIT" -o "$AUDITDIR/perm/p.txt" --overwrite "$AUDITDIR/perm/d" > /dev/null
check "replaced output file keeps its permissions test" "[ \"\$(stat -c '%a' '$AUDITDIR/perm/p.txt')\" == '640' ]"
echo old > "$AUDITDIR/perm/keep/real.txt"
ln -s keep/real.txt "$AUDITDIR/perm/link.txt"
"$HASHIT" -o "$AUDITDIR/perm/link.txt" --overwrite "$AUDITDIR/perm/d" > /dev/null
check "symlinked output updates the file it points to test" "[ -L '$AUDITDIR/perm/link.txt' ] && grep -q 'bytes' '$AUDITDIR/perm/keep/real.txt'"
(sleep 3; echo x) | "$HASHIT" -o "$AUDITDIR/perm/slow.txt" &
SLOWPID=$!
sleep 1
kill -INT $SLOWPID
wait $SLOWPID 2>/dev/null
check "interrupted run leaves no temporary file test" "! ls -a '$AUDITDIR/perm' | grep -q 'hashit.tmp'"
mkdir -p "$AUDITDIR/linked/real"
echo a > "$AUDITDIR/linked/real/a"
ln -s real "$AUDITDIR/linked/lnk"
"$HASHIT" -f sqlite -o "$AUDITDIR/linked/real/audit.db" "$AUDITDIR/linked/lnk/" > /dev/null
check "output reached through a symlinked path is not hashed test" \
    "[ \"\$(sqlite3 '$AUDITDIR/linked/real/audit.db' 'select count(*) from file_hashes')\" == '1' ] && '$HASHIT' -a '$AUDITDIR/linked/real/audit.db' '$AUDITDIR/linked/lnk/' | grep -q 'Audit passed'"
"$HASHIT" -f sqlite -o "$AUDITDIR/linked/x?y.db" "$AUDITDIR/linked/real" > /dev/null
check "sqlite output name containing a question mark test" "'$HASHIT' -a '$AUDITDIR/linked/x?y.db' '$AUDITDIR/linked/real' | grep -q 'Audit passed'"
check "empty hash list is an error test" "'$HASHIT' -c '' '$AUDITDIR/linked/real' > /dev/null 2>&1; [ \$? -eq 2 ]"

# audit matching
mkdir -p "$AUDITDIR/absmatch/X" "$AUDITDIR/absmatch/Z"
echo a > "$AUDITDIR/absmatch/X/a"
(cd "$AUDITDIR/absmatch/X" && "$HASHIT" -f hashdeep -o ../Z/m.txt . > /dev/null)
cp -r "$AUDITDIR/absmatch/X" "$AUDITDIR/absmatch/Y" && rm -r "$AUDITDIR/absmatch/X"
check "copied tree audited by absolute path test" "cd '$AUDITDIR/absmatch/Y' && '$HASHIT' -a ../Z/m.txt \"\$PWD\" | grep -q 'Audit passed'"
(cat "$AUDITDIR/absmatch/Z/m.txt"; tail -1 "$AUDITDIR/absmatch/Z/m.txt") > "$AUDITDIR/absmatch/dup.txt"
check "path listed twice in audit file test" "cd '$AUDITDIR/absmatch/Y' && '$HASHIT' -a ../dup.txt . 2>/dev/null | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/unread"
echo 1 > "$AUDITDIR/unread/a"
(cd "$AUDITDIR/unread" && "$HASHIT" -f hashdeep . > ../unread.txt)
echo 2 > "$AUDITDIR/unread/b" && chmod 000 "$AUDITDIR/unread/b"
if [ ! -r "$AUDITDIR/unread/b" ]; then
    check "unreadable new file is listed test" "cd '$AUDITDIR/unread' && '$HASHIT' -v -a ../unread.txt . 2>/dev/null | grep -q 'b: File could not be read'"
fi
chmod 644 "$AUDITDIR/unread/b"
if command -v rhash > /dev/null; then
    check "rhash bsd ed2k file test" "cd '$AUDITDIR/unread' && rhash --ed2k --bsd a > ../ed2k.txt && '$HASHIT' -a ../ed2k.txt a | grep -q 'Audit passed'"
fi

# explicitly supplied special files and broken links are errors, found while walking they are skipped
mkdir -p "$AUDITDIR/explicit"
echo 1 > "$AUDITDIR/explicit/a"
ln -s nowhere "$AUDITDIR/explicit/broken"
check "broken symlink found while walking is skipped test" "'$HASHIT' '$AUDITDIR/explicit' > /dev/null 2>&1"
printf '%s\n' "$AUDITDIR/explicit/broken" > "$AUDITDIR/explicit-list.txt"
check "broken symlink in input file is an error test" "'$HASHIT' -i '$AUDITDIR/explicit-list.txt' > /dev/null 2>&1; [ \$? -eq 2 ]"
check "explicit special file is an error test" "'$HASHIT' /dev/null > /dev/null 2>&1; [ \$? -eq 2 ]"
mkdir -p "$AUDITDIR/vendored/w/vendor" "$AUDITDIR/vendored/tgt"
echo t > "$AUDITDIR/vendored/tgt/t.me"
ln -s ../../tgt "$AUDITDIR/vendored/w/vendor/cache"
check "exclude-dir path end applies to symlinked directories test" \
    "! '$HASHIT' -c md5 -f sum --exclude-dir vendor/cache '$AUDITDIR/vendored/w' | grep -q 't.me'"

# round 6 regressions
mkdir -p "$AUDITDIR/r6/t/sub"
echo 1 > "$AUDITDIR/r6/t/x.txt"
echo 2 > "$AUDITDIR/r6/t/sub/x.txt"
(cd "$AUDITDIR/r6/t" && "$HASHIT" -f hashdeep . > sub/audit.txt)
check "record resolved through audit file directory does not take another record test" \
    "cd '$AUDITDIR/r6/t/sub' && '$HASHIT' -a audit.txt '$AUDITDIR/r6/t' | grep -q 'Audit passed'"
echo old > "$AUDITDIR/r6/ro.txt" && chmod 444 "$AUDITDIR/r6/ro.txt"
check "read only output file can be overwritten test" \
    "'$HASHIT' --overwrite -o '$AUDITDIR/r6/ro.txt' '$AUDITDIR/r6/t' > /dev/null && [ \"\$(stat -c '%a' '$AUDITDIR/r6/ro.txt')\" == '444' ]"
mkdir -p "$AUDITDIR/r6/hl"
echo important > "$AUDITDIR/r6/hl/f"
ln "$AUDITDIR/r6/hl/f" "$AUDITDIR/r6/hardlink.txt"
"$HASHIT" -c md5 -f sum --overwrite -o "$AUDITDIR/r6/hardlink.txt" "$AUDITDIR/r6/hl" > /dev/null
check "hard link of the output file is still hashed test" "grep -q 'hl/f' '$AUDITDIR/r6/hardlink.txt'"
(sleep 3; echo x) | "$HASHIT" -o "$AUDITDIR/r6/hup.txt" &
HUPPID=$!
sleep 1
kill -HUP $HUPPID
wait $HUPPID 2>/dev/null
check "hangup leaves no temporary file test" "! ls -a '$AUDITDIR/r6' | grep -q 'hashit.tmp'"
ln -s loopb "$AUDITDIR/r6/loopa"
ln -s loopa "$AUDITDIR/r6/loopb"
check "symlink loop output is refused test" "! '$HASHIT' --overwrite -o '$AUDITDIR/r6/loopa' '$AUDITDIR/r6/t' > /dev/null 2>&1 && [ -L '$AUDITDIR/r6/loopa' ]"
mkdir -p "$AUDITDIR/r6/dots/..cfg"
echo n > "$AUDITDIR/r6/dots/..notes"
echo x > "$AUDITDIR/r6/dots/..cfg/x"
(cd "$AUDITDIR/r6/dots" && "$HASHIT" -f sqlite -o ../dots.db . > /dev/null)
check "names starting with two dots audit from another directory test" "cd '$AUDITDIR/r6' && '$HASHIT' -a dots.db dots | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/r6/empty"
: > "$AUDITDIR/r6/empty.txt"
check "empty audit file passes against an empty directory test" "'$HASHIT' -a '$AUDITDIR/r6/empty.txt' '$AUDITDIR/r6/empty' | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/r6/orig" "$AUDITDIR/r6/copy"
echo a > "$AUDITDIR/r6/orig/a"
cp "$AUDITDIR/r6/orig/a" "$AUDITDIR/r6/copy/a"
(cd "$AUDITDIR/r6/orig" && "$HASHIT" -f hashdeep -o ../copy/m.txt . > /dev/null)
echo changed >> "$AUDITDIR/r6/orig/a"
check "two files matching one record are not both counted test" \
    "cd '$AUDITDIR/r6/orig' && '$HASHIT' -v -a ../copy/m.txt . ../copy | grep -q '^./a: File modified' && '$HASHIT' -a ../copy/m.txt . ../copy | grep -q 'Files modified: 1'"
mkdir -p "$AUDITDIR/r6/dupsum"
echo a > "$AUDITDIR/r6/dupsum/a"
(cd "$AUDITDIR/r6/dupsum" && md5sum a > ../dupsum.md5 && echo "00000000000000000000000000000000  a" >> ../dupsum.md5)
check "sum file listing a path twice uses the first test" "cd '$AUDITDIR/r6/dupsum' && '$HASHIT' -a ../dupsum.md5 . 2>/dev/null | grep -q 'Audit passed'"
check "format with audit warns test" "'$HASHIT' -f json -a '$AUDITDIR/r6/empty.txt' '$AUDITDIR/r6/empty' 2>&1 >/dev/null | grep -q 'format is ignored'"

# round 7
mkdir -p "$AUDITDIR/r7/t/deep" "$AUDITDIR/r7/in"
echo x > "$AUDITDIR/r7/in/x"
ln -s t/deep "$AUDITDIR/r7/tdl"
echo precious > "$AUDITDIR/r7/real2"
echo old > "$AUDITDIR/r7/t/real2"
ln -s tdl/../real2 "$AUDITDIR/r7/lnk2"
"$HASHIT" -c md5 -f sum -o "$AUDITDIR/r7/lnk2" --overwrite "$AUDITDIR/r7/in" > /dev/null
check "symlinked output through a symlinked directory replaces the right file test" \
    "grep -q precious '$AUDITDIR/r7/real2' && grep -q 'in/x' '$AUDITDIR/r7/t/real2'"
mkdir -p "$AUDITDIR/r7/c/sub"
echo A > "$AUDITDIR/r7/c/x"
echo B > "$AUDITDIR/r7/c/sub/x"
(cd "$AUDITDIR/r7/c" && "$HASHIT" -f hashdeep -o ../cm.txt . > /dev/null && "$HASHIT" -f sqlite -o ../cm.db . > /dev/null)
for AUDITFILE in cm.txt cm.db; do
    check "audit from a subdirectory matches the right record $AUDITFILE test" \
        "cd '$AUDITDIR/r7/c/sub' && '$HASHIT' -v -a ../../$AUDITFILE . | grep -q 'x: File expected but not found' && ! '$HASHIT' -v -a ../../$AUDITFILE . | grep -q 'File modified'"
    check "overlapping audit paths pass $AUDITFILE test" "cd '$AUDITDIR/r7/c/sub' && '$HASHIT' -a ../../$AUDITFILE . .. | grep -q 'Audit passed'"
done
(cd "$AUDITDIR/r7/in" && sha1sum x > ../x.sha1)
check "hashdeep format with a sha1 audit file test" "cd '$AUDITDIR/r7/in' && '$HASHIT' -f hashdeep -a ../x.sha1 . 2>/dev/null | grep -q 'Audit passed'"
echo OLD > "$AUDITDIR/r7/in/out.txt"
ln -s out.txt "$AUDITDIR/r7/in/lnk"
(cd "$AUDITDIR/r7/in" && "$HASHIT" -c md5 -f sum -o out.txt --overwrite . > /dev/null)
check "symlink to the output file is not hashed test" "! grep -q 'lnk' '$AUDITDIR/r7/in/out.txt'"
echo keep > "$AUDITDIR/r7/doc.txt"
check "output supplied as input is an error and kept test" "! '$HASHIT' -o '$AUDITDIR/r7/doc.txt' --overwrite '$AUDITDIR/r7/doc.txt' > /dev/null 2>&1 && grep -q keep '$AUDITDIR/r7/doc.txt'"
echo o > "$AUDITDIR/r7/m0" && chmod 000 "$AUDITDIR/r7/m0"
"$HASHIT" -o "$AUDITDIR/r7/m0" --overwrite "$AUDITDIR/r7/x.sha1" > /dev/null
check "output with mode 000 keeps it test" "[ \"\$(stat -c '%a' '$AUDITDIR/r7/m0')\" == '0' ]"
chmod 644 "$AUDITDIR/r7/m0"
mkdir -p "$AUDITDIR/r7/loop"
echo 1 > "$AUDITDIR/r7/loop/f"
ln -s self "$AUDITDIR/r7/loop/self"
check "symlink loop found while walking is a warning test" "'$HASHIT' '$AUDITDIR/r7/loop' > /dev/null 2>&1"
mkdir -p "$AUDITDIR/r7/big"
for i in $(seq 1 300); do echo $i > "$AUDITDIR/r7/big/$i"; done
"$HASHIT" -o "$AUDITDIR/r7/piped.txt" --debug --threads 1 "$AUDITDIR/r7/big" 2>&1 | head -c 10 > /dev/null
check "broken pipe leaves no temporary file test" "! ls -a '$AUDITDIR/r7' | grep -q 'hashit.tmp'"
mkdir -p "$AUDITDIR/r7/mv"
echo same > "$AUDITDIR/r7/mv/a"
(cd "$AUDITDIR/r7/mv" && "$HASHIT" -f hashdeep . > ../mv.txt)
rm "$AUDITDIR/r7/mv/a"
for n in 1 2 3 4; do mkdir -p "$AUDITDIR/r7/mv/d$n" && echo same > "$AUDITDIR/r7/mv/d$n/copy"; done
check "moved target is the first in path order test" "cd '$AUDITDIR/r7/mv' && '$HASHIT' -v --allow-moved -a ../mv.txt . | grep -q 'd1/copy: File moved'"
# files named like SQLite sidecars of a non SQLite output are unrelated and must be kept
mkdir -p "$AUDITDIR/sidecar"
echo precious > "$AUDITDIR/sidecar/notes-journal"
echo a > "$AUDITDIR/sidecar/a"
(cd "$AUDITDIR/sidecar" && "$HASHIT" -o notes --format sum . > /dev/null)
check "non sqlite output keeps files named like sidecars test" \
    "[ -f '$AUDITDIR/sidecar/notes-journal' ] && grep -q 'notes-journal' '$AUDITDIR/sidecar/notes'"

# round 8
mkdir -p "$AUDITDIR/r8/w" "$AUDITDIR/r8/t"
echo 1 > "$AUDITDIR/r8/t/a"
(cd "$AUDITDIR/r8/w" && "$HASHIT" -f sqlite -o ../t.db ../t > /dev/null)
check "sqlite database of a parent relative path audits test" "cd '$AUDITDIR/r8/w' && '$HASHIT' -a ../t.db ../t | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/r8/src/sub" "$AUDITDIR/r8/copy"
echo 1 > "$AUDITDIR/r8/src/a"
echo 2 > "$AUDITDIR/r8/src/sub/b"
(cd "$AUDITDIR/r8/src" && "$HASHIT" -f sqlite -o m.db . > /dev/null)
cp -r "$AUDITDIR/r8/src" "$AUDITDIR/r8/copy/"
check "copied sqlite database audits from inside the tree test" "cd '$AUDITDIR/r8/copy/src' && '$HASHIT' -a m.db . | grep -q 'Audit passed'"
check "copied sqlite database audits from outside the tree test" "cd / && '$HASHIT' -a '$AUDITDIR/r8/copy/src/m.db' '$AUDITDIR/r8/copy/src' | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/r8/glob"
echo 1 > "$AUDITDIR/r8/glob/a"
echo 2 > "$AUDITDIR/r8/glob/b"
(cd "$AUDITDIR/r8/glob" && "$HASHIT" -f sum -c md5 -o sums.md5 * > /dev/null)
check "audit file matched by a wildcard is skipped test" "cd '$AUDITDIR/r8/glob' && '$HASHIT' -a sums.md5 * 2>/dev/null | grep -q 'Audit passed'"
check "output matched by a wildcard is skipped test" "cd '$AUDITDIR/r8/glob' && '$HASHIT' -f sum -c md5 -o sums.md5 --overwrite * > /dev/null 2>&1 && ! grep -q sums.md5 sums.md5"
mkdir -p "$AUDITDIR/r8/real/sub"
ln -s real/sub "$AUDITDIR/r8/dl"
echo 1 > "$AUDITDIR/r8/real/sub/f"
(cd "$AUDITDIR/r8" && "$HASHIT" -o dl/../new.txt real/sub > /dev/null)
check "new output through a symlinked directory is created where it resolves test" "[ -f '$AUDITDIR/r8/real/new.txt' ] && [ ! -e '$AUDITDIR/r8/new.txt' ]"
mkdir -p "$AUDITDIR/r8/y1"
echo 1 > "$AUDITDIR/r8/y1/a"
(cd "$AUDITDIR/r8/y1" && "$HASHIT" -f hashdeep . > manifest.txt && ln -s manifest.txt mlink)
check "audit file through a differently named symlink test" "cd '$AUDITDIR/r8/y1' && '$HASHIT' -a mlink . | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/r8/y2/C" "$AUDITDIR/r8/y2/e"
echo 1 > "$AUDITDIR/r8/y2/C/a"
printf '%%%%%%%% HASHDEEP-1.0\n%%%%%%%% size,md5,sha256,filename\n## Invoked from: /nonexistent\n%s\n' "$(cd "$AUDITDIR/r8/y2/C" && "$HASHIT" -f hashdeep a | grep ',a$')" > "$AUDITDIR/r8/y2/C/manifest.txt"
ln -s ../C/manifest.txt "$AUDITDIR/r8/y2/e/m"
check "audit file symlinked from another directory uses its real directory test" "cd / && '$HASHIT' -a '$AUDITDIR/r8/y2/e/m' '$AUDITDIR/r8/y2/C' | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/r8/X" "$AUDITDIR/r8/Y"
echo 1 > "$AUDITDIR/r8/X/a"
"$HASHIT" -o "$AUDITDIR/r8/X/out" "$AUDITDIR/r8/X" > /dev/null
ln -s ../X "$AUDITDIR/r8/Y/dl"
ln "$AUDITDIR/r8/X/out" "$AUDITDIR/r8/X/hard"
"$HASHIT" -o "$AUDITDIR/r8/X/out" --overwrite "$AUDITDIR/r8/Y" > /dev/null
check "output through a symlinked directory is not hashed test" "! grep -q 'dl/out' '$AUDITDIR/r8/X/out' && grep -q 'dl/hard' '$AUDITDIR/r8/X/out'"
check "audit file read from a pipe test" "cd '$AUDITDIR/r8/glob' && cat sums.md5 | '$HASHIT' -a /dev/stdin a b | grep -q 'Audit passed'"
check "sqlite audit database from a pipe is an error test" "cat '$AUDITDIR/r8/t.db' | '$HASHIT' -a /dev/stdin '$AUDITDIR/r8/t' > /dev/null 2>&1; [ \$? -eq 2 ]"
check "broken pipe exits 141 test" "'$HASHIT' -o '$AUDITDIR/r8/piped.txt' --overwrite '$AUDITDIR/r7/big' 2>/dev/null | (exec 0<&-; sleep 0.2); [ \${PIPESTATUS[0]} -eq 141 ]"

# round 9
mkdir -p "$AUDITDIR/r9/slow"
truncate -s 400M "$AUDITDIR/r9/slow/f"
check "signal ignored when started stays ignored test" "sh -c \"trap '' INT HUP; '$HASHIT' -c all -o '$AUDITDIR/r9/ign.txt' '$AUDITDIR/r9/slow' > /dev/null & p=\\\$!; sleep 0.3; kill -INT \\\$p; kill -HUP \\\$p; wait \\\$p\" && [ -s '$AUDITDIR/r9/ign.txt' ]"
check "output created while running is not replaced test" "('$HASHIT' -c all -o '$AUDITDIR/r9/race.txt' '$AUDITDIR/r9/slow' > /dev/null 2>&1; echo \$? > '$AUDITDIR/r9/race.rc') & sleep 0.3; echo precious > '$AUDITDIR/r9/race.txt'; wait; [ \"\$(cat '$AUDITDIR/r9/race.txt')\" == precious ] && [ \"\$(cat '$AUDITDIR/r9/race.rc')\" == 2 ] && ! ls -a '$AUDITDIR/r9' | grep -q hashit.tmp"
mkdir -p "$AUDITDIR/r9/tree" "$AUDITDIR/r9/gone"
echo 1 > "$AUDITDIR/r9/tree/a"
"$HASHIT" -f hashdeep -o "$AUDITDIR/r9/m.hd" "$AUDITDIR/r9/tree" > /dev/null
"$HASHIT" -f sqlite -o "$AUDITDIR/r9/m.db" "$AUDITDIR/r9/tree" > /dev/null
check "audit from a deleted working directory test" "cd '$AUDITDIR/r9/gone' && rmdir '$AUDITDIR/r9/gone' && '$HASHIT' -a '$AUDITDIR/r9/m.hd' '$AUDITDIR/r9/tree' | grep -q 'Audit passed' && '$HASHIT' -a '$AUDITDIR/r9/m.db' '$AUDITDIR/r9/tree' | grep -q 'Audit passed'"
mkdir -p "$AUDITDIR/r9/sp/t/a"
echo 1 > "$AUDITDIR/r9/sp/t/a/x"
echo 2 > "$AUDITDIR/r9/sp/t/y"
(cd "$AUDITDIR/r9/sp/t" && "$HASHIT" -f sqlite -o ../m.db . > /dev/null)
sqlite3 "$AUDITDIR/r9/sp/m.db" "update file_hashes set filepath = 'a//x' where filepath = './a/x'; update file_hashes set filepath = '$AUDITDIR/r9/sp/t/y' where filepath = './y'"
check "sqlite database with unusual path spellings audits test" "cd '$AUDITDIR/r9/sp/t' && '$HASHIT' -a ../m.db . | grep -q 'Audit passed' && cd .. && '$HASHIT' -a m.db t | grep -q 'Audit passed'"
if unshare -r -m true 2> /dev/null; then
    mkdir -p "$AUDITDIR/r9/full/mnt" "$AUDITDIR/r9/full/src"
    for i in $(seq 1 3000); do echo $i > "$AUDITDIR/r9/full/src/$i"; done
    check "sqlite database which cannot be written in full is discarded test" "cd '$AUDITDIR/r9/full' && unshare -r -m sh -c \"mount -t tmpfs -o size=300k tmpfs mnt && ! '$HASHIT' -f sqlite -o mnt/n.db src > out.txt 2>&1 && [ -z \\\"\\\$(ls -A mnt)\\\" ]\" && ! grep -q 'results written' out.txt"
fi

# a manifest kept with the files can be audited after they are copied elsewhere
mkdir -p "$AUDITDIR/ship/orig/tree/sub" "$AUDITDIR/ship/copy"
echo 1 > "$AUDITDIR/ship/orig/tree/one"
echo 2 > "$AUDITDIR/ship/orig/tree/sub/two"
(cd "$AUDITDIR/ship/orig/tree" && "$HASHIT" -f hashdeep . > ../manifest.txt && mv ../manifest.txt .)
cp -r "$AUDITDIR/ship/orig/tree" "$AUDITDIR/ship/copy/"
check "copied manifest audits from inside the tree test" "cd '$AUDITDIR/ship/copy/tree' && '$HASHIT' -a manifest.txt . | grep -q 'Audit passed'"
check "copied manifest audits from outside the tree test" "cd / && '$HASHIT' -a '$AUDITDIR/ship/copy/tree/manifest.txt' '$AUDITDIR/ship/copy/tree' | grep -q 'Audit passed'"

head -c 300000 /dev/urandom > "$AUDITDIR/threshold.bin"
for SUMTOOL in sha224sum sha384sum; do
    HASHNAME=${SUMTOOL%sum}
    check "$HASHNAME matches $SUMTOOL test" \
        "[ \"\$('$HASHIT' -c $HASHNAME -f hashonly '$AUDITDIR/threshold.bin')\" == \"\$($SUMTOOL '$AUDITDIR/threshold.bin' | cut -d' ' -f1)\" ]"
    check "audit $SUMTOOL file test" \
        "cd '$AUDITDIR/media' && find . -type f -print0 | xargs -0 $SUMTOOL > ../sums.$HASHNAME && '$HASHIT' -a ../sums.$HASHNAME . | grep -q 'Audit passed'"
done

# files are hashed differently depending on size, every hash must agree across the thresholds
for HASHNAME in $("$HASHIT" --hashes | sed -n 's/.*(\(.*\))/\1/p'); do
    A=$("$HASHIT" -c "$HASHNAME" -f hashonly "$AUDITDIR/threshold.bin")
    B=$("$HASHIT" -c "$HASHNAME" -f hashonly --stream-size 0 "$AUDITDIR/threshold.bin")
    if [ -n "$A" ] && [ "$A" == "$B" ]; then
        echo -e "${GREEN}PASSED $HASHNAME parallel and stream agree test"
    else
        echo -e "${RED}======================================================="
        echo -e "FAILED $HASHNAME parallel and stream agree test"
        echo -e "======================================================="
        exit
    fi
done

# https://github.com/boyter/hashit/issues/20
audit_test "hashdeep can audit hashit --mtime output test" \
    "cd '$AUDITDIR/media' && '$HASHIT' --mtime -f hashdeep . > ../mtime-hashdeep.txt && grep -q '^## mtime ' ../mtime-hashdeep.txt && hashdeep -l -r -a -k ../mtime-hashdeep.txt . < /dev/null"
audit_test "hashit can audit hashit --mtime hashdeep output test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -a ../mtime-hashdeep.txt ."

# https://github.com/boyter/hashit/issues/27
audit_test "hashdeep can audit hashit sha256 only output test" \
    "cd '$AUDITDIR/media' && '$HASHIT' -c sha256 -f hashdeep . > ../sha256only.txt && hashdeep -l -r -a -k ../sha256only.txt . < /dev/null"

head -c 20000000 /dev/urandom > "$AUDITDIR/large-stdin.bin"
if [ "$("$HASHIT" -c sha256 -f hashonly < "$AUDITDIR/large-stdin.bin")" == "$(sha256sum < "$AUDITDIR/large-stdin.bin" | cut -d' ' -f1)" ]; then
    echo -e "${GREEN}PASSED large stdin hash test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED large stdin hash test"
    echo -e "======================================================="
    exit
fi

echo "$AUDITDIR/media/a.txt" > "$AUDITDIR/input-list.txt"
if echo "not this" | "$HASHIT" -c md5 -f sum -i "$AUDITDIR/input-list.txt" | grep -q "a.txt"; then
    echo -e "${GREEN}PASSED input file used when stdin is a pipe test"
else
    echo -e "${RED}======================================================="
    echo -e "FAILED input file used when stdin is a pipe test"
    echo -e "======================================================="
    exit
fi

rm -rf "$AUDITDIR"

echo -e "${NC}Cleaning up..."
rm ./hashit
rm -f ./audit.txt
rm -rf /tmp/hashit
rm -rf /tmp/hashit_sqlite
rm -rf /tmp/hashit_sqlite_missing
rm -rf /tmp/hashit_mismatch
rm -f /tmp/audit_sqlite.db
rm -f /tmp/audit_sqlite_missing.db
rm -f /tmp/audit_mismatch.db

echo -e "${GREEN}================================================="
echo -e "ALL TESTS PASSED"
echo -e "================================================="
