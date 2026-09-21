#!/usr/bin/env bash
# Local / CI helper: checksum + optional cosign sign-blob
set -euo pipefail
DIST="${1:-dist}"
mkdir -p "$DIST"
if ! ls "$DIST"/sde* >/dev/null 2>&1; then
  go build -o "$DIST/sde" ./cmd/sde
fi
(
  cd "$DIST"
  sha256sum sde* 2>/dev/null | grep -v '\.bundle$' | grep -v SHA256SUMS > SHA256SUMS.txt || shasum -a 256 sde* | grep -v bundle > SHA256SUMS.txt
)
if command -v cosign >/dev/null 2>&1; then
  while read -r _hash file; do
    [[ "$file" == *.bundle ]] && continue
    [[ "$file" == SHA256SUMS.txt ]] && continue
    cosign sign-blob --yes --bundle "${DIST}/${file}.cosign.bundle" "${DIST}/${file}"
  done < "${DIST}/SHA256SUMS.txt"
  cosign sign-blob --yes --bundle "${DIST}/SHA256SUMS.txt.cosign.bundle" "${DIST}/SHA256SUMS.txt"
  echo "SIGNED=cosign" > "${DIST}/SIGNING_STATUS.txt"
else
  echo "SIGNED=checksums-only (install cosign for blob signatures)" > "${DIST}/SIGNING_STATUS.txt"
fi
cat "${DIST}/SIGNING_STATUS.txt"
