// Package metrics defines canonical metric names and typed samples for SDE.
//
// Acquisition evaluations key off CUTOVER_WRITE_PAUSE_MS and related gauges.
package metrics

import "time"

// Canonical Prometheus-style metric names (no registry wiring required).
const (
	CutoverWritePauseMs     = "sde_cutover_write_pause_ms"
	CutoverPredictedPauseMs = "sde_cutover_predicted_pause_ms"
	SyncLagOps              = "sde_sync_lag_ops"
	SyncThroughputOps       = "sde_sync_throughput_ops_per_sec"
	SyncThroughputBytes     = "sde_sync_throughput_bytes_per_sec"
	VerifyDurationMs        = "sde_verify_duration_ms"
	RollbackLatencyMs       = "sde_rollback_latency_ms"
	RecoveryTimeMs          = "sde_recovery_time_ms"
	JournalReplayOps        = "sde_journal_replay_ops_per_sec"
	ObservationWindowMs     = "sde_observation_window_ms"
	RPOMutations            = "sde_rpo_mutations"
	RTOSeconds              = "sde_rto_seconds"
	DeployTotalMs           = "sde_deploy_total_ms"
	CorruptionDetected      = "sde_corruption_detected"
)

// Sample is one labeled observation.
type Sample struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Bundle groups samples from a single deploy evaluation.
type Bundle struct {
	DeployID string   `json:"deploy_id"`
	Samples  []Sample `json:"samples"`
}

// NewSample constructs a Sample with UTC timestamp.
func NewSample(name string, value float64, unit string, labels map[string]string) Sample {
	return Sample{
		Name:      name,
		Value:     value,
		Unit:      unit,
		Labels:    labels,
		Timestamp: time.Now().UTC(),
	}
}

// FromPause builds the acquisition-critical cutover pause samples.
func FromPause(deployID string, actualMs, predictedMs float64) Bundle {
	labels := map[string]string{"deploy_id": deployID}
	return Bundle{
		DeployID: deployID,
		Samples: []Sample{
			NewSample(CutoverWritePauseMs, actualMs, "ms", labels),
			NewSample(CutoverPredictedPauseMs, predictedMs, "ms", labels),
		},
	}
}

// Find returns the first sample with name, or false.
func (b Bundle) Find(name string) (Sample, bool) {
	for _, s := range b.Samples {
		if s.Name == name {
			return s, true
		}
	}
	return Sample{}, false
}
