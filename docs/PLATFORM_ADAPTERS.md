# Platform adapters (v1.1.0)

## Contracts

### PlatformAdapter

`create_candidate`, `attach_state`, `start_candidate`, `health_check`,
`prepare_cutover`, `activate_candidate`, `deactivate_previous`, `destroy_candidate`,
plus traffic/barrier helpers as implemented by each package.

### StorageAdapter

`checkpoint`, `journal`, `sync`, `verify`, `freeze_writes`, `unfreeze_writes`,
`rollback`, and **capability negotiation**.

## Implementations

| Adapter | Status |
|---------|--------|
| `local` | Full offline evaluation / demos |
| `railway` | ENGINE_PROVIDED clone/archive + live GraphQL when `SDE_RAILWAY_LIVE=1` |
| `fly` / `docker` / `kubernetes` / `nomad` | Scaffolded contracts + capability matrices |

## Capability negotiation

Adapters declare: `SNAPSHOT`, `COPY_ON_WRITE`, `CHANGE_TRACKING`, `FREEZE`,
`ATOMIC_RENAME`, `BLOCK_READ`, `BLOCK_WRITE`, `CHECKSUM`, `INCREMENTAL_SNAPSHOT`.

Core selects strategies from **actual** capabilities — no fake universal abstraction.

## Railway boundary

See [`RAILWAY_INTEGRATION_ARCHITECTURE.md`](RAILWAY_INTEGRATION_ARCHITECTURE.md)
and [`RAILWAY_GAP_MATRIX.md`](RAILWAY_GAP_MATRIX.md). Distinguish `RAILWAY_NATIVE`
vs `ENGINE_PROVIDED` in every public claim.
