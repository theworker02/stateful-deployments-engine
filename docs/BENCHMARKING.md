# Benchmarking (v1.0.0)

## Rules

- **Never hard-code** marketing numbers into docs or the Pages site  
- Site reads committed JSON under `site/data/` (copied from evaluation/benches)  
- Perf gates use **tolerances**, not flaky micro-timings  

## Suites

| Suite | Location |
|-------|----------|
| Cutover pause | `benchmarks/` |
| Chunk differential | `benchmarks/results/chunk_differential.json` |
| Perf gates | `benchmarks/results/perf_gates.json` · `benchmarks/perf_gates_test.go` |
| Workloads | `workloads/` (SMALL → LOG-LIKE) |
| Evaluate harness | `evaluation/BENCHMARKS.json` |

## Metrics of interest

- `CUTOVER_WRITE_PAUSE_MS` (predicted vs actual)  
- Archive / restore / journal / verify throughput  
- RPO (mutations pending at barrier) · RTO (restore / recover)  
- Dedup ratio (measured logical vs physical)  

## Running

```bash
go test ./benchmarks -bench=. -benchtime=3x
.\evaluate.ps1
```
