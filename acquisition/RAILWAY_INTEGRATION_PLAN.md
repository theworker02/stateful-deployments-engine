# Railway Integration Plan (technical)

**No deal value. No fabricated Railway internals.**  
Independent project — not affiliated with Railway.

## Phase A — Evaluate offline (today)

1. Run `evaluate.ps1` / `go test ./...`
2. Review `docs/RAILWAY_GAP_MATRIX.md` and `docs/RAILWAY_INTEGRATION_ARCHITECTURE.md`
3. Exercise ENGINE_PROVIDED archive/clone/fire-drill

## Phase B — Adapter live wiring (requires Railway credentials)

1. Implement GraphQL client behind `SDE_RAILWAY_LIVE=1`
2. Dual-service create + volume attach (respect one-volume/service)
3. Domain/TCP cutover via platform APIs
4. Keep RAILWAY_NATIVE backups distinct from ENGINE_PROVIDED escrow

## Phase C — Control-plane embedding

1. Host uses `internal/embed` LIBRARY/CONTROL-PLANE mode
2. Consume versioned events (`schemas/events/v1`)
3. Surface receipts in platform UI (OEM)

## Non-goals

- Replacing Railway Postgres PITR / logical dumps
- Claiming SDE as official Railway product
- Agent 3 cross-region networking
