// Package kubernetes stubs a Kubernetes StatefulSet / PVC PlatformAdapter.
package kubernetes

import (
	"context"
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// ErrNotWired until the Kubernetes client-go integration is connected.
var ErrNotWired = fmt.Errorf("kubernetes adapter: client-go not wired")

// Config selects cluster targeting.
type Config struct {
	KubeconfigPath string
	Namespace      string
	StatefulSet    string
	ServiceName    string
}

// Adapter implements PlatformAdapter (stub).
type Adapter struct{ cfg Config }

func New(cfg Config) *Adapter { return &Adapter{cfg: cfg} }

func (a *Adapter) Name() string { return "kubernetes" }

func (a *Adapter) Capabilities() []types.StorageCapability {
	return []types.StorageCapability{
		types.CapChecksum,
		types.CapFreeze,
		types.CapSnapshot, // CSI snapshots when available
		types.CapAtomicRename,
	}
}

func (a *Adapter) CreateCandidate(ctx context.Context, req adapter.CreateShadowRequest) (*types.DeploymentSlot, error) {
	return nil, ErrNotWired
}
func (a *Adapter) AttachState(ctx context.Context, slotID, stateRef string) error { return ErrNotWired }
func (a *Adapter) StartCandidate(ctx context.Context, slotID string) error       { return ErrNotWired }
func (a *Adapter) HealthCheck(ctx context.Context, slotID string) error          { return ErrNotWired }
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

var _ adapter.PlatformAdapter = (*Adapter)(nil)
