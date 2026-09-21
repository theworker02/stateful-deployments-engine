// Package local implements filesystem-backed Platform + Storage adapters for demos and benches.
package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/replay"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// Adapter is an in-process blue/green simulator with directory-per-slot state.
type Adapter struct {
	mu       sync.Mutex
	root     string
	barrier  chan struct{}
	barrierC sync.Mutex
	writeMu  sync.RWMutex
	journal  *journal.Journal
}

type meta struct {
	ActiveID  string                           `json:"active_id"`
	StandbyID string                           `json:"standby_id"`
	Slots     map[string]*types.DeploymentSlot `json:"slots"`
}

// Open initializes or loads a local adapter at root.
func Open(root string) (*Adapter, error) {
	if err := os.MkdirAll(filepath.Join(root, "slots"), 0o755); err != nil {
		return nil, err
	}
	a := &Adapter{root: root}
	if _, err := os.Stat(a.metaPath()); os.IsNotExist(err) {
		slot := &types.DeploymentSlot{
			ID:        "slot-active-0",
			Role:      types.RoleActive,
			ImageRef:  "local/app:initial",
			StatePath: filepath.Join(root, "slots", "slot-active-0", "state"),
			Healthy:   true,
			CreatedAt: time.Now().UTC(),
		}
		if err := os.MkdirAll(slot.StatePath, 0o755); err != nil {
			return nil, err
		}
		m := &meta{
			ActiveID: slot.ID,
			Slots:    map[string]*types.DeploymentSlot{slot.ID: slot},
		}
		if err := a.saveMeta(m); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (a *Adapter) Name() string { return "local" }

// BindJournal attaches the shared mutation journal (preferred over opening a second handle).
func (a *Adapter) BindJournal(j *journal.Journal) { a.journal = j }

func (a *Adapter) ensureJournal() (*journal.Journal, error) {
	if a.journal != nil {
		return a.journal, nil
	}
	j, err := journal.Open(filepath.Join(a.root, "journal"), 1)
	if err != nil {
		return nil, err
	}
	a.journal = j
	return j, nil
}

func (a *Adapter) Journal() *journal.Journal {
	j, _ := a.ensureJournal()
	return j
}

func (a *Adapter) metaPath() string { return filepath.Join(a.root, "meta.json") }

func (a *Adapter) loadMeta() (*meta, error) {
	b, err := os.ReadFile(a.metaPath())
	if err != nil {
		return nil, err
	}
	var m meta
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Slots == nil {
		m.Slots = map[string]*types.DeploymentSlot{}
	}
	return &m, nil
}

func (a *Adapter) saveMeta(m *meta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.metaPath(), b, 0o644)
}

func (a *Adapter) ActiveSlot(ctx context.Context) (*types.DeploymentSlot, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return nil, err
	}
	s, ok := m.Slots[m.ActiveID]
	if !ok {
		return nil, fmt.Errorf("active slot %q missing", m.ActiveID)
	}
	cp := *s
	return &cp, nil
}

func (a *Adapter) CreateShadow(ctx context.Context, req adapter.CreateShadowRequest) (*types.DeploymentSlot, error) {
	return a.CreateCandidate(ctx, req)
}

func (a *Adapter) CreateCandidate(ctx context.Context, req adapter.CreateShadowRequest) (*types.DeploymentSlot, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("slot-shadow-%d", time.Now().UnixNano())
	slot := &types.DeploymentSlot{
		ID:        id,
		Role:      types.RoleShadow,
		ImageRef:  req.ImageRef,
		StatePath: filepath.Join(a.root, "slots", id, "state"),
		Healthy:   false,
		CreatedAt: time.Now().UTC(),
	}
	if err := os.MkdirAll(slot.StatePath, 0o755); err != nil {
		return nil, err
	}
	m.Slots[id] = slot
	if err := a.saveMeta(m); err != nil {
		return nil, err
	}
	cp := *slot
	return &cp, nil
}

func (a *Adapter) AttachState(ctx context.Context, slotID, stateRef string) error {
	return nil
}

func (a *Adapter) StartShadow(ctx context.Context, slotID string) error {
	return a.StartCandidate(ctx, slotID)
}

func (a *Adapter) StartCandidate(ctx context.Context, slotID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return err
	}
	s, ok := m.Slots[slotID]
	if !ok {
		return fmt.Errorf("slot %q not found", slotID)
	}
	s.Healthy = true
	return a.saveMeta(m)
}

func (a *Adapter) HealthCheck(ctx context.Context, slotID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return err
	}
	s, ok := m.Slots[slotID]
	if !ok {
		return fmt.Errorf("slot %q not found", slotID)
	}
	if !s.Healthy {
		return fmt.Errorf("slot %q unhealthy", slotID)
	}
	return nil
}

func (a *Adapter) PrepareCutover(ctx context.Context, activeID, candidateID string) error {
	return nil
}

func (a *Adapter) EstablishWriteBarrier(ctx context.Context, activeSlotID string) (func() error, error) {
	return a.FreezeWrites(ctx, activeSlotID)
}

func (a *Adapter) FreezeWrites(ctx context.Context, slotID string) (func() error, error) {
	a.writeMu.Lock()
	a.barrierC.Lock()
	if a.barrier != nil {
		a.barrierC.Unlock()
		a.writeMu.Unlock()
		return nil, fmt.Errorf("barrier already held")
	}
	a.barrier = make(chan struct{})
	held := a.barrier
	a.barrierC.Unlock()
	release := func() error {
		a.barrierC.Lock()
		defer a.barrierC.Unlock()
		if a.barrier == held {
			close(a.barrier)
			a.barrier = nil
			a.writeMu.Unlock()
		}
		return nil
	}
	return release, nil
}

