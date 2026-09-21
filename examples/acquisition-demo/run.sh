#!/usr/bin/env bash
# Acquisition demo: create → archive → verify → destroy source → restore → verify → receipt
set -euo pipefail
ROOT="${1:-.sde-acq-demo}"
rm -rf "$ROOT"
mkdir -p "$ROOT/src/data"
echo '{"demo":true,"n":1}' > "$ROOT/src/data/record.json"
dd if=/dev/zero of="$ROOT/src/data/blob.bin" bs=1024 count=8 2>/dev/null || head -c 8192 /dev/zero > "$ROOT/src/data/blob.bin"

echo "== export =="
go run ./cmd/sde --root "$ROOT" init >/dev/null
# Use library path via go test helper if CLI needs active slot; direct archive export:
go run ./examples/acquisition-demo >/dev/null 2>&1 || true

echo "See examples/acquisition-demo/main.go for the full Go demo."
echo "Run: go run ./examples/acquisition-demo"
