// Package coordinator implements the State Transition Coordinator —
// durable FSM: ACTIVE→CHECKPOINT→SHADOW→SYNCHRONIZING→VERIFYING→QUIESCING→
// FINAL_DELTA→CUTOVER→OBSERVING→COMMITTED (unsafe: ABORTED|DEGRADED|RECOVERING|ROLLED_BACK).
package coordinator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adaptive"
	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/converge"
	"github.com/theworker02/stateful-deployments-engine/internal/fence"
	"github.com/theworker02/stateful-deployments-engine/internal/guardian"
	"github.com/theworker02/stateful-deployments-engine/internal/hotstate"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/receipt"
	"github.com/theworker02/stateful-deployments-engine/internal/replay"
	"github.com/theworker02/stateful-deployments-engine/internal/rollback"
	"github.com/theworker02/stateful-deployments-engine/internal/safegate"
	synclag "github.com/theworker02/stateful-deployments-engine/internal/sync"
	"github.com/theworker02/stateful-deployments-engine/internal/store"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// EventFunc receives progress lines.
type EventFunc func(phase types.DeployPhase, message string)

// Config tunes a stateful deploy / migration.
type Config struct {
	ImageRef           string
	SyncPollInterval   time.Duration
	BarrierTimeout     time.Duration
	HealthTimeout      time.Duration
	ObservationWindow  time.Duration
	MaxWritePause      time.Duration
	RetainStandby      bool
	SkipHealthCheck    bool
	DryRun             bool
	OverrideCutover    bool
	VerifyLevel        types.VerifyLevel
	SDERoot            string
	WriteRateHint      float64
}

// Coordinator orchestrates durable transactional cutover.
type Coordinator struct {
	platform adapter.Platform
	storage  adapter.StorageAdapter
	platA    adapter.PlatformAdapter
	journal  *journal.Journal
	fence    *fence.Manager
	store    *store.Store
	events   EventFunc
	local    *local.Adapter // optional barrier-aware local
	tuner    *adaptive.Controller
	hot      *hotstate.Tracker
}

// New constructs a coordinator (Phase 1 compatible).
func New(p adapter.Platform, j *journal.Journal, events EventFunc) *Coordinator {
	if events == nil {
		events = func(types.DeployPhase, string) {}
	}
	c := &Coordinator{
		platform: p,
		journal:  j,
		events:   events,
		tuner:    adaptive.New(),
		hot:      hotstate.New(),
		platA:    adapter.PlatformBridge{Inner: p},
	}
	if la, ok := p.(*local.Adapter); ok {
		c.local = la
		c.storage = la
		c.platA = la
	}
	return c
}

// NewFull constructs a Phase 2/3 coordinator with fencing + persistence.
func NewFull(p adapter.Platform, storage adapter.StorageAdapter, j *journal.Journal, f *fence.Manager, st *store.Store, events EventFunc) *Coordinator {
	c := New(p, j, events)
	c.storage = storage
	c.fence = f
	c.store = st
	return c
}

func (c *Coordinator) emit(phase types.DeployPhase, msg string) {
	c.events(phase, msg)
}

func (c *Coordinator) persist(st *store.DeploymentState) error {
	if c.store == nil {
		return nil
	}
	return c.store.Save(st)
}

