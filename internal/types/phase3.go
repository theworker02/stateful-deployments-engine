// Phase 3 domain types: migration planning, convergence, receipts, capabilities.
package types

import "time"

// MigrationConclusion is the planner's readiness verdict.
// UNKNOWN must never be treated as READY.
type MigrationConclusion string

const (
	MigrateReady              MigrationConclusion = "READY"
	MigrateReadyWithRisk      MigrationConclusion = "READY_WITH_RISK"
	MigrateLikelyNonConvergent MigrationConclusion = "LIKELY_NON_CONVERGENT"
	MigrateBlocked            MigrationConclusion = "BLOCKED"
	MigrateUnknown            MigrationConclusion = "UNKNOWN"
)

// MigrationPlan is the durable output of sde plan-migration.
type MigrationPlan struct {
	PlanID              string              `json:"plan_id"`
	SourceSlotID        string              `json:"source_slot_id"`
	TargetImage         string              `json:"target_image"`
	CreatedAt           time.Time           `json:"created_at"`
	StateBytes          int64               `json:"state_bytes"`
	ObjectCount         int                 `json:"object_count"`
	WriteRateOpsPerSec  float64             `json:"write_rate_ops_per_sec"`
	MutationRateBytesS  float64             `json:"mutation_rate_bytes_per_sec"`
	EstimatedBandwidthB float64             `json:"estimated_bandwidth_bytes_per_sec"`
	TargetCapacityBytes int64               `json:"target_capacity_bytes"`
	CheckpointCapable   bool                `json:"checkpoint_capable"`
	VerifyCapable       bool                `json:"verify_capable"`
	HealthRequired      bool                `json:"health_required"`
	HistoricalThroughput float64            `json:"historical_throughput_ops_s"`
	Strategies          []string            `json:"strategies"`
	Risks               []string            `json:"risks"`
	Confidence          float64             `json:"confidence"` // 0..1
	EstimatedCatchupMs  float64             `json:"estimated_catchup_ms"`
	EstimatedPauseMs    float64             `json:"estimated_pause_ms"`
	EstimatedRPO        uint64              `json:"estimated_rpo_mutations"`
	EstimatedRTOSec     float64             `json:"estimated_rto_seconds"`
	Conclusion          MigrationConclusion `json:"conclusion"`
	Rationale           string              `json:"rationale"`
}

// ConvergenceSnapshot captures live catch-up dynamics.
type ConvergenceSnapshot struct {
	WriteRateOpsPerSec        float64   `json:"write_rate"`
	ReplicationRateOpsPerSec  float64   `json:"replication_rate"`
	BacklogOps                uint64    `json:"backlog"`
	BacklogVelocityOpsPerSec  float64   `json:"backlog_velocity"` // negative = catching up
	EstimatedCatchupMs        float64   `json:"estimated_catchup_ms"`
	Converging                bool      `json:"converging"`
	NonConvergent             bool      `json:"non_convergent"`
	MitigationsApplied        []string  `json:"mitigations_applied,omitempty"`
	ObservedAt                time.Time `json:"observed_at"`
}

// ObjectTemperature classifies mutation heat.
type ObjectTemperature string

const (
	TempHot    ObjectTemperature = "HOT"
	TempWarm   ObjectTemperature = "WARM"
	TempCold   ObjectTemperature = "COLD"
	TempStatic ObjectTemperature = "STATIC"
)

// HotObject describes a path's observed heat for scheduling.
type HotObject struct {
	Path         string            `json:"path"`
	Temperature  ObjectTemperature `json:"temperature"`
	MutationCount uint64           `json:"mutation_count"`
	LastMutation time.Time         `json:"last_mutation"`
	BytesTouched int64             `json:"bytes_touched"`
}

// TuningDecision records an explainable adaptive sync change.
type TuningDecision struct {
	At          time.Time `json:"at"`
	Parameter   string    `json:"parameter"`
	OldValue    string    `json:"old_value"`
	NewValue    string    `json:"new_value"`
	Reason      string    `json:"reason"`
	LagOps      uint64    `json:"lag_ops"`
	WriteRate   float64   `json:"write_rate"`
	Throughput  float64   `json:"throughput"`
}

