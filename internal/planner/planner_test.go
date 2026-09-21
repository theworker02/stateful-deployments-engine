package planner_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/planner"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestPlanReadyEmptySource(t *testing.T) {
	root := t.TempDir()
	plat, err := local.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(filepath.Join(root, "journal"), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	p := &planner.Planner{Platform: plat, Storage: plat, Journal: j}
	plan, err := p.Plan(context.Background(), planner.Input{TargetImage: "app:v2"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Conclusion != types.MigrateReady && plan.Conclusion != types.MigrateReadyWithRisk {
		t.Fatalf("conclusion=%s rationale=%s", plan.Conclusion, plan.Rationale)
	}
	if !planner.IsReady(plan.Conclusion) {
		t.Fatal("IsReady false")
	}
}

func TestPlanNonConvergent(t *testing.T) {
	root := t.TempDir()
	plat, err := local.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(filepath.Join(root, "journal"), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	p := &planner.Planner{Platform: plat, Storage: plat, Journal: j}
	plan, err := p.Plan(context.Background(), planner.Input{
		TargetImage:          "app:v2",
		WriteRateOpsPerSec:   1000,
		HistoricalThroughput: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Conclusion != types.MigrateLikelyNonConvergent {
		t.Fatalf("got %s want LIKELY_NON_CONVERGENT (%s)", plan.Conclusion, plan.Rationale)
	}
	if planner.IsReady(types.MigrateUnknown) {
		t.Fatal("UNKNOWN must not be ready")
	}
}
