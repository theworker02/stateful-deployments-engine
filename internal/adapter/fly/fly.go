// Package fly is the Fly.io PlatformAdapter stub with a real capability contract.
package fly

import (
	"context"
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// ErrNotWired is returned until the Machines / Volumes API client is connected.
var ErrNotWired = fmt.Errorf("fly adapter: API client not wired")

// Config targets a Fly app and volume.
type Config struct {
	APIToken string
	AppName  string
	Region   string
	VolumeID string
}

// Adapter implements adapter.PlatformAdapter (stub).
type Adapter struct {
	cfg Config
}

// New constructs a Fly adapter.
func New(cfg Config) *Adapter { return &Adapter{cfg: cfg} }

func (a *Adapter) Name() string { return "fly" }

// Capabilities documents Fly volume features relevant to SDE.
func (a *Adapter) Capabilities() []types.StorageCapability {
	return []types.StorageCapability{
		types.CapChecksum,
		types.CapFreeze,
		types.CapSnapshot, // Fly volume snapshots exist; wiring TBD
	}
}

func (a *Adapter) CreateCandidate(ctx context.Context, req adapter.CreateShadowRequest) (*types.DeploymentSlot, error) {
	return nil, ErrNotWired
}
func (a *Adapter) AttachState(ctx context.Context, slotID, stateRef string) error {
	return ErrNotWired
}
func (a *Adapter) StartCandidate(ctx context.Context, slotID string) error { return ErrNotWired }
func (a *Adapter) HealthCheck(ctx context.Context, slotID string) error    { return ErrNotWired }
func (a *Adapter) PrepareCutover(ctx context.Context, activeID, candidateID string) error {
	return ErrNotWired
}
func (a *Adapter) ActivateCandidate(ctx context.Context, candidateID string) (*adapter.PromoteResult, error) {
	return nil, ErrNotWired
}
func (a *Adapter) DeactivatePrevious(ctx context.Context, previousID string) error { return ErrNotWired }
func (a *Adapter) DestroyCandidate(ctx context.Context, slotID string) error       { return ErrNotWired }
func (a *Adapter) ActiveSlot(ctx context.Context) (*types.DeploymentSlot, error) {
	return nil, ErrNotWired
}

// Ensure Adapter satisfies PlatformAdapter at compile time.
var _ adapter.PlatformAdapter = (*Adapter)(nil)