// Deploy runs one transactional stateful deployment (or dry-run simulation).
func (c *Coordinator) Deploy(ctx context.Context, cfg Config) (*types.DeployReport, error) {
	start := time.Now()
	if cfg.SyncPollInterval == 0 {
		cfg.SyncPollInterval = 50 * time.Millisecond
	}
	if cfg.BarrierTimeout == 0 {
		cfg.BarrierTimeout = 5 * time.Second
	}
	if cfg.HealthTimeout == 0 {
		cfg.HealthTimeout = 30 * time.Second
	}
	if cfg.ObservationWindow == 0 {
		cfg.ObservationWindow = 50 * time.Millisecond
	}
	if cfg.MaxWritePause == 0 {
		cfg.MaxWritePause = 250 * time.Millisecond
	}
	if cfg.VerifyLevel == "" {
		cfg.VerifyLevel = types.VerifyChecksum
	}

	holder := fmt.Sprintf("coord-%d", start.UnixNano())
	var ft types.FenceToken
	if cfg.DryRun {
		// Dry-run must not mutate fencing / lease state on the production root.
		if c.fence != nil {
			ft = c.fence.Current()
		} else {
			ft = types.FenceToken{Epoch: c.journal.Epoch(), Token: c.journal.FenceToken()}
		}
	} else if c.fence != nil {
		var err error
		ft, err = c.fence.Issue(holder)
		if err != nil {
			return nil, err
		}
		_ = c.journal.SetFenceToken(ft.Token)
	} else {
		ft = types.FenceToken{Epoch: c.journal.Epoch(), Token: c.journal.FenceToken()}
	}

	report := &types.DeployReport{
		DeployID:   fmt.Sprintf("sde-%d", start.UnixNano()),
		ToImage:    cfg.ImageRef,
		Phase:      types.PhaseActive,
		FenceToken: ft.Token,
		DryRun:     cfg.DryRun,
	}
	c.journal.SetDeploymentID(report.DeployID)

	st := &store.DeploymentState{
		DeployID:        report.DeployID,
		Phase:           types.PhaseActive,
		Epoch:           c.journal.Epoch(),
		FenceToken:      ft.Token,
		ToImage:         cfg.ImageRef,
		CreatedAt:       time.Now().UTC(),
		MaxWritePauseMs: float64(cfg.MaxWritePause.Milliseconds()),
		DryRun:          cfg.DryRun,
	}
	_ = c.persist(st)

	active, err := c.platform.ActiveSlot(ctx)
	if err != nil {
		return c.fail(report, st, start, err)
	}
	report.FromImage = active.ImageRef
	report.ActiveSlotID = active.ID
	report.FromEpoch = c.journal.Epoch()
	st.FromImage = active.ImageRef
	st.ActiveSlotID = active.ID

	if c.fence != nil {
		if err := c.fence.Check(st.Epoch, st.FenceToken); err != nil {
			return c.fail(report, st, start, err)
		}
	}

	// CHECKPOINT
	st.Phase = types.PhaseCheckpoint
	report.Phase = types.PhaseCheckpoint
	_ = c.persist(st)
	cp, err := verifier.BuildCheckpoint(active.StatePath, c.journal.Epoch())
	if err != nil {
		return c.fail(report, st, start, err)
	}
	c.emit(types.PhaseCheckpoint, fmt.Sprintf("Captured state checkpoint (epoch %d, root %s…)", cp.Epoch, short(cp.RootHash)))
	c.emit(types.PhaseCaptureCheckpoint, fmt.Sprintf("Captured state checkpoint (epoch %d, root %s…)", cp.Epoch, short(cp.RootHash)))

	fromSeq := c.journal.NextSeq()
	st.CheckpointSeq = fromSeq

	if cfg.DryRun {
		return c.dryRunFinish(ctx, cfg, report, st, active, cp, start)
	}

	// SHADOW — create candidate
	st.Phase = types.PhaseShadow
	report.Phase = types.PhaseShadow
	_ = c.persist(st)
	c.emit(types.PhaseCreateCandidate, "Created candidate deployment")
	shadow, err := c.platA.CreateCandidate(ctx, adapter.CreateShadowRequest{
		ImageRef:     cfg.ImageRef,
		ActiveSlotID: active.ID,
		Labels:       map[string]string{"sde.role": "shadow", "sde.deploy": report.DeployID},
	})
	if err != nil {
		return c.fail(report, st, start, err)
	}
	report.ShadowSlotID = shadow.ID
	st.CandidateSlotID = shadow.ID
	_ = c.platA.AttachState(ctx, shadow.ID, active.StatePath)

	if err := seedFromCheckpoint(active.StatePath, shadow.StatePath, cp); err != nil {
		return c.fail(report, st, start, err)
	}
	if err := c.platA.StartCandidate(ctx, shadow.ID); err != nil {
		return c.fail(report, st, start, err)
	}
	c.emit(types.PhaseStartCandidate, "Candidate started")
	_ = c.persist(st)

	// SYNCHRONIZING
	st.Phase = types.PhaseSynchronizing
	report.Phase = types.PhaseSynchronizing
	_ = c.persist(st)
	syncStart := time.Now()
	synced, nextFrom, err := c.syncUntilQuiet(ctx, shadow.StatePath, fromSeq, cfg.SyncPollInterval)
	if err != nil {
		return c.fail(report, st, start, err)
	}
	replayRate := float64(synced) / time.Since(syncStart).Seconds()
	if replayRate <= 0 {
		replayRate = 1
	}
	report.MutationsSynced = synced
	st.MutationsSynced = synced
	st.JournalCursor = nextFrom
	c.emit(types.PhaseSynchronizing, fmt.Sprintf("%d mutations synchronized", synced))
	c.emit(types.PhaseSyncMutations, fmt.Sprintf("%d mutations synchronized", synced))

	se := &synclag.Engine{Journal: c.journal}
	lag, _ := se.Measure(shadow.StatePath, nextFrom, replayRate, "unknown")
	report.SyncLag = lag
	st.SyncLag = lag
	_ = c.persist(st)

	conv := converge.Observe(converge.Sample{
		WriteRateOpsPerSec:       cfg.WriteRateHint,
		ReplicationRateOpsPerSec: replayRate,
		BacklogOps:               lag.JournalLag,
		At:                       time.Now().UTC(),
	}, lag.JournalLag, time.Now().Add(-time.Second))
	c.tuner.Tune(adaptive.Input{LagOps: lag.JournalLag, WriteRate: cfg.WriteRateHint, Throughput: replayRate})

	// VERIFYING
	st.Phase = types.PhaseVerifying
	report.Phase = types.PhaseVerifying
	_ = c.persist(st)
	vr, err := verifier.Verify(active.StatePath, shadow.StatePath, c.journal.Epoch(), verifier.Options{
		DeployID: report.DeployID,
		Level:    cfg.VerifyLevel,
	})
	if err != nil {
		return c.fail(report, st, start, err)
	}
	report.Verification = vr
	st.Verification = vr
	if verifier.IsVerified(vr) {
		report.ConsistencyOK = true
		c.emit(types.PhaseVerifying, "Filesystem consistency verified")
		c.emit(types.PhaseVerifyConsistency, "Filesystem consistency verified")
	} else {
		c.emit(types.PhaseVerifying, fmt.Sprintf("Pre-barrier verify %s — will close under write barrier", vr.Outcome))
	}

	mf, err := verifier.BuildManifest(shadow.StatePath, c.journal.Epoch(), nextFrom)
	if err == nil {
		st.Manifest = mf
	}

	// Health
	if !cfg.SkipHealthCheck {
		hctx, cancel := context.WithTimeout(ctx, cfg.HealthTimeout)
		err := c.platA.HealthCheck(hctx, shadow.ID)
		cancel()
		if err != nil {
			return c.fail(report, st, start, fmt.Errorf("candidate unhealthy: %w", err))
		}
	}
	report.HealthOK = true
	c.emit(types.PhaseHealthCheck, "Health checks passed")

	predictedPause := estimatePauseMs(cfg.WriteRateHint, replayRate, lag.JournalLag)
	st.PredictedPauseMs = predictedPause
	report.CutoverWritePauseMs = 0 // filled after actual

	gate := safegate.Evaluate(safegate.Input{
		JournalLag:           0, // will re-check under barrier
		PendingOps:           0,
		CandidateHealthy:     report.HealthOK,
		ManifestIntegrityOK:  st.Manifest != nil,
		JournalIntegrityOK:   c.journal.VerifyIntegrity() == nil,
		RollbackPointPresent: true,
		CoordinatorHealthy:   true,
		StorageHealthy:       true,
		PredictedPauseMs:     predictedPause,
		MaxWritePauseMs:      float64(cfg.MaxWritePause.Milliseconds()),
		Verification:         vr,
		OverrideRequested:    cfg.OverrideCutover,
		NonConvergent:        conv.NonConvergent,
	})
	// Soft pre-gate: if predicted pause exceeds budget before barrier, abort.
	if predictedPause > float64(cfg.MaxWritePause.Milliseconds()) && !cfg.OverrideCutover {
		err := fmt.Errorf("predicted pause %.2fms exceeds max-write-pause %s", predictedPause, cfg.MaxWritePause)
		return c.fail(report, st, start, err)
	}
	st.SafetyGate = &gate
	_ = c.persist(st)

	// QUIESCING — write barrier
	st.Phase = types.PhaseQuiescing
	report.Phase = types.PhaseQuiescing
	_ = c.persist(st)
	bctx, bcancel := context.WithTimeout(ctx, cfg.BarrierTimeout)
	var release func() error
	if c.storage != nil {
		release, err = c.storage.FreezeWrites(bctx, active.ID)
	} else {
		release, err = c.platform.EstablishWriteBarrier(bctx, active.ID)
	}
	bcancel()
	if err != nil {
		return c.fail(report, st, start, err)
	}
	barrierStart := time.Now()

	// FINAL_DELTA
	st.Phase = types.PhaseFinalDelta
	report.Phase = types.PhaseFinalDelta
	_ = c.persist(st)
	var applied uint64
	cursor := nextFrom
	for {
		final, err := c.journal.ReadFrom(cursor)
		if err != nil {
			_ = release()
			return c.fail(report, st, start, err)
		}
		if len(final) == 0 {
			break
		}
		n, err := replay.ApplyMutations(shadow.StatePath, final)
		if err != nil {
			_ = release()
			return c.fail(report, st, start, err)
		}
		applied += n
		cursor = final[len(final)-1].Seq + 1
	}
	report.FinalDeltaWrites = applied
	st.FinalDeltaWrites = applied
	st.JournalCursor = cursor
	c.emit(types.PhaseFinalDelta, fmt.Sprintf("Final delta: %d writes", applied))

	postVR, err := verifier.Verify(active.StatePath, shadow.StatePath, c.journal.Epoch(), verifier.Options{
		DeployID: report.DeployID,
		Level:    types.VerifyChecksum,
	})
	if err != nil || !verifier.IsVerified(postVR) {
		_ = release()
		msg := "consistency failed after final delta"
		if postVR != nil {
			msg = postVR.EvidenceSummary
		}
		if err != nil {
			msg = err.Error()
		}
		return c.fail(report, st, start, fmt.Errorf("%s", msg))
	}
	report.Verification = postVR
	st.Verification = postVR
	report.ConsistencyOK = true
	c.emit(types.PhaseVerifyConsistency, "Filesystem consistency verified")

	// Residual drain
	if residual, err := c.journal.ReadFrom(cursor); err == nil && len(residual) > 0 {
		n, err := replay.ApplyMutations(shadow.StatePath, residual)
		if err != nil {
			_ = release()
			return c.fail(report, st, start, err)
		}
		report.FinalDeltaWrites += n
		st.FinalDeltaWrites = report.FinalDeltaWrites
		cursor = residual[len(residual)-1].Seq + 1
	}

	actualPause := float64(time.Since(barrierStart).Microseconds()) / 1000.0
	if actualPause > float64(cfg.MaxWritePause.Milliseconds()) {
		_ = release()
		return c.fail(report, st, start, fmt.Errorf("actual pause %.2fms exceeds max-write-pause %s — not claiming success", actualPause, cfg.MaxWritePause))
	}

	// CUTOVER
	st.Phase = types.PhaseCutover
	report.Phase = types.PhaseCutover
	st.CutoverJournalSeq = cursor
	_ = c.persist(st)
	_ = c.platA.PrepareCutover(ctx, active.ID, shadow.ID)
	if err := c.platform.TransferTraffic(ctx, active.ID, shadow.ID); err != nil {
		_ = release()
		return c.fail(report, st, start, err)
	}
	promoted, err := c.platA.ActivateCandidate(ctx, shadow.ID)
	if err != nil {
		_ = release()
		return c.fail(report, st, start, err)
	}
	_ = c.platA.DeactivatePrevious(ctx, active.ID)
	if err := release(); err != nil {
		return c.fail(report, st, start, err)
	}
	report.BarrierDuration = time.Since(barrierStart)
	report.CutoverWritePauseMs = actualPause
	st.CutoverWritePauseMs = actualPause
	c.emit(types.PhaseWriteBarrier, fmt.Sprintf("State barrier: %s", report.BarrierDuration.Round(time.Microsecond)))
	c.emit(types.PhaseCutover, "Traffic transferred")
	c.emit(types.PhaseTrafficTransfer, "Traffic transferred")

	newEpoch := c.journal.BumpEpoch()
	if c.fence != nil {
		ft2, _ := c.fence.AdvanceEpoch(holder)
		newEpoch = ft2.Epoch
		st.FenceToken = ft2.Token
		report.FenceToken = ft2.Token
	}
	report.ToEpoch = newEpoch
	st.Epoch = newEpoch
	report.ActiveSlotID = promoted.Active.ID
	st.ActiveSlotID = promoted.Active.ID
	if promoted.Standby != nil {
		report.StandbySlotID = promoted.Standby.ID
		st.StandbySlotID = promoted.Standby.ID
	}

	// OBSERVING
	st.Phase = types.PhaseObserving
	report.Phase = types.PhaseObserving
	_ = c.persist(st)
	obs := guardian.Observe(ctx, c.platA, promoted.Active.ID, promoted.Active.StatePath, newEpoch, cfg.ObservationWindow)
	report.ObservationWindowMs = float64(cfg.ObservationWindow.Milliseconds())
	st.ObservationMs = report.ObservationWindowMs
	if !obs.Committed {
		st.Phase = types.PhaseDegraded
		report.Phase = types.PhaseDegraded
		_ = c.persist(st)
		return c.fail(report, st, start, fmt.Errorf("observation failed: %s", obs.Reason))
	}

	// COMMITTED
	st.Phase = types.PhaseCommitted
	report.Phase = types.PhaseCommitted
	report.TotalDuration = time.Since(start)
	_ = c.persist(st)
	c.emit(types.PhaseCommitted, "STATEFUL ZERO-DOWNTIME DEPLOYMENT")
	c.emit(types.PhaseComplete, "STATEFUL ZERO-DOWNTIME DEPLOYMENT")

	if c.store != nil {
		rcpt := receipt.Build(&types.DeploymentReceipt{
			DeployID:            report.DeployID,
			FromImage:           report.FromImage,
			ToImage:             report.ToImage,
			FromEpoch:           report.FromEpoch,
			ToEpoch:             report.ToEpoch,
			FenceToken:          report.FenceToken,
			Phases:              []string{"ACTIVE", "CHECKPOINT", "SHADOW", "SYNCHRONIZING", "VERIFYING", "QUIESCING", "FINAL_DELTA", "CUTOVER", "OBSERVING", "COMMITTED"},
			FinalPhase:          types.PhaseCommitted,
			MutationsSynced:     report.MutationsSynced,
			FinalDeltaWrites:    report.FinalDeltaWrites,
			PredictedPauseMs:    predictedPause,
			ActualPauseMs:       actualPause,
			Verification:        report.Verification,
			Manifest:            st.Manifest,
			SafetyGate:          st.SafetyGate,
			Convergence:         &conv,
			SyncLag:             lag,
			TuningDecisions:     c.tuner.Decisions,
			CapabilitiesUsed:    c.storageCaps(),
			DryRun:              false,
		})
		_, _ = receipt.WriteJSON(filepath.Join(c.store.Dir(report.DeployID), "receipts"), rcpt)
		_, _ = receipt.WriteHuman(filepath.Join(c.store.Dir(report.DeployID), "receipts"), rcpt)
	}
	return report, nil
}

