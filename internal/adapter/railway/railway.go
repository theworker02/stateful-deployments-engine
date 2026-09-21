package railway

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

// RecoveryMode documents which portability path an operator is using.
type RecoveryMode = archive.PortabilityKind

const (
	ModeRailwayNative  = archive.RailwayNative
	ModeEngineProvided = archive.EngineProvided
)

// Adapter talks to Railway via the public GraphQL API when SDE_RAILWAY_LIVE=1.
type Adapter struct {
	cfg    Config
	client *Client

	mu       sync.Mutex
	slots    map[string]*types.DeploymentSlot
	barrier  map[string]bool
	standby  string
	activeID string
}

// ErrNotWired is returned when live mode is disabled (SDE_RAILWAY_LIVE!=1).
var ErrNotWired = fmt.Errorf("railway adapter: live mode disabled (set SDE_RAILWAY_LIVE=1 with RAILWAY_TOKEN)")

func (a *Adapter) Name() string { return "railway" }

func (a *Adapter) live() bool { return os.Getenv("SDE_RAILWAY_LIVE") == "1" }

// LiveEnabled reports whether SDE_RAILWAY_LIVE=1.
func LiveEnabled() bool { return os.Getenv("SDE_RAILWAY_LIVE") == "1" }

func (a *Adapter) requireLive() error {
	if !a.live() {
		return ErrNotWired
	}
	if a.client == nil || a.client.Token == "" {
		return fmt.Errorf("railway: live mode requires RAILWAY_TOKEN (or RAILWAY_API_TOKEN)")
	}
	return nil
}

// ValidateAPI probes Railway GraphQL with the configured token (live mode only).
func (a *Adapter) ValidateAPI(ctx context.Context) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	return a.client.ValidateToken(ctx)
}

func (a *Adapter) ensureMaps() {
	if a.slots == nil {
		a.slots = map[string]*types.DeploymentSlot{}
	}
	if a.barrier == nil {
		a.barrier = map[string]bool{}
	}
	if a.activeID == "" {
		a.activeID = a.cfg.ActiveServiceID
	}
}

func (a *Adapter) Capabilities() []types.StorageCapability {
	return StorageCapabilities()
}

func (a *Adapter) NativeBackupLimitations() []string {
	return []string{
		"RAILWAY_NATIVE restores are constrained to the same project/environment model",
		"Deleting/wiping a volume may delete associated native volume backups",
		"Native backups are coupled to Railway volume lifecycle",
		"Cross-environment portability requires ENGINE_PROVIDED Portable State Archive",
		"SDE does not claim Railway endorsement of ENGINE_PROVIDED archives",
	}
}

func (a *Adapter) PreferredPortability() RecoveryMode { return ModeEngineProvided }

func (a *Adapter) statePathFor(role string) string {
	switch role {
	case "active":
		if p := os.Getenv("RAILWAY_ACTIVE_STATE_PATH"); p != "" {
			return p
		}
	case "shadow":
		if p := os.Getenv("RAILWAY_SHADOW_STATE_PATH"); p != "" {
			return p
		}
	}
	base := os.Getenv("RAILWAY_STATE_STAGING")
	if base == "" {
		base = filepath.Join(os.TempDir(), "sde-railway-staging")
	}
	return filepath.Join(base, role)
}

func (a *Adapter) CreateShadow(ctx context.Context, req adapter.CreateShadowRequest) (*types.DeploymentSlot, error) {
	return a.CreateCandidate(ctx, req)
}

func (a *Adapter) CreateCandidate(ctx context.Context, req adapter.CreateShadowRequest) (*types.DeploymentSlot, error) {
	if err := a.requireLive(); err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureMaps()

	name := fmt.Sprintf("sde-shadow-%d", time.Now().UnixNano())
	image := req.ImageRef
	id := a.cfg.ShadowServiceID
	var err error
	if id == "" {
		id, err = a.client.CreateService(ctx, a.cfg.ProjectID, name, image)
		if err != nil {
			return nil, fmt.Errorf("railway CreateCandidate: %w", err)
		}
		a.cfg.ShadowServiceID = id
		_, _ = a.client.CreateVolume(ctx, a.cfg.ProjectID, id, "/data", a.cfg.EnvironmentID)
	} else if _, _, err = a.client.GetService(ctx, id); err != nil {
		return nil, fmt.Errorf("railway CreateCandidate: shadow service: %w", err)
	}
	slot := &types.DeploymentSlot{
		ID:        id,
		Role:      types.RoleShadow,
		ImageRef:  image,
		StatePath: a.statePathFor("shadow"),
		Healthy:   false,
		CreatedAt: time.Now().UTC(),
	}
	_ = os.MkdirAll(slot.StatePath, 0o755)
	a.slots[id] = slot
	return slot, nil
}

