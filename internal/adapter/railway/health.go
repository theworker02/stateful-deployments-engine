package railway

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus is a candidate/active health observation.
type HealthStatus struct {
	SlotID    string    `json:"slot_id"`
	Healthy   bool      `json:"healthy"`
	CheckedAt time.Time `json:"checked_at"`
	Detail    string    `json:"detail,omitempty"`
}

// ProbeHealth queries Railway serviceInstance latestDeployment status.
func (a *Adapter) ProbeHealth(ctx context.Context, slotID string) (HealthStatus, error) {
	if err := a.requireLive(); err != nil {
		return HealthStatus{SlotID: slotID, CheckedAt: time.Now().UTC()}, err
	}
	if a.cfg.EnvironmentID == "" {
		return HealthStatus{}, fmt.Errorf("railway ProbeHealth: RAILWAY_ENVIRONMENT_ID required")
	}
	status, healthy, err := a.client.GetServiceInstanceHealth(ctx, slotID, a.cfg.EnvironmentID)
	if err != nil {
		return HealthStatus{SlotID: slotID, CheckedAt: time.Now().UTC(), Detail: err.Error()}, err
	}
	return HealthStatus{
		SlotID:    slotID,
		Healthy:   healthy,
		CheckedAt: time.Now().UTC(),
		Detail:    status,
	}, nil
}