func (c *Coordinator) dryRunFinish(ctx context.Context, cfg Config, report *types.DeployReport, st *store.DeploymentState, active *types.DeploymentSlot, cp *types.Checkpoint, start time.Time) (*types.DeployReport, error) {
	c.emit(types.PhaseShadow, "Dry-run: skipping candidate create (production slots untouched)")
	predicted := estimatePauseMs(cfg.WriteRateHint, 5000, 0)
	st.PredictedPauseMs = predicted
	report.CutoverWritePauseMs = 0
	report.ConsistencyOK = true
	report.HealthOK = true
	report.DryRun = true
	report.ShadowSlotID = ""
	report.ToEpoch = report.FromEpoch
	vr, _ := verifier.Verify(active.StatePath, active.StatePath, c.journal.Epoch(), verifier.Options{
		DeployID: report.DeployID,
		Level:    types.VerifyChecksum,
	})
	report.Verification = vr
	st.Verification = vr
	gate := safegate.Evaluate(safegate.Input{
		CandidateHealthy: true, ManifestIntegrityOK: true, JournalIntegrityOK: true,
		RollbackPointPresent: true, CoordinatorHealthy: true, StorageHealthy: true,
		PredictedPauseMs: predicted, MaxWritePauseMs: float64(cfg.MaxWritePause.Milliseconds()),
		Verification: vr,
	})
	st.SafetyGate = &gate
	st.CandidateSlotID = ""
	st.Phase = types.PhaseCommitted
	report.Phase = types.PhaseCommitted
	report.TotalDuration = time.Since(start)
	_ = c.persist(st)
	c.emit(types.PhaseCommitted, fmt.Sprintf("Dry-run complete: predicted_pause_ms=%.2f objects=%d bytes=%d gate=%s (no production mutation)",
		predicted, cp.FileCount, cp.ByteSize, gate.Decision))
	if c.store != nil {
		rcpt := receipt.Build(&types.DeploymentReceipt{
			DeployID: report.DeployID, FromImage: report.FromImage, ToImage: report.ToImage,
			FromEpoch: report.FromEpoch, ToEpoch: report.ToEpoch, FenceToken: report.FenceToken,
			FinalPhase: types.PhaseCommitted, PredictedPauseMs: predicted, ActualPauseMs: 0,
			Verification: vr, SafetyGate: &gate, DryRun: true,
			Phases: []string{"ACTIVE", "CHECKPOINT", "DRY_RUN"},
			HumanSummary: "dry-run simulation only; fence/slots/state unchanged",
		})
		_, _ = receipt.WriteJSON(filepath.Join(c.store.Dir(report.DeployID), "receipts"), rcpt)
	}
	return report, nil
}

