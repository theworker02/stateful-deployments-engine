# Portable State Archive (PSA) Format

**Version:** 1  
**Magic:** `SDEPSA`  
**Portability:** `ENGINE_PROVIDED` (never `RAILWAY_NATIVE`)

## Purpose

A platform-neutral, deterministic, checksummed archive of application state suitable for
cross-environment restore, escrow backups, and disaster recovery — **independent** of any
single deployment provider's native backup APIs.

SDE does **not** claim Railway endorsement. Railway-native volume backups remain a separate
`RAILWAY_NATIVE` path with same-project/environment constraints.

## Layout

```
<archive>/
  archive.json          # Manifest (this spec)
  chunks/<sha256>.chk   # Content-addressed fixed-size blocks (default 64 KiB)
  journal.jsonl         # Optional mutation journal snapshot
```

## Manifest fields

| Field | Meaning |
|-------|---------|
| `magic` | Must be `SDEPSA` |
| `format_version` | Writer version |
| `min_reader_version` | Forward-compatible gate |
| `archive_id` | Unique identity |
| `portability` | Always `ENGINE_PROVIDED` for PSA |
| `source_epoch` / `journal_position` | Recovery granularity (not sub-mutation PITR) |
| `logical_bytes` / `physical_bytes` / `dedup_bytes_saved` | **Measured** only |
| `root_digest` | Hash of sorted object index |
| `chunk_digests` | Chunk ID → sha256 (ID is the hash) |
| `object_index` | Paths, sizes, chunk ID lists |
| `verification_algorithm` | `sha256-chunk-v1` |

## Properties

- **Deterministic** object ordering (sorted paths)
- **Checksummed** chunks + root digest
- **Streamable / chunked** content-addressed blocks
- **Resumable** export can skip already-stored chunk hashes
- **Integrity-verifiable** via `sde verify-archive`
- **Forward-version-aware** via `min_reader_version`
- **Platform-neutral** directory format (no AWS/Railway types in core)

## Non-goals

- Sub-mutation point-in-time recovery finer than epoch/journal position
- Automatic anonymization of arbitrary binaries (see sensitive hooks)
- Claiming identity with `RAILWAY_NATIVE` backups
