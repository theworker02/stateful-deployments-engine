// Package observe implements the post-cutover observation window.
//
// After traffic transfer, SDE watches candidate health and error signals before
// sealing COMMITTED. Failures during observation may trigger rollback if safe.
package observe

import (
	"context"
	"time"
)

// Config controls the observation window.
type Config struct {
	Duration       time.Duration
	PollInterval   time.Duration
	MaxErrorEvents int
}

// DefaultConfig returns conservative acquisition-demo defaults.
func DefaultConfig() Config {
	return Config{
		Duration:       5 * time.Second,
		PollInterval:   200 * time.Millisecond,
		MaxErrorEvents: 0,
	}
}

// Probe is called each poll; return ok=false to signal unhealthy.
type Probe func(ctx context.Context) (ok bool, detail string, err error)

// Result summarizes an observation window.
type Result struct {
	StartedAt    time.Time `json:"started_at"`
	EndedAt      time.Time `json:"ended_at"`
	DurationMs   float64   `json:"duration_ms"`
	Polls        int       `json:"polls"`
	HealthyPolls int       `json:"healthy_polls"`
	ErrorEvents  int       `json:"error_events"`
	LastDetail   string    `json:"last_detail,omitempty"`
	Passed       bool      `json:"passed"`
	AbortedEarly bool      `json:"aborted_early"`
}

// Run executes the observation window until duration elapses or health fails.
func Run(ctx context.Context, cfg Config, probe Probe) (Result, error) {
	if cfg.Duration <= 0 {
		cfg.Duration = DefaultConfig().Duration
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = DefaultConfig().PollInterval
	}
	res := Result{StartedAt: time.Now().UTC()}
	deadline := res.StartedAt.Add(cfg.Duration)
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		res.Polls++
		ok, detail, err := probe(ctx)
		res.LastDetail = detail
		if err != nil {
			res.ErrorEvents++
			res.EndedAt = time.Now().UTC()
			res.DurationMs = float64(res.EndedAt.Sub(res.StartedAt).Milliseconds())
			res.AbortedEarly = true
			res.Passed = false
			return res, err
		}
		if ok {
			res.HealthyPolls++
		} else {
			res.ErrorEvents++
			if res.ErrorEvents > cfg.MaxErrorEvents {
				res.EndedAt = time.Now().UTC()
				res.DurationMs = float64(res.EndedAt.Sub(res.StartedAt).Milliseconds())
				res.AbortedEarly = true
				res.Passed = false
				return res, nil
			}
		}
		now := time.Now().UTC()
		if !now.Before(deadline) {
			res.EndedAt = now
			res.DurationMs = float64(res.EndedAt.Sub(res.StartedAt).Milliseconds())
			res.Passed = res.ErrorEvents <= cfg.MaxErrorEvents
			return res, nil
		}
		select {
		case <-ctx.Done():
			res.EndedAt = time.Now().UTC()
			res.DurationMs = float64(res.EndedAt.Sub(res.StartedAt).Milliseconds())
			res.AbortedEarly = true
			res.Passed = false
			return res, ctx.Err()
		case <-ticker.C:
		}
	}
}
