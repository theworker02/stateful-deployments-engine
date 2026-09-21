# Security Review (acquisition) — v1.1.0

Statuses: **VERIFIED** | **MITIGATED** | **OPEN** | **NOT TESTED** | **DOCUMENTED**

| Area | Status | Evidence |
|------|--------|----------|
| Stale epoch fencing | VERIFIED | `internal/fence`, journal fence checks |
| Journal corruption detection | VERIFIED | CRC trailers + fuzz seeds |
| UNKNOWN ≠ VERIFIED | VERIFIED | `internal/invariants`, verifier |
| Rollback won't destroy ack writes | VERIFIED | rollback.Classify + chaos |
| No secrets in repo | VERIFIED | `.env.example` only |
| PSA integrity | VERIFIED | archive verify + independence test |
| Railway live GraphQL path | VERIFIED | Real HTTP client + httptest suite; enable with `SDE_RAILWAY_LIVE=1` |
| Dependency / stdlib CVEs | DOCUMENTED | `SECURITY_GOVULNCHECK.txt` (upgrade Go patch on build hosts) |
| Supply-chain signing | VERIFIED | Cosign keyless on tag release workflow; local `scripts/sign-release.*` |
| Cross-region network attack surface | N/A | Not implemented |

Overall: **READY** for acquisition diligence of the offline ENGINE_PROVIDED product and live GraphQL adapter surface.
