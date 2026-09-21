// Package invariants locks the documented safety properties with executable tests.
package invariants_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/fence"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/safegate"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// ACKNOWLEDGED_WRITES_MUST_NOT_BE_SILENTLY_LOST:
// a synced journal append is readable after reopen.
func TestAcknowledgedWritesNotSilentlyLost(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := j.Append(types.Mutation{Kind: types.OpWrite, Path: "ack.txt", Payload: []byte("durable")})
	if err != nil {
		t.Fatal(err)
	}
	j.Close()

	j2, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j2.Close()
	all, err := j2.ReadFrom(seq)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || string(all[0].Payload) != "durable" {
		t.Fatalf("ACKNOWLEDGED_WRITES_MUST_NOT_BE_SILENTLY_LOST violated: %+v", all)
	}
}

// STALE_EPOCH_MUST_NOT_MODIFY_NEWER_STATE
func TestStaleEpochMustNotModifyNewerState(t *testing.T) {
	root := t.TempDir()
	fm, err := fence.Open(root, "holder")
	if err != nil {
		t.Fatal(err)
	}
	cur, err := fm.AdvanceEpoch("holder")
	if err != nil {
		t.Fatal(err)
	}
	if err := fm.Check(cur.Epoch-1, cur.Token); err == nil {
		t.Fatal("STALE_EPOCH_MUST_NOT_MODIFY_NEWER_STATE: expected rejection")
	}
}

// UNVERIFIED_STATE_MUST_NOT_BE_REPORTED_VERIFIED — UNKNOWN ≠ VERIFIED
func TestUnknownMustNotBeTreatedAsVerified(t *testing.T) {
	unknown := &types.VerificationReceipt{Outcome: types.VerifyUnknown}
	if verifier.IsVerified(unknown) {
		t.Fatal("UNKNOWN must never be treated as VERIFIED")
	}
	failed := &types.VerificationReceipt{Outcome: types.VerifyFailed}
	if verifier.IsVerified(failed) {
		t.Fatal("FAILED must never be treated as VERIFIED")
	}
	partial := &types.VerificationReceipt{Outcome: types.VerifyPartial}
	if verifier.IsVerified(partial) {
		t.Fatal("PARTIAL must never be treated as VERIFIED")
	}

	gate := safegate.Evaluate(safegate.Input{
		CandidateHealthy: true, ManifestIntegrityOK: true, JournalIntegrityOK: true,
		RollbackPointPresent: true, CoordinatorHealthy: true, StorageHealthy: true,
		MaxWritePauseMs: 250, Verification: unknown,
	})
	if gate.Decision == types.CutoverAllowed {
		t.Fatal("safety gate must not allow cutover on UNKNOWN verification")
	}
}

// FAILED_CUTOVER_MUST_HAVE_DETERMINISTIC_RECOVERY_PATH — fence persists across reopen
func TestFailedCutoverDeterministicFenceRecovery(t *testing.T) {
	root := t.TempDir()
	fm, err := fence.Open(root, "a")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := fm.Issue("a")
	if err != nil {
		t.Fatal(err)
	}
	fm2, err := fence.Open(root, "b")
	if err != nil {
		t.Fatal(err)
	}
	got := fm2.Current()
	if got.Token != tok.Token || got.Epoch != tok.Epoch {
		t.Fatalf("fence not durable: %+v vs %+v", got, tok)
	}
	// Ensure fence.json exists as the recovery artifact.
	if _, err := os.Stat(filepath.Join(root, "fence.json")); err != nil {
		t.Fatal(err)
	}
}
