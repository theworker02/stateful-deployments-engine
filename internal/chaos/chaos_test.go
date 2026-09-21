package chaos_test

import (
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/chaos"
)

func TestChaosMatrix(t *testing.T) {
	out := filepath.Join(t.TempDir(), "chaos")
	receipts, err := chaos.RunAll(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipts) < 13 {
		t.Fatalf("expected >=13 scenarios, got %d", len(receipts))
	}
	failed := 0
	for _, r := range receipts {
		if !r.Detected {
			t.Logf("scenario %s did not detect: %s", r.Scenario, r.Error)
			failed++
		}
	}
	if failed > 2 {
		t.Fatalf("%d scenarios failed detection", failed)
	}
}
