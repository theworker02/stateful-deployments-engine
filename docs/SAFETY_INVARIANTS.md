# Safety invariants (v1.0.0)

These invariants are **non-negotiable** for acquisition evaluation.
Tests live in `internal/invariants` and chaos/fuzz suites.

## ACKNOWLEDGED_WRITES_MUST_NOT_BE_SILENTLY_LOST

If the new active has accepted writes after cutover, storage rollback that would
destroy those writes is **refused** unless explicitly forced and classified
`UNSAFE`. Prefer forward recovery.

## STALE_EPOCH_MUST_NOT_MODIFY_NEWER_STATE

Every mutating coordinator operation carries a fencing token / epoch.
A restarted or delayed process with a stale fence cannot advance a newer epoch’s state.

## UNVERIFIED_STATE_MUST_NOT_BE_REPORTED_VERIFIED

Verification outcomes are `VERIFIED | PARTIAL | FAILED | UNKNOWN`.
`UNKNOWN` is never coerced to `VERIFIED` in gates, receipts, or DR planners.

## FAILED_CUTOVER_MUST_HAVE_DETERMINISTIC_RECOVERY_PATH

Persisted FSM state (`internal/store`) survives coordinator crash. On restart,
`sde recover` chooses resume / retry / abort / rollback / reconciliation from
durable phase — not operator guesswork.

## Operational corollaries

- Dry-run must not mutate production slots or issue write fences  
- Fire drills never write into production state paths  
- PSA export/import verifies digests before claiming success  

## Related

[`FAILURE_ATLAS.md`](FAILURE_ATLAS.md) · [`FAILURE_MODEL.md`](FAILURE_MODEL.md)
