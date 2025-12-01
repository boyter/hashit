-- SPDX-License-Identifier: MIT

create table if not exists file_hashes (
     filepath text primary key,
     crc32 text,
     xxhash64 text,
     md4 text,
     md5 text,
     sha1 text,
     sha256 text,
     sha512 text,
     blake2b_256 text,
     blake2b_512 text,
     blake3 text,
     sha3_224 text,
     sha3_256 text,
     sha3_384 text,
     sha3_512 text,
     ed2k text,
     size integer not null,
     mtime text null
);
