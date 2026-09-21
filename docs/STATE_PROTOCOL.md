# State protocol (v1.0.0)

## Epochs

Monotonic `epoch` advances on successful commit. All journal mutations carry the
active epoch. Rollback pins epoch to the standby’s epoch.

## Journal records

Append-only, ordered by `seq`. Fields include operation, path, checksums,
epoch, deployment id, and replay status. Unrecognized operations **fail closed**
(never silently discarded).

CRC trailers detect torn writes; reopen recovers truncation boundaries.

## Fencing

`internal/fence` issues tokens. Coordinators must present a current fence to
mutate durable deploy state. Stale processes are rejected.

## FSM persistence

`internal/store` writes `deployments/<id>/state.json` atomically (temp + rename)
at each major phase so crashes are recoverable.

## Phases

`ACTIVE → CHECKPOINT → SHADOW → SYNCHRONIZING → VERIFYING → QUIESCING →
FINAL_DELTA → CUTOVER → OBSERVING → COMMITTED`

Unsafe: `ABORTED | DEGRADED | RECOVERING | ROLLED_BACK`

See [`MIGRATION_PROTOCOL.md`](MIGRATION_PROTOCOL.md) and `internal/fsm`.
