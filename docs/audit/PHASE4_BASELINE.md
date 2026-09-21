# Phase 4 Baseline Audit

**Date:** 2026-09-21  
**Auditor:** Agent 2 (Stateful Deployments Engine)  
**Repo:** `github.com/theworker02/stateful-deployments-engine`  
**Baseline version observed:** `0.1.0-dev` / mid Phase 2–3 implementation (no PHASE2/PHASE3 reports yet)

## 1. Repository inventory (snapshot)

| Area | Status |
|------|--------|
| Core packages (`journal`, `coordinator`, `verifier`, `rollback`, `replay`, `agent`) | Present; Phase 2/3 evolved |
| Phase 2 additions (`fence`, `store`, `sync`, `chaos`, `receipt`, `guardian`, `safegate`) | Present |
| Phase 3 additions (`planner`, `converge`, `chunk`, `hotstate`, `adaptive`) | Present |
| Adapters (`local`, `railway`) | Present; Railway is stub/`ErrNotWired` |
| CLI `cmd/sde` | Present (plan, deploy, verify, recover, chaos, benchmark, demo…) |
| `PHASE2_REPORT.md` / `PHASE3_REPORT.md` | **Missing** at audit time |
| Benchmarks/results chaos artifacts | Code present; committed JSON results sparse |
| GitHub Pages `site/` | **Missing** |
| Brand assets | **Missing** |
| `acquisition/` package | Only root `ACQUISITION.md` |
| Portable archive / escrow / DR | **Not yet implemented** (Phase 4 scope) |

## 2. Test suite (pre-Phase-4)

- `go test ./...` was run; `internal/chaos` and journal tests previously green.
- Risk identified: `syncUntilQuiet` could hang under continuous live writers (fixed in Phase 4 prep with 2s deadline).
- Railway adapter has no live integration tests (expected; stub).

## 3. Stubs / mocks / incomplete surfaces

| Component | Reality |
|-----------|---------|
| `internal/adapter/railway` | Typed stub; returns `ErrNotWired` unless `SDE_RAILWAY_LIVE=1` (still unimplemented) |
| Railway GraphQL client | Not present |
| Cross-region networking | **Out of scope** (Agent 3); not present — correct |
| Dashboard / cloud control plane | Not present — correct for engine |
| Application-defined verify callback | Interface present; demos use checksum |

## 4. Documentation vs code

| Claim | Supported? |
|-------|------------|
| Transactional cutover FSM | Yes (`coordinator` + `store`) |
| Mutation journal + epochs/fencing | Yes (`journal`, `fence`) |
| Verification receipts | Yes (`verifier`) |
| Chaos matrix | Yes (`chaos`) |
| Railway live dual-service cutover | **No** — adapter stub only |
| Zero downtime absolute claim | Softened in code (measured `CUTOVER_WRITE_PAUSE_MS`) |
| Fine-grained PITR | **Not** advertised as finer than journal epochs — keep honest in Phase 4 |

## 5. Railway contamination check

- **Core packages do not import** `internal/adapter/railway`.
- Railway types/env vars isolated to `internal/adapter/railway`.
- Docs mention Railway as first production *target*, not as core dependency.
- Phase 4 must keep `RAILWAY_NATIVE` vs `ENGINE_PROVIDED` portability clearly separated.

## 6. Dependency licenses

| Module | License (typical) | Notes |
|--------|-------------------|-------|
| Go standard library | BSD-style | OK |
| `github.com/spf13/cobra` | Apache-2.0 | OK |
| `github.com/spf13/pflag` | BSD-3-Clause | OK (indirect) |
| `github.com/inconshreveable/mousetrap` | Apache-2.0 | OK (indirect) |

No copyleft dependencies observed in `go.mod`. No secrets in tracked files (`.env.example` only).

## 7. Branding / release config

- Version strings historically `0.1.0-dev` / `0.3.0-dev` mid-work — **must unify to v0.4.0**.
- No fake coverage badges observed.
- `LICENSE` proprietary pre-acquisition — preserve.
- `.github/FUNDING.yml` present for `theworker02`.

## 8. Gaps Phase 4 must close

1. Portable State Archive format + CLI
2. Escrow targets (local / S3-compatible / generic object store) — no AWS SDK hardcode in core
3. Continuous recovery points + DR planner + fire drill
4. Recovery independence automated test
5. Recovery catalog, cloning, sensitive-data hooks, receipts
6. Acquisition demo, GitHub Pages site, brand kit
7. Full docs tree + acquisition package + PHASE4_* artifacts
8. Honest PHASE2/PHASE3 reports (backfill)

## 9. Concurrent thickness agent

If additive packages arrive during Phase 4, **absorb and keep**; do not delete. Do not merge Agent 3 networking projects.

## 10. Audit conclusion

Baseline is a working local transactional engine with Phase 2/3 scaffolds and Railway-as-adapter-only. Phase 4 proceeds to add **engine-provided portability and DR** without claiming Railway-native cross-environment backup or endorsement.
