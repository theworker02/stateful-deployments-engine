# Local release signing helper (PowerShell)
# Prefers cosign keyless if available; otherwise produces SHA-256 only and notes signing status.
param(
  [string]$DistDir = "dist"
)

$ErrorActionPreference = "Stop"
if (-not (Test-Path $DistDir)) { New-Item -ItemType Directory -Path $DistDir | Out-Null }

$bins = Get-ChildItem $DistDir -File | Where-Object { $_.Name -like "sde*" -and $_.Extension -ne ".bundle" }
if (-not $bins) {
  Write-Host "No binaries in $DistDir — building sde.exe"
  go build -o "$DistDir/sde.exe" ./cmd/sde
  $bins = @(Get-Item "$DistDir/sde.exe")
}

$sums = Join-Path $DistDir "SHA256SUMS.txt"
$lines = @()
foreach ($b in $bins) {
  $h = (Get-FileHash $b.FullName -Algorithm SHA256).Hash.ToLower()
  $lines += "$h  $($b.Name)"
}
$lines | Set-Content -Encoding ascii $sums
Write-Host "Wrote $sums"

$cosign = Get-Command cosign -ErrorAction SilentlyContinue
if ($cosign) {
  Write-Host "Signing with cosign..."
  foreach ($b in $bins) {
    & cosign sign-blob --yes --bundle "$($b.FullName).cosign.bundle" $b.FullName
  }
  & cosign sign-blob --yes --bundle "$sums.cosign.bundle" $sums
  "SIGNED=cosign" | Set-Content -Encoding ascii (Join-Path $DistDir "SIGNING_STATUS.txt")
} else {
  "SIGNED=checksums-only (install cosign for blob signatures)" | Set-Content -Encoding ascii (Join-Path $DistDir "SIGNING_STATUS.txt")
  Write-Host "cosign not found — checksums written; CI release workflow signs via OIDC"
}

Get-Content (Join-Path $DistDir "SIGNING_STATUS.txt")
