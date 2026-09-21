#Requires -Version 5.1
<#
.SYNOPSIS
  Build SDE and optionally install sde.exe to a local bin directory.

.EXAMPLE
  .\scripts\install.ps1
  .\scripts\install.ps1 -Prefix "$env:USERPROFILE\bin" -AddToPath
#>
param(
  [string]$Prefix = (Join-Path $PSScriptRoot "..\dist"),
  [switch]$AddToPath,
  [switch]$SkipTest
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $RepoRoot

Write-Host "== SDE install ==" -ForegroundColor Cyan
Write-Host "repo: $RepoRoot"

if (-not $SkipTest) {
  Write-Host "== go test =="
  go test ./... -count=1
  if ($LASTEXITCODE -ne 0) { throw "tests failed" }
}

New-Item -ItemType Directory -Force -Path $Prefix | Out-Null
$Out = Join-Path $Prefix "sde.exe"
Write-Host "== build → $Out =="
go build -ldflags "-s -w" -o $Out ./cmd/sde
if ($LASTEXITCODE -ne 0) { throw "build failed" }

& $Out version
& $Out doctor --root .sde

if ($AddToPath) {
  $abs = (Resolve-Path $Prefix).Path
  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if ($userPath -notlike "*$abs*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$abs", "User")
    $env:Path = "$env:Path;$abs"
    Write-Host "Added to user PATH: $abs (restart shells to pick up)" -ForegroundColor Green
  } else {
    Write-Host "Already on user PATH: $abs"
  }
}

Write-Host ""
Write-Host "Next:" -ForegroundColor Cyan
Write-Host "  & `"$Out`" demo --root .sde"
Write-Host "  .\evaluate.ps1"
Write-Host "  docs\GETTING_STARTED.md"
