# v1.0.0 Release Report

## Goal
Ship **STABLE** offline ENGINE_PROVIDED product with a real GitHub Pages live demo and closed diligence gaps from v0.5.0.

## Delivered

| Item | Status |
|------|--------|
| Version 1.0.0 everywhere (`internal/version`) | Done |
| Interactive Pages live demo | Done (`site/demo.html`) |
| Architecture explorer | Done |
| Brand self-contained under `site/brand/` | Done |
| Pages workflow copies assets + bench JSON | Done |
| govulncheck evidence | Done |
| SHA-256 checksums | Done |
| README / CHANGELOG / release notes | Done |
| Readiness verdict | READY_WITH_DISCLOSED_ITEMS |

## Gaps intentionally remaining

- Live Railway GraphQL dual-service adapter (`ErrNotWired`)
- Cosign/GPG signed releases (checksums only)
- Executed legal agreements (drafts only)

## Verify

```
go test ./...
go build -o dist/sde.exe ./cmd/sde
./dist/sde.exe version
```
