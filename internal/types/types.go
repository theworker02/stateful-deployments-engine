// Package types defines the shared domain model for the state transition engine.
package types

import (
	"time"
)

// OpKind classifies a filesystem mutation recorded in the journal.
type OpKind string

const (
	OpCreate OpKind = "create"
	OpWrite  OpKind = "write"
	OpRename OpKind = "rename"
	OpDelete OpKind = "delete"
	OpChmod  OpKind = "chmod"
	OpTrunc  OpKind = "truncate"
)

// ReplayStatus tracks whether a mutation has been applied to a candidate.
type ReplayStatus string

const (
	ReplayPending   ReplayStatus = "pending"
	ReplayApplied   ReplayStatus = "applied"
	ReplayDuplicate ReplayStatus = "duplicate"
	ReplayFailed    ReplayStatus = "failed"
	ReplaySkipped   ReplayStatus = "skipped"
)

// Mutation is one append-only journal entry describing a state change.
type Mutation struct {
	Seq           uint64       `json:"seq"`
	Epoch         uint64       `json:"epoch"`
	FenceToken    uint64       `json:"fence_token,omitempty"`
	DeploymentID  string       `json:"deployment_id,omitempty"`
	Resource      string       `json:"resource,omitempty"` // volume / state resource id
	Kind          OpKind       `json:"kind"`
	Path          string       `json:"path"`
	DestPath      string       `json:"dest_path,omitempty"`
	Offset        int64        `json:"offset,omitempty"`
	Mode          uint32       `json:"mode,omitempty"`
	Size          int64        `json:"size,omitempty"`
	ContentHash   string       `json:"content_hash,omitempty"`
	PrevStateRef  string       `json:"prev_state_ref,omitempty"`
	NewStateRef   string       `json:"new_state_ref,omitempty"`
	TxnID         string       `json:"txn_id,omitempty"`
	ClientOpID    string       `json:"client_op_id,omitempty"` // duplicate detection key
	Payload       []byte       `json:"payload,omitempty"`
	Timestamp     time.Time    `json:"ts"`
	ReplayStatus  ReplayStatus `json:"replay_status,omitempty"`
	EntryChecksum string       `json:"entry_checksum,omitempty"`
}

// Checkpoint is a consistent snapshot of volume state at an epoch boundary.
type Checkpoint struct {
	Epoch     uint64            `json:"epoch"`
	RootHash  string            `json:"root_hash"`
	FileCount int               `json:"file_count"`
	ByteSize  int64             `json:"byte_size"`
	CreatedAt time.Time         `json:"created_at"`
	Manifest  map[string]string `json:"manifest,omitempty"` // path -> content hash
}

// JournalCheckpoint marks a durable replay cursor in the mutation log.
type JournalCheckpoint struct {
	Seq       uint64    `json:"seq"`
	Epoch     uint64    `json:"epoch"`
	CreatedAt time.Time `json:"created_at"`
	Note      string    `json:"note,omitempty"`
}

// DeploymentRole is the blue/green role of a deployment slot.
type DeploymentRole string

const (
	RoleActive  DeploymentRole = "active"
	RoleShadow  DeploymentRole = "shadow"
	RoleStandby DeploymentRole = "standby" // previous active retained for rollback
)

// DeploymentSlot is a platform-agnostic handle to a running candidate or active deploy.
type DeploymentSlot struct {
	ID        string         `json:"id"`
	Role      DeploymentRole `json:"role"`
	ImageRef  string         `json:"image_ref"`
	StatePath string         `json:"state_path"` // where this slot's volume/state lives
	Healthy   bool           `json:"healthy"`
	CreatedAt time.Time      `json:"created_at"`
}

// DeployPhase tracks progress through a transactional stateful deploy.
// Phase 2 central state machine (persisted; survives process restart):
//
//	ACTIVE → CHECKPOINT → SHADOW → SYNCHRONIZING → VERIFYING → QUIESCING →
//	FINAL_DELTA → CUTOVER → OBSERVING → COMMITTED
//
// Unsafe paths: ABORTED | DEGRADED | RECOVERING | ROLLED_BACK
type DeployPhase string

