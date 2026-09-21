# Recovery (v1.0.0)

## Goals

Restore application state from an ENGINE_PROVIDED Portable State Archive into a
**new** environment without depending on the original project/volume lifetime.

## Flows

### Independence restore (headline)

1. Export / escrow PSA while source exists  
2. Verify archive  
3. Destroy (or lose) source environment  
4. Provision new target via `RecoveryTargetAdapter`  
5. Import / restore  
6. Optional journal replay to selected epoch  
7. Verify root digest  
8. Seal `RECOVERY_RECEIPT.json`  

Automated by `examples/acquisition-demo` and `cmd/evaluate-harness`.

### Classified rollback

After a failed cutover/observation:

| Class | Meaning |
|-------|---------|
| `SAFE` | No ack writes on new active — standby restore OK |
| `REQUIRES_RECONCILIATION` / forward | Ack writes exist — do not destroy them |
| `UNSAFE` | Blind storage rollback refused |

See `internal/rollback.Classify` and `internal/recover`.

### Forward recovery

When rollback would destroy acknowledged writes, SDE refuses and classifies
`FORWARD_RECOVERY_REQUIRED` — recover into a new candidate from the newest
recoverable state instead of silently rewinding.

## Commands

```bash
sde restore --archive <dir> [--epoch N]
sde recovery-points
sde dr-plan
sde fire-drill --archive <dir>
sde rollback
sde recover
```

## Related

[`FIRE_DRILLS.md`](FIRE_DRILLS.md) · [`DR.md`](DR.md) · [`STATE_ARCHIVE_FORMAT.md`](STATE_ARCHIVE_FORMAT.md)