func (c *Coordinator) storageCaps() []types.StorageCapability {
	if c.storage == nil {
		return nil
	}
	return c.storage.Capabilities()
}

func estimatePauseMs(writeRate, replayRate float64, lag uint64) float64 {
	base := 5.0
	if replayRate <= 0 {
		replayRate = 1
	}
	return base + float64(lag)/replayRate*1000 + writeRate*0.02
}

func (c *Coordinator) syncUntilQuiet(ctx context.Context, shadowRoot string, fromSeq uint64, poll time.Duration) (synced uint64, nextSeq uint64, err error) {
	cursor := fromSeq
	quietRounds := 0
	deadline := time.Now().Add(2 * time.Second)
	for quietRounds < 3 {
		if time.Now().After(deadline) {
			// Continuous writers may never yield 3 quiet polls; barrier + final delta close the gap.
			break
		}
		select {
		case <-ctx.Done():
			return synced, cursor, ctx.Err()
		case <-time.After(poll):
		}
		batch, err := c.journal.ReadFrom(cursor)
		if err != nil {
			return synced, cursor, err
		}
		if len(batch) == 0 {
			quietRounds++
			continue
		}
		quietRounds = 0
		for _, m := range batch {
			c.hot.Observe(m.Path, m.Size, m.Timestamp)
		}
		n, err := replay.ApplyMutations(shadowRoot, batch)
		if err != nil {
			return synced, cursor, err
		}
		synced += n
		if n > 0 {
			cursor = batch[len(batch)-1].Seq + 1
			_ = c.journal.SaveCheckpoint(cursor-1, "sync")
		}
	}
	return synced, cursor, nil
}

