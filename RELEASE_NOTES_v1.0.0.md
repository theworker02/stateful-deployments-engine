# Release Notes — v1.0.0 (STABLE)

**Date:** 2026-09-21  
**Designation:** First stable release of the Stateful Deployments Engine.

Independent project — not affiliated with Railway.

## Highlights

- **Stable product version** with single authoritative `internal/version` constant
- **GitHub Pages live demo** — interactive cutover + restore simulation, architecture explorer, evidence-backed benchmarks page (`site/`)
- **Acquisition candidate → stable**: evaluate harness, failure atlas, portable archives, fire drills, embed API remain first-class
- **govulncheck** executed and recorded (`acquisition/SECURITY_GOVULNCHECK.txt`)
- **Release checksums** for `dist/sde.exe` in `dist/SHA256SUMS.txt`

## Correct Railway framing

Not “Railway lacks backups.” Railway provides volumes, backups, Postgres PITR, logical dumps, live resize.
SDE is a **general-purpose state lifecycle engine** for arbitrary volume-backed workloads with
ENGINE_PROVIDED portable recovery and transactional cutover.

## Known limitations (disclosed)

| Item | Status |
|------|--------|
| Railway GraphQL live dual-service adapter | OPEN (`ErrNotWired`) — offline ENGINE_PROVIDED path complete |
| Go stdlib vulns on toolchain go1.26.3 | Documented; recommend upgrading Go patch release |
| Cryptographic release signing | Checksums only — not claimed signed |
| Legal assignment | Drafts for attorney review only |

## Verify

```bash
go test ./...
go build -o dist/sde.exe ./cmd/sde
./dist/sde.exe version   # sde 1.0.0
.\evaluate.ps1
```

Open `site/index.html` for the live demo locally, or enable GitHub Pages via `.github/workflows/pages.yml`.
