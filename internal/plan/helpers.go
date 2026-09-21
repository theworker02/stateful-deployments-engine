// Package plan provides migration plan helpers and dry-run scaffolding.
//
// Planning logic lives in internal/planner; this package owns serialization,
// readiness predicates, and sample plan builders for docs/examples.
package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// IsReady reports whether conclusion is READY or READY_WITH_RISK.
// UNKNOWN is never ready.
func IsReady(p *types.MigrationPlan) bool {
	if p == nil {
		return false
	}
	switch p.Conclusion {
	case types.MigrateReady, types.MigrateReadyWithRisk:
		return true
	default:
		return false
	}
}

// WriteJSON persists a plan.
func WriteJSON(path string, p *types.MigrationPlan) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// LoadJSON reads a plan.
func LoadJSON(path string) (*types.MigrationPlan, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p types.MigrationPlan
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// SampleDryRun returns a realistic dry-run plan for docs/examples.
func SampleDryRun(image string, stateBytes int64) *types.MigrationPlan {
	if image == "" {
		image = "app:candidate"
	}
	return &types.MigrationPlan{
		PlanID:               fmt.Sprintf("plan-dry-%d", time.Now().UnixNano()),
		SourceSlotID:         "slot-active-0",
		TargetImage:          image,
		CreatedAt:            time.Now().UTC(),
		StateBytes:           stateBytes,
		ObjectCount:          128,
		WriteRateOpsPerSec:   120,
		MutationRateBytesS:   2 * 1024 * 1024,
		EstimatedBandwidthB:  32 * 1024 * 1024,
		TargetCapacityBytes:  stateBytes * 2,
		CheckpointCapable:    true,
		VerifyCapable:        true,
		HealthRequired:       true,
		HistoricalThroughput: 5000,
		Strategies:           []string{"baseline-copy", "journal-replay", "write-barrier-freeze", "checksum-verify"},
		Risks:                []string{"write rate may spike during sync"},
		Confidence:           0.82,
		EstimatedCatchupMs:   4500,
		EstimatedPauseMs:     55,
		EstimatedRPO:         0,
		EstimatedRTOSec:      12,
		Conclusion:           types.MigrateReadyWithRisk,
		Rationale:            "dry-run sample: catch-up projected under bandwidth; pause within 250ms budget",
	}
}
