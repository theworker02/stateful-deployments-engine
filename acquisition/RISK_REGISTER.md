# Risk register (acquisition diligence) — v1.1.1 FROZEN

Honest residual risks. **No vanity scores. No invented financial exposure.**

Statuses: **OPEN** · **MITIGATED** · **ACCEPTED** · **CLOSED**

| ID | Risk | Status | Notes / mitigation |
|----|------|--------|-------------------|
| R1 | Live dual-service pilot not yet run with production Railway creds | OPEN | Adapter + httptest present; needs operator pilot. Non-blocking for offline verdict. |
| R2 | Commercial APA / assignment not counsel-executed | MITIGATED | Sole-author declaration EXECUTED; APA templates in `legal-review/` for counsel |
| R3 | Stdlib / toolchain advisories on build hosts | DOCUMENTED | `SECURITY_GOVULNCHECK.txt` — upgrade Go patch releases on CI/build |
| R4 | Evaluator confuses Pages simulation with live control plane | MITIGATED | Labeled simulation chips; proof page points to CLI transcripts |
| R5 | Evaluator confuses ENGINE_PROVIDED with RAILWAY_NATIVE | MITIGATED | Gap matrix + clone docs + FAQ |
| R6 | Secrets leak via `.env` | MITIGATED | `.gitignore`; only `.env.example` committed |
| R7 | Supply-chain unsigned binaries | CLOSED | Cosign keyless on `v1.1.1` release assets |
| R8 | Multi-region networking gaps | ACCEPTED | Explicit non-goal (Agent 3) |
| R9 | Adapter scaffolds (Fly/K8s/…) incomplete vs Railway/local | ACCEPTED | Disclosed; Railway+local are diligence focus |
| R10 | Deal value / customer claims misread from public docs | MITIGATED | Public tree forbids fabricated value; FAQ restates |

## Residual diligence asks for buyer

1. Schedule live Railway pilot under NDA/creds
2. Counsel review of `legal-review/COMMERCIAL_ASSIGNMENT_DRAFT.md`
3. Confirm Go version policy for production builds

Update this register when pilot or counsel milestones close.
