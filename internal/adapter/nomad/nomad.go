// Package nomad stubs a HashiCorp Nomad PlatformAdapter.
package nomad

import (
	"context"
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// ErrNotWired until the Nomad HTTP API client is connected.
var ErrNotWired = fmt.Errorf("nomad adapter: HTTP API not wired")

// Config targets a Nomad job/group.
type Config struct {
	Address   string
	Token     string
	Namespace string
	JobID     string
	Group     string
}

// Adapter implements PlatformAdapter (stub).
type Adapter struct{ cfg Config }

func New(cfg Config) *Adapter { return &Adapter{cfg: cfg} }

func (a *Adapter) Name() string { return "nomad" }

func (a *Adapter) Capabilities() []types.StorageCapability {
	return []types.StorageCapability{types.CapChecksum, types.CapFreeze, types.CapBlockRead}
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

// HostVolumePlan maps Nomad host_volume / CSI to SDE slots.
type HostVolumePlan struct {
	ActiveVolume  string   `json:"active_volume"`
	ShadowVolume  string   `json:"shadow_volume"`
	Plugin        string   `json:"plugin"`
	Notes         []string `json:"notes"`
}

// DefaultHostVolumePlan returns a dual host_volume layout.
func DefaultHostVolumePlan(job string) HostVolumePlan {
	return HostVolumePlan{
		ActiveVolume: job + "-data-a",
		ShadowVolume: job + "-data-b",
		Plugin:       "host_volume|csi",
		Notes:        []string{"canary + journal sync; avoid single-volume in-place replace"},
	}
}
