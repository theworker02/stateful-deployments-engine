# Railway Gap Matrix (factual)

**Independent project — not affiliated with Railway.**  
This matrix compares SDE’s ENGINE_PROVIDED capabilities to **documented** Railway platform features. It does **not** invent Railway deficiencies.

Sources (operator-facing docs): Railway Volumes / Backups / Postgres references as publicly documented. Verify against current Railway docs at diligence time.

## What Railway already provides

| Capability | Railway (documented) | SDE relationship |
|------------|----------------------|------------------|
| Persistent volumes | Yes — attach volumes to services | SDE coordinates *state lifecycle around* volumes; does not replace volumes |
| Incremental / scheduled backups | Yes — volume backup features exist | Complementary: SDE archives are **ENGINE_PROVIDED**, portable by design |
| Postgres PITR | Yes — for Railway Postgres | Out of scope for SDE core; SDE targets **arbitrary volume-backed workloads**, not DB-native HA |
| Portable Postgres logical dumps | Yes — logical dump tooling | Complementary for DB apps; SDE covers non-DB / generic FS state |
| Live volume resize | Yes | Orthogonal; SDE does not claim resize |

## Documented Railway constraints (still relevant)

| Constraint | Implication for SDE pitch |
|------------|---------------------------|
| Native backup restore is same project / environment | Cross-env / escrow / offline evaluate needs **ENGINE_PROVIDED** Portable State Archives |
| Wiping a volume removes volume backups | Operators who need independent recovery artifacts should escrow ENGINE_PROVIDED archives off-volume |
| One volume per service; replicas incompatible with volumes | Motivates dual-service / dual-volume **deployment-safe** cutover patterns via adapters |

## What SDE adds (do not over-claim)

| SDE capability | Not a claim that Railway lacks |
|----------------|--------------------------------|
| Transactional deploy FSM for stateful cutover | Not “Railway has no deploys” |
| Mutation journal + verified barrier cutover | Not “Railway has no backups” |
| ENGINE_PROVIDED PSA export/import/escrow | Independent recovery artifacts beyond volume-coupled backups |
| Chaos / fire-drill / recovery receipts | Evaluation harness for acquisition diligence |
| Generic volume workload fixtures | Beyond DB-specific tooling |

## Modes (must stay distinct)

| Mode | Meaning |
|------|---------|
| `RAILWAY_NATIVE` | Platform volume/backup APIs — subject to Railway project/env/lifecycle rules |
| `ENGINE_PROVIDED` | SDE Portable State Archive — platform-neutral; **not Railway-endorsed** |

## Anti-pitch (do not use)

- ❌ “Railway doesn’t have backups”
- ❌ “Only SDE can restore state”
- ❌ Implied Railway endorsement or partnership
