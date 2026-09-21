// Package guardian monitors TARGET_ACTIVE_UNCOMMITTED through the observation window.
package guardian

import (
	"context"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// Result of an observation window.
type Result struct {
	Committed bool
	Phase     types.DeployPhase
	Reason    string
	Errors    []string
}

// Observe watches candidate/active health and integrity for window duration.
func Observe(ctx context.Context, platform adapter.PlatformAdapter, slotID, statePath string, epoch uint64, window time.Duration) Result {
	if window <= 0 {
		window = 50 * time.Millisecond
	}
	deadline := time.Now().Add(window)
	var errs []string
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return Result{Committed: false, Phase: types.PhaseDegraded, Reason: ctx.Err().Error(), Errors: errs}
		case <-time.After(10 * time.Millisecond):
		}
		if err := platform.HealthCheck(ctx, slotID); err != nil {
			errs = append(errs, "health: "+err.Error())
			return Result{Committed: false, Phase: types.PhaseDegraded, Reason: "health failure during observation", Errors: errs}
		}
	}
	// Final integrity self-check (root reachable).
	if _, err := verifier.BuildCheckpoint(statePath, epoch); err != nil {
		errs = append(errs, "integrity: "+err.Error())
		return Result{Committed: false, Phase: types.PhaseDegraded, Reason: "integrity failure during observation", Errors: errs}
	}
	return Result{Committed: true, Phase: types.PhaseCommitted, Reason: "observation window passed"}
}