func (a *Adapter) UnfreezeWrites(ctx context.Context, slotID string) error {
	return nil
}

func (a *Adapter) BeginWrite(ctx context.Context) error {
	for {
		if err := a.WaitBarrier(ctx); err != nil {
			return err
		}
		if a.writeMu.TryRLock() {
			a.barrierC.Lock()
			blocked := a.barrier != nil
			a.barrierC.Unlock()
			if blocked {
				a.writeMu.RUnlock()
				continue
			}
			return nil
		}
		if err := a.WaitBarrier(ctx); err != nil {
			return err
		}
	}
}

func (a *Adapter) EndWrite() { a.writeMu.RUnlock() }

func (a *Adapter) WaitBarrier(ctx context.Context) error {
	a.barrierC.Lock()
	ch := a.barrier
	a.barrierC.Unlock()
	if ch == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

func (a *Adapter) TransferTraffic(ctx context.Context, fromSlotID, toSlotID string) error {
	return nil
}

func (a *Adapter) PromoteShadow(ctx context.Context, shadowSlotID string) (*adapter.PromoteResult, error) {
	return a.ActivateCandidate(ctx, shadowSlotID)
}

func (a *Adapter) ActivateCandidate(ctx context.Context, candidateID string) (*adapter.PromoteResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return nil, err
	}
	shadow, ok := m.Slots[candidateID]
	if !ok {
		return nil, fmt.Errorf("shadow %q not found", candidateID)
	}
	prev := m.Slots[m.ActiveID]
	if prev != nil {
		prev.Role = types.RoleStandby
		m.StandbyID = prev.ID
	}
	shadow.Role = types.RoleActive
	shadow.Healthy = true
	m.ActiveID = shadow.ID
	if err := a.saveMeta(m); err != nil {
		return nil, err
	}
	activeCopy := *shadow
	var standbyCopy *types.DeploymentSlot
	if prev != nil {
		sc := *prev
		standbyCopy = &sc
	}
	return &adapter.PromoteResult{Active: &activeCopy, Standby: standbyCopy}, nil
}

func (a *Adapter) DeactivatePrevious(ctx context.Context, previousID string) error {
	return nil
}

func (a *Adapter) Rollback(ctx context.Context, req adapter.RollbackRequest) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return err
	}
	standbyID := req.StandbySlotID
	if standbyID == "" {
		standbyID = m.StandbyID
	}
	standby, ok := m.Slots[standbyID]
	if !ok {
		return fmt.Errorf("standby slot %q not found", standbyID)
	}
	cur := m.Slots[m.ActiveID]
	if cur != nil {
		cur.Role = types.RoleShadow
	}
	standby.Role = types.RoleActive
	if req.TargetImage != "" {
		standby.ImageRef = req.TargetImage
	}
	m.ActiveID = standby.ID
	m.StandbyID = ""
	if cur != nil {
		m.StandbyID = cur.ID
	}
	return a.saveMeta(m)
}

func (a *Adapter) DestroySlot(ctx context.Context, slotID string) error {
	return a.DestroyCandidate(ctx, slotID)
}

func (a *Adapter) DestroyCandidate(ctx context.Context, slotID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return err
	}
	if slotID == m.ActiveID {
		return fmt.Errorf("cannot destroy active slot")
	}
	delete(m.Slots, slotID)
	_ = os.RemoveAll(filepath.Join(a.root, "slots", slotID))
	return a.saveMeta(m)
}

// --- StorageAdapter ---

func (a *Adapter) Capabilities() []types.StorageCapability {
	return StorageCapabilities()
}

func (a *Adapter) Checkpoint(ctx context.Context, root string, epoch uint64) (*types.Checkpoint, error) {
	return verifier.BuildCheckpoint(root, epoch)
}

func (a *Adapter) JournalDir() string {
	return filepath.Join(a.root, "journal")
}

func (a *Adapter) Sync(ctx context.Context, srcRoot, dstRoot string, fromSeq uint64) (uint64, uint64, error) {
	j, err := a.ensureJournal()
	if err != nil {
		return 0, fromSeq, err
	}
	batch, err := j.ReadFrom(fromSeq)
	if err != nil {
		return 0, fromSeq, err
	}
	if len(batch) == 0 {
		return 0, fromSeq, nil
	}
	n, err := replay.ApplyMutations(dstRoot, batch)
	if err != nil {
		return 0, fromSeq, err
	}
	next := batch[len(batch)-1].Seq + 1
	return n, next, nil
}

func (a *Adapter) Verify(ctx context.Context, activeRoot, candidateRoot string, epoch uint64, level types.VerifyLevel) (*types.VerificationReceipt, error) {
	return verifier.Verify(activeRoot, candidateRoot, epoch, verifier.Options{Level: level})
}

func (a *Adapter) Root() string { return a.root }

// SlotByID returns a slot copy.
func (a *Adapter) SlotByID(id string) (*types.DeploymentSlot, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return nil, err
	}
	s, ok := m.Slots[id]
	if !ok {
		return nil, fmt.Errorf("slot %q not found", id)
	}
	cp := *s
	return &cp, nil
}

// MarkUnhealthy forces a health failure (chaos).
func (a *Adapter) MarkUnhealthy(slotID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	m, err := a.loadMeta()
	if err != nil {
		return err
	}
	s, ok := m.Slots[slotID]
	if !ok {
		return fmt.Errorf("slot %q not found", slotID)
	}
	s.Healthy = false
	return a.saveMeta(m)
}
