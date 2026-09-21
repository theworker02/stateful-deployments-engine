# Release Notes — v1.1.0 (READY)

**Date:** 2026-09-21  
**Codename:** READY  
**Prior:** v1.0.0 STABLE

Polish release focused on **ease of access**, stale-claim cleanup, and operator UX.
Acquisition verdict remains **READY**.

## Highlights

- **`sde doctor`** — local workspace + Railway env readiness (JSON + human)
- **`sde railway status [--probe]`** — live-mode config + optional GraphQL token validation
- **Install scripts** — `scripts/install.ps1` / `scripts/install.sh` (+ `make install` / `make doctor`)
- **Getting started** — [`docs/GETTING_STARTED.md`](docs/GETTING_STARTED.md)
- Docs/FAQ/README no longer claim live GraphQL is `ErrNotWired` / OPEN
- Evaluate harness scripts write **1.1.0** (were stale at 0.5.0)

## Verify

```bash
./scripts/install.sh          # or .\scripts\install.ps1
./dist/sde doctor
./dist/sde version            # sde 1.1.0
go test ./...
```

## Signed releases

Tag `v1.1.0` triggers Cosign keyless signing via `.github/workflows/release.yml`.
