# Deployment receipts (v1.0.0)

Every critical operation emits an immutable, structured receipt.

## Types

| Receipt | Produced by |
|---------|-------------|
| Deploy / migration receipt | Coordinator commit / abort |
| Verification receipt | Verifier levels |
| `CLONE_RECEIPT.json` | `sde railway clone` / clone flows |
| `RECOVERY_RECEIPT.json` | Restore / fire-drill / independence test |
| Safety gate result | Pre-cutover gate |
| Chaos receipt | Fault harness |

## Common fields

- IDs (deploy, archive, epoch, fence)  
- Digests (source / candidate / restored)  
- Counts (mutations, bytes, objects)  
- Timings (`CUTOVER_WRITE_PAUSE_MS`, restore duration)  
- Classification (rollback safety, DR verdict)  
- Human summary string  

## Schema

JSON Schemas under `schemas/` (where present) and Go types in `internal/receipt`,
`internal/recovery`, `internal/types`.

## Rule

Receipts are **evidence**, not marketing. Failed operations still emit receipts with
`FAILED` / error fields — silence is not success.
