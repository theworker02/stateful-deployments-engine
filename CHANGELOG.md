# Changelog

All notable changes to Stateful Deployments Engine (SDE) are documented here.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).  
SDE is **proprietary** — see `LICENSE`.

## [Unreleased]

### Planned

- Live dual-service pilot runbook with recorded credentials-free dry-run evidence
- Counsel-reviewed definitive APA (templates remain in `legal-review/`)

## [1.1.1] — 2026-09-21

### Added

- Acquisition **FROZEN** designation (`internal/version` codename)
- End-frame freeze on proof GIFs (~3s) + interactive demo freeze on terminal states
- Polished Pages media frames / figcaptions / FROZEN chips
- `RELEASE_NOTES_v1.1.1.md` · SPDX `sde-1.1.1.spdx.json`
- Expanded acquisition pack: `EVALUATOR_BRIEF`, `ACQUISITION_FAQ`, `INTEGRATION_PLAYBOOK`, `RISK_REGISTER`
- Thickened technical DD, security review, Railway plan, transfer checklist, overview

### Changed

- Version → **1.1.1** / codename **FROZEN**
- Demo pacing held for readability; site proof presentation tightened

## [1.1.0] — 2026-09-21

### Added

- `sde doctor` (+ `--json` / `--probe`) — install & Railway env readiness
- `sde railway status [--probe]` — live-mode config + GraphQL token validation
- `scripts/install.ps1` / `scripts/install.sh` — one-command build + doctor
- `make install` / `make doctor`
- `docs/GETTING_STARTED.md` — five-minute onboarding path
- SPDX SBOM `acquisition/sbom/sde-1.1.0.spdx.json`
- `RELEASE_NOTES_v1.1.0.md`
- **Proof media:** `assets/demo/*.gif`, screenshots, CLI transcripts; `site/proof.html`
- `acquisition/OUTREACH_RAILWAY.md` — Railway BD contact pack

### Changed

- Version → **1.1.0** / codename **READY**
- FAQ / README / adapters / integration docs: live GraphQL no longer marked OPEN/`ErrNotWired`-only
- `evaluate.ps1` / `evaluate.sh` write version **1.1.0** (were stale at 0.5.0)
- Site chips and diligence index refreshed for 1.1.0

### Fixed

- Stale “Railway live still OPEN” claims across user-facing docs

## [1.0.0] — 2026-09-21

### Added

- Authoritative `internal/version` (`1.0.0`)
- GitHub Pages **live demo**: interactive FSM cutover/restore, failure inject, architecture explorer
- `site/brand/` self-contained assets for Pages deploy; workflow copies brand + bench JSON
- `govulncheck` evidence: `acquisition/SECURITY_GOVULNCHECK.txt`
- `dist/SHA256SUMS.txt` release checksums
- `RELEASE_NOTES_v1.0.0.md`, readiness refresh for STABLE
- **Live Railway GraphQL client** (`SDE_RAILWAY_LIVE=1`) — service/volume/deploy/barrier/domain ops
- **Cosign keyless release signing** (`.github/workflows/release.yml`, `scripts/sign-release.*`)
- **Executed Sole Author IP Declaration** (`acquisition/legal/SOLE_AUTHOR_IP_DECLARATION.md`)

### Changed

- Designation **STABLE** (offline ENGINE_PROVIDED production-ready)
- README / site / CLI / evaluate harness version strings → **1.0.0**
- Pages site visual redesign (Fraunces / Source Serif / Plex Mono)
- Thickened stub docs (`docs/*`), acquisition index, failure atlas, comparison
- Added tests for embed, planner, guardian, store, adaptive, rollback, recovery

### Security

- Recorded stdlib toolchain advisories from govulncheck; recommend Go patch upgrade on build hosts

## [0.5.0] — 2026-09-21

### Added

- Acquisition-candidate packaging: gap matrix, integration API, Railway integration architecture
- Universal volume fixtures (sqlite-like, media, append-log, kv-fs, generic-dir)
- `sde railway clone` + `CLONE_RECEIPT.json` (ENGINE_PROVIDED)
- Versioned deployment events (`schemas/events/v1`) + `internal/embed` library mode
- Failure atlas, journal crash-consistency fuzz, perf regression gates
- `evaluate.ps1` / `evaluate.sh` / `make evaluate` / `cmd/evaluate-harness`
