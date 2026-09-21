# Security Review (acquisition) — v1.1.1 FROZEN

Statuses: **VERIFIED** | **MITIGATED** | **OPEN** | **NOT TESTED** | **DOCUMENTED** | **N/A**

Independent project — not affiliated with Railway.

## Summary

Overall: **READY** for acquisition diligence of the offline ENGINE_PROVIDED product
and the opt-in live GraphQL adapter surface. Freeze tag: `v1.1.1`.

## Control matrix

| Area | Status | Evidence |
|------|--------|----------|
| Stale epoch fencing | VERIFIED | `internal/fence`, journal fence checks |
| Journal corruption detection | VERIFIED | CRC trailers + fuzz seeds |
| UNKNOWN ≠ VERIFIED | VERIFIED | `internal/invariants`, verifier |
| Rollback won’t destroy ack writes | VERIFIED | `rollback.Classify` + chaos |
| No secrets in repo | VERIFIED | `.env.example` only; `.gitignore` |
| PSA integrity | VERIFIED | archive verify + independence restore |
| Railway live GraphQL path | VERIFIED | Real HTTP client + httptest; `SDE_RAILWAY_LIVE=1` |
| Live mode default-off | VERIFIED | `ErrNotWired` when live unset |
| Dependency / stdlib CVEs | DOCUMENTED | `SECURITY_GOVULNCHECK.txt` |
| Supply-chain signing | VERIFIED | Cosign keyless on tag release; `v1.1.1` assets |
| Checksums | VERIFIED | `SHA256SUMS.txt` on release |
| Cross-region network attack surface | N/A | Not implemented |
| Live prod dual-service pilot | OPEN | Optional follow-up — see `RISK_REGISTER.md` R1 |

## Threat themes (pointer)

Full narrative: [`../SECURITY_MODEL.md`](../SECURITY_MODEL.md), [`../SECURITY.md`](../SECURITY.md).

| Theme | SDE stance |
|-------|------------|
| Split-brain writers | Epoch fence; single-writer barrier |
| False “verified” | UNKNOWN never promoted to VERIFIED |
| Archive tampering | Digest verify before restore claims |
| Credential theft | Tokens only via env; not in git |
| Supply chain | Cosign keyless + SPDX SBOM |

## Recommended buyer actions

1. Re-run `govulncheck` on the exact Go toolchain used in their CI
2. Verify Cosign bundles for `v1.1.1` assets before trusting binaries
3. Keep Railway tokens in their secret store — never paste into this repo
