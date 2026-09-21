# Disaster recovery (v1.0.0)

## Planner

```bash
sde dr-plan --root .sde
```

Returns one of:

| Verdict | Meaning |
|---------|---------|
| `RECOVERABLE` | Valid archive + integrity + compatible target |
| `PARTIALLY_RECOVERABLE` | Gaps / older epoch / missing journal segment |
| `NOT_RECOVERABLE` | No usable archive or failed integrity |
| `UNKNOWN` | Insufficient evidence — **never treated as RECOVERABLE** |

Evidence includes latest archive age, epoch, digest, last fire-drill, unprotected
journal gap, and target capability negotiation.

## Readiness scorecard

```bash
sde readiness --root .sde
```

Factual dimensions only (no marketing scores): archive integrity, age, recovery
epoch, last fire drill, measured restore duration, unprotected mutations,
target compatibility.

## Escrow

Archives may be stored via:

- `LOCAL_FILESYSTEM`
- `S3_COMPATIBLE` / generic object store (HTTP path-style; **no AWS SDK in core**)

Escrow targets are ENGINE_PROVIDED — they are not Railway-native backups.

## Related

[`FIRE_DRILLS.md`](FIRE_DRILLS.md) · [`RECOVERY.md`](RECOVERY.md) · `internal/dr`
