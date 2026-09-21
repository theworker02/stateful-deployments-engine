# Evaluator brief (15–45 minutes) — v1.1.1 FROZEN

Audience: platform engineer or corp-dev technical reader evaluating SDE as an
acquisition / deep-integration candidate.

**Not affiliated with Railway.** No deal value claimed.

---

## Positioning (90 seconds)

| Say | Don’t say |
|-----|-----------|
| General-purpose state lifecycle engine for volume-backed apps | “Railway backup tool” |
| Complements volumes/backups/PITR/dumps/resize | “Railway lacks backups” |
| ENGINE_PROVIDED archives survive source destroy | “Native volume clone” without qualifier |
| Live GraphQL is opt-in (`SDE_RAILWAY_LIVE=1`) | “Always wired to Railway production” |
| Verdict READY at freeze tag `v1.1.1` | Invented ARR / customers / valuation |

---

## Fast path (15 minutes)

### 0–3 min — Clone & install

```powershell
git clone https://github.com/theworker02/stateful-deployments-engine
cd stateful-deployments-engine
git checkout v1.1.1
.\scripts\install.ps1
.\dist\sde.exe version    # sde 1.1.1
.\dist\sde.exe doctor
```

### 3–8 min — Local transactional deploy

```powershell
.\dist\sde.exe demo --root .sde-demo
.\dist\sde.exe status --root .sde-demo
.\dist\sde.exe readiness --root .sde-demo
```

Note measured `CUTOVER_WRITE_PAUSE_MS` / barrier (local warm target: tens of ms).
Sample proof transcript: barrier ~16 ms, pause ~15 ms, 129 mutations (`assets/demo/`).

### 8–12 min — Independence restore

```powershell
go run ./examples/acquisition-demo
# or export → verify-archive → fire-drill
```

Emphasize **ENGINE_PROVIDED ≠ RAILWAY_NATIVE**.

### 12–15 min — Evidence pack

```powershell
.\evaluate.ps1
```

Open:

- https://theworker02.github.io/stateful-deployments-engine/
- https://theworker02.github.io/stateful-deployments-engine/proof.html
- `ACQUISITION_READINESS_REPORT.md`
- `DUE_DILIGENCE_INDEX.md`

---

## Extended path (45 minutes)

| Block | Focus |
|-------|--------|
| +10 min | `docs/RAILWAY_GAP_MATRIX.md` + `RAILWAY_INTEGRATION_PLAN.md` |
| +10 min | `internal/adapter/railway` live client + `docs/RAILWAY_LIVE.md` |
| +10 min | Safety invariants + failure atlas |
| +10 min | Cosign assets on release `v1.1.1` + SECURITY_REVIEW |

---

## Questions buyers usually ask

See [`ACQUISITION_FAQ.md`](ACQUISITION_FAQ.md).

## Contact

Matthew Looney · matthewlooney5@gmail.com · [@theworker02](https://github.com/theworker02)
