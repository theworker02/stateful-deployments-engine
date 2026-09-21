// Package fence provides monotonic epochs and fencing tokens so a stale
// coordinator cannot mutate state belonging to a newer epoch/lease.
package fence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Manager persists epoch + fencing token under the SDE root.
type Manager struct {
	mu   sync.Mutex
	path string
	cur  types.FenceToken
}

type persisted struct {
	Epoch    uint64 `json:"epoch"`
	Token    uint64 `json:"token"`
	HolderID string `json:"holder_id"`
	IssuedAt time.Time `json:"issued_at"`
}

// Open loads or initializes fencing state at root/fence.json.
func Open(root, holderID string) (*Manager, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	m := &Manager{path: filepath.Join(root, "fence.json")}
	b, err := os.ReadFile(m.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		m.cur = types.FenceToken{
			Epoch:    1,
			Token:    1,
			HolderID: holderID,
			IssuedAt: time.Now().UTC(),
		}
		return m, m.save()
	}
	var p persisted
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	m.cur = types.FenceToken{
		Epoch:    p.Epoch,
		Token:    p.Token,
		HolderID: p.HolderID,
		IssuedAt: p.IssuedAt,
	}
	return m, nil
}

func (m *Manager) save() error {
	p := persisted{
		Epoch:    m.cur.Epoch,
		Token:    m.cur.Token,
		HolderID: m.cur.HolderID,
		IssuedAt: m.cur.IssuedAt,
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, b, 0o644)
}

// Current returns a copy of the active fence.
func (m *Manager) Current() types.FenceToken {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cur
}

// Issue bumps the fencing token for a new coordinator lease.
func (m *Manager) Issue(holderID string) (types.FenceToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cur.Token++
	m.cur.HolderID = holderID
	m.cur.IssuedAt = time.Now().UTC()
	if err := m.save(); err != nil {
		return types.FenceToken{}, err
	}
	return m.cur, nil
}

// AdvanceEpoch increments epoch and issues a fresh fence token.
func (m *Manager) AdvanceEpoch(holderID string) (types.FenceToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cur.Epoch++
	m.cur.Token++
	m.cur.HolderID = holderID
	m.cur.IssuedAt = time.Now().UTC()
	if err := m.save(); err != nil {
		return types.FenceToken{}, err
	}
	return m.cur, nil
}

// Check rejects stale tokens / epochs.
func (m *Manager) Check(epoch, token uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if epoch < m.cur.Epoch {
		return fmt.Errorf("STALE_EPOCH_MUST_NOT_MODIFY_NEWER_STATE: epoch %d < %d", epoch, m.cur.Epoch)
	}
	if epoch == m.cur.Epoch && token < m.cur.Token {
		return fmt.Errorf("stale fence token %d < %d for epoch %d", token, m.cur.Token, epoch)
	}
	return nil
}

// SetEpoch pins epoch (rollback) without lowering the fence token.
func (m *Manager) SetEpoch(epoch uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cur.Epoch = epoch
	m.cur.Token++
	m.cur.IssuedAt = time.Now().UTC()
	return m.save()
}