// StateManifest is a cryptographically verifiable description of state.
type StateManifest struct {
	Epoch               uint64            `json:"epoch"`
	ObjectCount         int               `json:"object_count"`
	TotalBytes          int64             `json:"total_bytes"`
	RootDigest          string            `json:"root_digest"`
	JournalPosition     uint64            `json:"journal_position"`
	VerificationAlgo    string            `json:"verification_algorithm"`
	MerkleRoot          string            `json:"merkle_root,omitempty"`
	ObjectDigests       map[string]string `json:"object_digests,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
}

// CutoverDecision is the pre-cutover safety gate result.
type CutoverDecision string

const (
	CutoverAllowed          CutoverDecision = "CUTOVER_ALLOWED"
	CutoverBlocked          CutoverDecision = "CUTOVER_BLOCKED"
	CutoverRequiresOverride CutoverDecision = "CUTOVER_REQUIRES_OVERRIDE"
)

// SafetyGateResult lists exact reasons for allow/block/override.
type SafetyGateResult struct {
	Decision              CutoverDecision `json:"decision"`
	Reasons               []string        `json:"reasons"`
	JournalLag            uint64          `json:"journal_lag"`
	UnverifiedObjects     int             `json:"unverified_objects"`
	CandidateHealthy      bool            `json:"candidate_healthy"`
	ManifestIntegrityOK   bool            `json:"manifest_integrity_ok"`
	JournalIntegrityOK    bool            `json:"journal_integrity_ok"`
	RollbackPointPresent  bool            `json:"rollback_point_present"`
	PendingOps            uint64          `json:"pending_ops"`
	CoordinatorHealthy    bool            `json:"coordinator_healthy"`
	StorageHealthy        bool            `json:"storage_healthy"`
	PredictedPauseMs      float64         `json:"predicted_pause_ms"`
	MaxWritePauseMs       float64         `json:"max_write_pause_ms"`
	OverrideRequested     bool            `json:"override_requested"`
}

// RecoveryClassification expands rollback safety for forward recovery.
type RecoveryClassification string

const (
	ClassRollbackSafe              RecoveryClassification = "ROLLBACK_SAFE"
	ClassForwardRecoveryRequired   RecoveryClassification = "FORWARD_RECOVERY_REQUIRED"
	ClassManualReconciliationReq   RecoveryClassification = "MANUAL_RECONCILIATION_REQUIRED"
)

// StorageCapability enumerates adapter-reported storage features.
type StorageCapability string

const (
	CapSnapshot            StorageCapability = "SNAPSHOT"
	CapCopyOnWrite         StorageCapability = "COPY_ON_WRITE"
	CapChangeTracking      StorageCapability = "CHANGE_TRACKING"
	CapFreeze              StorageCapability = "FREEZE"
	CapAtomicRename        StorageCapability = "ATOMIC_RENAME"
	CapBlockRead           StorageCapability = "BLOCK_READ"
	CapBlockWrite          StorageCapability = "BLOCK_WRITE"
	CapChecksum            StorageCapability = "CHECKSUM"
	CapIncrementalSnapshot StorageCapability = "INCREMENTAL_SNAPSHOT"
)

// DeploymentReceipt is an immutable structured evidence record.
type DeploymentReceipt struct {
	ReceiptID              string                 `json:"receipt_id"`
	DeployID               string                 `json:"deploy_id"`
	CreatedAt              time.Time              `json:"created_at"`
	FromImage              string                 `json:"from_image"`
	ToImage                string                 `json:"to_image"`
	FromEpoch              uint64                 `json:"from_epoch"`
	ToEpoch                uint64                 `json:"to_epoch"`
	FenceToken             uint64                 `json:"fence_token"`
	Phases                 []string               `json:"phases"`
	FinalPhase             DeployPhase            `json:"final_phase"`
	MutationsSynced        uint64                 `json:"mutations_synced"`
	FinalDeltaWrites       uint64                 `json:"final_delta_writes"`
	PredictedPauseMs       float64                `json:"predicted_pause_ms"`
	ActualPauseMs          float64                `json:"actual_pause_ms"`
	CutoverWritePauseMs    float64                `json:"cutover_write_pause_ms"`
	RPOMutations           uint64                 `json:"rpo_mutations"`
	RTOSeconds             float64                `json:"rto_seconds"`
	Verification           *VerificationReceipt   `json:"verification,omitempty"`
	Manifest               *StateManifest         `json:"manifest,omitempty"`
	SafetyGate             *SafetyGateResult      `json:"safety_gate,omitempty"`
	Convergence            *ConvergenceSnapshot   `json:"convergence,omitempty"`
	SyncLag                *SyncLagMetrics        `json:"sync_lag,omitempty"`
	TuningDecisions        []TuningDecision       `json:"tuning_decisions,omitempty"`
	RollbackClassification RecoveryClassification `json:"rollback_classification,omitempty"`
	CapabilitiesUsed       []StorageCapability    `json:"capabilities_used,omitempty"`
	DryRun                 bool                   `json:"dry_run"`
	HumanSummary           string                 `json:"human_summary"`
	EvidenceHash           string                 `json:"evidence_hash"`
}
