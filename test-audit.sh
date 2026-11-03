#!/bin/bash

# Audit scenarios checked against hashdeep and md5sum where they apply. Audits which are known to be
# wrong are reported as known bugs without failing, and fail once they are right so they can be
# moved to the scenarios which must pass. Needs hashdeep, the coreutils sum tools and sqlite3

GREEN='\033[1;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Building application..."
WORK=$(mktemp -d)
go build -o "$WORK/hashit" . || exit 1
HASHIT="$WORK/hashit"
KNOWN=0

fail() {
    echo -e "${RED}======================================================="
    echo -e "FAILED $1"
    echo -e "=======================================================${NC}"
    rm -rf "$WORK"
    exit 1
}

# expect_rc description exit-code command: the command must exit with the code
expect_rc() {
    (eval "$3") > "$WORK/out" 2>&1
    local rc=$?
    if [ "$rc" == "$2" ]; then
        echo -e "${GREEN}PASSED $1"
    else
        cat "$WORK/out"
        fail "$1, expected exit $2 got $rc"
    fi
}

# known_bug description exit-code command: the command should exit with the code but does not yet
known_bug() {
    (eval "$3") > "$WORK/out" 2>&1
    local rc=$?
    if [ "$rc" == "$2" ]; then
        fail "$1 is fixed, move it to the scenarios which must pass"
    fi
    echo -e "${YELLOW}KNOWN BUG $1, expected exit $2 got $rc"
    KNOWN=$((KNOWN + 1))
}

# A sum file written in the directory it lists, as in https://github.com/boyter/hashit/issues/12
mkdir -p "$WORK/user/path/sub"
cd "$WORK/user/path" || exit 1
echo a > a; echo b > sub/b
"$HASHIT" -c md5 -r -p --threads 2 -f sum -o hashes.md5 . > /dev/null 2>&1
"$HASHIT" -c md5 -r -p --threads 2 -f sum -o hashes2.md5 > /dev/null 2>&1
expect_rc "audit with a second sum file beside the first fails test" 1 "'$HASHIT' --threads 2 -a hashes.md5 ."
expect_rc "audit with a second sum file beside the first passes with --allow-new test" 0 "'$HASHIT' --threads 2 --allow-new -a hashes.md5 ."
expect_rc "audit of the second sum file passes test" 0 "'$HASHIT' --threads 2 -a hashes2.md5"
expect_rc "md5sum agrees the sum file passes test" 0 "md5sum --quiet -c hashes.md5"
echo changed >> sub/b
expect_rc "modified file fails with --allow-new test" 1 "'$HASHIT' --allow-new -a hashes.md5 ."
expect_rc "modified file fails with --allow-new and --allow-moved test" 1 "'$HASHIT' --allow-new --allow-moved -a hashes.md5 ."
echo b > sub/b
mv sub/b sub/b2
expect_rc "moved file fails with --allow-new test" 1 "'$HASHIT' --allow-new -a hashes.md5 ."
expect_rc "moved file passes with --allow-new and --allow-moved test" 0 "'$HASHIT' --allow-new --allow-moved -a hashes.md5 ."
rm sub/b2
expect_rc "missing file fails with --allow-new and --allow-moved test" 1 "'$HASHIT' --allow-new --allow-moved -a hashes.md5 ."

# A sum file of every hash the coreutils tools check is verified by them, and audited by hashit,
# and fails both once a file is modified
mkdir -p "$WORK/tools/t/sub"
cd "$WORK/tools/t" || exit 1
echo a > a; echo b > sub/b; echo "with space" > "sub/c d"
for pair in sha1:sha1sum:sha1 sha224:sha224sum:sha224 sha256:sha256sum:sha256 sha384:sha384sum:sha384 \
    sha512:sha512sum:sha512 blake2b512:b2sum:b2 blake2b256:b2sum:b2; do
    IFS=: read -r hash tool ext <<< "$pair"
    "$HASHIT" -c "$hash" -f sum -o "../$hash.$ext" . > /dev/null 2>&1
    expect_rc "$tool verifies the $hash sum file test" 0 "$tool --quiet -c '../$hash.$ext'"
    expect_rc "$hash sum file passes test" 0 "'$HASHIT' -a '../$hash.$ext' ."
