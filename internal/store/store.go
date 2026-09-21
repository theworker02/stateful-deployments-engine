// Package store persists deployment FSM state so cutovers survive coordinator restart.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// DeploymentState is the durable FSM record under .sde/deployments/<id>/state.json.
type DeploymentState struct {
	DeployID           string                  `json:"deploy_id"`
	Phase              types.DeployPhase       `json:"phase"`
	Epoch              uint64                  `json:"epoch"`
	FenceToken         uint64                  `json:"fence_token"`
	FromImage          string                  `json:"from_image"`
	ToImage            string                  `json:"to_image"`
	ActiveSlotID       string                  `json:"active_slot_id"`
	CandidateSlotID    string                  `json:"candidate_slot_id"`
	StandbySlotID      string                  `json:"standby_slot_id,omitempty"`
	JournalCursor      uint64                  `json:"journal_cursor"`
	CheckpointSeq      uint64                  `json:"checkpoint_seq"`
	MutationsSynced    uint64                  `json:"mutations_synced"`
	FinalDeltaWrites   uint64                  `json:"final_delta_writes"`
	CutoverWritePauseMs float64                `json:"cutover_write_pause_ms"`
	PredictedPauseMs   float64                 `json:"predicted_pause_ms"`
	ObservationMs      float64                 `json:"observation_ms"`
	MaxWritePauseMs    float64                 `json:"max_write_pause_ms"`
	DryRun             bool                    `json:"dry_run"`
	WritesAfterCutover uint64                  `json:"writes_after_cutover"`
	CutoverJournalSeq  uint64                  `json:"cutover_journal_seq"`
	SyncLag            *types.SyncLagMetrics   `json:"sync_lag,omitempty"`
	Verification       *types.VerificationReceipt `json:"verification,omitempty"`
	Manifest           *types.StateManifest    `json:"manifest,omitempty"`
	SafetyGate         *types.SafetyGateResult `json:"safety_gate,omitempty"`
	Error              string                  `json:"error,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// Store manages deployment state directories under root/deployments.
type Store struct {
	root string
}

// New creates a store rooted at sdeRoot.
func New(sdeRoot string) *Store {
	return &Store{root: filepath.Join(sdeRoot, "deployments")}
}

func (s *Store) dir(id string) string {
	return filepath.Join(s.root, id)
}

func (s *Store) path(id string) string {
	return filepath.Join(s.dir(id), "state.json")
}

// Save atomically writes deployment state.
func (s *Store) Save(st *DeploymentState) error {
	if st.DeployID == "" {
		return fmt.Errorf("deploy_id required")
	}
	if err := os.MkdirAll(s.dir(st.DeployID), 0o755); err != nil {
		return err
	}
	st.UpdatedAt = time.Now().UTC()
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path(st.DeployID) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(st.DeployID))
}

// Load reads a deployment state.
func (s *Store) Load(id string) (*DeploymentState, error) {
	b, err := os.ReadFile(s.path(id))
	if err != nil {
		return nil, err
	}
	var st DeploymentState
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// List returns deploy IDs with persisted state.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(s.path(e.Name())); err == nil {
				ids = append(ids, e.Name())
			}
		}
	}
	return ids, nil
}

// LatestNonTerminal returns the most recently updated non-terminal deploy, if any.
func (s *Store) LatestNonTerminal() (*DeploymentState, error) {
	ids, err := s.List()
	if err != nil {
		return nil, err
	}
	var best *DeploymentState
	for _, id := range ids {
		st, err := s.Load(id)
		if err != nil {
			continue
		}
		if st.Phase.IsTerminal() {
			continue
		}
		if best == nil || st.UpdatedAt.After(best.UpdatedAt) {
			best = st
		}
	}
	return best, nil
}

// Dir returns the on-disk directory for a deployment.
func (s *Store) Dir(id string) string { return s.dir(id) }
