# Railway Integration Strategy

**Proprietary** — Stateful Deployments Engine (SDE)  
First production adapter target. Companion to `ARCHITECTURE.md` and
`ACQUISITION.md`.

---

## 1. Why Railway first

Railway documents volume constraints that make **stateful redeploy downtime**
explicit and user-visible. That creates a sharp product wedge for a state
transition engine — while SDE’s core remains platform-independent.

SDE is **not** positioned as a Railway backup, PITR, or volume-resize utility.
Those solve durability and capacity. SDE solves **deploy-time continuity** of
volume-local state.

---

## 2. Documented Railway volume constraints

From the [Railway Volumes Reference](https://docs.railway.com/volumes/reference):

1. **Each service can only have a single volume**
2. **Replicas cannot be used with volumes**
3. **Multiple deployments cannot be active and mounted to the same service
   volume** — to prevent data corruption — which means **redeploy downtime
   even when a healthcheck is configured**

Operators commonly discover that drain/overlap settings and healthchecks do not
remove mount exclusivity. Community guidance often suggests moving files to
S3-compatible buckets or isolating storage in another service — workarounds that
abandon local FS semantics.

SDE accepts the mount exclusivity rule and designs around it.

---

## 3. Dual-service / dual-volume pattern

```
[stable public domain / custom hostname]
                 │
          ┌──────┴──────┐
          ▼             ▼
       Active         Shadow          ← two Railway services
       Volume A       Volume B        ← two volumes (one each — legal)
```

| Role | Railway object | SDE behavior |
|------|----------------|--------------|
| **Active** | Service + volume A | Serves traffic; writers journal mutations |
| **Shadow** | Service + volume B | Candidate image; receives replayed mutations |
| **Standby** | Previous active after promote | Retained for rollback (volume intact) |
| **Public endpoint** | Domain bound to active | Re-pointed on cutover |

**Never** attach volume A to two live deployments. Seeding and sync copy **data**,
not mounts.

---

## 4. Cutover sequence on Railway

Mapped to the core protocol (`docs/PROTOCOL.md`):

1. **CreateShadow** — ensure shadow service exists (clone config); ensure its
   **own** volume; deploy candidate image
2. **Checkpoint + seed** — content-addressed snapshot from A → initial populate B
   (via volume browse/SSH/backup-restore primitives as available — implementation
   detail of the live client)
3. **StartShadow** — candidate running against B
4. **Sync** — replay journal into B until quiet
5. **Verify** — manifest compare (pre-barrier may allow residual drift)
6. **HealthCheck** — Railway-compatible health on shadow
7. **EstablishWriteBarrier** — fence writers via agent/sidecar
8. **Final delta + hard verify**
9. **TransferTraffic** — rebind public domain / routing to shadow service
10. **PromoteShadow** — roles: shadow→active, active→standby
11. **Release barrier**

Abort paths leave the original active serving; destroy or idle the shadow.

---

## 5. Environment variables (planned live client)

| Variable | Purpose |
|----------|---------|
| `RAILWAY_TOKEN` | API authentication |
| `RAILWAY_PROJECT_ID` | Target project |
| `RAILWAY_ENVIRONMENT_ID` | Environment |
| `RAILWAY_SERVICE_ID` | Active stateful service |
| `RAILWAY_SHADOW_SERVICE_ID` | Optional pre-created shadow |
| `RAILWAY_PUBLIC_DOMAIN` | Hostname to re-point |
| `SDE_RAILWAY_LIVE=1` | Enable live client (otherwise `ErrNotWired`) |

Implementation status: **typed stub** in `internal/adapter/railway` with TODOs.
Core and local adapter are developed independently of Railway API access.

---

## 6. Complementary Railway investments

SDE sits **beside**:

- Volume backups and restore
- Point-in-time recovery (where offered)
- Live volume resize
- HA / managed databases
- S3-compatible object storage for apps that can leave local disk

Those increase the number of stateful apps on Railway; SDE improves **how those
apps deploy**.

---

## 7. Risks and open questions for diligence

| Topic | Note |
|-------|------|
| Domain / DNS TTL | Platform rebind may be fast; client resolvers may lag — measure separately from barrier ms |
| Seed throughput | First sync of large volumes dominates RTO; optimize checkpoint transport |
| Billing | Two services + two volumes during deploy/standby retain — product/packaging decision |
| Permissions | Token scopes for service create, volume attach, domain update |
| Agent install | How Railway users opt into journaling without painful DX |

---

## 8. Success criteria for “Railway live” milestone

- End-to-end dual-service cutover in a non-production project
- No attempt to double-mount one volume
- Deploy report shows verify gate and barrier duration
- Rollback re-points to standby with integrity verification
- Documented runbook for abort and crash-during-barrier

See `EVALUATION.md` §8 and `ROADMAP.md`.
