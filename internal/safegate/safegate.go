// Package safegate evaluates pre-cutover readiness with explicit allow/block/override reasons.
package safegate

import (
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// Input is the evidence set for the safety gate.
type Input struct {
	JournalLag           uint64
	PendingOps           uint64
	UnverifiedObjects    int
	CandidateHealthy     bool
	ManifestIntegrityOK  bool
	JournalIntegrityOK   bool
	RollbackPointPresent bool
	CoordinatorHealthy   bool
	StorageHealthy       bool
	PredictedPauseMs     float64
	MaxWritePauseMs      float64
	Verification         *types.VerificationReceipt
	OverrideRequested    bool
	NonConvergent        bool
}

// Evaluate decides cutover readiness. No silent overrides.
func Evaluate(in Input) types.SafetyGateResult {
	if in.MaxWritePauseMs <= 0 {
		in.MaxWritePauseMs = 250
	}
	r := types.SafetyGateResult{
		JournalLag:           in.JournalLag,
		UnverifiedObjects:    in.UnverifiedObjects,
		CandidateHealthy:     in.CandidateHealthy,
		ManifestIntegrityOK:  in.ManifestIntegrityOK,
		JournalIntegrityOK:   in.JournalIntegrityOK,
		RollbackPointPresent: in.RollbackPointPresent,
		PendingOps:           in.PendingOps,
		CoordinatorHealthy:   in.CoordinatorHealthy,
		StorageHealthy:       in.StorageHealthy,
		PredictedPauseMs:     in.PredictedPauseMs,
		MaxWritePauseMs:      in.MaxWritePauseMs,
		OverrideRequested:    in.OverrideRequested,
		Decision:             types.CutoverAllowed,
	}
	var reasons []string
	block := false
	overrideOnly := false

	if in.JournalLag > 0 || in.PendingOps > 0 {
		block = true
		reasons = append(reasons, fmt.Sprintf("journal lag %d / pending ops %d", in.JournalLag, in.PendingOps))
	}
	if !in.CandidateHealthy {
		block = true
		reasons = append(reasons, "candidate health check failed")
	}
	if !in.ManifestIntegrityOK {
		block = true
		reasons = append(reasons, "state manifest integrity failed")
	}
	if !in.JournalIntegrityOK {
		block = true
		reasons = append(reasons, "journal integrity failed")
	}
	if !in.RollbackPointPresent {
		block = true
		reasons = append(reasons, "no rollback point (standby/checkpoint) present")
	}
	if !in.CoordinatorHealthy || !in.StorageHealthy {
		block = true
		reasons = append(reasons, "coordinator or storage unhealthy")
	}
	if in.NonConvergent {
		block = true
		reasons = append(reasons, "convergence engine reports non-convergent backlog")
	}
	if in.Verification == nil || in.Verification.Outcome == types.VerifyUnknown {
		block = true
		reasons = append(reasons, "UNVERIFIED_STATE_MUST_NOT_BE_REPORTED_VERIFIED: verification UNKNOWN/missing")
	} else if !verifier.IsVerified(in.Verification) && in.Verification.Outcome != types.VerifyPartial {
		block = true
		reasons = append(reasons, "verification outcome is "+string(in.Verification.Outcome))
	} else if in.Verification.Outcome == types.VerifyPartial {
		overrideOnly = true
		reasons = append(reasons, "verification PARTIAL — requires explicit override")
	}
	if in.PredictedPauseMs > in.MaxWritePauseMs {
		block = true
		reasons = append(reasons, fmt.Sprintf("predicted pause %.2fms exceeds budget %.2fms", in.PredictedPauseMs, in.MaxWritePauseMs))
	}
	if in.UnverifiedObjects > 0 {
		overrideOnly = true
		reasons = append(reasons, fmt.Sprintf("%d unverified objects", in.UnverifiedObjects))
	}

	switch {
	case block && !(in.OverrideRequested && overrideOnly && !hardBlock(reasons)):
		r.Decision = types.CutoverBlocked
	case overrideOnly && !in.OverrideRequested:
		r.Decision = types.CutoverRequiresOverride
	case overrideOnly && in.OverrideRequested && !hardBlock(reasons):
		r.Decision = types.CutoverAllowed
		reasons = append(reasons, "override explicitly requested by operator")
	default:
		if len(reasons) == 0 {
			reasons = append(reasons, "all safety checks passed")
		}
		r.Decision = types.CutoverAllowed
	}
	r.Reasons = reasons
	return r
}

func hardBlock(reasons []string) bool {
	for _, r := range reasons {
		if contains(r, "journal integrity") || contains(r, "ACKNOWLEDGED") || contains(r, "UNKNOWN") ||
			contains(r, "exceeds budget") || contains(r, "non-convergent") || contains(r, "health check failed") {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
