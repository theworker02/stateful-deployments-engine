// Package docker provides a Docker Compose / engine PlatformAdapter stub.
package docker

import (
	"context"
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// ErrNotWired until Docker Engine API client is connected.
var ErrNotWired = fmt.Errorf("docker adapter: engine API not wired")

// Config targets a compose project or container set.
type Config struct {
	Host           string // DOCKER_HOST
	ComposeProject string
	ActiveService  string
	ShadowService  string
	VolumeName     string
}

// Adapter implements PlatformAdapter (stub).
type Adapter struct{ cfg Config }

func New(cfg Config) *Adapter { return &Adapter{cfg: cfg} }

func (a *Adapter) Name() string { return "docker" }

func (a *Adapter) Capabilities() []types.StorageCapability {
	return []types.StorageCapability{
		types.CapChecksum, types.CapFreeze, types.CapAtomicRename, types.CapBlockRead, types.CapBlockWrite,
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
