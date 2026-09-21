# Release Notes — v1.1.1 (FROZEN)

**Date:** 2026-09-21  
**Codename:** FROZEN  
**Prior:** v1.1.0 READY · v1.0.0 STABLE  
**Verdict:** Acquisition **READY** — this tag freezes the diligence surface for outreach.

Independent project — **not affiliated with Railway**.  
License: Proprietary Pre-Acquisition (`LICENSE`). No production use without a written agreement.

---

## Why this release

v1.1.1 is the **acquisition freeze** of Stateful Deployments Engine: polished proof media,
operator onboarding, live Railway GraphQL wiring (opt-in), Cosign-signed release pipeline,
executed sole-author IP declaration, expanded diligence documentation, and a public site.

**Use this tag as the evaluation baseline.** Subsequent work should be additive patches
unless an acquirer requests a different cut.

---

## Summary for busy readers

| Question | Answer |
|----------|--------|
| What is it? | Transactional state lifecycle engine for volume-backed workloads |
| Why Railway? | Preferred first adapter/acquirer — documented volume mount constraints |
| Replaces backups? | No — complements volumes/backups/PITR/dumps/resize |
| Offline eval? | Yes — default path needs no Railway credentials |
| Live GraphQL? | Opt-in via `SDE_RAILWAY_LIVE=1` |
| Ownership? | Sole-author declaration **EXECUTED** |
| Binaries signed? | Cosign keyless on this release |
| Diligence start | `acquisition/EVALUATOR_BRIEF.md` |

---

## Highlights

### Product
- Transactional deploy FSM: checkpoint → sync → verify → barrier → final Δ → cutover → observe → commit
- Portable State Archives (**ENGINE_PROVIDED**) with fire drills, escrow, and recovery receipts
- Embed modes: CLI · LIBRARY · CONTROL-PLANE (`internal/embed`)
- Railway adapter: offline ENGINE_PROVIDED by default; live GraphQL when `SDE_RAILWAY_LIVE=1`
- Safety invariants enforced in tests (no silent loss of acknowledged writes; UNKNOWN ≠ VERIFIED)

### Operator UX
- `sde doctor` / `sde doctor --json --probe`
- `sde railway status --probe`
- `scripts/install.ps1` · `scripts/install.sh` · `make install` · `make doctor`
- [`docs/GETTING_STARTED.md`](docs/GETTING_STARTED.md)

### Proof & polish
- Demo GIFs with readable pacing and **end-frame freeze** (~3s) before loop
- Interactive Pages demo freezes on COMMITTED / recovery / clone receipt
- Measured local cutover transcript: barrier **~16 ms**, write pause **~15 ms**, 129 mutations
- Site media frames, captions, and FROZEN chips

### Acquisition documentation (expanded in this freeze)
- `acquisition/EVALUATOR_BRIEF.md` — 15–45 minute script  
- `acquisition/ACQUISITION_FAQ.md` — buyer Q&A  
- `acquisition/INTEGRATION_PLAYBOOK.md` — OEM embedding phases  
- `acquisition/RISK_REGISTER.md` — residual risks  
- Thickened overview, technical DD, security review, Railway plan, transfer checklist  
- Ordered index: `acquisition/DUE_DILIGENCE_INDEX.md`

### Security & IP
- Cosign **keyless OIDC** signing on `v*` tags (`.github/workflows/release.yml`)
- Executed Sole Author IP Declaration (`acquisition/legal/`)
- SPDX SBOM: `acquisition/sbom/sde-1.1.1.spdx.json`
- Outreach pack + sent record for Railway BD (`team@railway.com`)

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
| Evaluator brief | `acquisition/EVALUATOR_BRIEF.md` |
| Integration playbook | `acquisition/INTEGRATION_PLAYBOOK.md` |
| Risk register | `acquisition/RISK_REGISTER.md` |
| Readiness | `ACQUISITION_READINESS_REPORT.md` |
| Outreach | `acquisition/OUTREACH_RAILWAY.md` |

---

## Release artifacts

Attached to this GitHub Release:

| Artifact | Purpose |
|----------|---------|
| `sde-linux-amd64` | Linux binary |
| `sde-windows-amd64.exe` | Windows binary |
| `sde-darwin-amd64` | macOS Intel |
| `sde-darwin-arm64` | macOS Apple Silicon |
| `SHA256SUMS.txt` | Checksums |
| `*.cosign.bundle` | Cosign keyless signatures |

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

### Artifact verification

```bash
# download release assets, then:
cosign verify-blob --bundle sde-linux-amd64.cosign.bundle sde-linux-amd64
sha256sum -c SHA256SUMS.txt
```

---

## Changelog since v1.0.0 (rolled up)

### v1.0.0 STABLE
- Offline ENGINE_PROVIDED production core, Pages demo, evaluate harness
- Live Railway GraphQL client, Cosign pipeline, executed ownership declaration

### v1.1.0 READY
- `sde doctor`, install scripts, getting started, railway status probe
- Stale OPEN/`ErrNotWired` doc claims cleared; evaluate scripts fixed to current version

### v1.1.1 FROZEN (this tag)
- Proof media freeze pacing + site polish
- Expanded acquisition package documentation
- Diligence freeze designation for outreach baseline

Full history: [`CHANGELOG.md`](CHANGELOG.md)

---

## Breaking / non-goals

- **Not** a Railway product or endorsement claim
- Does **not** replace Railway volumes/backups/Postgres PITR/dumps/resize
- Offline default still returns `ErrNotWired` until `SDE_RAILWAY_LIVE=1`
- Networking / tunnels / mesh remain out of scope
- Deal templates in `legal-review/` remain attorney-review drafts (ownership declaration is executed)
- No ARR, customers, or valuation asserted in public materials

---

## Known follow-ups (non-blocking)

1. Live dual-service pilot with production Railway credentials  
2. Counsel review / execution of definitive APA  
3. Go toolchain patch upgrades per `SECURITY_GOVULNCHECK.txt`  

See [`acquisition/RISK_REGISTER.md`](acquisition/RISK_REGISTER.md).

---

## Upgrade notes (1.1.0 → 1.1.1)

- Freeze + documentation + proof polish; no intentional public API breaks
- Re-run `scripts/generate-demo-gifs.py` if regenerating proof media locally
- Prefer this tag for diligence links shared with acquirers

---

## Links

- Repository: https://github.com/theworker02/stateful-deployments-engine  
- Release: https://github.com/theworker02/stateful-deployments-engine/releases/tag/v1.1.1  
- Live demo: https://theworker02.github.io/stateful-deployments-engine/  
- Proof: https://theworker02.github.io/stateful-deployments-engine/proof.html  
- Diligence: [`acquisition/DUE_DILIGENCE_INDEX.md`](acquisition/DUE_DILIGENCE_INDEX.md)  
- Changelog: [`CHANGELOG.md`](CHANGELOG.md)
