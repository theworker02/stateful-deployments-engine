# Integration playbook (acquirer OEM) — v1.1.1 FROZEN

How a platform team would embed SDE after acquisition or under commercial license.
**No fabricated Railway internals.** Independent project — not affiliated with Railway.

---

## Insertion model

```
Platform control plane / CLI
        │
        ▼
  internal/embed  (LIBRARY or CONTROL-PLANE)
        │
        ├── events (schemas/events/v1) → operator ledger / UI
        ├── receipts → deploy history
        └── PlatformAdapter (railway | local | …)
                 │
                 └── Storage capabilities negotiation
```

Prefer **embedding the coordinator** over reimplementing journal/verify/barrier logic.

---

## Phase 0 — Offline acceptance (week 0)

| Gate | Command / artifact |
|------|--------------------|
| Tests green | `go test ./...` |
| Demo cutover | `sde demo` |
| Evaluate pack | `evaluate.ps1` → receipts in `evaluation/` |
| Invariants | `docs/SAFETY_INVARIANTS.md` understood by owning eng |
| Freeze tag | checkout `v1.1.1` |

Exit: engineering memo that offline ENGINE_PROVIDED path meets bar.

---

## Phase 1 — Railway adapter pilot (credentials required)

1. Copy `.env.example` → `.env` (never commit)
2. Set `SDE_RAILWAY_LIVE=1`, token, `RAILWAY_PROJECT_ID`, `RAILWAY_ENVIRONMENT_ID`
3. `sde railway status --probe` / `sde doctor --probe`
4. Dual-service shadow path respecting **one volume per service**
5. Write barrier via service variables (`SDE_WRITE_BARRIER*`) as documented
6. Domain / traffic shift via public GraphQL helpers
7. Keep **RAILWAY_NATIVE** backups labeled separately from ENGINE_PROVIDED escrow

Docs: `docs/RAILWAY_LIVE.md`, `docs/RAILWAY_GAP_MATRIX.md`, `RAILWAY_INTEGRATION_PLAN.md`.

Exit: recorded pilot receipt (success or classified failure) — no silent UNKNOWN=VERIFIED.

---

## Phase 2 — Control-plane productization

1. Host UI surfaces phases, barrier ms, RPO, verify outcome, receipt ID
2. Subscribe to versioned events (`schemas/events/v1`)
3. Map platform “Stateful Deploy” button → `embed.Deploy(...)`
4. Store sealed receipts in platform object store / DB
5. Operator runbooks for rollback / fire-drill cadence

---

## Phase 3 — Multi-adapter expansion

Order suggested by maturity of scaffolds:

1. Docker Compose / local-adjacent
2. Fly
3. Kubernetes CSI-backed volumes
4. Nomad

Capability negotiation must drive strategy selection — never assume universal CoW.

---

## Non-goals for first OEM ship

- Cross-region mesh / tunnels (separate project)
- Replacing managed Postgres PITR
- Claiming SDE as historically “built by Railway”
- Concurrent mount of one volume across two live deployments

---

## Ownership after transfer

| Asset | Post-close owner |
|-------|------------------|
| Core protocol + Go tree | Acquirer |
| Brand “Stateful Deployments Engine” | Per APA |
| Railway trademarks | Remain Railway’s — do not co-brand without approval |
| Third-party deps | Per SBOM licenses |

See `TRANSFER_CHECKLIST.md` and `legal-review/`.
