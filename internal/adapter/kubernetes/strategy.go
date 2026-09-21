package kubernetes

// CutoverStrategy documents how SDE would cut over without claiming SharedVolume.
type CutoverStrategy struct {
	Mode  string   `json:"mode"`
	Steps []string `json:"steps"`
	Notes []string `json:"notes"`
}

// DefaultCutoverStrategy prefers PVC clone + Service selector flip.
func DefaultCutoverStrategy() CutoverStrategy {
	return CutoverStrategy{
		Mode: "dual-sts-or-bluegreen-service",
		Steps: []string{
			"create candidate StatefulSet with cloned/empty PVC",
			"journal/block sync until safety gate allows",
			"flip Service selector or Ingress backend",
			"retain prior STS as standby for rollback",
		},
		Notes: []string{
			"SDE does not replace CSI; it coordinates app+state epochs",
			"RWO volumes still cannot multi-attach — dual PVC is required",
		},
	}
}
