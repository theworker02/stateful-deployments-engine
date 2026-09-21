# Evaluation Summary

**Version:** 1.0.0  
**Harness:** `cmd/evaluate-harness` + `go test ./...`  
**Platform:** windows/amd64  

## Result

| Check | Outcome |
|-------|---------|
| Unit/integration tests | PASS (`go test ./...`) |
| Evaluate harness | PASS |
| Recovery receipt | `evaluation/RECOVERY_RECEIPT.json` |
| Migration receipt | `evaluation/MIGRATION_RECEIPT.json` |
| Benchmarks snapshot | `evaluation/BENCHMARKS.json` |
| Test results JSON | `evaluation/TEST_RESULTS.json` |

## Reproduce

```powershell
.\evaluate.ps1
# or
go test ./...
go run ./cmd/evaluate-harness -out evaluation -run evaluation/.run
```

Independent project — not affiliated with Railway.  
ENGINE_PROVIDED artifacts only; not RAILWAY_NATIVE backups.
