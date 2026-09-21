# Product overview (acquisition) — v1.1.0

**Stateful Deployments Engine (SDE)** is an independent, proprietary state lifecycle
engine for arbitrary volume-backed workloads.

It is **not affiliated with Railway** and does not claim Railway endorsement.

## What buyers evaluate

| Capability | Evidence |
|------------|----------|
| Transactional stateful cutover | `internal/coordinator`, `sde demo`, chaos suite |
| Mutation journal + fencing | `internal/journal`, `internal/fence`, fuzz tests |
| Portable State Archives | `internal/archive`, independence restore test |
| Fire drills / DR planner | `internal/dr`, `sde fire-drill`, `sde readiness` |
| Embeddable control-plane API | `internal/embed`, versioned events |
| Railway fit (adapter-only) | `docs/RAILWAY_GAP_MATRIX.md`, ENGINE_PROVIDED clone |
| Live demo | GitHub Pages `site/` (labeled simulation) |
| One-command diligence | `evaluate.ps1` → `evaluation/` |

## Correct pitch

Railway already provides volumes, backups, Postgres PITR, logical dumps, and live resize.
SDE supplies **general-purpose** deployment-safe state mobility, independently escrowed
recovery artifacts, and verified restoration for **generic filesystem state** — not a
replacement for Railway-native DB backups.

## Version & verdict

- **v1.1.0 READY** (offline ENGINE_PROVIDED + live GraphQL wiring + operator UX)
- Acquisition verdict: **READY_WITH_DISCLOSED_ITEMS** — see `ACQUISITION_READINESS_REPORT.md`

## Package index

See [`README.md`](README.md) in this folder for the full diligence file list.
