// Package safety provides cutover gate helpers and policy presets.
//
// Core evaluation lives in internal/safegate; this package adds operator-facing
// policies, dry-run simulation, and serialization helpers.
package safety

import (
	"encoding/json"

	"github.com/theworker02/stateful-deployments-engine/internal/safegate"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Policy is an operator-configured gate profile.
type Policy struct {
	Name            string  `json:"name"`
	MaxWritePauseMs float64 `json:"max_write_pause_ms"`
	AllowOverride   bool    `json:"allow_override"`
	RequireRollback bool    `json:"require_rollback_point"`
}

// Strict is the default acquisition / production-leaning policy.
func Strict() Policy {
	return Policy{Name: "strict", MaxWritePauseMs: 250, AllowOverride: false, RequireRollback: true}
}

// EvalDemo is a looser policy for local demos.
func EvalDemo() Policy {
	return Policy{Name: "eval-demo", MaxWritePauseMs: 2000, AllowOverride: true, RequireRollback: true}
}

// Evaluate applies policy defaults onto a safegate.Input and runs the gate.
func Evaluate(p Policy, in safegate.Input) types.SafetyGateResult {
	if in.MaxWritePauseMs <= 0 {
		in.MaxWritePauseMs = p.MaxWritePauseMs
	}
	if p.RequireRollback && !in.RollbackPointPresent {
		in.RollbackPointPresent = false
	}
	if !p.AllowOverride {
		in.OverrideRequested = false
	}
	return safegate.Evaluate(in)
}

// IsAllowed reports CUTOVER_ALLOWED.
func IsAllowed(r types.SafetyGateResult) bool {
	return r.Decision == types.CutoverAllowed
}

// MarshalResult JSON-encodes a gate result.
func MarshalResult(r types.SafetyGateResult) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
