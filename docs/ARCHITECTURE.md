# Architecture (v1.1.0)

Canonical product architecture for evaluators and acquirers.
Root overview also lives at [`../ARCHITECTURE.md`](../ARCHITECTURE.md).

**Independent project — not affiliated with Railway.**

## Positioning

SDE is a **platform-independent state lifecycle engine** for arbitrary volume-backed
workloads. Railway is the first **adapter** target, not the product identity.

Value proposition (correct framing):

> General-purpose state lifecycle: transactional cutover, mutation journals,
> ENGINE_PROVIDED portable archives, verified recovery, and fire drills —
> beyond database-specific backup tooling.

Not:

> “Railway doesn’t have backups.”

Railway already offers volumes, incremental/scheduled backups, Postgres PITR,
logical dumps, and live resize. Documented limits (same-project/env native restore;
wipe removes volume backups) still motivate portable, independently escrowed recovery.

## Layered design

```
┌──────────────────────────────────────────────────────────────┐
│  CLI  ·  LIBRARY embed  ·  CONTROL-PLANE (versioned events)  │
└────────────────────────────┬─────────────────────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────┐
│              State Transition Coordinator (FSM)              │
│  ACTIVE→CHECKPOINT→SHADOW→SYNC→VERIFY→QUIESCE→FINAL_DELTA→   │
│  CUTOVER→OBSERVE→COMMITTED  |  ABORT / DEGRADED / RECOVER    │
└───────┬───────────────┬───────────────┬──────────────────────┘
        │               │               │
   Journal+Fence   Verifier+Manifest  Receipts+Safety gate
        │               │               │
        └────── Replay / Chunk sync ────┘
                             │
              PlatformAdapter + StorageAdapter
         (local · railway · fly · docker · k8s · nomad)
                             │
              ENGINE_PROVIDED Portable State Archive
              (escrow · clone · fire-drill · DR plan)
```

## Core modules (`internal/`)

| Package | Role |
|---------|------|
| `coordinator` | Durable FSM cutover orchestration |
| `journal` | Append-only mutation log + CRC / fencing |
| `fence` / `epoch` | Stale-coordinator prevention |
| `verifier` / `manifest` / `chunk` | Consistency + differential transfer |
| `planner` / `converge` / `adaptive` / `heat` | Migration readiness + sync tuning |
| `safegate` / `guardian` / `observe` | Pre-cutover gate + post-cutover window |
| `rollback` / `recover` / `recovery` | Classified rollback + forward recovery |
| `archive` / `escrow` / `dr` / `recoverypoint` | PSA + DR |
| `embed` / `events` / `receipt` / `store` | Integration surface |
| `adapter/*` | Platform bindings (Railway adapter-only) |

## Portability classes

| Class | Meaning |
|-------|---------|
| `RAILWAY_NATIVE` | Platform-owned volume backups as Railway documents them |
| `ENGINE_PROVIDED` | SDE PSA / clone / escrow — independent of one project/env |

Core never embeds Railway GraphQL types. Live dual-service wiring remains an
explicit adapter milestone (offline default `ErrNotWired` until `SDE_RAILWAY_LIVE=1`).

## Safety invariants

See [`SAFETY_INVARIANTS.md`](SAFETY_INVARIANTS.md). Enforced in
`internal/invariants` and chaos/fuzz suites.

## Related docs

- [`INTEGRATION_API.md`](INTEGRATION_API.md) · [`STATE_PROTOCOL.md`](STATE_PROTOCOL.md)
- [`MIGRATION_PROTOCOL.md`](MIGRATION_PROTOCOL.md) · [`RAILWAY_GAP_MATRIX.md`](RAILWAY_GAP_MATRIX.md)
- [`FAILURE_ATLAS.md`](FAILURE_ATLAS.md) · [`ACQUISITION_DEMO.md`](ACQUISITION_DEMO.md)
