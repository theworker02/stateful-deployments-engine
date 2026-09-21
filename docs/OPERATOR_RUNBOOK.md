# Operator Runbook

Day-2 operations for the Stateful Deployments Engine (evaluation / lab use under proprietary license).

## Workspace layout

```
.sde/
  journal/          # append-only mutation log
  slots/            # active + shadow + standby state trees
  meta.json         # local adapter slot registry
  fence.json        # epoch + fencing token
  fsm/snapshot.json # durable deploy FSM cursor
  receipts/         # sealed DeploymentReceipt JSON
```

## Common commands

```bash
sde init
sde plan-migration --image app:2
sde deploy --image app:2
sde status
sde verify
sde rollback
sde recover
sde chaos --list   # if wired
```

## Pre-cutover checklist

1. `plan-migration` conclusion is `READY` or `READY_WITH_RISK` (never treat `UNKNOWN` as ready).
2. Storage capability matrix includes at least `CHECKSUM` + `FREEZE`.
3. Candidate health check passes.
4. Safety gate decision is `CUTOVER_ALLOWED`.
5. Standby / rollback point present.

## Crash during deploy

| Durable phase | Typical action |
|---------------|----------------|
| SYNCHRONIZING / VERIFYING | Resume forward recovery |
| QUIESCING / FINAL_DELTA | Unfreeze if needed; abort or resume |
| CUTOVER | Classify — may retry forward |
| OBSERVING (no new writes) | Rollback if standby intact |
| OBSERVING (writes accepted) | Manual reconciliation — do **not** silent rollback |

See `internal/recover` and `docs/PROTOCOL.md`.

## Sealed evidence

Every successful or failed deploy should leave a receipt under `receipts/`:

- JSON: machine-readable (`schemas/deployment-receipt.schema.json`)
- Evidence hash: tamper-evident seal (`internal/receipt`)

## Escalation

Production use requires a written Agreement. Security issues: `SECURITY.md`.
