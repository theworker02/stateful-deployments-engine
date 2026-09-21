# Failure atlas (v1.0.0)

Formal map from chaos / fault scenarios to invariants and recovery.
Authoritative test IDs live in `internal/chaos` and related `*_test.go` files.

| Failure | Detection | Invariant | Recovery | RPO | RTO | Test / receipt | Status |
|---------|-----------|-----------|----------|-----|-----|----------------|--------|
| Coordinator death mid-phase | Persisted FSM (`store`) | Deterministic recover path | `sde recover` resume/retry/abort | ≤ journal lag | seconds | chaos + store tests | Covered |
| Source death during sync | Health / I/O errors | Active may degrade | Abort or forward path | pending mutations | — | chaos matrix | Covered |
| Candidate death | HealthCheck fail | Active untouched | Destroy candidate | 0 on active | — | coordinator tests | Covered |
| Disk full | Write/export errors | Fail closed | Abort; retain last good archive | — | — | archive/export errors | Covered |
| Archive corruption | `verify-archive` | UNKNOWN≠VERIFIED | NOT_RECOVERABLE | — | — | archive tests | Covered |
| Journal corruption / truncate | CRC / parse | No silent discard | Resume last good seq | ≤ torn write | — | journal fuzz | Covered |
| Generic transport interrupt | Adapter I/O error | No false VERIFIED | Retry sync / abort | backlog | — | sync error paths | Covered |
| Verification mismatch | Receipt FAILED | Gate blocks cutover | Keep active | 0 | — | verifier tests | Covered |
| Cutover interruption | Phase persistence | Recoverable FSM | recover / rollback class | barrier window | — | crash fuzz | Covered |
| Rollback with ack writes | Classify() | No silent loss | FORWARD_RECOVERY_REQUIRED | ack writes kept | — | rollback tests | Covered |
| New writes after cutover | WritesAfterCutover counter | Same | Forward recovery | — | — | chaos | Covered |
| Lost source environment | Independence harness | PSA usable alone | Restore + receipt | epoch | measured | evaluate-harness | Covered |

## Multi-failure matrix

See `PHASE3_CHAOS_MATRIX.json` / `testdata/chaos/` and `sde chaos` for combination scenarios
(high write load + coordinator crash, disk pressure + corrupt journal, etc.).

## Related

[`FAILURE_MODEL.md`](FAILURE_MODEL.md) · [`SAFETY_INVARIANTS.md`](SAFETY_INVARIANTS.md)