done
echo changed >> "sub/c d"
for pair in sha1:sha1sum:sha1 sha224:sha224sum:sha224 sha256:sha256sum:sha256 sha384:sha384sum:sha384 \
    sha512:sha512sum:sha512 blake2b512:b2sum:b2 blake2b256:b2sum:b2; do
    IFS=: read -r hash tool ext <<< "$pair"
    expect_rc "$tool fails a modified file in the $hash sum file test" 1 "$tool --quiet -c '../$hash.$ext'"
    expect_rc "$hash sum file fails a modified file test" 1 "'$HASHIT' -a '../$hash.$ext' ."
done

# A manifest kept below the root of a tree with a name repeated at the root and beside the manifest.
# Moving a file onto the path of a known file is a modification, as hashdeep reports
for fmt in hashdeep sqlite sum; do
    case $fmt in hashdeep) M=meta/m.hd ;; sqlite) M=meta/m.db ;; sum) M=meta/m.md5 ;; esac
    mkdir -p "$WORK/meta-$fmt/orig/meta"
    cd "$WORK/meta-$fmt/orig" || exit 1
    echo top > README; echo meta > meta/README; echo w > w; echo y > y
    "$HASHIT" -c md5 -f $fmt -o $M . > /dev/null
    cp -r "$WORK/meta-$fmt/orig" "$WORK/meta-$fmt/copy"
    expect_rc "$fmt manifest below the root passes in a copy test" 0 "cd '$WORK/meta-$fmt/copy' && '$HASHIT' -a $M ."
    for tree in orig copy; do
        (cd "$WORK/meta-$fmt/$tree" && mv meta/README meta/README.old && mv w meta/README)
    done
    expect_rc "$fmt file moved onto a known path fails with --allow-moved test" 1 "cd '$WORK/meta-$fmt/orig' && '$HASHIT' --allow-moved -a $M ."
    expect_rc "$fmt file moved onto a known path fails with --allow-moved in a copy test" 1 "cd '$WORK/meta-$fmt/copy' && '$HASHIT' --allow-moved -a $M ."
done
expect_rc "hashdeep reports a file moved onto a known path test" 1 "cd '$WORK/meta-hashdeep/copy' && hashdeep -r -l -a -k meta/m.hd ."

# As above with no other files, so the manifest's directory and the working directory each name
# meta/README, and the audit must fail through either
for fmt in hashdeep sqlite sum; do
    case $fmt in hashdeep) M=meta/m.hd ;; sqlite) M=meta/m.db ;; sum) M=meta/m.md5 ;; esac
    mkdir -p "$WORK/tie-$fmt/orig/meta"
    cd "$WORK/tie-$fmt/orig" || exit 1
    echo top > README; echo meta > meta/README
    "$HASHIT" -c md5 -f $fmt -o $M . > /dev/null
    cp -r "$WORK/tie-$fmt/orig" "$WORK/tie-$fmt/copy"
    cd "$WORK/tie-$fmt/copy" || exit 1
    expect_rc "$fmt manifest below the root naming as many files through either directory passes in a copy test" 0 "'$HASHIT' -a $M ."
    mv meta/README meta/README.old && mv README meta/README
    expect_rc "$fmt file moved onto a known path naming as many files fails with --allow-moved in a copy test" 1 "'$HASHIT' --allow-moved -a $M ."
done
expect_rc "hashdeep reports a file moved onto a known path naming as many files test" 1 "cd '$WORK/tie-hashdeep/copy' && hashdeep -r -l -a -k meta/m.hd ."
expect_rc "md5sum fails a file moved onto a known path naming as many files test" 1 "cd '$WORK/tie-sum/copy' && md5sum --quiet -c meta/m.md5"

# A database not written by hashit with paths such as ././README, which are looked up through an
# index, kept below the root of a copied tree, where ././README and ././meta/README both name
# meta/README. Each directory must find its own record
mkdir -p "$WORK/spelled/orig/meta"
cd "$WORK/spelled/orig" || exit 1
echo top > README; echo meta > meta/README; echo w > w; echo x > x
"$HASHIT" -c md5 -f sqlite -o meta/m.db . > /dev/null
sqlite3 meta/m.db "update file_hashes set filepath = '././' || substr(filepath, 3)"
cp -r "$WORK/spelled/orig" "$WORK/spelled/copy"
cd "$WORK/spelled/copy" || exit 1
expect_rc "sqlite paths through the index pass in a copy test" 0 "'$HASHIT' -a meta/m.db ."
cp meta/README meta/README.bak && mv w meta/README
expect_rc "sqlite paths through the index fail a file moved onto a known path with a backup test" 1 "'$HASHIT' --allow-moved -a meta/m.db ."

