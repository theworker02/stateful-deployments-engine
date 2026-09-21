# Acquisition Readiness Report

**Product:** Stateful Deployments Engine (SDE)  
**Version:** v1.1.1 FROZEN  
**Date:** 2026-09-21  
**Verdict:** **READY** (surface frozen at this tag for diligence)

Independent project — not affiliated with Railway. No vanity numeric score.

## Previously disclosed items — resolved

| Item | Resolution |
|------|------------|
| Railway GraphQL live adapter | **FIXED** — live client at `internal/adapter/railway` (`SDE_RAILWAY_LIVE=1`); httptest-tested; see `docs/RAILWAY_LIVE.md` |
| Release signing | **FIXED** — Cosign keyless signing in `.github/workflows/release.yml` + `scripts/sign-release.*`; checksums + cosign bundles on tagged releases |
| Legal / ownership | **FIXED** — Sole Author IP Declaration **EXECUTED** at `acquisition/legal/SOLE_AUTHOR_IP_DECLARATION.md` (+ `EXECUTED_OWNERSHIP_RECORD.json`). Commercial purchase templates remain in `legal-review/` for deal counsel |

## v1.1.0 operator UX (this release)

| Item | Status |
|------|--------|
| `sde doctor` / install scripts / getting started | Shipped |
| `sde railway status --probe` | Shipped |
| Stale OPEN/`ErrNotWired` doc claims | Cleared |

## Dimensions

| Dimension | Assessment | Evidence |
|-----------|------------|----------|
| FUNCTIONAL | Ready | Tests, demo, evaluate, live Railway GraphQL wiring |
| PERFORMANCE | Ready | Perf gates + evaluation benchmarks |
| SECURITY | Ready | SECURITY_REVIEW + govulncheck log + signed release pipeline |
| DOCUMENTATION | Ready | GETTING_STARTED + docs/ + site + RAILWAY_LIVE |
| IP PROVENANCE | Ready | Executed sole-author declaration + SBOM 1.1.0 |
| DEPENDENCIES | Ready | SPDX 1.1.0 |
| BRANDING | Ready | Original brand assets |
| BUILD REPRODUCIBILITY | Ready | Go modules + Makefile + install scripts + signed release workflow |
| PLATFORM INTEGRATION | Ready | Public API + live GraphQL adapter + status probe |
| TRANSFERABILITY | Ready | TRANSFER_CHECKLIST + executed ownership record |
| EASE OF ACCESS | Ready | install → doctor → demo → evaluate path |

## Optional follow-ups (non-blocking)

- Run a live dual-service pilot with production Railway credentials  
- Counsel review of definitive asset-purchase agreement templates  
- Upgrade CI Go toolchain when patch releases address stdlib advisories  

## Diligence entrypoint

[`acquisition/DUE_DILIGENCE_INDEX.md`](acquisition/DUE_DILIGENCE_INDEX.md)
