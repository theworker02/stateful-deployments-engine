# Release Notes — v0.5.0 (Acquisition Candidate)

**Stateful Deployments Engine** — diligence-ready offline evaluation pack.

## Highlights

- Correct Railway framing (gap matrix; ENGINE_PROVIDED vs RAILWAY_NATIVE)
- One-command evaluate harness
- `sde railway clone` with CLONE_RECEIPT
- Failure atlas + fuzz + perf gates
- SBOM / IP / transfer / security review pack

## Build

```bash
go test ./...
go build -o sde.exe ./cmd/sde
./sde.exe version   # sde 0.5.0
```

Checksums: run `Get-FileHash sde.exe` / `sha256sum sde` after build and record in diligence notes.
**Not claiming signed releases** unless signatures are produced separately.

## Compatibility

- Go 1.22+
- Local adapter full; Railway live still stub

Independent project — not affiliated with Railway.
