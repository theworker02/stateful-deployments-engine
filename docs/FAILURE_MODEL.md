# Failure model (v1.0.0)

## Principles

1. Prefer fail-closed over silent success  
2. Persist enough state to recover after process death  
3. Classify recovery safety before mutating storage  
4. Emit machine-readable receipts for every critical path  

## Failure domains

| Domain | Examples | Primary response |
|--------|----------|------------------|
| Coordinator | Process kill mid-phase | `sde recover` from `store` |
| Journal | CRC fail, truncation, duplicate seq | Reject / resume from last good |
| Verification | Hash mismatch | Abort cutover; keep active |
| Candidate | Unhealthy shadow | Destroy candidate; active untouched |
| Post-cutover | Health fail + new writes | Forward recovery classification |
| Archive | Corruption / missing chunks | `NOT_RECOVERABLE` / partial |
| Source loss | Env wiped | Restore from escrowed PSA |

## Detailed atlas

See [`FAILURE_ATLAS.md`](FAILURE_ATLAS.md) for per-scenario detection, RPO/RTO,
test IDs, and receipt expectations.

## Chaos

`sde chaos` and `internal/chaos` exercise multi-failure combinations.
Seeds for crash fuzzing persist under testdata for reproduction.
