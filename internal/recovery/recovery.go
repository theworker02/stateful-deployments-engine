// Package recovery implements restore orchestration, receipts, and RecoveryTargetAdapter.
package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Mode distinguishes engine-provided vs platform-native recovery paths.
type Mode string

const (
	ModeEngineProvided Mode = "ENGINE_PROVIDED"
	ModeRailwayNative  Mode = "RAILWAY_NATIVE"
)

// TargetAdapter restores into a deployment environment after capability negotiation.
type TargetAdapter interface {
	Name() string
	Capabilities() []types.StorageCapability
	PrepareRestore(ctx context.Context) (statePath string, err error)
	FinalizeRestore(ctx context.Context, statePath string, epoch uint64) error
}

// LocalTarget restores into a filesystem directory (ENGINE_PROVIDED).
type LocalTarget struct {
	Root string
}

func (l *LocalTarget) Name() string { return "local-recovery" }

func (l *LocalTarget) Capabilities() []types.StorageCapability {
	return []types.StorageCapability{types.CapChecksum, types.CapBlockRead, types.CapBlockWrite, types.CapAtomicRename}
}

func (l *LocalTarget) PrepareRestore(ctx context.Context) (string, error) {
	p := filepath.Join(l.Root, "restored-state")
	return p, os.MkdirAll(p, 0o755)
}

func (l *LocalTarget) FinalizeRestore(ctx context.Context, statePath string, epoch uint64) error {
	meta := map[string]interface{}{"epoch": epoch, "state_path": statePath, "mode": ModeEngineProvided}
	b, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(filepath.Join(l.Root, "restore-meta.json"), b, 0o644)
}

// Ensure LocalTarget strategies use negotiated caps.
func SelectRestoreStrategies(caps []types.StorageCapability) []string {
	return adapter.SelectStrategies(caps)
}

// Receipt is RECOVERY_RECEIPT.json.
type Receipt struct {
	ReceiptID          string    `json:"receipt_id"`
	CreatedAt          time.Time `json:"created_at"`
	Mode               Mode      `json:"mode"`
	ArchiveID          string    `json:"archive_id"`
	SourceDigest       string    `json:"source_digest"`
	RestoredDigest     string    `json:"restored_digest"`
	Epoch              uint64    `json:"epoch"`
	JournalPosition    uint64    `json:"journal_position"`
	TargetName         string    `json:"target_name"`
	TargetPath         string    `json:"target_path"`
	SourceDestroyed    bool      `json:"source_destroyed"`
	VerifyOutcome      string    `json:"verify_outcome"`
	DurationMs         float64   `json:"duration_ms"`
	CapabilitiesUsed   []types.StorageCapability `json:"capabilities_used"`
	Strategies         []string  `json:"strategies"`
	EvidenceHash       string    `json:"evidence_hash"`
	HumanSummary       string    `json:"human_summary"`
	Error              string    `json:"error,omitempty"`
}

// Restore imports archive into target and writes a recovery receipt.
func Restore(ctx context.Context, archiveDir string, target TargetAdapter, receiptDir string, sourceDestroyed bool) (*Receipt, error) {
	start := time.Now()
	m, err := archive.Inspect(archiveDir)
	if err != nil {
		return nil, err
	}
	caps := target.Capabilities()
	r := &Receipt{
		ReceiptID:        fmt.Sprintf("recv-%d", start.UnixNano()),
		CreatedAt:        start.UTC(),
		Mode:             ModeEngineProvided,
		ArchiveID:        m.ArchiveID,
		SourceDigest:     m.RootDigest,
		Epoch:            m.SourceEpoch,
		JournalPosition:  m.JournalPosition,
		TargetName:       target.Name(),
		SourceDestroyed:  sourceDestroyed,
		CapabilitiesUsed: caps,
		Strategies:       SelectRestoreStrategies(caps),
	}
	statePath, err := target.PrepareRestore(ctx)
	if err != nil {
		r.Error = err.Error()
		return seal(r, receiptDir), err
	}
	r.TargetPath = statePath
	if _, err := archive.Import(archiveDir, statePath); err != nil {
		r.Error = err.Error()
		return seal(r, receiptDir), err
	}
	// Verify by re-export digest
	tmp := filepath.Join(filepath.Dir(statePath), ".verify-export")
	_ = os.RemoveAll(tmp)
	m2, _, err := archive.Export(statePath, tmp, m.SourceEpoch, m.JournalPosition, "")
	_ = os.RemoveAll(tmp)
	if err != nil {
		r.Error = err.Error()
		r.VerifyOutcome = string(types.VerifyUnknown)
		return seal(r, receiptDir), err
	}
	r.RestoredDigest = m2.RootDigest
	if m2.RootDigest != m.RootDigest {
		r.VerifyOutcome = string(types.VerifyFailed)
		r.Error = "restored digest mismatch"
		return seal(r, receiptDir), fmt.Errorf("%s", r.Error)
	}
	r.VerifyOutcome = string(types.VerifyVerified)
	if err := target.FinalizeRestore(ctx, statePath, m.SourceEpoch); err != nil {
		r.Error = err.Error()
		return seal(r, receiptDir), err
	}
	r.DurationMs = float64(time.Since(start).Microseconds()) / 1000
	r.HumanSummary = fmt.Sprintf("ENGINE_PROVIDED restore archive=%s epoch=%d digest_ok=true target=%s",
		m.ArchiveID, m.SourceEpoch, target.Name())
	return seal(r, receiptDir), nil
}

func seal(r *Receipt, dir string) *Receipt {
	r.EvidenceHash = ""
	b, _ := json.Marshal(r)
	sum := sha256.Sum256(b)
	r.EvidenceHash = hex.EncodeToString(sum[:])
	if dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		out, _ := json.MarshalIndent(r, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "RECOVERY_RECEIPT.json"), out, 0o644)
		_ = os.WriteFile(filepath.Join(dir, r.ReceiptID+".json"), out, 0o644)
	}
	return r
}
