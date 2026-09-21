// Package planner produces MigrationPlan readiness verdicts for live state mobility.
package planner

import (
	"context"
	"fmt"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// Input gathers observations used to plan a migration.
type Input struct {
	TargetImage          string
	WriteRateOpsPerSec   float64
	MutationRateBytesS   float64
	EstimatedBandwidthB  float64
	TargetCapacityBytes  int64
	HistoricalThroughput float64
	HealthRequired       bool
	SampleWindow         time.Duration
}

// Planner inspects source state + storage capabilities.
type Planner struct {
	Platform adapter.PlatformAdapter
	Storage  adapter.StorageAdapter
	Journal  *journal.Journal
}

// Plan inspects the active source and returns a MigrationPlan.
// UNKNOWN is never treated as READY by callers of IsReady.
func (p *Planner) Plan(ctx context.Context, in Input) (*types.MigrationPlan, error) {
	active, err := p.Platform.ActiveSlot(ctx)
	if err != nil {
		return &types.MigrationPlan{
			PlanID:     fmt.Sprintf("plan-%d", time.Now().UnixNano()),
			CreatedAt:  time.Now().UTC(),
			Conclusion: types.MigrateUnknown,
			Rationale:  "unable to inspect active slot: " + err.Error(),
			Confidence: 0,
		}, nil
	}
	cp, err := verifier.BuildCheckpoint(active.StatePath, p.Journal.Epoch())
	if err != nil {
		return &types.MigrationPlan{
			PlanID:       fmt.Sprintf("plan-%d", time.Now().UnixNano()),
			SourceSlotID: active.ID,
			CreatedAt:    time.Now().UTC(),
			Conclusion:   types.MigrateUnknown,
			Rationale:    "checkpoint failed: " + err.Error(),
			Confidence:   0,
		}, nil
	}

	caps := p.Storage.Capabilities()
	strategies := adapter.SelectStrategies(caps)
	bw := in.EstimatedBandwidthB
	if bw <= 0 {
		bw = 32 * 1024 * 1024 // assume 32 MiB/s local default for planning only
	}
	hist := in.HistoricalThroughput
	if hist <= 0 {
		hist = 5000
	}
	writeRate := in.WriteRateOpsPerSec
	catchupMs := 0.0
	if hist > writeRate && cp.FileCount > 0 {
		// rough: baseline copy time + catch-up of live writes during copy
		copySec := float64(cp.ByteSize) / bw
		catchupMs = (copySec*writeRate)/hist*1000 + copySec*1000
	} else if hist <= writeRate && writeRate > 0 {
		catchupMs = -1
	}

	plan := &types.MigrationPlan{
		PlanID:               fmt.Sprintf("plan-%d", time.Now().UnixNano()),
		SourceSlotID:         active.ID,
		TargetImage:          in.TargetImage,
		CreatedAt:            time.Now().UTC(),
		StateBytes:           cp.ByteSize,
		ObjectCount:          cp.FileCount,
		WriteRateOpsPerSec:   writeRate,
		MutationRateBytesS:   in.MutationRateBytesS,
		EstimatedBandwidthB:  bw,
		TargetCapacityBytes:  in.TargetCapacityBytes,
		CheckpointCapable:    adapter.HasCapability(caps, types.CapChecksum) || adapter.HasCapability(caps, types.CapSnapshot),
		VerifyCapable:        adapter.HasCapability(caps, types.CapChecksum),
		HealthRequired:       in.HealthRequired,
		HistoricalThroughput: hist,
		Strategies:           strategies,
		EstimatedCatchupMs:   catchupMs,
		EstimatedPauseMs:     estimatePause(writeRate, hist),
		EstimatedRPO:         0,
		EstimatedRTOSec:      catchupMs / 1000,
		Confidence:           0.7,
	}

	var risks []string
	conclusion := types.MigrateReady
	rationale := "source inspectable; strategies selected from capabilities"

	if in.TargetCapacityBytes > 0 && cp.ByteSize > in.TargetCapacityBytes {
		conclusion = types.MigrateBlocked
		risks = append(risks, "target capacity smaller than source state")
		rationale = "BLOCKED: insufficient target capacity"
		plan.Confidence = 0.95
	} else if !plan.VerifyCapable {
		conclusion = types.MigrateBlocked
		risks = append(risks, "storage cannot checksum-verify")
		rationale = "BLOCKED: verification capability missing"
	} else if writeRate > 0 && hist > 0 && writeRate >= hist {
		conclusion = types.MigrateLikelyNonConvergent
		risks = append(risks, "write rate >= historical replication rate")
		rationale = "LIKELY_NON_CONVERGENT: backlog will not drain without mitigation"
		plan.Confidence = 0.8
	} else if writeRate > hist*0.7 {
		conclusion = types.MigrateReadyWithRisk
		risks = append(risks, "write rate close to replication capacity")
		rationale = "READY_WITH_RISK: may need quiescence or hot-object prioritization"
		plan.Confidence = 0.6
	} else if cp.FileCount == 0 && cp.ByteSize == 0 && writeRate == 0 {
		// empty state is fine
		conclusion = types.MigrateReady
		rationale = "empty or idle source; ready"
		plan.Confidence = 0.9
	}

	if plan.EstimatedPauseMs > 250 {
		risks = append(risks, fmt.Sprintf("predicted pause %.1fms exceeds default 250ms budget", plan.EstimatedPauseMs))
		if conclusion == types.MigrateReady {
			conclusion = types.MigrateReadyWithRisk
		}
	}

	plan.Risks = risks
	plan.Conclusion = conclusion
	plan.Rationale = rationale
	return plan, nil
}

func estimatePause(writeRate, hist float64) float64 {
	// Barrier holds while final delta drains residual; scale with write pressure.
	base := 5.0
	if hist <= 0 {
		return base + writeRate*0.05
	}
	return base + (writeRate/hist)*40
}

// IsReady reports whether conclusion allows proceeding (never true for UNKNOWN).
func IsReady(c types.MigrationConclusion) bool {
	return c == types.MigrateReady || c == types.MigrateReadyWithRisk
}
