// Package sync quantifies shadow catch-up; file-size equality is never "complete".
package sync

import (
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// Engine computes lag metrics for a candidate relative to the journal cursor.
type Engine struct {
	Journal *journal.Journal
}

// Measure returns SyncLagMetrics. ReadyForCutover requires zero journal lag
// AND (optionally) a prior verification — never size equality alone.
func (e *Engine) Measure(candidateRoot string, cursor uint64, replayRateOps float64, integrity string) (*types.SyncLagMetrics, error) {
	pending, err := e.Journal.ReadFrom(cursor)
	if err != nil {
		return nil, err
	}
	var bytesRem int64
	var oldestSeq uint64
	var oldestAt time.Time
	for i, m := range pending {
		bytesRem += m.Size
		if len(m.Payload) > int(m.Size) {
			bytesRem += int64(len(m.Payload)) - m.Size
		}
		if i == 0 {
			oldestSeq = m.Seq
			oldestAt = m.Timestamp
		}
	}
	lag := uint64(len(pending))
	estMs := 0.0
	if replayRateOps > 0 && lag > 0 {
		estMs = float64(lag) / replayRateOps * 1000
	} else if lag > 0 {
		estMs = -1 // unknown
	}
	if integrity == "" {
		integrity = "unknown"
	}
	// Size match of roots is intentionally NOT used as readiness.
	ready := lag == 0 && integrity == "ok"
	return &types.SyncLagMetrics{
		BytesRemaining:         bytesRem,
		OperationsRemaining:    lag,
		JournalLag:             lag,
		OldestPendingMutation:  oldestSeq,
		OldestPendingAt:        oldestAt,
		EstimatedCatchupTimeMs: estMs,
		CandidateIntegrity:     integrity,
		ReplayRateOpsPerSec:    replayRateOps,
		ReadyForCutover:        ready,
	}, nil
}

// IntegrityFromVerify maps a receipt to an integrity label without treating UNKNOWN as ok.
func IntegrityFromVerify(r *types.VerificationReceipt) string {
	if r == nil {
		return "unknown"
	}
	switch r.Outcome {
	case types.VerifyVerified:
		return "ok"
	case types.VerifyPartial:
		return "degraded"
	case types.VerifyFailed:
		return "degraded"
	default:
		return "unknown"
	}
}

// CompareSizesOnly is documented as insufficient for cutover readiness.
func CompareSizesOnly(activeRoot, candidateRoot string) (same bool, err error) {
	a, err := verifier.BuildCheckpoint(activeRoot, 0)
	if err != nil {
		return false, err
	}
	c, err := verifier.BuildCheckpoint(candidateRoot, 0)
	if err != nil {
		return false, err
	}
	return a.ByteSize == c.ByteSize && a.FileCount == c.FileCount, nil
}
