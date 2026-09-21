package fsm

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func FuzzValidateTransition(f *testing.F) {
	f.Add(string(types.PhaseActive), string(types.PhaseCheckpoint))
	f.Add(string(types.PhaseActive), string(types.PhaseCutover))
	f.Add(string(types.PhaseObserving), string(types.PhaseCommitted))
	f.Fuzz(func(t *testing.T, fromS, toS string) {
		from := types.DeployPhase(fromS)
		to := types.DeployPhase(toS)
		ok := CanTransition(from, to)
		err := ValidateTransition(from, to)
		if ok != (err == nil) {
			t.Fatalf("CanTransition=%v ValidateTransition err=%v for %q→%q", ok, err, fromS, toS)
		}
	})
}
