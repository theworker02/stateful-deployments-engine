# Acquisition Demo (~15 minutes)

**Independent project — not affiliated with Railway.**  
Audience: Railway (or other platform) engineer evaluating SDE as an acquisition candidate.

## Goal

Show deployment-safe state mobility + ENGINE_PROVIDED independent recovery for a **generic volume-backed workload** (not “Railway lacks backups”).

## Script

### 0–2 min — Position

- Railway already has volumes, backups, Postgres PITR, dumps, resize.
- Documented constraints still matter: same-project/env native restore; wipe deletes volume backups.
- SDE: general-purpose **state lifecycle engine** — transactional cutover, verified archives, fire drills.

### 2–5 min — Clone & test

```powershell
git clone <repo>
cd stateful-deployments-engine
.\scripts\install.ps1
.\dist\sde.exe version   # 1.1.0
.\dist\sde.exe doctor
```

### 5–8 min — Local transactional deploy

```powershell
.\dist\sde.exe demo --root .sde-demo
.\dist\sde.exe status --root .sde-demo
.\dist\sde.exe readiness --root .sde-demo
```

Point at measured `CUTOVER_WRITE_PAUSE_MS` (not zero-downtime absolute).

### 8–12 min — Destroy-source recovery (ENGINE_PROVIDED)

```powershell
go run ./examples/acquisition-demo
# or
./sde.exe export --root .sde-demo --out .sde-demo/archives/a
./sde.exe verify-archive .sde-demo/archives/a
./sde.exe railway clone --source <state> --dest .sde-demo/clone --mode read-only
```

Show `RECOVERY_RECEIPT.json` / `CLONE_RECEIPT.json`. Emphasize **ENGINE_PROVIDED ≠ RAILWAY_NATIVE**.

### 12–15 min — Evaluate harness + docs

```powershell
.\evaluate.ps1
# or: make evaluate
```

Open: `docs/RAILWAY_GAP_MATRIX.md`, `docs/FAILURE_ATLAS.md`, `ACQUISITION_READINESS_REPORT.md`.

## Do not say

- “Railway doesn’t have backups”
- Anything implying affiliation or endorsement
