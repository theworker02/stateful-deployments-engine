# Security Policy

## Status and scope

Stateful Deployments Engine (SDE) is **proprietary software** under the
Proprietary Pre-Acquisition License (`LICENSE`). It is currently **v0.1.0-dev**.

**Do not use SDE in production** until you have a written acquisition or
commercial license agreement with the Owner **and** the Owner (or licensee)
explicitly designates a release as production-supported.

This policy covers security issues in:

- the State Transition Coordinator, journal, replay, verifier, rollback;
- adapters (including Railway stub / future live client);
- CLI, agent/writer paths, and bench harnesses as they relate to integrity.

## Reporting a vulnerability

Please use **coordinated disclosure**. Do **not** open a public GitHub issue
for security-sensitive reports.

**Preferred contact**

- Email: matthewlooney5@gmail.com  
  Subject line: `[SDE SECURITY] short description`
- GitHub: privately message or email the account [@theworker02](https://github.com/theworker02)

**Please include**

1. Affected component / package path (e.g. `internal/journal`)
2. Description of the issue and security impact
3. Steps to reproduce (PoC preferred; keep payloads minimal)
4. Affected commit / tag if known
5. Whether you plan public disclosure and any preferred timeline

## What to expect

| Step | Target |
|------|--------|
| Acknowledgement | Within 5 business days |
| Initial severity triage | Within 10 business days of acknowledgement |
| Fix / mitigation plan | Communicated after triage; timing depends on severity and stage (pre-sale / post-license) |

We may request clarification or a private follow-up. Credit will be offered in
release notes if you wish, unless you prefer anonymity.

## Coordinated disclosure

We ask that you:

- give us a reasonable window to assess and fix before public disclosure;
- avoid exploiting issues beyond what is needed to demonstrate impact;
- refrain from accessing data that is not yours.

If the Software is under active acquisition diligence, treat reports as
**confidential** under the same posture as `ACQUISITION.md`.

## Safe harbor (good-faith research)

Good-faith security research that complies with this policy and applicable law
is welcomed. Do not:

- disrupt production systems (there should be none under this license without
  Agreement);
- exfiltrate data beyond the minimum needed for the report;
- demand payment as a condition of disclosure (optional bug bounties may be
  offered later; none are committed at v0.1.0-dev).

## Cryptography and integrity notes

SDE relies on content-addressed manifests (SHA-256) for cutover gates. Issues
that allow cutover despite divergent state, journal replay that corrupts silent
data, barrier bypass, or rollback to an unverified epoch are **high priority**.

See `SECURITY_MODEL.md` for the threat model.

## Out of scope (examples)

- Denial of service against local demo paths with unbounded disk fill (unless it
  bypasses a documented integrity gate)
- Issues solely in third-party dependencies — report upstream when appropriate;
  tell us if SDE’s usage amplifies impact (`THIRD_PARTY.md`)
- Social engineering of the Owner’s personal accounts

## License reminder

Viewing source for evaluation does not authorize production deployment. Report
security issues even if you only evaluated the Software under `LICENSE` §3.
