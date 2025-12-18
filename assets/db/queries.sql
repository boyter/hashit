-- SPDX-License-Identifier: MIT

-- name: FileHashInsertReplace :one
insert or replace into file_hashes (
    filepath, crc32, xxhash64, md4, md5, sha1, sha256, sha512,
    blake2b_256, blake2b_512, blake3, sha3_224, sha3_256, sha3_384, sha3_512, ed2k,
    size, mtime
) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
returning *;

-- name: FileHashByFilePath :one
select * from file_hashes where filepath = ?;

-- name: ListFilePathsPaged :many
SELECT filepath FROM file_hashes LIMIT ? OFFSET ?;

-- name: FileHashByMD5 :one
SELECT * FROM file_hashes WHERE md5 = ? limit 1;

-- name: FileHashBySHA1 :one
SELECT * FROM file_hashes WHERE sha1 = ? limit 1;

-- name: FileHashBySHA256 :one
SELECT * FROM file_hashes WHERE sha256 = ? limit 1;

-- name: FileHashBySHA512 :one
SELECT * FROM file_hashes WHERE sha512 = ? limit 1;