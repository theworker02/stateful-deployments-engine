package safety

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/safegate"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestStrictBlocksLag(t *testing.T) {
	r := Evaluate(Strict(), safegate.Input{
		JournalLag:           5,
		CandidateHealthy:     true,
		ManifestIntegrityOK:  true,
		JournalIntegrityOK:   true,
		RollbackPointPresent: true,
		CoordinatorHealthy:   true,
		StorageHealthy:       true,
		PredictedPauseMs:     40,
	})
	if r.Decision != types.CutoverBlocked {
		t.Fatalf("%+v", r)
	}
}

func TestStrictAllowsClean(t *testing.T) {
	r := Evaluate(Strict(), safegate.Input{
		CandidateHealthy:     true,
		ManifestIntegrityOK:  true,
		JournalIntegrityOK:   true,
		RollbackPointPresent: true,
		CoordinatorHealthy:   true,
		StorageHealthy:       true,
		PredictedPauseMs:     40,
		Verification:         &types.VerificationReceipt{Outcome: types.VerifyVerified},
	})
	if !IsAllowed(r) {
		t.Fatalf("%+v", r)
	}
}