# A hashdeep file without an invoked from line audited from a subdirectory, where the working
# directory names sub/b as record ./b
mkdir -p "$WORK/noinv/tree/sub"
cd "$WORK/noinv/tree" || exit 1
echo BBB > b; echo SUBB > sub/b; echo C > sub/c
hashdeep -l -r . | grep -v '^## Invoked from' > m.txt
expect_rc "hashdeep file without invoked from audited from a subdirectory passes test" 0 "cd sub && '$HASHIT' -a ../m.txt .."
cp sub/b sub/b.orig
echo MODIFIED > sub/b
expect_rc "hashdeep file without invoked from fails a modified file test" 1 "'$HASHIT' --allow-new --allow-moved -a m.txt ."
expect_rc "hashdeep fails a modified file test" 1 "hashdeep -l -r -a -k m.txt ."
expect_rc "hashdeep file without invoked from fails a modified file audited from a subdirectory test" 1 \
    "cd sub && '$HASHIT' --allow-new --allow-moved -a ../m.txt .."

# A sum file whose tree has a directory of the tree's name holding a file of the sum file's name.
# Through the working directory that record is the sum file, through the sum file's directory it
# is the other file
mkdir -p "$WORK/nested/Q/data/data"
cd "$WORK/nested/Q/data" || exit 1
echo x > x; echo inner > data/hashes.md5
"$HASHIT" -c md5 -f sum -o hashes.md5 . > /dev/null
expect_rc "sum file listing a file of its own name passes test" 0 "'$HASHIT' -a hashes.md5 ."
expect_rc "md5sum passes a sum file listing a file of its own name test" 0 "md5sum --quiet -c hashes.md5"
expect_rc "sum file listing a file of its own name passes from the parent test" 0 "cd .. && '$HASHIT' -a data/hashes.md5 data"
rm data/hashes.md5
expect_rc "deleted file of the sum file's name fails test" 1 "'$HASHIT' -a hashes.md5 ."
expect_rc "md5sum fails a deleted file of the sum file's name test" 1 "md5sum --quiet -c hashes.md5"
expect_rc "deleted file of the sum file's name fails from the parent test" 1 "cd .. && '$HASHIT' -a data/hashes.md5 data"

# A manifest kept outside the tree it was written from, which holds an empty file of the
# manifest's name. Through the manifest's directory that line names the manifest itself, but the
# audit is matched through the tree, where it is another file which must still be audited
mkdir -p "$WORK/own/sum/t" "$WORK/own/sum/m" "$WORK/own/hd/t" "$WORK/own/hd/m"
cd "$WORK/own/sum/t" || exit 1
: > hashes.md5; echo a > a; echo b > b
md5sum hashes.md5 a b > ../m/hashes.md5
expect_rc "empty file of the sum file's name kept elsewhere passes test" 0 "'$HASHIT' -a ../m/hashes.md5 ."
expect_rc "md5sum passes an empty file of the sum file's name kept elsewhere test" 0 "md5sum --quiet -c ../m/hashes.md5"
echo changed > hashes.md5
expect_rc "modified empty file of the sum file's name fails with --allow-new test" 1 "'$HASHIT' --allow-new -a ../m/hashes.md5 ."
rm hashes.md5
expect_rc "deleted empty file of the sum file's name fails test" 1 "'$HASHIT' -a ../m/hashes.md5 ."
expect_rc "md5sum fails a deleted empty file of the sum file's name test" 1 "md5sum --quiet -c ../m/hashes.md5"
cd "$WORK/own/hd/t" || exit 1
: > m.hd; echo a > a
"$HASHIT" -c md5 -f hashdeep -o ../m/m.hd . > /dev/null
expect_rc "empty file of the hashdeep file's name kept elsewhere passes test" 0 "'$HASHIT' -a ../m/m.hd ."
expect_rc "hashdeep passes an empty file of the hashdeep file's name kept elsewhere test" 0 "hashdeep -l -r -a -k ../m/m.hd ."
rm m.hd
expect_rc "deleted empty file of the hashdeep file's name fails test" 1 "'$HASHIT' -a ../m/m.hd ."
expect_rc "hashdeep fails a deleted empty file of the hashdeep file's name test" 1 "hashdeep -l -r -a -k ../m/m.hd ."

