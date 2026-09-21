// Package capability negotiates storage features between platform and planner.
package capability

import (
	"sort"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Matrix is a negotiated capability set with rationale.
type Matrix struct {
	Offered    []types.StorageCapability `json:"offered"`
	Required   []types.StorageCapability `json:"required"`
	Negotiated []types.StorageCapability `json:"negotiated"`
	Missing    []types.StorageCapability `json:"missing"`
	Strategies []string                  `json:"strategies"`
	Notes      []string                  `json:"notes,omitempty"`
}

// MinimalRequired is the floor for a transactional deploy.
func MinimalRequired() []types.StorageCapability {
	return []types.StorageCapability{types.CapChecksum, types.CapFreeze}
}

// Desired prefers richer transfer modes when available.
func Desired() []types.StorageCapability {
	return []types.StorageCapability{
		types.CapSnapshot,
		types.CapIncrementalSnapshot,
		types.CapChangeTracking,
		types.CapBlockRead,
		types.CapBlockWrite,
		types.CapCopyOnWrite,
		types.CapChecksum,
		types.CapFreeze,
		types.CapAtomicRename,
	}
}

// Negotiate intersects offered with desired and checks required.
func Negotiate(offered, required []types.StorageCapability) Matrix {
	if required == nil {
		required = MinimalRequired()
	}
	set := make(map[types.StorageCapability]bool, len(offered))
	for _, c := range offered {
		set[c] = true
	}
	m := Matrix{Offered: append([]types.StorageCapability(nil), offered...), Required: append([]types.StorageCapability(nil), required...)}
	for _, c := range Desired() {
		if set[c] {
			m.Negotiated = append(m.Negotiated, c)
		}
	}
	for _, c := range required {
		if !set[c] {
			m.Missing = append(m.Missing, c)
			m.Notes = append(m.Notes, "missing required capability: "+string(c))
		}
	}
	m.Strategies = strategiesFrom(m.Negotiated)
	sort.Slice(m.Negotiated, func(i, j int) bool { return m.Negotiated[i] < m.Negotiated[j] })
	return m
}

func strategiesFrom(caps []types.StorageCapability) []string {
	has := func(c types.StorageCapability) bool {
		for _, x := range caps {
			if x == c {
				return true
			}
		}
		return false
	}
	var out []string
	if has(types.CapSnapshot) || has(types.CapIncrementalSnapshot) {
		out = append(out, "snapshot-seed")
	} else {
		out = append(out, "baseline-copy")
	}
	if has(types.CapChangeTracking) || has(types.CapBlockRead) {
		out = append(out, "block-differential")
	} else {
		out = append(out, "journal-replay")
	}
	if has(types.CapFreeze) {
		out = append(out, "write-barrier-freeze")
	}
	if has(types.CapChecksum) {
		out = append(out, "checksum-verify")
	}
	if has(types.CapAtomicRename) {
		out = append(out, "atomic-pointer-cutover")
	}
	if has(types.CapCopyOnWrite) {
		out = append(out, "cow-candidate")
	}
	return out
}

// Satisfied reports whether all required capabilities are present.
func (m Matrix) Satisfied() bool {
	return len(m.Missing) == 0
}