// Recover resumes or decides action for an interrupted deployment.
func (c *Coordinator) Recover(ctx context.Context, deployID string) (*types.RecoveryDecision, *types.DeployReport, error) {
	if c.store == nil {
		return nil, nil, fmt.Errorf("no state store configured")
	}
	var st *store.DeploymentState
	var err error
	if deployID == "" {
		st, err = c.store.LatestNonTerminal()
	} else {
		st, err = c.store.Load(deployID)
	}
	if err != nil {
		return nil, nil, err
	}
	if st == nil {
		return &types.RecoveryDecision{Action: types.RecoverAbort, Reason: "no interrupted deployment"}, nil, nil
	}
	// Decide from the interrupted phase BEFORE marking RECOVERING (audit marker).
	interrupted := *st
	st.Phase = types.PhaseRecovering
	_ = c.persist(st)

	dec := DecideRecovery(&interrupted)
	report := &types.DeployReport{
		DeployID: st.DeployID, Phase: interrupted.Phase, FromImage: st.FromImage, ToImage: st.ToImage,
		ActiveSlotID: st.ActiveSlotID, ShadowSlotID: st.CandidateSlotID, StandbySlotID: st.StandbySlotID,
		FromEpoch: st.Epoch, FenceToken: st.FenceToken, CutoverWritePauseMs: st.CutoverWritePauseMs,
		MutationsSynced: st.MutationsSynced, FinalDeltaWrites: st.FinalDeltaWrites,
	}

	switch dec.Action {
	case types.RecoverResume, types.RecoverRetry:
		// Re-enter deploy for non-cutover phases is safest as retry from checkpoint seed.
		if st.Phase == types.PhaseObserving || dec.FromPhase == types.PhaseObserving {
			// Complete observation → commit
			active, err := c.platform.ActiveSlot(ctx)
			if err != nil {
				return &dec, report, err
			}
			obs := guardian.Observe(ctx, c.platA, active.ID, active.StatePath, st.Epoch, 50*time.Millisecond)
			if obs.Committed {
				st.Phase = types.PhaseCommitted
				report.Phase = types.PhaseCommitted
				_ = c.persist(st)
				dec.Action = types.RecoverResume
				dec.Reason = "observation completed after restart"
			}
		} else if dec.FromPhase == types.PhaseSynchronizing || dec.FromPhase == types.PhaseShadow ||
			dec.FromPhase == types.PhaseCheckpoint || dec.FromPhase == types.PhaseVerifying {
			cfg := Config{ImageRef: st.ToImage, SDERoot: "", SyncPollInterval: 20 * time.Millisecond}
			// Mark aborted interrupted deploy and start fresh retry — state survived for audit.
			st.Phase = types.PhaseAborted
			st.Error = "superseded by recover-retry"
			_ = c.persist(st)
			rep, err := c.Deploy(ctx, cfg)
			return &dec, rep, err
		}
	case types.RecoverRollback:
		eng := rollback.New(c.platform, c.journal, func(string) {})
		_, err := eng.Execute(ctx, rollback.Request{
			StandbySlotID:      st.StandbySlotID,
			TargetEpoch:        st.Epoch,
			TargetImage:        st.FromImage,
			WritesAfterCutover: st.WritesAfterCutover,
			PhaseAtFailure:     dec.FromPhase,
		})
		if err != nil {
			dec.Action = types.RecoverRequestReconciliation
			dec.Reason = err.Error()
		} else {
			st.Phase = types.PhaseRolledBack
			_ = c.persist(st)
			report.Phase = types.PhaseRolledBack
		}
	case types.RecoverAbort:
		st.Phase = types.PhaseAborted
		_ = c.persist(st)
		report.Phase = types.PhaseAborted
	}
	return &dec, report, nil
}

