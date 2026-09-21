# Phase 5 Report — Acquisition Candidate (v0.5.0)

**Date:** 2026-09-21  
**Mindset:** Diligence density over feature count. Independent project — not affiliated with Railway.

## Correct value prop (locked)

SDE is a **general-purpose state lifecycle engine for arbitrary volume-backed workloads**:
deployment-safe state mobility, independent ENGINE_PROVIDED recovery artifacts, verified
restoration/migration beyond DB-specific tooling.

Railway already has volumes, backups, Postgres PITR, dumps, resize. Documented constraints
(same-project/env restore; wipe deletes volume backups) remain relevant. **Do not** pitch
“Railway doesn’t have backups.”

## What landed (25-section map)

| # | Deliverable | Path / notes |
|---|-------------|--------------|
| 1 | Gap matrix | `docs/RAILWAY_GAP_MATRIX.md` |
| 2 | Volume fixtures | `workloads/fixtures.go` |
| 3–4 | Railway clone + receipt | `internal/adapter/railway/clone.go`, `sde railway clone` |
| 5–7 | Integration API + events + embed | `docs/INTEGRATION_API.md`, `schemas/events/v1`, `internal/embed` |
| 8 | Railway integration architecture | `docs/RAILWAY_INTEGRATION_ARCHITECTURE.md` |
| 9 | Failure atlas | `docs/FAILURE_ATLAS.md` |
| 10 | Journal fuzz | `internal/journal/fuzz_crash_test.go` |
| 11 | Perf gates | `benchmarks/perf_gates_test.go` |
| 12 | Evaluate harness | `evaluate.ps1`, `evaluate.sh`, `make evaluate`, `cmd/evaluate-harness` |
| 13 | Acquisition demo script | `docs/ACQUISITION_DEMO.md` |
| 14 | Pages upgrade | `site/*` failure-atlas, railway-fit, security, due-diligence |
| 15 | README | tech-first v0.5.0 |
| 16 | Brand guidelines | `docs/BRAND_GUIDELINES.md` |
| 17–19 | SBOM / IP / provenance / assets | `acquisition/sbom`, IP_MANIFEST, PROVENANCE, ASSET_REGISTER |
| 20–22 | Integration plan / transfer / legal drafts | `acquisition/*` |
| 23 | Security review | `acquisition/SECURITY_REVIEW.md` |
| 24 | Release eng | CHANGELOG, RELEASE_NOTES_v0.5.0.md |
| 25 | Readiness verdict | `ACQUISITION_READINESS_REPORT.md` |

## Verification

- `go test ./...` — PASS
- `go run ./cmd/evaluate-harness` — PASS
- `sde version` — `sde 0.5.0`

## Explicit non-goals held

- No Agent 3 networking
- No fake Railway endorsement
- No fabricated valuation / signed-release claims
- Proprietary license preserved
