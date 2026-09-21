package embed_test

import (
	"context"
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/embed"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestOpenLocalDryRun(t *testing.T) {
	root := t.TempDir()
	eng, err := embed.OpenLocal(root, embed.ModeLibrary)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()
	if eng.Mode != embed.ModeLibrary {
		t.Fatalf("mode=%s", eng.Mode)
	}
	rep, err := eng.Deploy(context.Background(), "app:v2", true, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if rep == nil {
		t.Fatal("nil report")
	}
	// Dry-run should not leave production in failed mystery state.
	if rep.Phase == types.PhaseFailed && rep.Error == "" {
		t.Fatalf("failed without error: %+v", rep)
	}
}
