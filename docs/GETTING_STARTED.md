# Getting started (v1.1.0)

Five-minute path from clone → working demo. No Railway credentials required.

## 1. Install

```powershell
# Windows
.\scripts\install.ps1
# optional: .\scripts\install.ps1 -AddToPath
```

```bash
# macOS / Linux
chmod +x scripts/install.sh
./scripts/install.sh
# optional: export PATH="$PWD/dist:$PATH"
```

Or manually:

```bash
go test ./...
go build -o sde.exe ./cmd/sde   # omit .exe on Unix
./sde.exe version               # → sde 1.1.0
```

## 2. Doctor

```bash
./sde.exe doctor
./sde.exe doctor --json
```

Expect `ready: yes` for offline use. Workspace warnings are OK until `init`/`demo`.

## 3. Local demo

```bash
./sde.exe demo --root .sde
./sde.exe status --root .sde
./sde.exe readiness --root .sde
```

## 4. One-command evaluation (acquisition evidence)

```powershell
.\evaluate.ps1
```

```bash
./evaluate.sh
# or: make evaluate
```

Writes `evaluation/SUMMARY.md`, receipts, and benchmarks.

## 5. Live Railway (optional)

```bash
cp .env.example .env
# edit: SDE_RAILWAY_LIVE=1, RAILWAY_TOKEN, project/env IDs
./sde.exe railway status --probe
./sde.exe doctor --probe
```

Details: [`RAILWAY_LIVE.md`](RAILWAY_LIVE.md).

## Next

| Goal | Doc / command |
|------|----------------|
| CLI map | [`CLI.md`](CLI.md) · `sde --help` |
| Embed in Go | [`INTEGRATION_API.md`](INTEGRATION_API.md) |
| 15‑min diligence demo | [`ACQUISITION_DEMO.md`](ACQUISITION_DEMO.md) |
| Browser simulation | [`../site/`](../site/) |
| Shell completions | `sde completion powershell` / `bash` / `zsh` |
