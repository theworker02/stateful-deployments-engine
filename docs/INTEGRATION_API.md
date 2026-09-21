# Integration API

Public surface for embedding SDE as a **library**, **CLI**, or **control-plane** component.

**License:** Proprietary Pre-Acquisition (`LICENSE`). Independent project; not affiliated with Railway.

## Modes

| Mode | Entry | Use |
|------|-------|-----|
| CLI | `cmd/sde` | Operator demos, evaluate harness, fire drills |
| LIBRARY | `sdk/go/sdesdk` + selected `internal/` for engine hosts | App journaling; embedders that vendor under license |
| CONTROL-PLANE | Host process owns slots/health; calls `coordinator` + adapters | Platform OEM integration (requires platform credentials) |

Headless library path: construct `coordinator.NewFull(...)` with a `Platform`/`StorageAdapter`, journal, fence, store — no TTY required.

## Deployment transaction API (stable shapes)

### Start

```go
report, err := coord.Deploy(ctx, coordinator.Config{
    ImageRef: "app:next",
    MaxWritePause: 250 * time.Millisecond,
    DryRun: false,
})
```

### Observe

- Persisted FSM: `.sde/deployments/<id>/state.json`
- Events: versioned schema in `schemas/events/` and `internal/events`
- Receipts: deployment + recovery JSON under deploy/receipt dirs

### Recover

```go
dec, report, err := coord.Recover(ctx, deployID)
```

### Archive / restore (ENGINE_PROVIDED)

```go
m, stats, err := archive.Export(stateRoot, outDir, epoch, journalPos, journalPath)
_, err = archive.Verify(outDir)
_, err = archive.Import(outDir, destRoot)
rcpt, err := recovery.Restore(ctx, outDir, target, receiptDir, sourceDestroyed)
```

## Versioning

- Engine version: **1.1.0**
- Event schema: `schemas/events/v1/` (`schema_version: 1`)
- Breaking changes require a minor/major bump and CHANGELOG entry

## Non-goals in the public API

- Cross-region networking / tunnels / mesh (Agent 3)
- Implying Railway endorsement of ENGINE_PROVIDED archives
- Enabling live GraphQL without explicit `SDE_RAILWAY_LIVE=1` (offline default returns `ErrNotWired`)
