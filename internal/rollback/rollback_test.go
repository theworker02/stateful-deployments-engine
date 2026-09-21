package rollback_test

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/rollback"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestClassifySafe(t *testing.T) {
	s, c, _ := rollback.Classify(0, types.PhaseObserving)
	if s != types.RollbackSafe {
		t.Fatalf("safety=%s class=%s", s, c)
	}
}

func TestClassifyForwardWhenWrites(t *testing.T) {
	s, c, reason := rollback.Classify(3, types.PhaseObserving)
	if s == types.RollbackSafe {
		t.Fatal("must not be SAFE with ack writes")
	}
	if c != types.ClassForwardRecoveryRequired && c != types.ClassManualReconciliationReq {
		t.Fatalf("class=%s reason=%s", c, reason)
	}
}