const (
	PhaseActive         DeployPhase = "ACTIVE"
	PhaseCheckpoint     DeployPhase = "CHECKPOINT"
	PhaseShadow         DeployPhase = "SHADOW"
	PhaseSynchronizing  DeployPhase = "SYNCHRONIZING"
	PhaseVerifying      DeployPhase = "VERIFYING"
	PhaseQuiescing      DeployPhase = "QUIESCING"
	PhaseFinalDelta     DeployPhase = "FINAL_DELTA"
	PhaseCutover        DeployPhase = "CUTOVER"
	PhaseObserving      DeployPhase = "OBSERVING"
	PhaseCommitted      DeployPhase = "COMMITTED"
	PhaseAborted        DeployPhase = "ABORTED"
	PhaseDegraded       DeployPhase = "DEGRADED"
	PhaseRecovering     DeployPhase = "RECOVERING"
	PhaseRolledBack     DeployPhase = "ROLLED_BACK"

	// Legacy aliases (Phase 1 CLI/tests).
	PhaseIdle              DeployPhase = "idle"
	PhaseCreateCandidate   DeployPhase = "create_candidate"
	PhaseCaptureCheckpoint DeployPhase = "capture_checkpoint"
	PhaseStartCandidate    DeployPhase = "start_candidate"
	PhaseSyncMutations     DeployPhase = "sync_mutations"
	PhaseVerifyConsistency DeployPhase = "verify_consistency"
	PhaseHealthCheck       DeployPhase = "health_check"
	PhaseWriteBarrier      DeployPhase = "write_barrier"
	PhaseTrafficTransfer   DeployPhase = "traffic_transfer"
	PhaseComplete          DeployPhase = "COMMITTED"
	PhaseFailed            DeployPhase = "ABORTED"
)

// IsTerminal reports whether the phase ends the deploy lifecycle.
func (p DeployPhase) IsTerminal() bool {
	switch p {
	case PhaseCommitted, PhaseAborted, PhaseRolledBack:
		return true
	default:
		return false
	}
}

// IsUnsafe reports unsafe / recovery branches of the FSM.
func (p DeployPhase) IsUnsafe() bool {
	switch p {
	case PhaseAborted, PhaseDegraded, PhaseRecovering, PhaseRolledBack:
		return true
	default:
		return false
	}
}

// VerifyLevel selects how thoroughly candidate state is compared to active.
type VerifyLevel string

const (
	VerifyMetadata           VerifyLevel = "METADATA"
	VerifyChecksum           VerifyLevel = "CHECKSUM"
	VerifyStructural         VerifyLevel = "STRUCTURAL"
	VerifyApplicationDefined VerifyLevel = "APPLICATION_DEFINED"
)

// VerifyOutcome is the result of a consistency check.
// UNKNOWN must never be treated as VERIFIED.
type VerifyOutcome string

const (
	VerifyVerified VerifyOutcome = "VERIFIED"
	VerifyPartial  VerifyOutcome = "PARTIAL"
	VerifyFailed   VerifyOutcome = "FAILED"
	VerifyUnknown  VerifyOutcome = "UNKNOWN"
)

// VerificationReceipt is durable evidence from a consistency check.
type VerificationReceipt struct {
	DeployID        string        `json:"deploy_id"`
	Level           VerifyLevel   `json:"level"`
	Outcome         VerifyOutcome `json:"outcome"`
	Epoch           uint64        `json:"epoch"`
	ActiveRootHash  string        `json:"active_root_hash,omitempty"`
	CandidateHash   string        `json:"candidate_root_hash,omitempty"`
	FileCountActive int           `json:"file_count_active"`
	FileCountCand   int           `json:"file_count_candidate"`
	BytesCompared   int64         `json:"bytes_compared"`
	MissingPaths    []string      `json:"missing_paths,omitempty"`
	ExtraPaths      []string      `json:"extra_paths,omitempty"`
	HashMismatches  []string      `json:"hash_mismatches,omitempty"`
	StructuralNotes []string      `json:"structural_notes,omitempty"`
	AppDefinedOK    *bool         `json:"app_defined_ok,omitempty"`
	EvidenceSummary string        `json:"evidence_summary"`
	CheckedAt       time.Time     `json:"checked_at"`
	DurationMs      float64       `json:"duration_ms"`
}

// SyncLagMetrics exposes candidate catch-up progress (not file-size equality).
type SyncLagMetrics struct {
	BytesRemaining         int64     `json:"bytes_remaining"`
	OperationsRemaining    uint64    `json:"operations_remaining"`
	JournalLag             uint64    `json:"journal_lag"` // active next_seq - candidate cursor
	OldestPendingMutation  uint64    `json:"oldest_pending_mutation,omitempty"`
	OldestPendingAt        time.Time `json:"oldest_pending_at,omitempty"`
	EstimatedCatchupTimeMs float64   `json:"estimated_catchup_time_ms"`
	CandidateIntegrity     string    `json:"candidate_integrity"` // ok|unknown|degraded
	ReplayRateOpsPerSec    float64   `json:"replay_rate_ops_per_sec,omitempty"`
	ReadyForCutover        bool      `json:"ready_for_cutover"`
}

