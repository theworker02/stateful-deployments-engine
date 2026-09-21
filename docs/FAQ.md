# FAQ (v1.1.0)

### Is this a Railway product?

No. Independent project — **not affiliated with Railway**.

### Does Railway already have backups?

Yes — volumes, incremental/scheduled backups, Postgres PITR, logical dumps, live resize.
SDE targets **generic volume state lifecycle**, portable ENGINE_PROVIDED archives, and
transactional cutover beyond DB-specific tooling.

### Is the GitHub Pages demo a live control plane?

No. It is a **labeled browser simulation**. Real evidence: `evaluate.ps1`.

### Can I use this in production under the current license?

Only under a written acquisition or commercial license. See `LICENSE`
(Proprietary Pre-Acquisition). Evaluation viewing is allowed.

### How do I enable live Railway GraphQL?

Set `SDE_RAILWAY_LIVE=1` plus `RAILWAY_TOKEN` (and project/environment IDs).
Then run `sde railway status --probe` or `sde doctor --probe`.
Without live mode, adapters return `ErrNotWired` by design (safe offline default).
See [`RAILWAY_LIVE.md`](RAILWAY_LIVE.md).

### What does v1.1.0 “READY” mean?

Offline core + local adapter + PSA/DR/fire-drill + embed API + evaluate harness +
live GraphQL wiring + signed release pipeline + operator onboarding (`doctor` /
install scripts) are acquisition-ready. A production pilot with real Railway
credentials remains an optional follow-up, not a blocker.

### Fastest path to a green evaluate?

[`GETTING_STARTED.md`](GETTING_STARTED.md) → `scripts/install.*` → `evaluate.ps1` / `evaluate.sh`.
