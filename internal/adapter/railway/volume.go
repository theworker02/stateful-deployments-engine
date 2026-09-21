package railway

import (
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// VolumeModel documents Railway volume constraints relevant to SDE.
type VolumeModel struct {
	MaxVolumesPerService int
	ConcurrentMounts     int
	SupportsCoW          bool
	SupportsSnapshots    bool
	SupportsReplicas     bool
	Notes                []string
}

// DefaultVolumeModel reflects Railway's documented single-volume / no-replica model.
func DefaultVolumeModel() VolumeModel {
	return VolumeModel{
		MaxVolumesPerService: 1,
		ConcurrentMounts:     1,
		SupportsCoW:          false,
		SupportsSnapshots:    false,
		SupportsReplicas:     false,
		Notes: []string{
			"one volume per service",
			"replicas cannot be used with volumes",
			"SDE uses dual-service dual-volume sync instead of shared mounts",
		},
	}
}

// StorageCapabilities returns the negotiated capability matrix for Railway volumes.
func StorageCapabilities() []types.StorageCapability {
	return []types.StorageCapability{types.CapChecksum, types.CapFreeze}
}
