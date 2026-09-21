# Quick start (v1.1.0)

Prefer [`GETTING_STARTED.md`](GETTING_STARTED.md) for the five-minute path.

## Prerequisites

- Go 1.22+ (1.26.4+ recommended for toolchain CVE patches)
- Windows PowerShell or bash

## Install

```powershell
.\scripts\install.ps1
```

```bash
./scripts/install.sh
```

## Build & test

```bash
go test ./...
go build -o sde.exe ./cmd/sde
./sde.exe version          # → sde 1.1.0
./sde.exe doctor
```

## Local transactional demo

```bash
./sde.exe demo --root .sde-demo
./sde.exe status --root .sde-demo
./sde.exe readiness --root .sde-demo
```

## One-command evaluation (acquisition)

```powershell
.\evaluate.ps1
# bash: ./evaluate.sh
# make evaluate
```

Writes `evaluation/SUMMARY.md`, receipts, and benchmarks.

## Portable archive round-trip

```bash
./sde.exe init --root .sde
# ... generate state via demo or workloads ...
./sde.exe export --root .sde --out ./archives/psa-1
./sde.exe verify-archive --archive ./archives/psa-1
./sde.exe fire-drill --archive ./archives/psa-1
```

## Embed (library mode)

```go
eng, err := embed.OpenLocal(".sde", embed.ModeLibrary)
defer eng.Close()
report, err := eng.Deploy(ctx, "app:v2", false, 250*time.Millisecond)
```

See [`INTEGRATION_API.md`](INTEGRATION_API.md).

## Live demo site

Open [`../site/index.html`](../site/index.html) or enable GitHub Pages
(`.github/workflows/pages.yml`). Browser demo is a **labeled simulation**;
evidence comes from `evaluate.ps1`.

## Next

- [`CLI.md`](CLI.md) · [`ACQUISITION_DEMO.md`](ACQUISITION_DEMO.md) (~15 min)
- [`RAILWAY_LIVE.md`](RAILWAY_LIVE.md) · [`RAILWAY_GAP_MATRIX.md`](RAILWAY_GAP_MATRIX.md)
