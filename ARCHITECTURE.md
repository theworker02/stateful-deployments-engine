# Architecture

## Positioning

SDE is a **platform-independent state transition engine**. Adapters bind it to
hosting providers; the core never assumes Railway APIs, volume semantics, or
DNS primitives.

Acquisition pitch:

> We've built a platform-independent state transition engine that enables
> blue/green deployment semantics for persistent container workloads.
> Railway's current volume architecture is our first production integration.

Not:

> I noticed Railway can't replicate volumes.

## Core components

```
┌─────────────────────────────────────────────────────────────┐
│                    State Transition Coordinator             │
│  create → checkpoint → start → sync → verify → barrier →    │
│  final-delta → traffic → promote → retain-standby           │
└───────────────┬─────────────────────────────┬───────────────┘
                │                             │
        Mutation Journal              Consistency Verifier
        (append-only ops)             (content-addressed
                                       merkle/manifest)
                │                             │
                └──────── Replay Engine ──────┘
                              │
                     Platform Adapter
              (local | railway | fly | k8s | …)
```

### Mutation journal

Append-only JSONL log of filesystem ops (`create`, `write`, `rename`, `delete`,
`chmod`, `truncate`). Each entry has `(epoch, seq)`. Epochs advance on successful
cutover; rollback pins epoch to the retained standby.

### Consistency verifier

Walks active and shadow state trees, builds `path → sha256` manifests, and
compares root hashes. Cutover is refused if post-barrier manifests diverge.

### Cutover protocol

1. **Create candidate** — adapter provisions shadow slot + independent state store
2. **Checkpoint** — content-addressed snapshot of active state; seed shadow
3. **Start candidate** — boot new image against seeded (possibly incomplete) state
4. **Sync** — continuously replay journal into shadow until quiet
5. **Verify** — pre-barrier check (drift OK if residual mutations exist)
6. **Health** — adapter health probes on shadow
7. **Write barrier** — fence active writers (agent/sidecar)
8. **Final delta** — replay residual journal; hard-verify manifests
9. **Traffic transfer + promote** — ingress → shadow; previous active → standby
10. **Release barrier** — writers resume against new active

Barrier duration is the operator-visible "downtime" — target **tens of milliseconds**
for warm local state, higher under large final deltas / slow storage.

### Rollback engine

Keeps the previous slot as **standby** with its volume intact. Rollback:

- re-points traffic to standby
- sets journal epoch to the standby epoch
- rebuilds checkpoint and optionally compares to a pinned root hash

```
railway rollback   (conceptual UX)

Candidate deployment unhealthy.

Restoring:
  application:  7f21c9a → 98ca12e
  state epoch:  883102  → 883044

Recovery completed: 1.7s
Data integrity: VERIFIED
```

## Railway adapter strategy

From [Railway volumes reference](https://docs.railway.com/volumes/reference):

- one volume per service
- no replicas with volumes
- no concurrent mounts of the same service volume → redeploy downtime

SDE does **not** try to double-mount one Railway volume. Instead:

```
[stable public domain]
         │
    ┌────┴────┐
    ▼         ▼
 Active     Shadow          ← two Railway services
 Volume A   Volume B        ← two volumes (legal)
```

Coordinator seeds B from A, journals mutations, barriers, verifies, then
repoints the domain. Active becomes standby for rollback.

This sits beside Railway's investment in volume backups, PITR, live resize, and
HA databases — complementary infrastructure, not a competing backup utility.

## Adapter interface

See `internal/adapter/adapter.go`:

- `CreateShadow` / `StartShadow` / `HealthCheck`
- `EstablishWriteBarrier` / `TransferTraffic` / `PromoteShadow`
- `Rollback` / `DestroySlot` / `ActiveSlot`

Planned adapters: **railway**, **fly**, **docker**, **kubernetes**, **nomad**.

Detailed cutover state machine: [`docs/PROTOCOL.md`](./docs/PROTOCOL.md).  
Railway adapter strategy: [`docs/RAILWAY.md`](./docs/RAILWAY.md).  
Threat model: [`SECURITY_MODEL.md`](./SECURITY_MODEL.md).  
Acquisition brief: [`ACQUISITION.md`](./ACQUISITION.md).

## Failure model (initial)

| Failure | Behavior |
|---------|----------|
| Shadow unhealthy pre-cutover | Abort; active untouched; destroy shadow |
| Verify fail under barrier | Release barrier; abort; active continues |
| Crash mid-barrier | Active fenced until operator/ intervenes (agent TTL planned) |
| Crash post-promote pre-release | Traffic on new active; barrier release on restart |
| Corruption under fault injection | Bench harness records rate; verify must catch before promote |

## Eval metrics

Exposed via deploy reports + `go test ./benchmarks`:

| Metric | Meaning |
|--------|---------|
| Cutover pause | Write-barrier hold time |
| RPO | Mutations not on shadow at barrier entry |
| RTO | Full deploy wall time / rollback latency |
| Sync throughput | Ops/s and MB/s during catch-up |
| Correctness | Post-barrier manifest equality |
| Rollback latency | Standby restore + integrity verify |

## Non-goals (v0)

- Transparent FUSE interception without an agent (may come later)
- Cross-region volume migration
- Replacing database-native replication (Postgres HA etc.)
- Multi-writer active-active
