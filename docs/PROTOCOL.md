# Cutover Protocol

**Proprietary** — Stateful Deployments Engine (SDE)  
Detailed state machine for transactional stateful deployment. Normative behavior
for the State Transition Coordinator; adapters implement platform primitives
only.

Code references: `internal/types` (`DeployPhase`), `internal/coordinator`.

---

## 1. Roles

| Role | Meaning |
|------|---------|
| **Active** | Currently serving traffic; source of truth for live writes |
| **Shadow** | Candidate deployment receiving synced state |
| **Standby** | Previous active retained after promote for rollback |

Journal **epoch** increments on successful cutover. Rollback pins epoch to the
standby’s epoch.

---

## 2. Phase state machine

```
                     ┌─────────────┐
                     │    idle     │
                     └──────┬──────┘
                            │ Deploy()
                            ▼
                 create_candidate
                            │
                            ▼
               capture_checkpoint
                            │
                            ▼
                 start_candidate
                            │
                            ▼
                 sync_mutations ◄──┐
                            │      │ (poll until quiet / deadline)
                            └──────┘
                            │
                            ▼
               verify_consistency  (pre-barrier; residual OK)
                            │
                            ▼
                  health_check
                            │
              ┌─────────────┴─────────────┐
              │ fail                      │ ok
              ▼                           ▼
           failed                 write_barrier
           (abort)                        │
                                          ▼
                                    final_delta
                                          │
                                          ▼
                              verify_consistency (hard)
                                          │
                              ┌───────────┴───────────┐
                              │ fail                  │ ok
                              ▼                       ▼
                     release barrier            traffic_transfer
                     → failed                         │
                                                      ▼
                                                   promote
                                                   (standby retain)
                                                      │
                                                      ▼
                                                 release barrier
                                                      │
                                                      ▼
                                                   complete
```

On many abort paths: **release barrier if held**, destroy or idle shadow, leave
active untouched → phase `failed`.

Rollback is a separate flow → phase `rolled_back` on success.

---

## 3. Phase dictionary

| Phase | Coordinator actions | Adapter hooks |
|-------|---------------------|---------------|
| `idle` | No in-flight deploy | — |
| `create_candidate` | Allocate shadow slot identity | `CreateShadow` |
| `capture_checkpoint` | Build content-addressed snapshot of active; seed shadow | State copy helpers |
| `start_candidate` | Boot candidate image against seeded state | `StartShadow` |
| `sync_mutations` | Replay journal ops to shadow until catch-up | — |
| `verify_consistency` | Compare manifests (pre-barrier may allow drift) | — |
| `health_check` | Probe candidate readiness | `HealthCheck` |
| `write_barrier` | Fence active writers | `EstablishWriteBarrier` |
| `final_delta` | Replay residual journal under barrier | — |
| `traffic_transfer` | Move ingress to shadow | `TransferTraffic` |
| `complete` | Promote roles; retain standby; release barrier | `PromoteShadow` |
| `failed` | Abort; clean up shadow as policy dictates | `DestroySlot` optional |
| `rolled_back` | Traffic → standby; epoch pin; integrity verify | `Rollback` |

Note: implementation may order barrier vs final-delta labeling slightly
differently in logs; **invariant** is: hard verify and traffic transfer occur
only while writers are fenced, and barrier releases only after promote or abort.

---

## 4. Invariants

1. **Single writer fence** — under barrier, no unjournaled productive writes
   (agent-enforced).
2. **Hard verify before traffic** — post-final-delta manifests must match.
3. **Active untouched on abort** — pre-promote failures do not re-point traffic.
4. **Epoch monotonicity** — epoch increases only on successful complete; rollback
   sets epoch to standby’s.
5. **No shared exclusive mount** — adapters must not violate platform rules
   (Railway: one volume ↔ one live deployment).
6. **Standby retention** — promote keeps previous active available for rollback
   until intentionally destroyed.

---

## 5. Failure handling (summary)

| Failure | Behavior |
|---------|----------|
| Shadow unhealthy pre-cutover | Abort; active untouched; destroy/idle shadow |
| Verify fail under barrier | Release barrier; abort; active continues |
| Crash mid-barrier | Active fenced until operator/agent TTL (TTL **planned**) |
| Crash post-promote pre-release | Traffic on new active; barrier release on restart |
| Rollback integrity fail | Do not claim success; surface error in `RollbackReport` |

Full table: `ARCHITECTURE.md`. Threat detail: `SECURITY_MODEL.md`.

---

## 6. Operator-visible ledger

Successful deploys should expose at least:

- `deploy_id`, from/to image, from/to epoch
- `mutations_synced`, `final_delta_writes`
- `barrier_duration`, `total_duration`
- `consistency_ok`, `health_ok`
- active / shadow / standby slot IDs

This ledger is the UX hook for “STATEFUL ZERO-DOWNTIME DEPLOYMENT” in platform
CLIs and dashboards.

---

## 7. Non-goals of this protocol (v0)

- Active-active multi-writer
- Transparent FUSE without agent (future optional)
- Cross-region synchronous cutover
- Replacing DB logical replication protocols
