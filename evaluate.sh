#!/usr/bin/env bash
# One-command acquisition evaluation
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
OUT="$ROOT/evaluation"
RUN="$OUT/.run"
VER="1.1.0"
mkdir -p "$OUT"
rm -rf "$RUN"
mkdir -p "$RUN"

echo "== build =="
(cd "$ROOT" && go build -o sde ./cmd/sde)
"$ROOT/sde" version

echo "== doctor =="
"$ROOT/sde" doctor --root "$RUN" || true

echo "== tests =="
set +e
(cd "$ROOT" && go test ./... -count=1)
TEST_EXIT=$?
set -e

echo "== acquisition demo (recovery receipt) =="
(cd "$ROOT" && go run ./examples/acquisition-demo) | tee "$OUT/demo.log"

echo "== perf gates =="
(cd "$ROOT" && go test ./benchmarks -run TestPerfRegressionGates -count=1)

echo "== evaluate harness =="
(cd "$ROOT" && go run ./cmd/evaluate-harness -out "$OUT" -run "$RUN") || true

cat > "$OUT/SUMMARY.md" <<EOF
# Evaluation Summary

**Version:** $VER
**go test exit:** $TEST_EXIT

Independent project — not affiliated with Railway.

Artifacts in this directory: TEST_RESULTS.json, BENCHMARKS.json, RECOVERY_RECEIPT.json, MIGRATION_RECEIPT.json.
EOF

printf '{"result":"%s","exit_code":%s,"version":"%s"}\n' \
  "$( [ $TEST_EXIT -eq 0 ] && echo PASS || echo FAIL )" "$TEST_EXIT" "$VER" > "$OUT/TEST_RESULTS.json"

exit $TEST_EXIT