func (a *Adapter) AttachState(ctx context.Context, slotID, stateRef string) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureMaps()
	if s, ok := a.slots[slotID]; ok && stateRef != "" {
		s.StatePath = stateRef
	}
	return nil
}

func (a *Adapter) StartShadow(ctx context.Context, slotID string) error {
	return a.StartCandidate(ctx, slotID)
}

func (a *Adapter) StartCandidate(ctx context.Context, slotID string) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	if a.cfg.EnvironmentID == "" {
		return fmt.Errorf("railway StartCandidate: RAILWAY_ENVIRONMENT_ID required")
	}
	if err := a.client.DeployService(ctx, slotID, a.cfg.EnvironmentID); err != nil {
		return fmt.Errorf("railway StartCandidate: %w", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureMaps()
	if s, ok := a.slots[slotID]; ok {
		s.Healthy = true
	}
	return nil
}

func (a *Adapter) HealthCheck(ctx context.Context, slotID string) error {
	st, err := a.ProbeHealth(ctx, slotID)
	if err != nil {
		return err
	}
	if !st.Healthy {
		return fmt.Errorf("railway health: slot %s status=%s", slotID, st.Detail)
	}
	return nil
}

func (a *Adapter) PrepareCutover(ctx context.Context, activeID, candidateID string) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	if _, _, err := a.client.GetService(ctx, activeID); err != nil {
		return err
	}
	_, _, err := a.client.GetService(ctx, candidateID)
	return err
}

func (a *Adapter) EstablishWriteBarrier(ctx context.Context, activeSlotID string) (func() error, error) {
	if err := a.requireLive(); err != nil {
		return nil, err
	}
	vars := map[string]string{
		"SDE_WRITE_BARRIER":    "1",
		"SDE_WRITE_BARRIER_AT": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := a.client.UpsertVariables(ctx, a.cfg.ProjectID, a.cfg.EnvironmentID, activeSlotID, vars); err != nil {
		return nil, fmt.Errorf("railway write barrier: %w", err)
	}
	a.mu.Lock()
	a.ensureMaps()
	a.barrier[activeSlotID] = true
	a.mu.Unlock()
	return func() error { return a.UnfreezeWrites(ctx, activeSlotID) }, nil
}

func (a *Adapter) FreezeWrites(ctx context.Context, slotID string) (func() error, error) {
	return a.EstablishWriteBarrier(ctx, slotID)
}

func (a *Adapter) UnfreezeWrites(ctx context.Context, slotID string) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	if err := a.client.UpsertVariables(ctx, a.cfg.ProjectID, a.cfg.EnvironmentID, slotID, map[string]string{
		"SDE_WRITE_BARRIER": "0",
	}); err != nil {
		return err
	}
	a.mu.Lock()
	delete(a.barrier, slotID)
	a.mu.Unlock()
	return nil
}

func (a *Adapter) TransferTraffic(ctx context.Context, fromSlotID, toSlotID string) error {
	return a.ExecuteDomainSwap(ctx, fromSlotID, toSlotID)
}

func (a *Adapter) PromoteShadow(ctx context.Context, shadowSlotID string) (*adapter.PromoteResult, error) {
	return a.ActivateCandidate(ctx, shadowSlotID)
}

func (a *Adapter) ActivateCandidate(ctx context.Context, candidateID string) (*adapter.PromoteResult, error) {
	if err := a.requireLive(); err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureMaps()
	prev := a.activeID
	if prev == "" {
		prev = a.cfg.ActiveServiceID
	}
	cand := a.slots[candidateID]
	if cand == nil {
		cand = &types.DeploymentSlot{
			ID: candidateID, Role: types.RoleActive, StatePath: a.statePathFor("shadow"),
			Healthy: true, CreatedAt: time.Now().UTC(),
		}
	}
	cand.Role = types.RoleActive
	var standby *types.DeploymentSlot
	if prev != "" && prev != candidateID {
		standby = &types.DeploymentSlot{
			ID: prev, Role: types.RoleStandby, StatePath: a.statePathFor("active"),
			Healthy: true, CreatedAt: time.Now().UTC(),
		}
		a.standby = prev
		a.slots[prev] = standby
	}
	a.activeID = candidateID
	a.cfg.ActiveServiceID = candidateID
	a.slots[candidateID] = cand
	cp := *cand
	return &adapter.PromoteResult{Active: &cp, Standby: standby}, nil
}

