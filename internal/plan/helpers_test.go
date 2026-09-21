package plan

import (
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestSampleDryRunRoundTrip(t *testing.T) {
	p := SampleDryRun("app:2", 10<<20)
	if !IsReady(p) {
		t.Fatal("expected ready-with-risk")
	}
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := WriteJSON(path, p); err != nil {
		t.Fatal(err)
	}
	got, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Conclusion != types.MigrateReadyWithRisk {
		t.Fatalf("%s", got.Conclusion)
	}
	unk := &types.MigrationPlan{Conclusion: types.MigrateUnknown}
	if IsReady(unk) {
		t.Fatal("UNKNOWN must not be ready")
	}
}
