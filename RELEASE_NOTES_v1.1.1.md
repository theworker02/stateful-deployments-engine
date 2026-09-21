# Release Notes — v1.1.1 (FROZEN)

**Date:** 2026-09-21  
**Codename:** FROZEN  
**Prior:** v1.1.0 READY  
**Verdict:** Acquisition **READY** — this tag freezes the diligence surface for outreach.

Independent project — **not affiliated with Railway**.  
License: Proprietary Pre-Acquisition (`LICENSE`). No production use without a written agreement.

---

## Why this release

v1.1.1 is the **acquisition freeze** of Stateful Deployments Engine: polished proof media,
operator onboarding, live Railway GraphQL wiring (opt-in), Cosign-signed release pipeline,
executed sole-author IP declaration, and a public diligence site.

Use this tag as the evaluation baseline. Subsequent work should be additive patches unless
an acquirer requests a different cut.

---

## Highlights

### Product
- Transactional deploy FSM: checkpoint → sync → verify → barrier → final Δ → cutover → observe → commit
- Portable State Archives (**ENGINE_PROVIDED**) with fire drills, escrow, and recovery receipts
- Embed modes: CLI · LIBRARY · CONTROL-PLANE (`internal/embed`)
- Railway adapter: offline ENGINE_PROVIDED by default; live GraphQL when `SDE_RAILWAY_LIVE=1`

### Operator UX
- `sde doctor` / `sde doctor --json --probe`
- `sde railway status --probe`
- `scripts/install.ps1` · `scripts/install.sh` · `make install` · `make doctor`
- [`docs/GETTING_STARTED.md`](docs/GETTING_STARTED.md)

### Proof & polish
- Demo GIFs with slower pacing and **end-frame freeze** (~3s) before loop
- Interactive Pages demo freezes on COMMITTED / recovery / clone receipt
- Measured local cutover transcript: barrier **~16 ms**, write pause **~15 ms**, 129 mutations
- Site media frames, captions, and FROZEN chips

### Security & IP
- Cosign **keyless OIDC** signing on `v*` tags (`.github/workflows/release.yml`)
- Executed Sole Author IP Declaration (`acquisition/legal/`)
- SPDX SBOM: `acquisition/sbom/sde-1.1.1.spdx.json`

---

## What’s included (surface area)

| Area | Path / command |
|------|----------------|
| CLI | `cmd/sde` — `sde version` → `sde 1.1.1` |
| Evaluate | `evaluate.ps1` / `evaluate.sh` / `make evaluate` |
| Live Railway | `docs/RAILWAY_LIVE.md`, `SDE_RAILWAY_LIVE=1` |
| Pages demo | https://theworker02.github.io/stateful-deployments-engine/ |
| Proof pack | https://theworker02.github.io/stateful-deployments-engine/proof.html |
| Diligence index | `acquisition/DUE_DILIGENCE_INDEX.md` |
| Readiness | `ACQUISITION_READINESS_REPORT.md` |
| Outreach | `acquisition/OUTREACH_RAILWAY.md` (sent record on file) |

---

## Install / verify

```powershell
git checkout v1.1.1
.\scripts\install.ps1
.\dist\sde.exe version          # sde 1.1.1
.\dist\sde.exe doctor
.\dist\sde.exe demo --root .sde
.\evaluate.ps1
go test ./...
```

```bash
git checkout v1.1.1
./scripts/install.sh
./dist/sde version
./dist/sde doctor
./evaluate.sh
go test ./...
```

### Artifact verification (after CI attaches binaries)

```bash
# download release assets, then:
cosign verify-blob --bundle sde-linux-amd64.cosign.bundle sde-linux-amd64
sha256sum -c SHA256SUMS.txt
```

---

## Breaking / non-goals

- **Not** a Railway product or endorsement claim
- Does **not** replace Railway volumes/backups/Postgres PITR/dumps/resize
- Offline default still returns `ErrNotWired` until `SDE_RAILWAY_LIVE=1`
- Networking / tunnels / mesh remain out of scope
- Deal templates in `legal-review/` remain attorney-review drafts (ownership declaration is executed)

---

## Upgrade notes (1.1.0 → 1.1.1)

- Version/codename only for freeze + polish; no intentional API breaks
- Re-run `scripts/generate-demo-gifs.py` if regenerating proof media locally
- Prefer this tag for diligence links shared with acquirers

---

## Links

- Repository: https://github.com/theworker02/stateful-deployments-engine
- Release tag: https://github.com/theworker02/stateful-deployments-engine/releases/tag/v1.1.1
- Live demo: https://theworker02.github.io/stateful-deployments-engine/
- Changelog: [`CHANGELOG.md`](CHANGELOG.md)
