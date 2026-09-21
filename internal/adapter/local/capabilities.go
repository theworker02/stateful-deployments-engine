package local

import (
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// StorageCapabilities reports what the local filesystem adapter can negotiate.
func StorageCapabilities() []types.StorageCapability {
	return []types.StorageCapability{
		types.CapChecksum,
		types.CapFreeze,
		types.CapAtomicRename,
		types.CapBlockRead,
		types.CapBlockWrite,
		types.CapChangeTracking,
	}
}
