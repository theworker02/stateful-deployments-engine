// Package receipt builds immutable deployment evidence records.
package receipt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Build constructs a DeploymentReceipt and seals it with EvidenceHash.
func Build(r *types.DeploymentReceipt) *types.DeploymentReceipt {
	if r.ReceiptID == "" {
		r.ReceiptID = fmt.Sprintf("rcpt-%d", time.Now().UnixNano())
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	r.CutoverWritePauseMs = r.ActualPauseMs
	if r.HumanSummary == "" {
		r.HumanSummary = fmt.Sprintf(
			"deploy %s %s→%s phase=%s pause_ms=%.2f (predicted %.2f) mutations=%d final_delta=%d dry_run=%v",
			r.DeployID, r.FromImage, r.ToImage, r.FinalPhase, r.ActualPauseMs, r.PredictedPauseMs,
			r.MutationsSynced, r.FinalDeltaWrites, r.DryRun,
		)
	}
	r.EvidenceHash = ""
	b, _ := json.Marshal(r)
	sum := sha256.Sum256(b)
	r.EvidenceHash = hex.EncodeToString(sum[:])
	return r
}

// WriteJSON writes the sealed receipt to dir/<receipt_id>.json.
func WriteJSON(dir string, r *types.DeploymentReceipt) (string, error) {
	Build(r)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, r.ReceiptID+".json")
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, b, 0o644)
}

// WriteHuman writes a text summary alongside the JSON receipt.
func WriteHuman(dir string, r *types.DeploymentReceipt) (string, error) {
	Build(r)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, r.ReceiptID+".txt")
	body := r.HumanSummary + "\nEvidenceHash: " + r.EvidenceHash + "\n"
	return path, os.WriteFile(path, []byte(body), 0o644)
}
