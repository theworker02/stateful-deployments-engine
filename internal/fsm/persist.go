package fsm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Snapshot is a durable FSM cursor for crash recovery.
type Snapshot struct {
	DeployID     string            `json:"deploy_id"`
	Phase        types.DeployPhase `json:"phase"`
	Epoch        uint64            `json:"epoch"`
	FenceToken   uint64            `json:"fence_token"`
	ActiveSlot   string            `json:"active_slot_id"`
	CandidateSlot string           `json:"candidate_slot_id,omitempty"`
	StandbySlot  string            `json:"standby_slot_id,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at"`
	LastError    string            `json:"last_error,omitempty"`
	History      []types.DeployPhase `json:"history,omitempty"`
}

// Path returns the conventional FSM snapshot path under an SDE root.
func Path(root string) string {
	return filepath.Join(root, "fsm", "snapshot.json")
}

// Save writes snapshot atomically (write temp + rename).
func Save(root string, snap Snapshot) error {
	if snap.UpdatedAt.IsZero() {
		snap.UpdatedAt = time.Now().UTC()
	}
	dir := filepath.Join(root, "fsm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, "snapshot.json.tmp")
	final := Path(root)
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// Load reads a snapshot; returns empty Snapshot and os.ErrNotExist if missing.
func Load(root string) (Snapshot, error) {
	b, err := os.ReadFile(Path(root))
	if err != nil {
		return Snapshot{}, err
	}
	var snap Snapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

// ApplyTransition validates and advances a snapshot.
func ApplyTransition(snap *Snapshot, to types.DeployPhase) error {
	if err := ValidateTransition(snap.Phase, to); err != nil {
		return err
	}
	if snap.Phase != to {
		snap.History = append(snap.History, snap.Phase)
	}
	snap.Phase = to
	snap.UpdatedAt = time.Now().UTC()
	return nil
}