// FenceToken is a monotonic coordinator lease tied to an epoch.
type FenceToken struct {
	Epoch     uint64    `json:"epoch"`
	Token     uint64    `json:"token"`
	HolderID  string    `json:"holder_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// RollbackSafety classifies whether rolling back storage is safe.
type RollbackSafety string

const (
	RollbackSafe                   RollbackSafety = "SAFE"
	RollbackRequiresReconciliation RollbackSafety = "REQUIRES_RECONCILIATION"
	RollbackUnsafe                 RollbackSafety = "UNSAFE"
)

// DeployReport is the transactional deploy ledger shown to operators / platforms.
type DeployReport struct {
	DeployID             string             `json:"deploy_id"`
	FromImage            string             `json:"from_image"`
	ToImage              string             `json:"to_image"`
	FromEpoch            uint64             `json:"from_epoch"`
	ToEpoch              uint64             `json:"to_epoch"`
	FenceToken           uint64             `json:"fence_token,omitempty"`
	MutationsSynced      uint64             `json:"mutations_synced"`
	FinalDeltaWrites     uint64             `json:"final_delta_writes"`
	BarrierDuration      time.Duration      `json:"barrier_duration"`
	CutoverWritePauseMs  float64            `json:"cutover_write_pause_ms"`
	TotalDuration        time.Duration      `json:"total_duration"`
	ConsistencyOK        bool               `json:"consistency_ok"`
	HealthOK             bool               `json:"health_ok"`
	Phase                DeployPhase        `json:"phase"`
	ActiveSlotID         string             `json:"active_slot_id"`
	ShadowSlotID         string             `json:"shadow_slot_id"`
	StandbySlotID        string             `json:"standby_slot_id,omitempty"`
	SyncLag              *SyncLagMetrics    `json:"sync_lag,omitempty"`
	Verification         *VerificationReceipt `json:"verification,omitempty"`
	ObservationWindowMs  float64            `json:"observation_window_ms,omitempty"`
	DryRun               bool               `json:"dry_run,omitempty"`
	Error                string             `json:"error,omitempty"`
}

// RollbackReport records a verified recovery to a prior state epoch.
type RollbackReport struct {
	FromImage       string         `json:"from_image"`
	ToImage         string         `json:"to_image"`
	FromEpoch       uint64         `json:"from_epoch"`
	ToEpoch         uint64         `json:"to_epoch"`
	Duration        time.Duration  `json:"duration"`
	IntegrityOK     bool           `json:"integrity_ok"`
	Safety          RollbackSafety `json:"safety"`
	SafetyReason    string         `json:"safety_reason"`
	WritesAfterCutover uint64      `json:"writes_after_cutover,omitempty"`
	Phase           DeployPhase    `json:"phase,omitempty"`
	Error           string         `json:"error,omitempty"`
}

// RecoveryAction is chosen after crash / restart inspection.
type RecoveryAction string

const (
	RecoverResume               RecoveryAction = "resume"
	RecoverRetry                RecoveryAction = "retry"
	RecoverAbort                RecoveryAction = "abort"
	RecoverRollback             RecoveryAction = "rollback"
	RecoverRequestReconciliation RecoveryAction = "request_reconciliation"
)

// RecoveryDecision is the crash-recovery planner output.
type RecoveryDecision struct {
	DeployID    string         `json:"deploy_id"`
	FromPhase   DeployPhase    `json:"from_phase"`
	Action      RecoveryAction `json:"action"`
	Reason      string         `json:"reason"`
	TargetPhase DeployPhase    `json:"target_phase,omitempty"`
}

// MetricsSnapshot is used by the bench harness and acquisition pitch evaluations.
type MetricsSnapshot struct {
	CutoverPauseMs     float64 `json:"cutover_pause_ms"`
	RPOMutations       uint64  `json:"rpo_mutations"`
	RTOSeconds         float64 `json:"rto_seconds"`
	SyncThroughputOpsS float64 `json:"sync_throughput_ops_s"`
	SyncThroughputMBS  float64 `json:"sync_throughput_mb_s"`
	VerifyDurationMs   float64 `json:"verify_duration_ms"`
	RollbackLatencyMs  float64 `json:"rollback_latency_ms"`
	RecoveryTimeMs     float64 `json:"recovery_time_ms"`
	JournalReplayOpsS  float64 `json:"journal_replay_ops_s"`
	CorruptionDetected bool    `json:"corruption_detected"`
}

// ChaosReceipt is a machine-readable fault-injection outcome.
type ChaosReceipt struct {
	Scenario     string                 `json:"scenario"`
	InjectedAt   time.Time              `json:"injected_at"`
	Detected     bool                   `json:"detected"`
	Recovered    bool                   `json:"recovered"`
	ActionTaken  RecoveryAction         `json:"action_taken,omitempty"`
	FinalPhase   DeployPhase            `json:"final_phase"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	DurationMs   float64                `json:"duration_ms"`
}
