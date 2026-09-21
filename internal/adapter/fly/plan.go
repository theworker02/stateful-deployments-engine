package fly

// MachinePlan documents how SDE would map to Fly Machines + Volumes.
type MachinePlan struct {
	ActiveMachine  string   `json:"active_machine"`
	ShadowMachine  string   `json:"shadow_machine"`
	VolumeStrategy string   `json:"volume_strategy"`
	Notes          []string `json:"notes"`
}

// DefaultMachinePlan returns the dual-machine dual-volume approach.
func DefaultMachinePlan(app string) MachinePlan {
	return MachinePlan{
		ActiveMachine:  app + "-active",
		ShadowMachine:  app + "-shadow",
		VolumeStrategy: "clone-volume-then-journal-sync",
		Notes: []string{
			"Fly volumes are region-scoped; candidate should share region with active",
			"Traffic cutover via Fly proxy / DNS — networking details owned elsewhere",
		},
	}
}
