package receipt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Validate checks structural integrity of a sealed receipt.
func Validate(r *types.DeploymentReceipt) error {
	if r == nil {
		return fmt.Errorf("receipt: nil")
	}
	if r.ReceiptID == "" {
		return fmt.Errorf("receipt: missing receipt_id")
	}
	if r.DeployID == "" {
		return fmt.Errorf("receipt: missing deploy_id")
	}
	if r.EvidenceHash == "" {
		return fmt.Errorf("receipt: missing evidence_hash (seal with Build)")
	}
	if r.CreatedAt.IsZero() {
		return fmt.Errorf("receipt: missing created_at")
	}
	// Recompute hash and compare.
	want := r.EvidenceHash
	clone := *r
	Build(&clone)
	if clone.EvidenceHash != want {
		return fmt.Errorf("receipt: evidence_hash mismatch (tamper or schema drift)")
	}
	return nil
}

// FromReport builds a receipt skeleton from a DeployReport.
func FromReport(rep types.DeployReport, dryRun bool) *types.DeploymentReceipt {
	return Build(&types.DeploymentReceipt{
		DeployID:            rep.DeployID,
		CreatedAt:           time.Now().UTC(),
		FromImage:           rep.FromImage,
		ToImage:             rep.ToImage,
		FromEpoch:           rep.FromEpoch,
		ToEpoch:             rep.ToEpoch,
		FenceToken:          rep.FenceToken,
		FinalPhase:          rep.Phase,
		MutationsSynced:     rep.MutationsSynced,
		FinalDeltaWrites:    rep.FinalDeltaWrites,
		ActualPauseMs:       rep.CutoverWritePauseMs,
		CutoverWritePauseMs: rep.CutoverWritePauseMs,
		Verification:        rep.Verification,
		SyncLag:             rep.SyncLag,
		DryRun:              dryRun,
	})
}

// MustMarshal returns indented JSON or panics (tests/fixtures only).
func MustMarshal(r *types.DeploymentReceipt) []byte {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		panic(err)
	}
	return b
}
