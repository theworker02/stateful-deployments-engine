package events

// Versioned deployment lifecycle events (schema_version = 1).
const SchemaVersion = 1

const (
	KindDeploymentStarted   Kind = "DeploymentStarted"
	KindDeploymentPlanReady Kind = "DeploymentPlanReady"
	KindCheckpointCaptured  Kind = "CheckpointCaptured"
	KindShadowReady         Kind = "ShadowReady"
	KindSyncProgress        Kind = "SyncProgress"
	KindVerificationDone    Kind = "VerificationDone"
	KindCutoverStarted      Kind = "CutoverStarted"
	KindCutoverCompleted    Kind = "CutoverCompleted"
	KindObservationPassed   Kind = "ObservationPassed"
	KindDeploymentCommitted Kind = "DeploymentCommitted"
	KindDeploymentAborted   Kind = "DeploymentAborted"
	KindRecoveryStarted     Kind = "RecoveryStarted"
	KindRecoveryCompleted   Kind = "RecoveryCompleted"
)

// Envelope is the versioned wire format for control-plane consumers.
type Envelope struct {
	SchemaVersion int                    `json:"schema_version"`
	Event         Event                  `json:"event"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// Wrap attaches schema_version to an event.
func Wrap(e Event) Envelope {
	return Envelope{SchemaVersion: SchemaVersion, Event: e}
}
