# Release Notes — v0.4.0

**Stateful Deployments Engine** — Portable state + disaster recovery (ENGINE_PROVIDED).

## Highlights

- Portable State Archives: export / verify / import / clone
- Escrow catalog with LOCAL_FILESYSTEM / S3_COMPATIBLE / GENERIC_OBJECT_STORE targets
- Recovery points, DR planner, fire drills, recovery receipts
- Headline automated recovery independence test
- Docs site + brand + acquisition package
- Clear separation: `RAILWAY_NATIVE` ≠ `ENGINE_PROVIDED`

## Install / build

```bash
go build -o sde.exe ./cmd/sde
./sde.exe version   # sde 0.4.0
```

## Compatibility

- Go 1.22+
- Local adapter fully functional; Railway remains stub (`ErrNotWired`)

## Not in this release

- Live Railway GraphQL cutover
- Cross-region networking / mesh / tunnels
- Sub-mutation PITR finer than epoch/journal position
