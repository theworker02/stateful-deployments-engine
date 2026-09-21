# Railway Integration Architecture

**Independent project — not affiliated with Railway.**

## Two layers

### PUBLIC API TODAY (evaluable without Railway credentials)

- Local adapter + PSA archive/escrow/DR/fire-drill
- CLI: `deploy`, `export`, `restore`, `railway clone` (local simulation of clone semantics)
- `go test ./...`, `./evaluate.sh` / `evaluate.ps1`
- Docs: gap matrix, failure atlas, acquisition demo

### INTERNAL PLATFORM (needs Railway access)

- Live GraphQL / service clone / domain cutover behind `SDE_RAILWAY_LIVE=1`
- Dual-service + dual-volume orchestration (`internal/adapter/railway`)
- Still adapter-only — core never imports Railway SDKs

## Value prop (correct)

SDE is a **general-purpose state lifecycle engine for arbitrary volume-backed workloads**:
deployment-safe state mobility, independent recovery artifacts, verified restoration/migration
beyond DB-specific tooling.

Railway already has volumes, backups, Postgres PITR, logical dumps, resize. SDE does not pitch
“Railway lacks backups.” SDE pitches **ENGINE_PROVIDED** portability and transactional cutover
for generic FS/state workloads under documented volume constraints (one volume/service, no volume replicas, wipe deletes volume backups).

## Data flow (conceptual)

```
App writes ──► journal/agent ──► active volume
                      │
              shadow sync/replay
                      │
         verify → barrier → cutover → observe → commit
                      │
         optional: PSA export → escrow (ENGINE_PROVIDED)
```

## RAILWAY_NATIVE vs ENGINE_PROVIDED

| | RAILWAY_NATIVE | ENGINE_PROVIDED |
|--|----------------|-----------------|
| Coupling | Volume / project lifecycle | Archive directory / object store |
| Cross-env | Constrained by platform rules | Designed for offline / escrow / new target |
| Endorsement | Platform feature | Not Railway-endorsed |
