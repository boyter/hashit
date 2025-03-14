

-- name: FileHashInsertReplace :one
insert or replace into file_hashes (
    filepath, crc32, xxhash64, md4, md5, sha1, sha256, sha512,
    blake2b_256, blake2b_512, blake3, sha3_224, sha3_256, sha3_384, sha3_512,
    size, modified
) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
returning *;

-- name: FileHashByFilePath :one
select * from file_hashes where filepath = ?;