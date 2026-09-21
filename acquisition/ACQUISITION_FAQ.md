# Acquisition FAQ — v1.1.1 FROZEN

Independent project — **not affiliated with Railway**.

## Product

### What is SDE?

A state lifecycle engine: transactional cutover for volume-backed workloads,
mutation journals, content verification, write barriers, Portable State Archives,
fire drills, and an embeddable control-plane API.

### Is this a Railway product?

No. Independent proprietary software. Railway is the preferred first acquisition /
integration target because its documented volume constraints make the problem concrete.

### Does Railway already have backups?

Yes — volumes, incremental/scheduled backups, Postgres PITR, logical dumps, live resize.
SDE does not replace those. It targets deploy-time continuity and portable recovery for
**generic filesystem state**.

### What does ENGINE_PROVIDED mean?

Archives and orchestration produced by SDE that remain usable after the source
environment is destroyed. Distinct from **RAILWAY_NATIVE** volume backups tied to
Railway’s volume lifecycle.

### Is the GitHub Pages demo a live control plane?

No. Labeled browser simulation. Real evidence: `evaluate.ps1`, `sde demo`, release artifacts.

## Diligence

### What is the acquisition verdict?

**READY** at freeze tag `v1.1.1`. See `ACQUISITION_READINESS_REPORT.md`.

### What was previously disclosed as open?

Live Railway GraphQL, signed releases, and legal ownership drafts — all addressed in
v1.0.0–v1.1.1 (live client, Cosign, executed sole-author declaration). Remaining
follow-ups are optional: live pilot with credentials, counsel APA, Go patch upgrades.

### Can we run without Railway credentials?

Yes. Offline ENGINE_PROVIDED path is the default diligence path.

### How do we enable live GraphQL?

`SDE_RAILWAY_LIVE=1` + token/project/env IDs. See `docs/RAILWAY_LIVE.md`.
`sde railway status --probe` / `sde doctor --probe`.

## Legal / commercial

### What license applies today?

Proprietary Pre-Acquisition (`LICENSE`). Evaluation viewing allowed; production use
requires a written agreement.

### Is ownership clear?

Sole Author IP Declaration is **EXECUTED** (`legal/SOLE_AUTHOR_IP_DECLARATION.md`).
Commercial purchase agreement templates remain in `legal-review/` for attorney review.

### Do you invent deal value?

No. Public materials contain no ARR, customers, or valuation.

## Supply chain

### Are binaries signed?

Yes on `v*` tags — Cosign keyless OIDC. See release `v1.1.1` assets + bundles.

### Where is the SBOM?

`acquisition/sbom/sde-1.1.1.spdx.json`

## Next steps

1. Run [`EVALUATOR_BRIEF.md`](EVALUATOR_BRIEF.md)
2. Read [`INTEGRATION_PLAYBOOK.md`](INTEGRATION_PLAYBOOK.md)
3. Contact matthewlooney5@gmail.com for NDA / term-sheet discussion
