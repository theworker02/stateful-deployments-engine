package docker

// VolumeBind describes how bind mounts / named volumes map to SDE slots.
type VolumeBind struct {
	Name       string `json:"name"`
	MountPath  string `json:"mount_path"`
	ReadWrite  bool   `json:"read_write"`
	SlotRole   string `json:"slot_role"`
}

// DualVolumeLayout is the local-compose equivalent of Railway dual-service.
func DualVolumeLayout(project string) []VolumeBind {
	return []VolumeBind{
		{Name: project + "_active_data", MountPath: "/data", ReadWrite: true, SlotRole: "active"},
		{Name: project + "_shadow_data", MountPath: "/data", ReadWrite: true, SlotRole: "shadow"},
	}
}
