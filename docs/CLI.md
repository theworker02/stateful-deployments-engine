# CLI reference (v1.1.0)

Binary: `sde` — version from `internal/version` → **1.1.0**.

```bash
go build -o sde.exe ./cmd/sde
./sde.exe --help
```

Global flag: `--root` (default `.sde`) — journal, slots, fence, deployments, catalog.

## Workspace

| Command | Purpose |
|---------|---------|
| `init` | Create local workspace |
| `doctor [--json] [--probe]` | Install / env readiness |
| `status` | Active slot, epoch, journal head |
| `readiness` | DR / recovery scorecard |
| `modules` | Package surface listing |
| `version` | `sde 1.1.0` |
| `completion` | Shell completions |

## Deployment transaction

| Command | Purpose |
|---------|---------|
| `plan` / `plan-migration` | Readiness plan (never treats UNKNOWN as READY) |
| `deploy [--dry-run] [--max-write-pause]` | Transactional cutover |
| `verify` | Consistency verification receipt |
| `cutover` | Explicit cutover step when split |
| `rollback` / `recover` | Classified recovery |

## Archives & DR

| Command | Purpose |
|---------|---------|
| `export` / `inspect-archive` / `verify-archive` / `import` | PSA |
| `restore --archive [--epoch]` | Restore to target |
| `escrow` / catalog (`archives`) | Escrow + catalog prune |
| `recovery-points` / `dr-plan` / `fire-drill` / `clone` | DR |
| `railway clone` | ENGINE_PROVIDED env clone + `CLONE_RECEIPT.json` |
| `railway status [--probe]` | Live-mode config + optional GraphQL validate |

## Ops

`demo` · `chaos` · `benchmark`

## Related

[`GETTING_STARTED.md`](GETTING_STARTED.md) · [`QUICKSTART.md`](QUICKSTART.md) · [`INTEGRATION_API.md`](INTEGRATION_API.md)
