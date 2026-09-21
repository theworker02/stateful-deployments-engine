# Upgrades notes (v0.4.1)

Post–Phase 4 hardening pass focused on reliability, honest dry-run semantics,
escrow completeness, operator UX, and acquisition-facing evidence — without
opening Agent 3 networking or claiming Railway endorsement.

## What shipped

| Upgrade | Why |
|---------|-----|
| `Recover` phase decision fix | Interrupted phase was overwritten to `RECOVERING` before decide |
| Journal CRC trailers + truncation recovery | Survive crash mid-write; detect bit-flips on resume |
| Concurrent append mutex coverage | Prove append serialization under load |
| Dry-run without fence/slot mutation | Simulation must not look like a real cutover |
| Escrow `ObjectStore` + HTTP path-style client | Make `S3_COMPATIBLE` real via interface (MinIO-friendly; no AWS SDK) |
| Fire-drill → catalog `last_fire_drill` + DR scorecard | Operators can prove drills ran |
| `sde readiness` / richer `status` / catalog table | Acquisition demos need a scorecard, not only JSON dumps |
| `export --progress` + shell `completion` | Operator UX polish |
| Chunk handle caching + measured bench JSON | Real differential numbers under `benchmarks/results/` |
| `internal/invariants` tests | Lock ACKNOWLEDGED_WRITES / STALE_EPOCH / UNKNOWN≠VERIFIED |
| `clone-redact` + `sqlite-restore-cycle` examples | Show redaction honesty + DB-like restore |
| CI `test.yml` + `Makefile` | Make green builds visible |

## Version

Bumped **0.4.0 → 0.4.1** (patch): reliability/UX/docs, no new major surface area.

## Explicit non-goals (unchanged)

- No Railway GraphQL live wiring
- No Agent 3 tunnels/mesh/mTLS/overlay/private DNS
- No fabricated benchmark claims
- Proprietary pre-acquisition license retained

## How to verify

```bash
go test ./...
go run ./examples/acquisition-demo
go run ./examples/clone-redact
go run ./examples/sqlite-restore-cycle
go test ./internal/chunk -run TestWriteDifferentialBenchResults -count=1
./sde readiness --root .sde
```
