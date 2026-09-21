# Railway outreach — acquisition inquiry

**Status:** Ready to send  
**Target:** Railway (preferred first platform acquirer / integration partner)  
**Owner:** Matthew Looney (`theworker02`) · matthewlooney5@gmail.com  
**Product:** Stateful Deployments Engine (SDE) v1.1.0 READY  

This is **not** a securities offering and does **not** invent revenue, customers, or deal value.

## Why Railway

Railway’s documented volume model (one volume per service, no replicas with volumes,
no concurrent mounts) creates redeploy downtime for stateful apps. Railway already
ships volumes, backups, Postgres PITR, dumps, and resize — SDE does **not** claim
to replace those. SDE targets **generic FS/state mobility** and transactional
blue/green semantics around attached volumes.

## Channels

| Channel | Address / URL | Use |
|---------|---------------|-----|
| Sales / BD | `team@railway.com` | Primary acquisition / commercial inquiry |
| Partnerships | `partners@railway.com` | Secondary (program is OSS/open-core oriented; SDE is proprietary — frame as technology acquisition, not partner template) |
| Form | https://railway.com/partners | Optional; not a substitute for BD email |

## Assets to attach / link

- Live demo site (GitHub Pages after deploy)
- Repository: `https://github.com/theworker02/stateful-deployments-engine`
- `ACQUISITION_READINESS_REPORT.md` — verdict READY
- `acquisition/DUE_DILIGENCE_INDEX.md`
- Demo GIFs: `assets/demo/*.gif`
- `RELEASE_NOTES_v1.1.0.md`

## Email (sent copy)

See `OUTREACH_EMAIL_SENT.md` after transmission. Draft body lives below for reuse.

---

### Subject

Acquisition inquiry: Stateful Deployments Engine (volume-backed transactional cutover) — v1.1.0 READY

### Body

Hello Railway team,

I’m Matthew Looney (GitHub: theworker02). I’ve built an independent product —
Stateful Deployments Engine (SDE) v1.1.0 — and I’m reaching out because Railway
is the preferred first acquisition / deep-integration target.

What it is:
SDE is a state lifecycle engine for volume-backed workloads: mutation journals,
content-verified sync, write barrier, bounded cutover, observation window, and
sealed receipts — plus ENGINE_PROVIDED Portable State Archives that remain usable
after a source environment is destroyed.

What it is not:
Not a Railway-native backup clone, not a claim that Railway lacks backups, and
not affiliated with Railway. Your volumes, backups, Postgres PITR, dumps, and
live resize already cover durability/capacity. SDE targets deploy-time continuity
and portable recovery for generic filesystem state under the documented
one-volume / no-concurrent-mount constraints.

Diligence-ready:
- Acquisition verdict: READY (ACQUISITION_READINESS_REPORT.md)
- Live GraphQL adapter (opt-in), Cosign-signed release pipeline, executed
  sole-author IP declaration
- Evaluate harness: evaluate.ps1 / evaluate.sh
- Interactive demo + proof GIFs on the project Pages site

Repo: https://github.com/theworker02/stateful-deployments-engine
Demo: https://theworker02.github.io/stateful-deployments-engine/

I’d welcome a short conversation with whoever owns platform product / corp-dev
for stateful services. Happy to walk a 15-minute technical demo and share the
diligence index under NDA if useful.

Thank you,
Matthew Looney
matthewlooney5@gmail.com
https://github.com/theworker02
