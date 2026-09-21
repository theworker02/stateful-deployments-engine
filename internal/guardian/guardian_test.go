package guardian_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/guardian"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestObserveCommitsHealthySlot(t *testing.T) {
	root := t.TempDir()
	plat, err := local.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	active, err := plat.ActiveSlot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_ = filepath.Join(active.StatePath)
	res := guardian.Observe(context.Background(), plat, active.ID, active.StatePath, 1, 30*time.Millisecond)
	if !res.Committed || res.Phase != types.PhaseCommitted {
		t.Fatalf("result=%+v", res)
	}
}
