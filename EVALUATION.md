# Evaluation Guide

**Proprietary / Pre-Acquisition** — for platform engineering diligence.  
Run these procedures yourself. Do **not** treat marketing claims or third-party
screenshots as evidence. SDE is not licensed for production use without a
written Agreement (`LICENSE`).

---

## 1. Goals

A platform eng team evaluating SDE should leave with answers to:

1. Does the coordinator refuse unsafe cutovers?
2. What is the **write-barrier pause** under sustained mutations (local)?
3. What are **RPO** (pending mutations at barrier) and **RTO** (full deploy /
   rollback) in the harness?
4. Is the architecture adapter-shaped for Railway (and later others)?
5. What is still stubbed vs proven?

---

## 2. Prerequisites

- Go **1.22+**
- Git checkout of this repository
- No cloud credentials required for local eval
- Optional later: Railway token + project for live adapter (when wired)

```bash
go version
go test ./...
go build -o sde ./cmd/sde
```

---

## 3. Functional walkthrough (local)

```bash
# Unix-style path; on Windows use a writable directory, e.g. %TEMP%\sde-demo
go run ./cmd/sde demo --root /tmp/sde-demo
go run ./cmd/sde status --root /tmp/sde-demo
go run ./cmd/sde rollback --root /tmp/sde-demo
```

Also see `examples/local-demo.md`.

**Expect**

- Phased deploy ledger (create → … → traffic transfer)
- Consistency verification before promote
- Rollback that restores prior image **and** state epoch with integrity check

**Fail the eval if**

- Cutover proceeds when manifests would diverge (inject artificial drift if
  extending tests)
- Rollback skips integrity verification
- Demo requires undocumented cloud calls

---

## 4. Benchmarks

```bash
go test ./benchmarks -bench=CutoverPause -benchtime=5x -v
```

### Metrics to record

| Metric | Source | Interpretation |
|--------|--------|----------------|
| **Cutover pause** | Barrier duration / bench output | Operator-visible write pause |
| **RPO** | Mutations pending at barrier entry | Should be drained by final delta |
| **RTO** | Total deploy duration | End-to-end wall time |
| **Sync throughput** | Ops/s, MB/s during catch-up | Workload-dependent |
| **Correctness** | `ConsistencyOK` / manifest equality | Hard gate |
| **Rollback latency** | Rollback report duration | Standby restore + verify |
| **Fault injection** | Planned harness | Corruption must be detected pre-promote |

`internal/types.MetricsSnapshot` and `DeployReport` define the fields the
product intends to expose to a platform UI.

### How to report results

When sharing with corp-dev / Owner:

- Machine class (CPU, disk type: SSD/HDD/ramdisk)
- OS and Go version
- Exact command lines and commit SHA
- Raw bench logs (attach files)
- Note: local numbers **do not** transfer to Railway disk + DNS cutover

---

## 5. Suggested acceptance criteria (diligence)

These are **engineering gates for continued diligence**, not contractual SLAs.

| ID | Criterion |
|----|-----------|
| A1 | `go test ./...` passes on a clean checkout |
| A2 | Local demo completes a full transactional deploy |
| A3 | Local rollback completes with integrity verification |
| A4 | Bench runs without panic; reports barrier-related timing |
| A5 | Verifier logic is inspectable and used as a cutover gate |
| A6 | Railway adapter documents dual-service strategy and does **not** claim concurrent single-volume mounts |
| A7 | License/`NOTICE` clearly proprietary; no accidental Apache claim in README |
| A8 | Failure model for verify-fail-under-barrier leaves active serving (`ARCHITECTURE.md`) |

**Explicit non-criteria at v0.1.0-dev**

- Live Railway cutover (stub until `SDE_RAILWAY_LIVE` client lands)
- Multi-GB volume soak numbers
- Formal security audit
- Production SRE runbooks

---

## 6. Fault and correctness experiments (recommended)

Extend or temporarily patch tests to:

1. **Divergent shadow** — mutate shadow off-journal; assert deploy fails verify
2. **Health fail** — force unhealthy shadow; assert abort, active untouched
3. **Mid-sync kill** — interrupt process; assert active data intact
4. **Rollback after promote** — deploy then rollback; compare epoch and sample file hashes

Upstream near-term work tracks systematic fault injection in `ROADMAP.md`.

---

## 7. Architecture review checklist

- [ ] Core has no hard dependency on Railway APIs
- [ ] Adapter interface covers barrier, traffic, promote, rollback
- [ ] Journal is append-only with epoch/seq
- [ ] Cutover phases match `docs/PROTOCOL.md`
- [ ] Threat model read (`SECURITY_MODEL.md`)
- [ ] Acquisition framing understood (`ACQUISITION.md`) — coordinator IP, not backups

---

## 8. Railway live eval (when available)

When the live adapter is implemented:

1. Use a **throwaway** Railway project
2. Set `RAILWAY_TOKEN`, project/service IDs, `SDE_RAILWAY_LIVE=1`
3. Confirm two services / two volumes; stable public domain re-point
4. Measure barrier pause **and** DNS/propagation effects separately
5. Destroy resources after eval

Procedure details: `docs/RAILWAY.md`.

---

## 9. Contact

Questions on eval methodology or to schedule a diligence walkthrough:

- GitHub: [@theworker02](https://github.com/theworker02)
- Email: matthewlooney5@gmail.com
