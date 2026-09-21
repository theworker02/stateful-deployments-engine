# Corruption fixtures

| File | Intent |
|------|--------|
| `truncated-journal.ndjson` | Last line cut mid-JSON — journal open must detect |
| `hash-mismatch-manifest.json` | Manifest merkle does not match object digests |

These are inputs for verifier / chaos tests — not live deploy artifacts.
