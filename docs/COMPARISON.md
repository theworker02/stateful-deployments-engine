# Comparison (v1.0.0)

How SDE relates to adjacent tools — for diligence, not marketing slam pieces.

| Approach | What it solves | What SDE adds / differs |
|----------|----------------|-------------------------|
| `rsync` / volume copy | Bulk file copy | Journaled live delta, fencing, cutover barrier, receipts |
| CRIU | Process checkpoint/restore | Deployment transaction + portable archive lifecycle (not full process CRIU) |
| DB native HA / PITR | Engine-specific durability | Works for **arbitrary FS state**, not only Postgres/MySQL |
| Platform volume backups | Provider snapshots | ENGINE_PROVIDED archives that survive project/env wipe; fire drills |
| K8s StatefulSet rolling | Pod replace with PVC | Explicit blue/green state sync + classified rollback |
| Backup utilities only | Point restore | Plus **deployment-safe cutover** while source stays hot |

SDE does **not** replace Railway-native Postgres backups/PITR. It complements them for
generic volume workloads and transactional deploy semantics.
