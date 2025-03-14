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
     size integer not null,
     modified integer not null
);

create index if not exists idx_crc32 ON file_hashes (crc32);
create index if not exists idx_xxhash64 ON file_hashes (xxhash64);
create index if not exists idx_md4 ON file_hashes (md4);
create index if not exists idx_md5 ON file_hashes (md5);
create index if not exists idx_sha1 ON file_hashes (sha1);
create index if not exists idx_sha256 ON file_hashes (sha256);
create index if not exists idx_sha512 ON file_hashes (sha512);
create index if not exists idx_blake2b_256 ON file_hashes (blake2b_256);
create index if not exists idx_blake2b_512 ON file_hashes (blake2b_512);
create index if not exists idx_blake3 ON file_hashes (blake3);
create index if not exists idx_sha3_224 ON file_hashes (sha3_224);
create index if not exists idx_sha3_256 ON file_hashes (sha3_256);
create index if not exists idx_sha3_384 ON file_hashes (sha3_384);
create index if not exists idx_sha3_512 ON file_hashes (sha3_512);