func (a *Adapter) DeactivatePrevious(ctx context.Context, previousID string) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	return a.client.UpsertVariables(ctx, a.cfg.ProjectID, a.cfg.EnvironmentID, previousID, map[string]string{
		"SDE_ROLE": "standby",
	})
}

func (a *Adapter) Rollback(ctx context.Context, req adapter.RollbackRequest) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	target := req.StandbySlotID
	if target == "" {
		a.mu.Lock()
		target = a.standby
		a.mu.Unlock()
	}
	if target == "" {
		return fmt.Errorf("railway Rollback: no standby slot")
	}
	if len(req.TargetImage) > 7 && req.TargetImage[:7] == "deploy:" {
		return a.client.RollbackDeployment(ctx, req.TargetImage[7:])
	}
	if err := a.client.DeployService(ctx, target, a.cfg.EnvironmentID); err != nil {
		return err
	}
	a.mu.Lock()
	a.activeID = target
	a.cfg.ActiveServiceID = target
	a.mu.Unlock()
	return nil
}

func (a *Adapter) DestroySlot(ctx context.Context, slotID string) error {
	return a.DestroyCandidate(ctx, slotID)
}

func (a *Adapter) DestroyCandidate(ctx context.Context, slotID string) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	if slotID == a.cfg.ActiveServiceID {
		return fmt.Errorf("railway: refuse to destroy active service")
	}
	if err := a.client.DeleteService(ctx, slotID); err != nil {
		return err
	}
	a.mu.Lock()
	delete(a.slots, slotID)
	a.mu.Unlock()
	return nil
}

func (a *Adapter) ActiveSlot(ctx context.Context) (*types.DeploymentSlot, error) {
	if err := a.requireLive(); err != nil {
		return nil, err
	}
	id := a.cfg.ActiveServiceID
	if id == "" {
		return nil, fmt.Errorf("railway ActiveSlot: RAILWAY_SERVICE_ID required")
	}
	_, name, err := a.client.GetService(ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.DeploymentSlot{
		ID: id, Role: types.RoleActive, ImageRef: name,
		StatePath: a.statePathFor("active"), Healthy: true, CreatedAt: time.Now().UTC(),
	}, nil
}

func (a *Adapter) Checkpoint(ctx context.Context, root string, epoch uint64) (*types.Checkpoint, error) {
	if err := a.requireLive(); err != nil {
		return nil, err
	}
	if root == "" {
		root = a.statePathFor("active")
	}
	return verifier.BuildCheckpoint(root, epoch)
}

func (a *Adapter) JournalDir() string {
	if d := os.Getenv("RAILWAY_JOURNAL_DIR"); d != "" {
		return d
	}
	return filepath.Join(filepath.Dir(a.statePathFor("active")), "journal")
}

func (a *Adapter) Sync(ctx context.Context, srcRoot, dstRoot string, fromSeq uint64) (uint64, uint64, error) {
	if err := a.requireLive(); err != nil {
		return 0, fromSeq, err
	}
	if srcRoot == "" {
		srcRoot = a.statePathFor("active")
	}
	if dstRoot == "" {
		dstRoot = a.statePathFor("shadow")
	}
	_ = os.MkdirAll(dstRoot, 0o755)
	n, err := copyMissing(srcRoot, dstRoot)
	return n, fromSeq, err
}

func (a *Adapter) Verify(ctx context.Context, activeRoot, candidateRoot string, epoch uint64, level types.VerifyLevel) (*types.VerificationReceipt, error) {
	if err := a.requireLive(); err != nil {
		return nil, err
	}
	if activeRoot == "" {
		activeRoot = a.statePathFor("active")
	}
	if candidateRoot == "" {
		candidateRoot = a.statePathFor("shadow")
	}
	r, _, _, err := verifier.CompareRoots(activeRoot, candidateRoot, epoch)
	if err != nil {
		return nil, err
	}
	outcome := types.VerifyVerified
	if !r.OK {
		outcome = types.VerifyFailed
	}
	return &types.VerificationReceipt{Level: level, Outcome: outcome, Epoch: epoch}, nil
}

func copyMissing(src, dst string) (uint64, error) {
	var n uint64
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, err := os.Stat(target); err == nil {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, b, 0o644); err != nil {
			return err
		}
		n++
		return nil
	})
	return n, err
}

const StrategyDoc = `
Railway integration (live GraphQL when SDE_RAILWAY_LIVE=1)
==========================================================
Endpoint: https://backboard.railway.com/graphql/v2
Dual-service / dual-volume; write barrier via SDE_WRITE_BARRIER variable.
ENGINE_PROVIDED PSA remains the cross-environment DR path.
`
