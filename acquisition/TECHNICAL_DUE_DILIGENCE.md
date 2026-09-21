# Technical due diligence checklist — v1.1.1 FROZEN

Independent project — **not affiliated with Railway**.  
Run evidence yourself; do not rely on marketing claims.

## How to use this checklist

1. Check out tag `v1.1.1`
2. Run the commands in each section
3. Mark items only when evidence is observed locally
4. Record gaps in `RISK_REGISTER.md` (do not invent severity scores)

---

## A. Build, version, operator UX

| Item | Status | How to verify |
|------|--------|----------------|
| Version string | [x] | `sde version` → `sde 1.1.1` |
| Codename FROZEN | [x] | `internal/version` / doctor output |
| Unit/integration tests | [x] | `go test ./...` |
| Install scripts | [x] | `scripts/install.ps1` or `install.sh` |
| Doctor (offline ready) | [x] | `sde doctor` → ready: yes |
| Shell completions | [x] | `sde completion powershell\|bash\|zsh` |
| Evaluate harness | [x] | `evaluate.ps1` / `evaluate.sh` → `evaluation/` |

## B. Core correctness

| Item | Status | How to verify |
|------|--------|----------------|
| Durable deploy FSM + crash recovery | [x] | `internal/fsm`, coordinator tests |
| Mutation journal + fencing | [x] | `internal/journal`, `internal/fence` |
| Verification receipts (UNKNOWN ≠ VERIFIED) | [x] | `internal/invariants`, verifier |
| Measured cutover pause | [x] | `sde demo` transcript; benches |
| Rollback safety classification | [x] | `internal/rollback`, chaos suite |
| Journal crash fuzz | [x] | `internal/journal` fuzz tests |
| Perf regression gates | [x] | `go test ./benchmarks -run TestPerfRegressionGates` |

## C. Portability & DR

| Item | Status | How to verify |
|------|--------|----------------|
| Portable State Archive (ENGINE_PROVIDED) | [x] | `sde export` / archive package |
| Recovery after source destruction | [x] | `examples/acquisition-demo`, RECOVERY_RECEIPT |
| Escrow catalog | [x] | `internal/escrow`, `sde escrow` |
| DR planner + fire drill | [x] | `sde fire-drill`, `sde readiness` |
| Env clone + CLONE_RECEIPT | [x] | `sde railway clone` |

## D. Platform adapters

| Item | Status | How to verify |
|------|--------|----------------|
| Local adapter (full offline) | [x] | `sde demo --root .sde` |
| Railway live GraphQL client | [x] | `internal/adapter/railway` + httptest; `docs/RAILWAY_LIVE.md` |
| Live mode gated offline by default | [x] | Without `SDE_RAILWAY_LIVE=1` → `ErrNotWired` |
| Railway status / doctor probe | [x] | `sde railway status --probe` (creds required) |
| Fly / Docker / K8s / Nomad scaffolds | [x] | Adapter packages + capability matrices |
| Multi-region networking | [ ] | Explicit non-goal (Agent 3) |

## E. Embed & schemas

| Item | Status | How to verify |
|------|--------|----------------|
| Library / control-plane embed | [x] | `internal/embed` |
| Versioned deployment events | [x] | `schemas/events/v1/` |
| Public SDK surface | [x] | `sdk/go/sdesdk` |

## F. Security & supply chain

| Item | Status | How to verify |
|------|--------|----------------|
| No secrets in tree | [x] | `.env.example` only; `.gitignore` |
| Cosign keyless release signing | [x] | Release assets `*.cosign.bundle` on `v1.1.1` |
| SHA256SUMS | [x] | Release `SHA256SUMS.txt` |
| govulncheck log | [x] | `SECURITY_GOVULNCHECK.txt` |
| SECURITY_REVIEW statuses | [x] | This folder |

## G. IP & transfer

| Item | Status | How to verify |
|------|--------|----------------|
| Executed sole-author declaration | [x] | `legal/SOLE_AUTHOR_IP_DECLARATION.md` |
| Ownership record JSON | [x] | `legal/EXECUTED_OWNERSHIP_RECORD.json` |
| SBOM 1.1.1 | [x] | `sbom/sde-1.1.1.spdx.json` |
| Transfer inventory | [x] | `TRANSFER_INVENTORY.json` |
| Commercial APA templates | [~] | `legal-review/` — **draft for counsel** |

---

## Suggested eng review order (½ day)

1. `go test ./...` + `sde demo` + `evaluate.ps1`
2. Read `docs/SAFETY_INVARIANTS.md` + `docs/FAILURE_ATLAS.md`
3. Walk `internal/coordinator` + `internal/adapter/railway/client.go`
4. Review `SECURITY_REVIEW.md` + release Cosign assets
5. Skim `INTEGRATION_PLAYBOOK.md` for OEM insertion points

## Known follow-ups (non-blocking)

- Live dual-service pilot with production Railway credentials (operator runbook)
- Counsel review of definitive asset-purchase agreement
- Go toolchain patch upgrades per govulncheck advisories on build hosts
