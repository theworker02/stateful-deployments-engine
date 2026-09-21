# Migration protocol (v1.0.0)

Live state mobility treats migration as a **durable transaction**.

## Planner

```bash
sde plan-migration --image app:v2
```

Inspects size, object count, write/mutation rates, bandwidth, capabilities,
historical throughput. Conclusions:

`READY | READY_WITH_RISK | LIKELY_NON_CONVERGENT | BLOCKED | UNKNOWN`

`UNKNOWN` is never auto-promoted to `READY`.

## Convergence

Measures write rate vs replication rate, backlog, backlog velocity, ETA.
If write rate ≥ replication rate, reports non-convergence and suggests
mitigations (hot-object priority, bounded quiescence, adaptive batching,
compression, unchanged-block elimination) — without inventing success.

## Bounded cutover

```bash
sde deploy --image app:v2 --max-write-pause 250ms
```

If predicted pause exceeds budget: continue sync, change strategy, or abort.
Records `predicted_pause_ms` and `actual_pause_ms` (`CUTOVER_WRITE_PAUSE_MS`).

## Pre-cutover safety gate

`CUTOVER_ALLOWED | CUTOVER_BLOCKED | CUTOVER_REQUIRES_OVERRIDE` with explicit reasons
(lag, unverified state, health, journal/manifest integrity, rollback point).

## Observation → commit

Post-cutover `TARGET_ACTIVE_UNCOMMITTED` window (`internal/guardian` /
`internal/observe`). Only then `COMMITTED`; else rollback analysis.

## Receipts

Every migration emits immutable JSON + human summary (`internal/receipt`).
