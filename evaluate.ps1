# evaluate.ps1 — one-command acquisition evaluation (Windows)
$ErrorActionPreference = "Stop"
$Root = Join-Path $PSScriptRoot "evaluation\.run"
$Out = Join-Path $PSScriptRoot "evaluation"
$Ver = "1.1.1"
New-Item -ItemType Directory -Force -Path $Out | Out-Null
Remove-Item -Recurse -Force $Root -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $Root | Out-Null

Write-Host "== build =="
go build -o sde.exe ./cmd/sde
.\sde.exe version

Write-Host "== doctor =="
.\sde.exe doctor --root $Root

Write-Host "== tests =="
go test ./... -count=1
$testExit = $LASTEXITCODE

Write-Host "== fixtures + archive + restore =="
go run ./cmd/evaluate-harness 2>$null
if ($LASTEXITCODE -ne 0) {
  go run ./examples/acquisition-demo
}

Write-Host "== perf gates =="
go test ./benchmarks -run TestPerfRegressionGates -count=1

$summary = @"
# Evaluation Summary

**Version:** $Ver
**Date:** $(Get-Date -Format o)
**Platform:** Windows
**go test exit:** $testExit

## Commands

``````
go test ./...
go build -o sde.exe ./cmd/sde
.\sde.exe doctor
go run ./examples/acquisition-demo
go test ./benchmarks -run TestPerfRegressionGates -count=1
``````

## Artifacts

See TEST_RESULTS.json, BENCHMARKS.json, RECOVERY_RECEIPT.json, MIGRATION_RECEIPT.json in this directory (populated by evaluate harness / demos).

Independent project — not affiliated with Railway.
"@
Set-Content -Path (Join-Path $Out "SUMMARY.md") -Value $summary -Encoding utf8
@{ result = $(if ($testExit -eq 0) { "PASS" } else { "FAIL" }); exit_code = $testExit; version = $Ver } | ConvertTo-Json | Set-Content (Join-Path $Out "TEST_RESULTS.json")
Write-Host "Wrote $Out\SUMMARY.md (test_exit=$testExit)"
exit $testExit
