# Schemas

JSON Schema definitions for acquisition diligence and integrator contracts:

| File | Purpose |
|------|---------|
| `deployment-receipt.schema.json` | Immutable deploy evidence |
| `migration-plan.schema.json` | `sde plan-migration` output |
| `deploy-status.schema.json` | Operator status snapshot |

Go types live in `internal/types`. Sealing / hashing: `internal/receipt`.