# A manifest listing a relative a and an absolute s/a, identical files. With a deleted, audited
# from s, the only file matching is named absolutely so no directory is chosen, and a must not be
# taken as the same file as s/a through the working directory s
for fmt in sum hashdeep; do
    mkdir -p "$WORK/nochosen-$fmt/t/s"
    cd "$WORK/nochosen-$fmt/t" || exit 1
    echo same > a; echo same > s/a
    if [ $fmt == sum ]; then
        M=hashes.md5; { md5sum a; md5sum "$PWD/s/a"; } > $M
    else
        M=m.hd; H=$(md5sum < a | cut -c1-32)
        printf '%%%%%%%% HASHDEEP-1.0\n%%%%%%%% size,md5,filename\n##\n5,%s,./a\n5,%s,%s\n' "$H" "$H" "$PWD/s/a" > $M
    fi
    expect_rc "$fmt relative and absolute identical files pass from a subdirectory test" 0 "cd s && '$HASHIT' -a ../$M .."
    rm a
    expect_rc "$fmt deleted file with an identical file listed absolutely fails from a subdirectory test" 1 "cd s && '$HASHIT' -a ../$M .."
    expect_rc "$fmt deleted file with an identical file listed absolutely fails test" 1 "'$HASHIT' -a $M ."
done
expect_rc "md5sum fails a deleted file with an identical file listed absolutely test" 1 "cd '$WORK/nochosen-sum/t' && md5sum --quiet -c hashes.md5"
expect_rc "hashdeep fails a deleted file with an identical file listed absolutely test" 1 "cd '$WORK/nochosen-hashdeep/t' && hashdeep -l -r -a -k m.hd ."

# A sum file at the root of the files it lists, audited from a subdirectory, with a name repeated
# in the subdirectory
mkdir -p "$WORK/root-sum/R/sub"
cd "$WORK/root-sum/R" || exit 1
echo top > README; echo sub > sub/README; echo c > sub/c
"$HASHIT" -c md5 -f sum -o hashes.md5 . > /dev/null
expect_rc "sum file at the root passes test" 0 "'$HASHIT' -a hashes.md5 ."
expect_rc "md5sum passes the sum file at the root test" 0 "md5sum --quiet -c hashes.md5"
expect_rc "sum file at the root passes from a subdirectory with a repeated name test" 0 "cd sub && '$HASHIT' -a ../hashes.md5 .."
echo changed >> sub/README
expect_rc "sum file at the root fails a modified file from a subdirectory test" 1 "cd sub && '$HASHIT' --allow-new --allow-moved -a ../hashes.md5 .."

# A SQLite database is only read by an audit
mkdir -p "$WORK/ro/t"
cd "$WORK/ro" || exit 1
echo a > t/a
"$HASHIT" -f sqlite -o m.db t > /dev/null
BEFORE=$(md5sum m.db; ls -a)
expect_rc "sqlite audit passes test" 0 "'$HASHIT' -a m.db t"
echo b > t/b
expect_rc "sqlite audit with a new file fails test" 1 "'$HASHIT' -a m.db t"
expect_rc "sqlite audit leaves the database unchanged test" 0 "[ \"\$(md5sum m.db; ls -a)\" == \"$BEFORE\" ]"

# A SQLite database with an extra column whose name contains collate is still created by hashit
cp m.db collate.db
sqlite3 collate.db "alter table file_hashes add column collateral text"
expect_rc "sqlite audit accepts a column named collateral test" 0 "'$HASHIT' --allow-new -a collate.db t"
cp m.db real-collate.db
sqlite3 real-collate.db "alter table file_hashes add column note text collate nocase"
expect_rc "sqlite audit refuses a collate clause test" 2 "'$HASHIT' -a real-collate.db t"

cd / || exit 1
rm -rf "$WORK"

echo -e "${GREEN}================================================="
echo -e "ALL AUDIT SCENARIOS PASSED, ${KNOWN} KNOWN BUGS"
echo -e "=================================================${NC}"
