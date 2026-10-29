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