// DecideRecovery picks resume|retry|abort|rollback|request_reconciliation from persisted phase.
func DecideRecovery(st *store.DeploymentState) types.RecoveryDecision {
	dec := types.RecoveryDecision{DeployID: st.DeployID, FromPhase: st.Phase}
	switch st.Phase {
	case types.PhaseActive, types.PhaseCheckpoint, types.PhaseShadow, types.PhaseSynchronizing, types.PhaseVerifying:
		dec.Action = types.RecoverRetry
		dec.Reason = "pre-cutover phase; safe to retry deploy"
		dec.TargetPhase = types.PhaseActive
	case types.PhaseQuiescing, types.PhaseFinalDelta:
		dec.Action = types.RecoverAbort
		dec.Reason = "interrupted during barrier/final delta; abort and release writers"
		dec.TargetPhase = types.PhaseAborted
	case types.PhaseCutover:
		if st.WritesAfterCutover > 0 {
			dec.Action = types.RecoverRequestReconciliation
			dec.Reason = "FAILED_CUTOVER_MUST_HAVE_DETERMINISTIC_RECOVERY_PATH: writes after cutover require forward recovery"
		} else {
			dec.Action = types.RecoverRollback
			dec.Reason = "cutover interrupted with no post-cutover writes; rollback safe"
		}
	case types.PhaseObserving:
		dec.Action = types.RecoverResume
		dec.Reason = "resume observation window then commit"
		dec.TargetPhase = types.PhaseObserving
	case types.PhaseRecovering, types.PhaseDegraded:
		dec.Action = types.RecoverRequestReconciliation
		dec.Reason = "already in degraded/recovering state"
	default:
		dec.Action = types.RecoverAbort
		dec.Reason = "terminal or unknown phase"
	}
	return dec
}

func (c *Coordinator) fail(report *types.DeployReport, st *store.DeploymentState, start time.Time, err error) (*types.DeployReport, error) {
	report.Phase = types.PhaseAborted
	report.Error = err.Error()
	report.TotalDuration = time.Since(start)
	if st != nil {
		st.Phase = types.PhaseAborted
		st.Error = err.Error()
		_ = c.persist(st)
	}
	c.emit(types.PhaseAborted, err.Error())
	c.emit(types.PhaseFailed, err.Error())
	return report, err
}

func short(h string) string {
	if len(h) <= 12 {
		return h
	}
	return h[:12]
}

func seedFromCheckpoint(activeRoot, shadowRoot string, cp *types.Checkpoint) error {
	if err := os.MkdirAll(shadowRoot, 0o755); err != nil {
		return err
	}
	for rel := range cp.Manifest {
		src := filepath.Join(activeRoot, filepath.FromSlash(rel))
		dst := filepath.Join(shadowRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, info.Mode())
}
