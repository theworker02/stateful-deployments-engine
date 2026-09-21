# State archive format (v1.0.0)

Canonical specification: [`spec/STATE_ARCHIVE_FORMAT.md`](spec/STATE_ARCHIVE_FORMAT.md).

## Operator summary

| Property | Detail |
|----------|--------|
| Magic | `SDEPSA` |
| Portability | Always `ENGINE_PROVIDED` |
| Layout | `archive.json` + `chunks/` + optional `journal.jsonl` |
| Integrity | Per-chunk SHA-256 + root digest |
| CLI | `export` · `inspect-archive` · `verify-archive` · `import` |

## What it is not

- Not a Railway-native volume backup  
- Not sub-mutation PITR finer than epoch/journal position  
- Not automatic PII anonymization (hooks only)

## Deduplication

Unchanged chunks are content-addressed; logical vs physical bytes and dedup
savings are **measured**, never fabricated (`archive` stats on export).
