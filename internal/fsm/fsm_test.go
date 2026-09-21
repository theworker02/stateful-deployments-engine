package fsm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestHappyPathTransitions(t *testing.T) {
	for i := 0; i < len(HappyPath)-1; i++ {
		from, to := HappyPath[i], HappyPath[i+1]
		if err := ValidateTransition(from, to); err != nil {
			t.Fatalf("%s→%s: %v", from, to, err)
		}
	}
}

func TestIllegalTransition(t *testing.T) {
	if err := ValidateTransition(types.PhaseActive, types.PhaseCutover); err == nil {
		t.Fatal("expected illegal")
	}
}

func TestPersistRoundTrip(t *testing.T) {
	root := t.TempDir()
	snap := Snapshot{
		DeployID: "d1",
		Phase:    types.PhaseSynchronizing,
		Epoch:    3,
	}
	if err := Save(root, snap); err != nil {
		t.Fatal(err)
	}
	got, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.DeployID != "d1" || got.Phase != types.PhaseSynchronizing {
		t.Fatalf("%+v", got)
	}
	if err := ApplyTransition(&got, types.PhaseVerifying); err != nil {
		t.Fatal(err)
	}
	if err := Save(root, got); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "fsm", "snapshot.json")); err != nil {
		t.Fatal(err)
	}
}
