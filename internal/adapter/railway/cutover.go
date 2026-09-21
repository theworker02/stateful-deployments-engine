package railway

import (
	"context"
	"fmt"
)

// CutoverPlan describes domain / traffic cutover for dual-service Railway.
type CutoverPlan struct {
	ActiveServiceID  string  `json:"active_service_id"`
	ShadowServiceID  string  `json:"shadow_service_id"`
	PublicDomain     string  `json:"public_domain"`
	Strategy         string  `json:"strategy"`
	PredictedPauseMs float64 `json:"predicted_pause_ms"`
}

// BuildCutoverPlan documents the intended cutover approach (no side effects).
func (a *Adapter) BuildCutoverPlan() CutoverPlan {
	return CutoverPlan{
		ActiveServiceID:  a.cfg.ActiveServiceID,
		ShadowServiceID:  a.cfg.ShadowServiceID,
		PublicDomain:     a.cfg.PublicDomain,
		Strategy:         "service-domain-create + variable flip",
		PredictedPauseMs: 50,
	}
}

// ExecuteDomainSwap shifts traffic toward the candidate by ensuring it has a
// Railway service domain and marking it as the SDE active target via variables.
// Custom domain DNS edits (when PublicDomain is set) remain operator-owned:
// SDE records intent and attaches a service domain; it does not silently rewrite
// third-party DNS.
func (a *Adapter) ExecuteDomainSwap(ctx context.Context, fromService, toService string) error {
	if err := a.requireLive(); err != nil {
		return err
	}
	if a.cfg.EnvironmentID == "" {
		return fmt.Errorf("railway ExecuteDomainSwap: RAILWAY_ENVIRONMENT_ID required")
	}
	domain, err := a.client.CreateServiceDomain(ctx, toService, a.cfg.EnvironmentID)
	if err != nil {
		// Domain may already exist — continue with variable flip.
		domain = "(existing-or-error: " + err.Error() + ")"
	}
	if err := a.client.UpsertVariables(ctx, a.cfg.ProjectID, a.cfg.EnvironmentID, toService, map[string]string{
		"SDE_ROLE":           "active",
		"SDE_PUBLIC_DOMAIN":  a.cfg.PublicDomain,
		"SDE_SERVICE_DOMAIN": domain,
	}); err != nil {
		return fmt.Errorf("railway ExecuteDomainSwap: %w", err)
	}
	_ = a.client.UpsertVariables(ctx, a.cfg.ProjectID, a.cfg.EnvironmentID, fromService, map[string]string{
		"SDE_ROLE": "standby",
	})
	return nil
}
