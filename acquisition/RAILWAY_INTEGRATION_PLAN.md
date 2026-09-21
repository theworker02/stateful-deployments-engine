# Railway Integration Plan (technical) — v1.1.1 FROZEN

**No deal value. No fabricated Railway internals.**  
Independent project — not affiliated with Railway.

## Current state (freeze)

| Layer | Status |
|-------|--------|
| ENGINE_PROVIDED archive / clone / fire-drill | Complete offline |
| Live GraphQL client | Implemented (`internal/adapter/railway`) |
| Enablement | `SDE_RAILWAY_LIVE=1` + token + project/env IDs |
| Probe UX | `sde railway status --probe`, `sde doctor --probe` |
| Production dual-service pilot | Optional follow-up (credentials) |

---

## Phase A — Evaluate offline (today / freeze tag)

1. `git checkout v1.1.1`
2. Run `evaluate.ps1` / `go test ./...`
3. Review `docs/RAILWAY_GAP_MATRIX.md` and `docs/RAILWAY_INTEGRATION_ARCHITECTURE.md`
4. Exercise ENGINE_PROVIDED archive/clone/fire-drill
5. Review proof: `assets/demo/`, Pages `/proof.html`

## Phase B — Adapter live pilot (requires Railway credentials)

1. Configure env from `.env.example` (see `docs/RAILWAY_LIVE.md`)
2. Validate token: `sde railway status --probe`
3. Dual-service create + volume attach (respect one-volume/service)
4. Domain/TCP cutover via public GraphQL helpers already wired
5. Keep RAILWAY_NATIVE backups distinct from ENGINE_PROVIDED escrow
6. Record pilot outcome in buyer diligence notes (do not invent success)

## Phase C — Control-plane embedding

1. Host uses `internal/embed` LIBRARY/CONTROL-PLANE mode
2. Consume versioned events (`schemas/events/v1`)
3. Surface receipts in platform UI (OEM)
4. Follow [`INTEGRATION_PLAYBOOK.md`](INTEGRATION_PLAYBOOK.md)

## Non-goals

- Replacing Railway Postgres PITR / logical dumps
- Claiming SDE as official Railway product
- Agent 3 cross-region networking
- Concurrent mount of one volume across two live deployments

## Citations (buyer should re-check)

- [Railway Volumes Reference — limitations](https://docs.railway.com/volumes/reference)
- Public GraphQL: `https://backboard.railway.com/graphql/v2`
