package safegate

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestEvaluateTable(t *testing.T) {
	cases := []struct {
		name string
		in   Input
		want types.CutoverDecision
	}{
		{
			name: "clean",
			in: Input{
				CandidateHealthy: true, ManifestIntegrityOK: true, JournalIntegrityOK: true,
				RollbackPointPresent: true, CoordinatorHealthy: true, StorageHealthy: true,
				PredictedPauseMs: 40,
				Verification:     &types.VerificationReceipt{Outcome: types.VerifyVerified},
			},
			want: types.CutoverAllowed,
		},
		{
			name: "lag",
			in: Input{
				JournalLag: 1, CandidateHealthy: true, ManifestIntegrityOK: true, JournalIntegrityOK: true,
				RollbackPointPresent: true, CoordinatorHealthy: true, StorageHealthy: true,
				Verification: &types.VerificationReceipt{Outcome: types.VerifyVerified},
			},
			want: types.CutoverBlocked,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := Evaluate(tc.in)
			if r.Decision != tc.want {
				t.Fatalf("got %s want %s reasons=%v", r.Decision, tc.want, r.Reasons)
			}
		})
	}
}
