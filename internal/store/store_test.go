package store_test

import (
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/store"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestSaveLoadListLatest(t *testing.T) {
	root := t.TempDir()
	s := store.New(root)
	st := &store.DeploymentState{
		DeployID:  "dep-1",
		Phase:     types.PhaseSynchronizing,
		Epoch:     2,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.Save(st); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load("dep-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase != types.PhaseSynchronizing {
		t.Fatalf("phase=%s", got.Phase)
	}
	ids, err := s.List()
	if err != nil || len(ids) != 1 {
		t.Fatalf("list=%v err=%v", ids, err)
	}
	latest, err := s.LatestNonTerminal()
	if err != nil || latest == nil || latest.DeployID != "dep-1" {
		t.Fatalf("latest=%v err=%v", latest, err)
	}
	st.Phase = types.PhaseCommitted
	_ = s.Save(st)
	latest, err = s.LatestNonTerminal()
	if err != nil {
		t.Fatal(err)
	}
	if latest != nil {
		t.Fatalf("expected nil after terminal, got %+v", latest)
	}
}
