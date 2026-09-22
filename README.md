# Stateful Deployments Engine (SDE)


---

## License & acquisition

This project is **proprietary**. Production use, redistribution, and commercial deployment require a written commercial license or completed acquisition. See [LICENSE](./LICENSE) and [ACQUISITION.md](./ACQUISITION.md). Contact [@theworker02](https://github.com/theworker02).


<p align="center">
  <img src="assets/brand/logo.svg" alt="Stateful Deployments Engine" width="320"/>
</p>

<p align="center">
  <img alt="version" src="https://img.shields.io/badge/version-1.1.1-0B3D2E"/>
  <img alt="status" src="https://img.shields.io/badge/status-FROZEN-1F7A5C"/>
  <img alt="license" src="https://img.shields.io/badge/license-Proprietary%20Pre--Acquisition-red"/>
  <img alt="go" src="https://img.shields.io/badge/go-1.22+-00ADD8"/>
  <a href=".github/workflows/test.yml"><img alt="ci" src="https://img.shields.io/badge/ci-test.yml-1F7A5C"/></a>
  <a href=".github/workflows/pages.yml"><img alt="pages" src="https://img.shields.io/badge/live%20demo-GitHub%20Pages-1F7A5C"/></a>
</p>

**Independent project â€” not affiliated with Railway.**

General-purpose **state lifecycle engine** for arbitrary volume-backed workloads:
transactional cutover, mutation journals, portable State Archives, fire drills, and
verified recovery â€” beyond database-specific backup tooling.

## Get started in 5 minutes

```powershell
.\scripts\install.ps1
.\dist\sde.exe doctor
.\dist\sde.exe demo --root .sde
.\evaluate.ps1
```

```bash
./scripts/install.sh
./dist/sde doctor
./dist/sde demo --root .sde
./evaluate.sh
```

Full path: [`docs/GETTING_STARTED.md`](docs/GETTING_STARTED.md).

## Proof (live demo GIFs)

Measured local cutover on evaluation host: **barrier ~16 ms**, **write pause ~15 ms**, 129 mutations synced.

<p align="center">
  <img src="assets/demo/cutover-pipeline.gif" alt="Cutover pipeline" width="720"/>
</p>

<p align="center">
  <img src="assets/demo/cli-doctor-demo.gif" alt="CLI doctor and demo" width="720"/>
</p>

<p align="center">
  <img src="assets/demo/restore-independence.gif" alt="Source-independent restore" width="720"/>
</p>

Transcripts: [`assets/demo/`](assets/demo/) Â· Pages proof: [`site/proof.html`](site/proof.html)

## Live demo (GitHub Pages)

Interactive browser simulation + proof GIFs:

- **https://theworker02.github.io/stateful-deployments-engine/**
- Site source: [`site/`](site/)

## What it does

| Capability | Notes |
|------------|--------|
| Transactional deploy | Checkpoint â†’ sync â†’ verify â†’ barrier â†’ cutover â†’ observe â†’ commit |
| Portable State Archive | ENGINE_PROVIDED; usable after source env destruction |
| Fire drills / DR plan | Isolated restore verification + readiness scorecard |
| Embed modes | CLI Â· LIBRARY Â· CONTROL-PLANE (`internal/embed`) |
| Railway | Live GraphQL when `SDE_RAILWAY_LIVE=1`; offline ENGINE_PROVIDED by default |
| Operator UX | `sde doctor` Â· `sde railway status --probe` Â· install scripts |

Railway already offers volumes, backups, Postgres PITR, logical dumps, and live resize.
SDE targets **generic FS/state mobility** and independent escrowed recovery. See
[`docs/RAILWAY_GAP_MATRIX.md`](docs/RAILWAY_GAP_MATRIX.md).

**License:** Proprietary Pre-Acquisition â€” [`LICENSE`](./LICENSE).  
**Acquisition:** [`ACQUISITION_READINESS_REPORT.md`](./ACQUISITION_READINESS_REPORT.md) â€” **READY**

## Architecture

```
App â†’ journal â†’ ACTIVE volume
             â†˜ shadow sync â†’ verify â†’ barrier â†’ cutover â†’ observe â†’ commit
             â†˜ ENGINE_PROVIDED PSA â†’ escrow / clone / fire-drill
```

Networking / tunnels / mesh: **out of scope** (separate project).

## Documentation

- [Getting started](docs/GETTING_STARTED.md) Â· [Quick start](docs/QUICKSTART.md) Â· [CLI](docs/CLI.md)
- [15-minute demo](docs/ACQUISITION_DEMO.md) Â· [Integration API](docs/INTEGRATION_API.md)
- [Railway live](docs/RAILWAY_LIVE.md) Â· [Failure atlas](docs/FAILURE_ATLAS.md)
- [Site](site/) Â· [Brand](docs/BRAND_GUIDELINES.md) Â· [Changelog](CHANGELOG.md)

**Status:** **v1.1.1 FROZEN** â€” acquisition verdict **READY**. Tag `v1.1.1` is the diligence freeze.
