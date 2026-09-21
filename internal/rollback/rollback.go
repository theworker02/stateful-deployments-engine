// Package rollback restores application + state epoch with safety classification.
package rollback

import (
	"context"
	"fmt"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// Engine performs verified rollbacks to a retained standby slot.
type Engine struct {
	platform adapter.Platform
	journal  *journal.Journal
	events   func(string)
}

// New creates a rollback engine.
func New(p adapter.Platform, j *journal.Journal, events func(string)) *Engine {
	if events == nil {
		events = func(string) {}
	}
	return &Engine{platform: p, journal: j, events: events}
}

// Request describes the recovery target.
type Request struct {
	StandbySlotID      string
	TargetEpoch        uint64
	TargetImage        string
	ExpectedHash       string
	WritesAfterCutover uint64
	CutoverJournalSeq  uint64
	PhaseAtFailure     types.DeployPhase
	ForceUnsafe        bool
}

// Classify decides whether storage rollback is safe given acknowledged writes.
func Classify(writesAfterCutover uint64, phase types.DeployPhase) (types.RollbackSafety, types.RecoveryClassification, string) {
	if writesAfterCutover == 0 {
		return types.RollbackSafe, types.ClassRollbackSafe,
			"no acknowledged writes on new active after cutover"
	}
	switch phase {
	case types.PhaseObserving, types.PhaseCutover:
		return types.RollbackRequiresReconciliation, types.ClassForwardRecoveryRequired,
			fmt.Sprintf("%d writes acknowledged on new active; rollback would destroy them — forward recovery required", writesAfterCutover)
	default:
		return types.RollbackUnsafe, types.ClassManualReconciliationReq,
			fmt.Sprintf("%d acknowledged writes on new active; blind storage rollback is UNSAFE", writesAfterCutover)
	}
}

// Execute restores traffic to standby only when classified SAFE (or ForceUnsafe).
func (e *Engine) Execute(ctx context.Context, req Request) (*types.RollbackReport, error) {
	start := time.Now()
	active, err := e.platform.ActiveSlot(ctx)
	if err != nil {
		return &types.RollbackReport{Error: err.Error()}, err
	}

	safety, class, reason := Classify(req.WritesAfterCutover, req.PhaseAtFailure)
	report := &types.RollbackReport{
		FromImage:          active.ImageRef,
		ToImage:            req.TargetImage,
		FromEpoch:          e.journal.Epoch(),
		ToEpoch:            req.TargetEpoch,
		Safety:             safety,
		SafetyReason:       reason,
		WritesAfterCutover: req.WritesAfterCutover,
		Phase:              types.PhaseRolledBack,
	}

	if safety != types.RollbackSafe && !req.ForceUnsafe {
		report.Error = fmt.Sprintf("rollback refused (%s / %s): %s", safety, class, reason)
		return report, fmt.Errorf("%s", report.Error)
	}

	e.events(fmt.Sprintf("Restoring:\n  application:  %s → %s\n  state epoch:  %d  → %d\n  safety: %s",
		shortID(active.ImageRef), shortID(req.TargetImage), report.FromEpoch, req.TargetEpoch, safety))

	if err := e.platform.Rollback(ctx, adapter.RollbackRequest{
		TargetEpoch:   req.TargetEpoch,
		TargetImage:   req.TargetImage,
		StandbySlotID: req.StandbySlotID,
	}); err != nil {
		report.Error = err.Error()
		return report, err
	}

	e.journal.SetEpoch(req.TargetEpoch)

	restored, err := e.platform.ActiveSlot(ctx)
	if err != nil {
		report.Error = err.Error()
		return report, err
	}
	cp, err := verifier.BuildCheckpoint(restored.StatePath, req.TargetEpoch)
	if err != nil {
		report.Error = err.Error()
		return report, err
	}
	if req.ExpectedHash != "" && cp.RootHash != req.ExpectedHash {
		report.IntegrityOK = false
		report.Error = fmt.Sprintf("integrity mismatch: got %s want %s", cp.RootHash, req.ExpectedHash)
		return report, fmt.Errorf("%s", report.Error)
	}
	report.IntegrityOK = true
	report.Duration = time.Since(start)
	e.events(fmt.Sprintf("Recovery completed: %s\nData integrity: VERIFIED\nClassification: %s",
		report.Duration.Round(time.Millisecond), class))
	return report, nil
}

func shortID(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}
