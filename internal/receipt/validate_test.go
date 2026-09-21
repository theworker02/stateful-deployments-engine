package receipt

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestBuildValidateRoundTrip(t *testing.T) {
	r := Build(&types.DeploymentReceipt{
		DeployID:        "deploy-1",
		FromImage:       "app:1",
		ToImage:         "app:2",
		FromEpoch:       1,
		ToEpoch:         2,
		FinalPhase:      types.PhaseCommitted,
		ActualPauseMs:   42.5,
		MutationsSynced: 100,
		CreatedAt:       time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	})
	if err := Validate(r); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path, err := WriteJSON(dir, r)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(path) != ".json" {
		t.Fatalf("path=%s", path)
	}
	// Tamper
	r.ActualPauseMs = 99
	if err := Validate(r); err == nil {
		t.Fatal("expected tamper detect")
	}
}
