# Live Railway GraphQL adapter (v1.0+)

Independent project — not affiliated with Railway.

## Enable live mode

```bash
export SDE_RAILWAY_LIVE=1
export RAILWAY_TOKEN="<account-or-workspace-token>"   # or RAILWAY_API_TOKEN / RAILWAY_PROJECT_TOKEN
export RAILWAY_PROJECT_ID="<project-id>"
export RAILWAY_ENVIRONMENT_ID="<environment-id>"
export RAILWAY_SERVICE_ID="<active-service-id>"
# optional:
export RAILWAY_SHADOW_SERVICE_ID="<existing-shadow>"  # else created via serviceCreate
export RAILWAY_PUBLIC_DOMAIN="app.example.com"
export RAILWAY_ACTIVE_STATE_PATH="/path/to/active/volume/staging"
export RAILWAY_SHADOW_STATE_PATH="/path/to/shadow/volume/staging"
```

Endpoint: `https://backboard.railway.com/graphql/v2`

## What live mode does

| Operation | GraphQL / behavior |
|-----------|-------------------|
| Auth check | `me` or `projectToken` |
| Active slot | `service(id)` |
| Create candidate | `serviceCreate` (+ best-effort `volumeCreate`) |
| Start candidate | `serviceInstanceDeployV2` |
| Health | `serviceInstance.latestDeployment.status` |
| Write barrier | `variableCollectionUpsert` → `SDE_WRITE_BARRIER` |
| Traffic shift | `serviceDomainCreate` + role variables |
| Rollback | `deploymentRollback` or redeploy standby |

Volume **byte sync** uses staged paths (`RAILWAY_*_STATE_PATH` / `RAILWAY_STATE_STAGING`)
for ENGINE_PROVIDED journal/PSA workflows. Cross-environment DR remains PSA-based.

## Offline mode

Without `SDE_RAILWAY_LIVE=1`, methods return `ErrNotWired`. ENGINE_PROVIDED
`CloneVolume` / PSA paths still work without live API.
