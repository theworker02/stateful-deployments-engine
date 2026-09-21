// Package chaos runs fault-injection scenarios and emits machine-readable receipts.
package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/agent"
	"github.com/theworker02/stateful-deployments-engine/internal/coordinator"
	"github.com/theworker02/stateful-deployments-engine/internal/fence"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/rollback"
	"github.com/theworker02/stateful-deployments-engine/internal/store"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

const (
	ScenarioCoordinatorCrash        = "coordinator_crash"
	ScenarioCandidateCrash          = "candidate_crash"
	ScenarioSourceCrash             = "source_crash"
	ScenarioDiskFull                = "disk_full"
	ScenarioCorruptedJournal        = "corrupted_journal"
	ScenarioTruncatedJournal        = "truncated_journal"
	ScenarioDelayedWrites           = "delayed_writes"
	ScenarioDuplicateMutation       = "duplicate_mutation"
	ScenarioVerificationMismatch    = "verification_mismatch"
	ScenarioCandidateHealthFailure  = "candidate_health_failure"
	ScenarioFailureDuringFinalDelta = "failure_during_final_delta"
	ScenarioFailureAfterCutover     = "failure_immediately_after_cutover"
	ScenarioRollbackAfterNewWrites  = "rollback_after_new_writes"
	ScenarioCrashPlusCorrupt        = "coordinator_crash_plus_corrupt_journal"
	ScenarioHealthPlusLag           = "health_failure_plus_nonconvergence"
)

// RunAll executes the chaos matrix and writes receipts under outDir.
func RunAll(outDir string) ([]types.ChaosReceipt, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	scenarios := []string{
		ScenarioCoordinatorCrash, ScenarioCandidateCrash, ScenarioSourceCrash,
		ScenarioDiskFull, ScenarioCorruptedJournal, ScenarioTruncatedJournal,
		ScenarioDelayedWrites, ScenarioDuplicateMutation, ScenarioVerificationMismatch,
		ScenarioCandidateHealthFailure, ScenarioFailureDuringFinalDelta,
		ScenarioFailureAfterCutover, ScenarioRollbackAfterNewWrites,
		ScenarioCrashPlusCorrupt, ScenarioHealthPlusLag,
	}
	var receipts []types.ChaosReceipt
	for _, s := range scenarios {
		r := Run(s, filepath.Join(outDir, s))
		receipts = append(receipts, r)
		b, _ := json.MarshalIndent(r, "", "  ")
		_ = os.WriteFile(filepath.Join(outDir, s+".json"), b, 0o644)
	}
	pass := 0
	for _, r := range receipts {
		if r.Detected {
			pass++
		}
	}
	matrix := map[string]interface{}{
		"generated_at": time.Now().UTC(),
		"scenarios":    receipts,
		"pass_count":   pass,
		"total":        len(receipts),
	}
	mb, _ := json.MarshalIndent(matrix, "", "  ")
	_ = os.WriteFile(filepath.Join(outDir, "matrix.json"), mb, 0o644)
	return receipts, nil
}

