// Package recover classifies forward-recovery vs rollback after interrupted cutovers.
package recover

import (
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Evidence is the durable state inspected after crash/restart.
type Evidence struct {
	Phase                types.DeployPhase
	WritesAfterCutover   uint64
	CandidatePromoted    bool
	StandbyIntact        bool
	ManifestMatch        bool
	JournalIntegrityOK   bool
	ObservationFailed    bool
}

// Classify maps evidence to RecoveryClassification + RecoveryDecision.
func Classify(deployID string, e Evidence) (types.RecoveryClassification, types.RecoveryDecision) {
	dec := types.RecoveryDecision{
		DeployID:  deployID,
		FromPhase: e.Phase,
	}
	switch e.Phase {
	case types.PhaseCutover, types.PhaseObserving, types.PhaseRecovering:
		if e.WritesAfterCutover > 0 && e.CandidatePromoted {
			dec.Action = types.RecoverRequestReconciliation
			dec.Reason = "writes accepted on new epoch after cutover; rollback unsafe"
			return types.ClassManualReconciliationReq, dec
		}
		if e.CandidatePromoted && e.ObservationFailed && e.StandbyIntact && e.WritesAfterCutover == 0 {
			dec.Action = types.RecoverRollback
			dec.Reason = "post-cutover unhealthy with intact standby and zero new writes"
			dec.TargetPhase = types.PhaseRolledBack
			return types.ClassRollbackSafe, dec
		}
		if !e.CandidatePromoted && e.JournalIntegrityOK {
			dec.Action = types.RecoverResume
			dec.Reason = "cutover incomplete; journal intact — resume forward"
			dec.TargetPhase = types.PhaseSynchronizing
			return types.ClassForwardRecoveryRequired, dec
		}
		dec.Action = types.RecoverRetry
		dec.Reason = "interrupted mid-cutover; retry forward recovery"
		return types.ClassForwardRecoveryRequired, dec
	case types.PhaseSynchronizing, types.PhaseVerifying, types.PhaseQuiescing, types.PhaseFinalDelta:
		if !e.JournalIntegrityOK || !e.ManifestMatch {
			dec.Action = types.RecoverAbort
			dec.Reason = "integrity failure before cutover"
			dec.TargetPhase = types.PhaseAborted
			return types.ClassRollbackSafe, dec
		}
		dec.Action = types.RecoverResume
		dec.Reason = "pre-cutover interrupt; resume"
		dec.TargetPhase = e.Phase
		return types.ClassForwardRecoveryRequired, dec
	default:
		dec.Action = types.RecoverAbort
		dec.Reason = fmt.Sprintf("no recovery playbook for phase %s", e.Phase)
		return types.ClassManualReconciliationReq, dec
	}
}

// IsRollbackSafe is a convenience predicate.
func IsRollbackSafe(c types.RecoveryClassification) bool {
	return c == types.ClassRollbackSafe
}
