create table file_hashes (
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
     size integer,
     modified integer
);

CREATE INDEX idx_crc32 ON file_hashes (crc32);
CREATE INDEX idx_xxhash64 ON file_hashes (xxhash64);
CREATE INDEX idx_md4 ON file_hashes (md4);
CREATE INDEX idx_md5 ON file_hashes (md5);
CREATE INDEX idx_sha1 ON file_hashes (sha1);
CREATE INDEX idx_sha256 ON file_hashes (sha256);
CREATE INDEX idx_sha512 ON file_hashes (sha512);
CREATE INDEX idx_blake2b_256 ON file_hashes (blake2b_256);
CREATE INDEX idx_blake2b_512 ON file_hashes (blake2b_512);
CREATE INDEX idx_blake3 ON file_hashes (blake3);
CREATE INDEX idx_sha3_224 ON file_hashes (sha3_224);
CREATE INDEX idx_sha3_256 ON file_hashes (sha3_256);
CREATE INDEX idx_sha3_384 ON file_hashes (sha3_384);
CREATE INDEX idx_sha3_512 ON file_hashes (sha3_512);