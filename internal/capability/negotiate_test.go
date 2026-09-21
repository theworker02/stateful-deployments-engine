package capability

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestNegotiateLocalLike(t *testing.T) {
	m := Negotiate([]types.StorageCapability{
		types.CapChecksum, types.CapFreeze, types.CapAtomicRename, types.CapBlockRead,
	}, nil)
	if !m.Satisfied() {
		t.Fatalf("missing=%v", m.Missing)
	}
	if len(m.Strategies) < 2 {
		t.Fatalf("strategies=%v", m.Strategies)
	}
}

func TestNegotiateMissingFreeze(t *testing.T) {
	m := Negotiate([]types.StorageCapability{types.CapChecksum}, MinimalRequired())
	if m.Satisfied() {
		t.Fatal("expected unsatisfied")
	}
}
