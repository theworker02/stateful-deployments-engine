// Package adapter defines platform-agnostic PlatformAdapter and StorageAdapter contracts.
//
// Core never imports Railway/Fly/K8s specifics — adapters translate platform
// primitives into slots, health, traffic cutover, and storage operations.
package adapter

import (
	"context"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Platform is the Phase-1 compatibility surface (still used by demos).
type Platform interface {
	Name() string
	CreateShadow(ctx context.Context, req CreateShadowRequest) (*types.DeploymentSlot, error)
	StartShadow(ctx context.Context, slotID string) error
	HealthCheck(ctx context.Context, slotID string) error
	EstablishWriteBarrier(ctx context.Context, activeSlotID string) (release func() error, err error)
	TransferTraffic(ctx context.Context, fromSlotID, toSlotID string) error
	PromoteShadow(ctx context.Context, shadowSlotID string) (*PromoteResult, error)
	Rollback(ctx context.Context, req RollbackRequest) error
	DestroySlot(ctx context.Context, slotID string) error
	ActiveSlot(ctx context.Context) (*types.DeploymentSlot, error)
}

// PlatformAdapter is the Phase-2/3 platform contract.
type PlatformAdapter interface {
	Name() string
	CreateCandidate(ctx context.Context, req CreateShadowRequest) (*types.DeploymentSlot, error)
	AttachState(ctx context.Context, slotID, stateRef string) error
	StartCandidate(ctx context.Context, slotID string) error
	HealthCheck(ctx context.Context, slotID string) error
	PrepareCutover(ctx context.Context, activeID, candidateID string) error
	ActivateCandidate(ctx context.Context, candidateID string) (*PromoteResult, error)
	DeactivatePrevious(ctx context.Context, previousID string) error
	DestroyCandidate(ctx context.Context, slotID string) error
	ActiveSlot(ctx context.Context) (*types.DeploymentSlot, error)
}

// StorageAdapter is the Phase-2/3 storage contract (platform-independent).
type StorageAdapter interface {
	Name() string
	Capabilities() []types.StorageCapability
	Checkpoint(ctx context.Context, root string, epoch uint64) (*types.Checkpoint, error)
	JournalDir() string
	Sync(ctx context.Context, srcRoot, dstRoot string, fromSeq uint64) (applied uint64, nextSeq uint64, err error)
	Verify(ctx context.Context, activeRoot, candidateRoot string, epoch uint64, level types.VerifyLevel) (*types.VerificationReceipt, error)
	FreezeWrites(ctx context.Context, slotID string) (unfreeze func() error, err error)
	UnfreezeWrites(ctx context.Context, slotID string) error
	Rollback(ctx context.Context, req RollbackRequest) error
}

// CreateShadowRequest describes the candidate to spin up.
type CreateShadowRequest struct {
	ImageRef     string
	ActiveSlotID string
	Checkpoint   *types.Checkpoint
	Labels       map[string]string
}

// PromoteResult is returned after shadow becomes active.
type PromoteResult struct {
	Active  *types.DeploymentSlot
	Standby *types.DeploymentSlot
}

// RollbackRequest selects which epoch/image to restore.
type RollbackRequest struct {
	TargetEpoch   uint64
	TargetImage   string
	StandbySlotID string
}

// PlatformBridge adapts a legacy Platform to PlatformAdapter.
type PlatformBridge struct {
	Inner Platform
}

func (b PlatformBridge) Name() string { return b.Inner.Name() }

func (b PlatformBridge) CreateCandidate(ctx context.Context, req CreateShadowRequest) (*types.DeploymentSlot, error) {
	return b.Inner.CreateShadow(ctx, req)
}

func (b PlatformBridge) AttachState(ctx context.Context, slotID, stateRef string) error {
	return nil // local/fs adapters attach at create time
}

func (b PlatformBridge) StartCandidate(ctx context.Context, slotID string) error {
	return b.Inner.StartShadow(ctx, slotID)
}

func (b PlatformBridge) HealthCheck(ctx context.Context, slotID string) error {
	return b.Inner.HealthCheck(ctx, slotID)
}

func (b PlatformBridge) PrepareCutover(ctx context.Context, activeID, candidateID string) error {
	return nil
}

func (b PlatformBridge) ActivateCandidate(ctx context.Context, candidateID string) (*PromoteResult, error) {
	if err := b.Inner.TransferTraffic(ctx, "", candidateID); err != nil {
		return nil, err
	}
	return b.Inner.PromoteShadow(ctx, candidateID)
}

func (b PlatformBridge) DeactivatePrevious(ctx context.Context, previousID string) error {
	return nil
}

func (b PlatformBridge) DestroyCandidate(ctx context.Context, slotID string) error {
	return b.Inner.DestroySlot(ctx, slotID)
}

func (b PlatformBridge) ActiveSlot(ctx context.Context) (*types.DeploymentSlot, error) {
	return b.Inner.ActiveSlot(ctx)
}

// HasCapability reports whether caps contains c.
func HasCapability(caps []types.StorageCapability, c types.StorageCapability) bool {
	for _, x := range caps {
		if x == c {
			return true
		}
	}
	return false
}

// SelectStrategies chooses migration strategies from negotiated capabilities.
func SelectStrategies(caps []types.StorageCapability) []string {
	var out []string
	if HasCapability(caps, types.CapSnapshot) || HasCapability(caps, types.CapIncrementalSnapshot) {
		out = append(out, "snapshot-seed")
	} else {
		out = append(out, "baseline-copy")
	}
	if HasCapability(caps, types.CapChangeTracking) || HasCapability(caps, types.CapBlockRead) {
		out = append(out, "block-differential")
	} else {
		out = append(out, "journal-replay")
	}
	if HasCapability(caps, types.CapFreeze) {
		out = append(out, "write-barrier-freeze")
	}
	if HasCapability(caps, types.CapChecksum) {
		out = append(out, "checksum-verify")
	}
	if HasCapability(caps, types.CapAtomicRename) {
		out = append(out, "atomic-pointer-cutover")
	}
	if HasCapability(caps, types.CapCopyOnWrite) {
		out = append(out, "cow-candidate")
	}
	return out
}
