// Package integration hosts cross-package smoke tests (local stack).
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/capability"
	"github.com/theworker02/stateful-deployments-engine/internal/epoch"
	"github.com/theworker02/stateful-deployments-engine/internal/fsm"
	"github.com/theworker02/stateful-deployments-engine/internal/manifest"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestLocalStackSmoke(t *testing.T) {
	root := t.TempDir()
	a, err := local.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	m := capability.Negotiate(a.Capabilities(), nil)
	if !m.Satisfied() {
		t.Fatalf("missing=%v", m.Missing)
	}
	es, err := epoch.Open(filepath.Join(root, "epochs"))
	if err != nil {
		t.Fatal(err)
	}
	rec, err := es.Advance("vol", "test", "smoke")
	if err != nil || rec.Epoch != 1 {
		t.Fatalf("%+v %v", rec, err)
	}
	active, err := a.ActiveSlot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(active.StatePath, "hello.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	man, err := (manifest.Builder{}).Build(active.StatePath, rec.Epoch, 0)
	if err != nil || man.ObjectCount < 1 {
		t.Fatalf("%+v %v", man, err)
	}
	snap := fsm.Snapshot{DeployID: "smoke", Phase: types.PhaseActive, Epoch: rec.Epoch}
	if err := fsm.ApplyTransition(&snap, types.PhaseCheckpoint); err != nil {
		t.Fatal(err)
	}
	if err := fsm.Save(root, snap); err != nil {
		t.Fatal(err)
	}
}
