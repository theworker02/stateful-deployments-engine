package fsm

import (
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Transition documents one legal edge in the deploy FSM.
type Transition struct {
	From   types.DeployPhase `json:"from"`
	To     types.DeployPhase `json:"to"`
	Guard  string            `json:"guard"`  // human-readable precondition
	Effect string            `json:"effect"` // side effect summary
}

// HappyPath is the primary success path (excluding unsafe branches).
var HappyPath = []types.DeployPhase{
	types.PhaseActive,
	types.PhaseCheckpoint,
	types.PhaseShadow,
	types.PhaseSynchronizing,
	types.PhaseVerifying,
	types.PhaseQuiescing,
	types.PhaseFinalDelta,
	types.PhaseCutover,
	types.PhaseObserving,
	types.PhaseCommitted,
}

// Table is the exhaustive legal transition set used for validation.
var Table = []Transition{
	{types.PhaseActive, types.PhaseCheckpoint, "deploy requested", "capture consistent checkpoint"},
	{types.PhaseCheckpoint, types.PhaseShadow, "checkpoint sealed", "create candidate slot"},
	{types.PhaseShadow, types.PhaseSynchronizing, "candidate started", "begin journal/block sync"},
	{types.PhaseSynchronizing, types.PhaseVerifying, "lag within bound", "run consistency verifier"},
	{types.PhaseVerifying, types.PhaseQuiescing, "verification VERIFIED", "prepare write barrier"},
	{types.PhaseQuiescing, types.PhaseFinalDelta, "writes frozen", "drain final mutations"},
	{types.PhaseFinalDelta, types.PhaseCutover, "safety gate ALLOWED", "transfer traffic / promote"},
	{types.PhaseCutover, types.PhaseObserving, "candidate active", "post-cutover observation window"},
	{types.PhaseObserving, types.PhaseCommitted, "observation healthy", "seal receipt; retire standby policy"},

	// Abort / degrade edges from most non-terminal phases.
	{types.PhaseCheckpoint, types.PhaseAborted, "checkpoint failed", "abort deploy"},
	{types.PhaseShadow, types.PhaseAborted, "candidate start failed", "destroy candidate"},
	{types.PhaseSynchronizing, types.PhaseAborted, "non-convergent / operator abort", "destroy candidate"},
	{types.PhaseSynchronizing, types.PhaseDegraded, "partial sync integrity unknown", "hold; do not cut over"},
	{types.PhaseVerifying, types.PhaseAborted, "verification FAILED", "abort"},
	{types.PhaseVerifying, types.PhaseDegraded, "verification UNKNOWN/PARTIAL", "block cutover"},
	{types.PhaseQuiescing, types.PhaseAborted, "barrier failed", "unfreeze; abort"},
	{types.PhaseFinalDelta, types.PhaseAborted, "safety gate BLOCKED", "unfreeze; abort"},
	{types.PhaseCutover, types.PhaseRecovering, "cutover interrupted", "classify recovery"},
	{types.PhaseObserving, types.PhaseRecovering, "post-cutover unhealthy", "rollback or reconcile"},
	{types.PhaseObserving, types.PhaseRolledBack, "observation failed; rollback safe", "restore prior epoch"},
	{types.PhaseDegraded, types.PhaseRecovering, "operator/recovery kick", "inspect durable state"},
	{types.PhaseRecovering, types.PhaseSynchronizing, "forward recovery resume", "retry sync"},
	{types.PhaseRecovering, types.PhaseAborted, "unrecoverable", "abort"},
	{types.PhaseRecovering, types.PhaseRolledBack, "rollback chosen", "restore standby"},
	{types.PhaseAborted, types.PhaseActive, "cleanup complete", "return to idle active"},
	{types.PhaseRolledBack, types.PhaseActive, "rollback verified", "return to active"},
}

// allowed builds a quick lookup map.
func allowed() map[types.DeployPhase]map[types.DeployPhase]Transition {
	m := make(map[types.DeployPhase]map[types.DeployPhase]Transition)
	for _, t := range Table {
		if m[t.From] == nil {
			m[t.From] = make(map[types.DeployPhase]Transition)
		}
		m[t.From][t.To] = t
	}
	return m
}

var edges = allowed()

// CanTransition reports whether from→to is legal.
func CanTransition(from, to types.DeployPhase) bool {
	if from == to {
		return true
	}
	_, ok := edges[from][to]
	return ok
}

// ValidateTransition returns an error if the edge is illegal.
func ValidateTransition(from, to types.DeployPhase) error {
	if CanTransition(from, to) {
		return nil
	}
	return fmt.Errorf("fsm: illegal transition %s → %s", from, to)
}

// Lookup returns the transition metadata if present.
func Lookup(from, to types.DeployPhase) (Transition, bool) {
	t, ok := edges[from][to]
	return t, ok
}
