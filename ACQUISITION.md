# ACQUISITION.md

**CONFIDENTIAL — Proprietary Pre-Acquisition Material**  
Stateful Deployments Engine (SDE)  
Owner: Matthew Looney / GitHub [`theworker02`](https://github.com/theworker02)  
Status: **v1.1.1 FROZEN** — acquisition diligence freeze (core + live GraphQL opt-in + CLI + evaluate + Pages proof + expanded acquisition pack)  
License: Proprietary — see `LICENSE`. No production use without a written Agreement.

This document is an acquisition and diligence brief for platform-engineering and
corporate-development readers. It is not a solicitation of investment securities,
not legal advice, and not a representation of revenue, customers, or production
SLAs. Metrics below are **evaluation targets and harness capabilities**, not
claimed production results.

---

## 1. Executive summary / thesis

**Thesis:** Platforms that attach persistent volumes to application containers
still force a hard trade-off — either accept redeploy downtime, or push state
out of the app into databases/object storage. SDE closes that gap with a
**transactional state transition engine**: mutation journal → sync → content-
addressed verify → write barrier → final delta → traffic cutover → standby
rollback — so stateful workloads get blue/green semantics without concurrent
mounts of a single volume.

**Positioning:** Platform-independent core IP (State Transition Coordinator).
**Railway is the first adapter and preferred acquisition / integration target**,
because Railway’s documented volume model makes the problem concrete and the
timing strong — not because SDE is a “Railway backup utility.”

**What a buyer acquires:** A protocol + reference implementation that can become
a first-class platform feature (“stateful zero-downtime deploy”) across Railway
first, then Fly, Docker, Kubernetes, Nomad, and others via adapters.

---

## 2. Problem

### 2.1 Platform constraint (Railway, documented)

Railway’s volume reference documents constraints that force downtime for
volume-backed services:

| Constraint | Implication |
|------------|-------------|
| **One volume per service** | Cannot attach a second volume to the same service for blue/green |
| **No replicas with volumes** | Horizontal scale and classic rolling updates are unavailable |
| **No concurrent mounts** of the same service volume across deployments | Old deployment must stop before the new one mounts → **redeploy downtime even with healthchecks** |

Primary citation: [Railway Volumes Reference — limitations](https://docs.railway.com/volumes/reference)

> “To prevent data corruption, we prevent multiple deployments from being active
> and mounted to the same service. This means that there will be a small amount
> of downtime when re-deploying a service that has a volume attached, even if
> there is a healthcheck endpoint configured.”

Community and Station threads confirm operators hit this regularly: healthchecks,
drain seconds, and overlap settings do not remove the mount exclusivity
constraint. Suggested workarounds are typically “move files to S3-compatible
buckets” or “split storage into a separate service” — i.e., **abandon local
stateful UX**, not fix it.

### 2.2 Broader market problem

The same pattern appears wherever platforms treat “one disk ↔ one running
revision” as a safety invariant:

- App-owned SQLite / embedded DBs / local queues / media caches / ML artifact
  dirs on attached volumes
- Self-hosted CMS, analytics, game servers, internal tools that “just use `/data`”
- Teams that chose a PaaS for DX and then discovered volume deploys are stop-the-world

Backups, PITR, and live volume resize solve **durability and capacity** — not
**deploy-time continuity of local state**.

---

## 3. Solution / product

SDE presents a **transactional stateful deploy UX**:

```
Deploying myapp

✓ Created candidate deployment
✓ Captured state checkpoint
✓ Candidate started
✓ N mutations synchronized
✓ Filesystem consistency verified
✓ Health checks passed
✓ Final delta: M writes
✓ State barrier: ~tens of ms (local warm target)
✓ Traffic transferred

STATEFUL ZERO-DOWNTIME DEPLOYMENT
```

Rollback restores **application revision and state epoch**, with integrity
verification against a content-addressed checkpoint.

**Product shape for a platform buyer:**

1. Native “Stateful Deploy” path in the control plane / CLI
2. Sidecar or agent that journals filesystem mutations (or an approved SDK)
3. Dual-slot adapter that respects platform volume rules (never double-mounts
   one volume)
4. Operator-visible ledger: phases, barrier duration, RPO, verify result

---

## 4. Why the IP matters

### Core IP = State Transition Coordinator

| Component | Role |
|-----------|------|
| **Mutation journal** | Append-only `(epoch, seq)` log of create/write/rename/delete/chmod/truncate |
| **Sync / replay** | Continuously applies journal into shadow state until quiet |
| **Consistency verifier** | Content-addressed `path → sha256` manifests; refuses cutover on divergence |
| **Cutover coordinator** | Write barrier → final delta → hard verify → traffic → promote → retain standby |
| **Rollback engine** | Re-point traffic to standby; pin journal epoch; verify integrity |

**Not the IP:** naive `rsync`, one-shot volume snapshot restore, or “copy files
then restart.” Those lack transactional cutover, barrier semantics, epoch
rollback, and a platform-agnostic coordinator.

Adapters (Railway, Fly, …) are **integration surface**. The moat is the
protocol and correctness story under concurrent writes and failure.

See `ARCHITECTURE.md` and `docs/PROTOCOL.md`.

---

## 5. Architecture overview

```
┌─────────────────────────────────────────────────────────────┐
│                 State Transition Coordinator (CORE)         │
│  create → checkpoint → start → sync → verify → barrier →    │
│  final-delta → traffic → promote → retain-standby           │
└───────────────┬─────────────────────────────┬───────────────┘
                │                             │
        Mutation Journal              Consistency Verifier
                │                             │
                └──────── Replay Engine ──────┘
                              │
                     Platform Adapter
              (local | railway | fly | k8s | …)
```

| Layer | Today (v1.1.1 FROZEN) | Near-term |
|-------|--------------------|-----------|
| **Core** | Coordinator, journal, replay, verifier, rollback, PSA/DR | Hardened pilot evidence, expanded fault injection |
| **Adapters** | `local` full; `railway` live GraphQL (opt-in) | Production dual-service pilot |
| **Agent/SDK** | Local writer agent + `sdk/go/sdesdk` | Production sidecar packaging |
| **CLI** | Full deploy/DR/doctor/railway surface | Platform-native UX embedding |

**Railway-first strategy:** dual-service / dual-volume under a stable public
domain — legal under Railway’s “one volume per service” rule; never concurrent
mount of the same volume. Details: `docs/RAILWAY.md`.

---

## 6. Competitive differentiation

| Approach | What it solves | What it doesn’t |
|----------|----------------|-----------------|
| **Volume backups / PITR** | Durability, point-in-time restore | Deploy continuity; live cutover under writes |
| **rsync / file sync tools** | Bulk copy | Transactional barrier, epoch rollback, platform traffic cutover, verify-gated promote |
| **CRIU / process migration** | Checkpoint/restore of process memory | Portable PaaS UX; volume mount exclusivity; multi-platform adapters |
| **DB-native HA (Postgres, etc.)** | Replicated database state | App-local filesystem state; embedded DBs; non-DB artifacts |
| **Classic blue/green (stateless)** | Traffic cutover for immutable containers | Persistent local state attached to the old revision |
| **“Move to object storage”** | Enables replicas / zero-downtime on PaaS | Forces architecture change; loses local FS semantics |

SDE is **complementary** to backups, HA databases, and object storage — it
targets the remaining class: **apps that legitimately need volume-local state
and frequent deploys**.

---

## 7. Market / timing

Railway and peers are investing in the **stateful platform** surface: volumes,
backups, PITR, live resize, HA managed databases, S3-compatible buckets. That
investment increases the number of viable stateful apps on the platform — and
therefore the pain of **volume redeploy downtime**.

Strategic fit for Railway (or similar):

- Differentiates PaaS DX for stateful apps without weakening volume safety
- Reduces “just move to S3 / leave for K8s” churn from volume-backed users
- Ships as a platform feature with measurable cutover pause / RPO / verify gates
- Extends naturally once the core is owned: Fly, Docker Compose, K8s CSI, Nomad

No customer count, ARR, or pipeline figures are claimed in this brief.

---

## 8. Acquisition strategy

1. **Preferred:** Railway (or Railway-aligned corp-dev) acquires the IP and
   productizes Stateful Deploy as a first-party capability.
2. **Framing for diligence:** Platform-independent engine; Railway is the
   *first* production adapter and the sharpest wedge — not the entire TAM.
3. **Alternatives:** Commercial license to one platform; multi-platform OEM;
   acqui-hire of the owner with IP assignment; integration partnership with
   option to buy.

Non-goal of this pitch: positioning SDE as a third-party “Railway backup
plugin.” Backups are adjacent infrastructure; SDE is **deploy-time state
transition**.

---

## 9. Product roadmap / milestones

| Horizon | Outcomes |
|---------|----------|
| **Near** | Railway adapter live (`SDE_RAILWAY_LIVE`); end-to-end dual-service cutover; agent TTL on barrier crash; expanded benches + fault injection |
| **Mid** | Fly + Docker adapters; production sidecar packaging; operator dashboards / deploy ledger export; EVAL acceptance suite green on CI |
| **Long** | Kubernetes + Nomad adapters; optional FUSE interception; commercial packaging / platform SDK; multi-region considerations (explicit non-goal for v0) |

See `ROADMAP.md` for detail.

---

## 10. Technical moat & evaluation metrics

Buyers should run the harness themselves — do not take marketing numbers from
third parties. Harness entry points:

```bash
go test ./...
go test ./benchmarks -bench=CutoverPause -benchtime=5x -v
```

| Metric | Meaning | Notes |
|--------|---------|-------|
| **Cutover pause** | Write-barrier hold time | Operator-visible “downtime”; local warm target: tens of ms |
| **RPO** | Mutations not on shadow at barrier entry | Lower is better; final delta drains residual |
| **RTO** | Full deploy wall time / rollback latency | Includes sync + verify + cutover |
| **Sync throughput** | Ops/s and MB/s during catch-up | Workload-dependent |
| **Correctness** | Post-barrier manifest equality | Hard gate — cutover refused on fail |
| **Rollback latency** | Standby restore + integrity verify | Epoch pin + traffic re-point |
| **Fault injection** | Corruption / crash under barrier | Planned rate tracking in harness |

Detailed procedure and acceptance criteria: `EVALUATION.md`.  
Threat model: `SECURITY_MODEL.md`.

---

## 11. Suggested deal structures (high level — not legal advice)

| Structure | Sketch |
|-----------|--------|
| **Asset purchase** | Buyer acquires code, docs, brand, IP assignment; owner transition assistance |
| **Acqui-hire** | Employment / contractor engagement + IP assignment of SDE |
| **Commercial license** | Exclusive or non-exclusive platform license; optional railroad to purchase |
| **Integration partnership** | Paid integration milestone + option; SDE remains Owner-controlled until exercise |

Commercial terms, exclusivity, escrow, and reps/warranties belong in counsel-
drafted Agreements. Contact the Owner for a term sheet discussion.

---

## 12. Diligence checklist — what the buyer gets

| Asset | Location / status |
|-------|-------------------|
| Source (core + adapters + CLI) | `cmd/`, `internal/`, `benchmarks/` |
| Architecture & protocol | `ARCHITECTURE.md`, `docs/PROTOCOL.md` |
| Railway strategy | `docs/RAILWAY.md`, `internal/adapter/railway` |
| Security / threat model | `SECURITY.md`, `SECURITY_MODEL.md` |
| Eval & benches | `EVALUATION.md`, `benchmarks/` |
| License posture | `LICENSE`, `NOTICE`, `THIRD_PARTY.md` |
| Roadmap / changelog | `ROADMAP.md`, `CHANGELOG.md` |
| Brand / funding handle | Repo name; `.github/FUNDING.yml` (`theworker02`) |
| Tests | `go test ./...` (unit + coordinator + journal) |

**Not included unless negotiated:** production Railway credentials, customer
lists (none claimed), trademarks registration filings, patents (none asserted
in-repo).

---

## 13. Contact / next steps

1. Read `LICENSE`, `ARCHITECTURE.md`, `EVALUATION.md`, `docs/RAILWAY.md`.
2. Run local demo and benches (no cloud credentials required).
3. Request deeper diligence access / NDA if needed.
4. Contact for acquisition or commercial license:

   - GitHub: [@theworker02](https://github.com/theworker02)
   - Email: matthewlooney5@gmail.com
   - Repo: https://github.com/theworker02/stateful-deployments-engine

---

**CONFIDENTIAL.** © 2026 Matthew Looney / theworker02. All rights reserved.
Unauthorized production use or redistribution is prohibited under the
Proprietary Pre-Acquisition License.
