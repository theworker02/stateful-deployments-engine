#!/usr/bin/env bash
# Build SDE and install the binary to PREFIX (default: ./dist).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PREFIX="${PREFIX:-$ROOT/dist}"
SKIP_TEST="${SKIP_TEST:-0}"
cd "$ROOT"

echo "== SDE install =="
echo "repo: $ROOT"

if [[ "$SKIP_TEST" != "1" ]]; then
  echo "== go test =="
  go test ./... -count=1
fi

mkdir -p "$PREFIX"
OUT="$PREFIX/sde"
echo "== build → $OUT =="
go build -ldflags "-s -w" -o "$OUT" ./cmd/sde

"$OUT" version
"$OUT" doctor --root .sde || true

echo
echo "Next:"
echo "  $OUT demo --root .sde"
echo "  ./evaluate.sh"
echo "  docs/GETTING_STARTED.md"
echo
echo "Optional: export PATH=\"$PREFIX:\$PATH\""
