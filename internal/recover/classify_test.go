package recover

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestClassifyRollbackSafe(t *testing.T) {
	c, d := Classify("d1", Evidence{
		Phase:              types.PhaseObserving,
		CandidatePromoted:  true,
		StandbyIntact:      true,
		ObservationFailed:  true,
		WritesAfterCutover: 0,
		JournalIntegrityOK: true,
	})
	if c != types.ClassRollbackSafe || d.Action != types.RecoverRollback {
		t.Fatalf("%s %+v", c, d)
	}
}

func TestClassifyManualAfterWrites(t *testing.T) {
	c, d := Classify("d1", Evidence{
		Phase:              types.PhaseObserving,
		CandidatePromoted:  true,
		WritesAfterCutover: 3,
	})
	if c != types.ClassManualReconciliationReq || d.Action != types.RecoverRequestReconciliation {
		t.Fatalf("%s %+v", c, d)
	}
}
