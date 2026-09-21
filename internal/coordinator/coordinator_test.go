package coordinator_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/agent"
	"github.com/theworker02/stateful-deployments-engine/internal/coordinator"
	"github.com/theworker02/stateful-deployments-engine/internal/fence"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/store"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

func TestStatefulDeployWithLiveWrites(t *testing.T) {
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
	plat.BindJournal(j)
	fm, err := fence.Open(root, "test")
	if err != nil {
		t.Fatal(err)
	}
	st := store.New(root)

	active, err := plat.ActiveSlot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	w := &agent.Writer{Root: active.StatePath, Journal: j, Local: plat}
	for i := 0; i < 10; i++ {
		if _, err := w.WriteFile(context.Background(), fmt.Sprintf("f-%d.txt", i), []byte("seed"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		n := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(8 * time.Millisecond):
				n++
				// Use Background so barrier wait is not cancelled mid-deploy.
				_, _ = w.WriteFile(context.Background(), fmt.Sprintf("live-%d.txt", n), []byte(fmt.Sprintf("v%d", n)), 0o644)
			}
		}
	}()
	time.Sleep(40 * time.Millisecond)

	c := coordinator.NewFull(plat, plat, j, fm, st, func(types.DeployPhase, string) {})
	report, err := c.Deploy(context.Background(), coordinator.Config{
		ImageRef:         "app:v2",
		SyncPollInterval: 8 * time.Millisecond,
		MaxWritePause:    2 * time.Second,
	})
	cancel()
	if err != nil {
		t.Fatal(err)
	}
	if report.Phase != types.PhaseCommitted && report.Phase != types.PhaseComplete {
		t.Fatalf("phase=%s err=%s", report.Phase, report.Error)
	}
	if !report.ConsistencyOK || !report.HealthOK {
		t.Fatalf("consistency/health failed: %+v", report)
	}
	if report.BarrierDuration <= 0 || report.CutoverWritePauseMs <= 0 {
		t.Fatalf("expected positive barrier/pause, got barrier=%s pause=%v", report.BarrierDuration, report.CutoverWritePauseMs)
	}

	// State survived on disk
	loaded, err := st.Load(report.DeployID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Phase != types.PhaseCommitted {
		t.Fatalf("persisted phase=%s", loaded.Phase)
	}

	newActive, err := plat.ActiveSlot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if newActive.ImageRef != "app:v2" {
		t.Fatalf("image=%s", newActive.ImageRef)
	}
	if report.StandbySlotID == "" {
		t.Fatal("expected standby slot")
	}
	if report.ToEpoch < 2 {
		t.Fatalf("epoch=%d", report.ToEpoch)
	}

	cp, err := verifier.BuildCheckpoint(newActive.StatePath, report.ToEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if cp.FileCount < 10 {
		t.Fatalf("fileCount=%d", cp.FileCount)
	}
}

func TestCrashRecoveryDecision(t *testing.T) {
	st := &store.DeploymentState{DeployID: "x", Phase: types.PhaseSynchronizing}
	dec := coordinator.DecideRecovery(st)
	if dec.Action != types.RecoverRetry {
		t.Fatalf("action=%s", dec.Action)
	}
	st.Phase = types.PhaseCutover
	st.WritesAfterCutover = 1
	dec = coordinator.DecideRecovery(st)
	if dec.Action != types.RecoverRequestReconciliation {
		t.Fatalf("action=%s", dec.Action)
	}
}

func TestStaleEpochFenced(t *testing.T) {
	root := t.TempDir()
	fm, err := fence.Open(root, "a")
	if err != nil {
		t.Fatal(err)
	}
	ft, err := fm.Issue("a")
	if err != nil {
		t.Fatal(err)
	}
	if err := fm.Check(ft.Epoch, ft.Token); err != nil {
		t.Fatal(err)
	}
	if err := fm.Check(ft.Epoch, ft.Token-1); err == nil {
		t.Fatal("expected stale fence rejection")
	}
}

func TestVerifierDetectsDrift(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	if err := os.WriteFile(filepath.Join(a, "x.txt"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(b, "x.txt"), []byte("2"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, _, _, err := verifier.CompareRoots(a, b, 1)
	if err != nil {
		t.Fatal(err)
	}
	if r.OK {
		t.Fatal("expected mismatch")
	}
	vr, err := verifier.Verify(a, b, 1, verifier.Options{Level: types.VerifyChecksum})
	if err != nil {
		t.Fatal(err)
	}
	if vr.Outcome == types.VerifyVerified || vr.Outcome == types.VerifyUnknown && false {
		t.Fatalf("outcome=%s", vr.Outcome)
	}
	if verifier.IsVerified(vr) {
		t.Fatal("must not treat failed as verified")
	}
}

func TestDryRunDeploy(t *testing.T) {
	root := t.TempDir()
	plat, _ := local.Open(root)
	j, _ := journal.Open(filepath.Join(root, "journal"), 1)
	defer j.Close()
	plat.BindJournal(j)
	fm, _ := fence.Open(root, "t")
	fenceBefore := fm.Current().Token
	active, err := plat.ActiveSlot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(active.StatePath, "marker.txt"), []byte("prod"), 0o644)
	cpBefore, err := verifier.BuildCheckpoint(active.StatePath, j.Epoch())
	if err != nil {
		t.Fatal(err)
	}
	nextBefore := j.NextSeq()
	st := store.New(root)
	c := coordinator.NewFull(plat, plat, j, fm, st, nil)
	rep, err := c.Deploy(context.Background(), coordinator.Config{ImageRef: "x", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Phase.IsTerminal() {
		t.Fatalf("phase=%s", rep.Phase)
	}
	if !rep.DryRun {
		t.Fatal("expected dry_run=true on report")
	}
	if rep.ShadowSlotID != "" {
		t.Fatalf("dry-run must not create shadow slot, got %q", rep.ShadowSlotID)
	}
	if fm.Current().Token != fenceBefore {
		t.Fatalf("dry-run must not issue fence: before=%d after=%d", fenceBefore, fm.Current().Token)
	}
	if j.NextSeq() != nextBefore {
		t.Fatalf("dry-run must not append journal: before=%d after=%d", nextBefore, j.NextSeq())
	}
	cpAfter, err := verifier.BuildCheckpoint(active.StatePath, j.Epoch())
	if err != nil {
		t.Fatal(err)
	}
	if cpAfter.RootHash != cpBefore.RootHash {
		t.Fatal("dry-run must not mutate production state digest")
	}
	active2, _ := plat.ActiveSlot(context.Background())
	if active2.ID != active.ID {
		t.Fatal("dry-run must not change active slot")
	}
}
