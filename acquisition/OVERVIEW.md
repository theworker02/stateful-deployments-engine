# Product overview (acquisition) — v1.1.1 FROZEN

**Stateful Deployments Engine (SDE)** is an independent, proprietary state lifecycle
engine for arbitrary volume-backed workloads.

It is **not affiliated with Railway** and does not claim Railway endorsement.

**Freeze tag:** `v1.1.1` — use this tag as the diligence baseline.

## One-paragraph thesis

Platforms that attach persistent volumes still force a trade-off: redeploy downtime
or abandon local filesystem state. SDE provides a **transactional state transition
coordinator** — mutation journal → sync → content-addressed verify → write barrier →
final delta → traffic cutover → standby rollback — so stateful workloads get blue/green
semantics without concurrent mounts of a single volume.

## What buyers evaluate

| Capability | Evidence |
|------------|----------|
| Transactional stateful cutover | `internal/coordinator`, `sde demo`, chaos suite |
| Mutation journal + fencing | `internal/journal`, `internal/fence`, fuzz tests |
| Portable State Archives | `internal/archive`, independence restore test |
| Fire drills / DR planner | `internal/dr`, `sde fire-drill`, `sde readiness` |
| Embeddable control-plane API | `internal/embed`, versioned events |
| Railway fit (adapter) | Live GraphQL when `SDE_RAILWAY_LIVE=1`; ENGINE_PROVIDED clone offline |
| Operator onboarding | `sde doctor`, install scripts, `docs/GETTING_STARTED.md` |
| Live demo + proof | GitHub Pages `site/` + `assets/demo/` GIFs/transcripts |
| One-command diligence | `evaluate.ps1` → `evaluation/` |
| Signed supply chain | Cosign keyless on `v*` tags |
| Ownership | Executed sole-author declaration |

## Correct pitch

Railway already provides volumes, backups, Postgres PITR, logical dumps, and live resize.
SDE supplies **general-purpose** deployment-safe state mobility, independently escrowed
recovery artifacts, and verified restoration for **generic filesystem state** — not a
replacement for Railway-native DB backups.

## What a buyer acquires

1. Protocol + Go reference implementation of the State Transition Coordinator
2. Local + Railway adapters (Fly/Docker/K8s/Nomad scaffolds)
3. CLI, evaluate harness, embed library, event schemas
4. Brand assets, docs, Pages diligence site, proof media
5. Executed sole-author IP attestation + transfer inventory
6. Cosign-signed multi-platform release binaries at `v1.1.1`

## Version & verdict

- **v1.1.1 FROZEN** — acquisition surface locked for outreach
- Verdict: **READY** — see `../ACQUISITION_READINESS_REPORT.md`
- Prior: v1.1.0 READY → polish/freeze; v1.0.0 STABLE core

## 15-minute evaluator path

1. [`../docs/GETTING_STARTED.md`](../docs/GETTING_STARTED.md)
2. [`../docs/ACQUISITION_DEMO.md`](../docs/ACQUISITION_DEMO.md)
3. [`EVALUATOR_BRIEF.md`](EVALUATOR_BRIEF.md)
4. Run `evaluate.ps1` / `evaluate.sh`
5. Open https://theworker02.github.io/stateful-deployments-engine/proof.html

## Package index

See [`README.md`](README.md) and [`DUE_DILIGENCE_INDEX.md`](DUE_DILIGENCE_INDEX.md).