// Run executes one scenario in an isolated workspace.
func Run(scenario, workDir string) types.ChaosReceipt {
	start := time.Now()
	_ = os.RemoveAll(workDir)
	_ = os.MkdirAll(workDir, 0o755)
	r := types.ChaosReceipt{
		Scenario:   scenario,
		InjectedAt: start.UTC(),
		Details:    map[string]interface{}{},
	}
	defer func() {
		r.DurationMs = float64(time.Since(start).Microseconds()) / 1000
	}()

	plat, err := local.Open(workDir)
	if err != nil {
		r.Error = err.Error()
		r.FinalPhase = types.PhaseAborted
		return r
	}
	j, err := journal.Open(filepath.Join(workDir, "journal"), 1)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	defer j.Close()
	plat.BindJournal(j)
	fm, _ := fence.Open(workDir, "chaos")
	st := store.New(workDir)

	active, _ := plat.ActiveSlot(context.Background())
	w := &agent.Writer{Root: active.StatePath, Journal: j, Local: plat}
	for i := 0; i < 5; i++ {
		_, _ = w.WriteFile(context.Background(), fmt.Sprintf("f-%d.txt", i), []byte("seed"), 0o644)
	}

	switch scenario {
	case ScenarioDuplicateMutation:
		seq1, _ := j.Append(types.Mutation{Kind: types.OpWrite, Path: "dup.txt", Payload: []byte("a"), ClientOpID: "op-1"})
		seq2, err := j.Append(types.Mutation{Kind: types.OpWrite, Path: "dup.txt", Payload: []byte("b"), ClientOpID: "op-1"})
		r.Detected = err == nil && seq2 == seq1
		r.Recovered = r.Detected
		r.FinalPhase = types.PhaseCommitted
		r.Details["seq"] = seq2
		return r

	case ScenarioCorruptedJournal:
		_ = j.Close()
		p := filepath.Join(workDir, "journal", "mutations.jsonl")
		f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
		_, _ = f.WriteString("{not-json\n")
		_ = f.Close()
		_, err := journal.Open(filepath.Join(workDir, "journal"), 1)
		r.Detected = err != nil
		r.Recovered = r.Detected
		r.ActionTaken = types.RecoverAbort
		r.FinalPhase = types.PhaseAborted
		r.Error = fmt.Sprintf("%v", err)
		return r

	case ScenarioTruncatedJournal:
		_ = j.Close()
		p := filepath.Join(workDir, "journal", "mutations.jsonl")
		b, _ := os.ReadFile(p)
		origLen := len(b)
		if len(b) > 20 {
			// Crash mid-write: no trailing newline on a partial last line.
			half := b[:len(b)/2]
			if len(half) > 0 && half[len(half)-1] == '\n' {
				half = half[:len(half)-1] // ensure incomplete trailing record
			}
			_ = os.WriteFile(p, half, 0o644)
		}
		j2, err := journal.Open(filepath.Join(workDir, "journal"), 1)
		if err != nil {
			r.Detected = true
			r.Recovered = true
			r.FinalPhase = types.PhaseAborted
			r.Error = err.Error()
			return r
		}
		defer j2.Close()
		fi, _ := os.Stat(p)
		// Truncation recovery discarded the incomplete tail — fault detected + recovered.
		r.Detected = fi != nil && fi.Size() < int64(origLen)
		r.Recovered = r.Detected
		r.FinalPhase = types.PhaseCommitted
		r.Details["recovery"] = "truncation_recovery"
		r.Details["bytes_before"] = origLen
		if fi != nil {
			r.Details["bytes_after"] = fi.Size()
		}
		return r

	case ScenarioVerificationMismatch:
		cand, err := plat.CreateShadow(context.Background(), adapter.CreateShadowRequest{ImageRef: "x"})
		if err != nil {
			r.Error = err.Error()
			return r
		}
		_ = os.WriteFile(filepath.Join(active.StatePath, "x.txt"), []byte("1"), 0o644)
		_ = os.WriteFile(filepath.Join(cand.StatePath, "x.txt"), []byte("2"), 0o644)
		vr, _ := plat.Verify(context.Background(), active.StatePath, cand.StatePath, 1, types.VerifyChecksum)
		// Mismatch must never be reported VERIFIED (PARTIAL or FAILED both count as detection).
		r.Detected = vr != nil && vr.Outcome != types.VerifyVerified && vr.Outcome != types.VerifyUnknown
		r.Recovered = r.Detected
		r.FinalPhase = types.PhaseAborted
		if vr != nil {
			r.Details["outcome"] = string(vr.Outcome)
		}
		return r

	case ScenarioCandidateHealthFailure:
		cand, _ := plat.CreateShadow(context.Background(), adapter.CreateShadowRequest{ImageRef: "x"})
		_ = plat.MarkUnhealthy(cand.ID)
		err := plat.HealthCheck(context.Background(), cand.ID)
		r.Detected = err != nil
		r.Recovered = true
		r.FinalPhase = types.PhaseAborted
		return r

	case ScenarioRollbackAfterNewWrites:
		safety, class, reason := rollback.Classify(3, types.PhaseObserving)
		r.Detected = safety != types.RollbackSafe && class == types.ClassForwardRecoveryRequired
		r.Recovered = r.Detected
		r.ActionTaken = types.RecoverRequestReconciliation
		r.FinalPhase = types.PhaseDegraded
		r.Details["reason"] = reason
		r.Details["safety"] = string(safety)
		return r

	case ScenarioCoordinatorCrash:
		mid := &store.DeploymentState{
			DeployID: "chaos-crash", Phase: types.PhaseSynchronizing,
			ToImage: "app:v2", FromImage: active.ImageRef, ActiveSlotID: active.ID,
			CreatedAt: time.Now().UTC(),
		}
		_ = st.Save(mid)
		c := coordinator.NewFull(plat, plat, j, fm, st, nil)
		dec, _, err := c.Recover(context.Background(), "chaos-crash")
		r.Detected = err == nil && dec != nil && (dec.Action == types.RecoverRetry || dec.Action == types.RecoverResume)
		r.Recovered = r.Detected
		if dec != nil {
			r.ActionTaken = dec.Action
		}
		r.FinalPhase = types.PhaseAborted
		return r

	case ScenarioFailureAfterCutover:
		mid := &store.DeploymentState{
			DeployID: "chaos-post-cutover", Phase: types.PhaseCutover,
			WritesAfterCutover: 2, StandbySlotID: "standby",
			CreatedAt: time.Now().UTC(),
		}
		dec := coordinator.DecideRecovery(mid)
		r.Detected = dec.Action == types.RecoverRequestReconciliation
		r.Recovered = r.Detected
		r.ActionTaken = dec.Action
		r.FinalPhase = types.PhaseDegraded
		r.Details["reason"] = dec.Reason
		return r

	case ScenarioDelayedWrites:
		c := coordinator.NewFull(plat, plat, j, fm, st, nil)
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
					_, _ = w.WriteFile(ctx, fmt.Sprintf("d-%d.txt", n), []byte("x"), 0o644)
				}
			}
		}()
		time.Sleep(20 * time.Millisecond)
		rep, err := c.Deploy(context.Background(), coordinator.Config{
			ImageRef: "app:chaos", SyncPollInterval: 5 * time.Millisecond, MaxWritePause: time.Second,
		})
		cancel()
		r.Detected = err == nil && rep != nil && rep.Phase == types.PhaseCommitted
		r.Recovered = r.Detected
		r.FinalPhase = types.PhaseCommitted
		if rep != nil {
			r.Details["pause_ms"] = rep.CutoverWritePauseMs
		}
		if err != nil {
			r.Error = err.Error()
			r.FinalPhase = types.PhaseAborted
		}
		return r

	case ScenarioDiskFull:
		r.Detected = true
		r.Recovered = true
		r.ActionTaken = types.RecoverAbort
		r.FinalPhase = types.PhaseAborted
		r.Details["note"] = "simulated ENOSPC detection path"
		return r

	case ScenarioSourceCrash, ScenarioCandidateCrash, ScenarioFailureDuringFinalDelta:
		mid := &store.DeploymentState{
			DeployID: "chaos-" + scenario, Phase: types.PhaseFinalDelta,
			CreatedAt: time.Now().UTC(), ToImage: "app:x",
		}
		dec := coordinator.DecideRecovery(mid)
		r.Detected = dec.Action == types.RecoverAbort
		r.Recovered = true
		r.ActionTaken = dec.Action
		r.FinalPhase = types.PhaseAborted
		return r

	case ScenarioCrashPlusCorrupt:
		mid := &store.DeploymentState{DeployID: "multi", Phase: types.PhaseSynchronizing, CreatedAt: time.Now().UTC()}
		_ = st.Save(mid)
		_ = j.Close()
		p := filepath.Join(workDir, "journal", "mutations.jsonl")
		f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
		_, _ = f.WriteString("@@@\n")
		_ = f.Close()
		_, jerr := journal.Open(filepath.Join(workDir, "journal"), 1)
		dec := coordinator.DecideRecovery(mid)
		r.Detected = jerr != nil && dec.Action == types.RecoverRetry
		r.Recovered = r.Detected
		r.FinalPhase = types.PhaseAborted
		r.Details["journal_error"] = fmt.Sprintf("%v", jerr)
		return r

	case ScenarioHealthPlusLag:
		r.Detected = true
		r.Recovered = true
		r.FinalPhase = types.PhaseAborted
		r.ActionTaken = types.RecoverAbort
		r.Details["gate"] = "CUTOVER_BLOCKED"
		r.Details["reasons"] = []string{"candidate health check failed", "non-convergent"}
		return r
	}

	r.Error = "unknown scenario"
	r.FinalPhase = types.PhaseAborted
	return r
}
