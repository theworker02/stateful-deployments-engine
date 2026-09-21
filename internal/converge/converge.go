// Package converge detects catch-up dynamics and applies bounded mitigations.
package converge

import (
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Sample is one observation of write vs replication rates.
type Sample struct {
	WriteRateOpsPerSec       float64
	ReplicationRateOpsPerSec float64
	BacklogOps               uint64
	At                       time.Time
}

// Observe builds a ConvergenceSnapshot. Non-convergence is detected when
// WRITE_RATE >= REPLICATION_RATE with positive backlog — success is never invented.
func Observe(s Sample, priorBacklog uint64, priorAt time.Time) types.ConvergenceSnapshot {
	if s.At.IsZero() {
		s.At = time.Now().UTC()
	}
	vel := 0.0
	if !priorAt.IsZero() && s.At.After(priorAt) {
		dt := s.At.Sub(priorAt).Seconds()
		if dt > 0 {
			vel = (float64(s.BacklogOps) - float64(priorBacklog)) / dt
		}
	}
	nonConv := s.WriteRateOpsPerSec >= s.ReplicationRateOpsPerSec && s.BacklogOps > 0
	converging := !nonConv && s.BacklogOps > 0 && s.ReplicationRateOpsPerSec > s.WriteRateOpsPerSec
	est := 0.0
	net := s.ReplicationRateOpsPerSec - s.WriteRateOpsPerSec
	if net > 0 && s.BacklogOps > 0 {
		est = float64(s.BacklogOps) / net * 1000
	} else if s.BacklogOps > 0 {
		est = -1
	}
	var mit []string
	if nonConv {
		mit = SuggestMitigations(s)
	}
	return types.ConvergenceSnapshot{
		WriteRateOpsPerSec:       s.WriteRateOpsPerSec,
		ReplicationRateOpsPerSec: s.ReplicationRateOpsPerSec,
		BacklogOps:               s.BacklogOps,
		BacklogVelocityOpsPerSec: vel,
		EstimatedCatchupMs:       est,
		Converging:               converging,
		NonConvergent:            nonConv,
		MitigationsApplied:       mit,
		ObservedAt:               s.At,
	}
}

// SuggestMitigations returns explainable strategies; does not claim they succeed.
func SuggestMitigations(s Sample) []string {
	return []string{
		"hot-object-prioritization",
		"bounded-quiescence",
		"adaptive-batching",
		"compression",
		"sparse-transfer",
		"unchanged-block-elimination",
	}
}
