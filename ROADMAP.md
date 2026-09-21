# Roadmap

**CONFIDENTIAL / Proprietary** — Stateful Deployments Engine (SDE)  
License: `LICENSE`. This roadmap describes intended engineering direction; it is
not a commitment of dates, SLAs, or commercial terms.

Current status: **v0.1.0-dev** (core protocol + local adapter + CLI + benches).

---

## Near term

| Item | Detail |
|------|--------|
| **Railway adapter live** | Implement GraphQL / API client behind `SDE_RAILWAY_LIVE=1`; dual-service create, volume seed, domain cutover, standby retain |
| **Barrier crash safety** | Agent TTL so a fenced active cannot remain stuck indefinitely after coordinator crash mid-barrier |
| **Fault injection benches** | Inject corrupt writes / truncated journal / killed barrier; assert verifier refuses promote; track detection rate |
| **Eval suite CI** | Codify `EVALUATION.md` acceptance checks in automated tests where deterministic |
| **Docs polish** | Keep Railway citations and protocol docs aligned with code |

Success signal: a diligence engineer can run local benches *and* (with Railway
credentials) a controlled dual-service cutover in a non-production project.

## Mid term

| Item | Detail |
|------|--------|
| **Fly.io adapter** | Map coordinator slots to Fly Machines / volumes under Fly constraints |
| **Docker adapter** | Local/dev and self-hosted Docker Compose / Swarm-style dual volume cutover |
| **Production sidecar / agent** | Package mutation interception for supported runtimes without requiring app rewrite where possible |
| **Language SDKs (thin)** | Optional SDK hooks for apps that prefer explicit journaled writes |
| **Operator ledger export** | Machine-readable deploy reports (JSON) for platform UI embedding |
| **Kubernetes adapter (alpha)** | PVC-aware dual-workload pattern; CSI realities documented frankly |

Success signal: two platforms beyond local can execute the same coordinator
protocol with adapter-only changes.

## Long term

| Item | Detail |
|------|--------|
| **Nomad adapter** | Complete the planned adapter set from `ARCHITECTURE.md` |
| **Optional FUSE interception** | Transparent mutation capture where agent install is undesirable (explicit non-goal for v0) |
| **Commercial packaging** | Platform OEM SDK, support tiers under commercial license / post-acquisition |
| **Hardening** | Formal failure drills, soak tests, larger-state sync profiles |
| **Multi-region** | Remains an explicit **non-goal** until single-region cutover is production-proven |

## Non-goals (remain out of scope unless revisited)

- Replacing database-native replication (Postgres HA, etc.)
- Multi-writer active-active on one logical volume
- Cross-region volume migration as a v0 feature
- Positioning SDE as a backup/PITR product

## Alignment with acquisition

Roadmap priority is intentionally **Railway-first** while keeping the core
platform-independent. See `ACQUISITION.md` for deal framing and `docs/RAILWAY.md`
for integration strategy.

Updates will be reflected in `CHANGELOG.md` as milestones land.
