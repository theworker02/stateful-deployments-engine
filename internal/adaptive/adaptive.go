// Package adaptive tunes sync parameters with bounded, explainable decisions.
package adaptive

import (
	"fmt"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Params are the tunable sync knobs.
type Params struct {
	Concurrency    int
	ChunkSize      int
	BatchSize      int
	Compression    bool
	VerifyEveryN   int
	JournalFlushMs int
}

// Default returns conservative starting params.
func Default() Params {
	return Params{
		Concurrency:    2,
		ChunkSize:      64 * 1024,
		BatchSize:      64,
		Compression:    false,
		VerifyEveryN:   1000,
		JournalFlushMs: 10,
	}
}

// Controller records every tuning decision.
type Controller struct {
	Params    Params
	Decisions []types.TuningDecision
}

// New creates a controller with defaults.
func New() *Controller {
	return &Controller{Params: Default()}
}

// Input drives a tuning step.
type Input struct {
	LagOps     uint64
	WriteRate  float64
	Throughput float64
	CPUBusy    float64 // 0..1 estimated
	MemBusy    float64
	StorageLag float64
}

// Tune adjusts params within bounds and appends explainable decisions.
func (c *Controller) Tune(in Input) []types.TuningDecision {
	var made []types.TuningDecision
	at := time.Now().UTC()

	// Increase batching when lag high and CPU not saturated.
	if in.LagOps > 100 && in.CPUBusy < 0.85 && c.Params.BatchSize < 512 {
		old := c.Params.BatchSize
		c.Params.BatchSize *= 2
		if c.Params.BatchSize > 512 {
			c.Params.BatchSize = 512
		}
		d := types.TuningDecision{
			At: at, Parameter: "batch_size",
			OldValue: fmt.Sprintf("%d", old), NewValue: fmt.Sprintf("%d", c.Params.BatchSize),
			Reason: "high journal lag with spare CPU", LagOps: in.LagOps, WriteRate: in.WriteRate, Throughput: in.Throughput,
		}
		c.Decisions = append(c.Decisions, d)
		made = append(made, d)
	}

	// Enable compression when write rate high and throughput bandwidth-bound.
	if in.WriteRate > 100 && !c.Params.Compression && in.Throughput > 0 {
		c.Params.Compression = true
		d := types.TuningDecision{
			At: at, Parameter: "compression", OldValue: "false", NewValue: "true",
			Reason: "elevated write rate; reduce transfer bytes", LagOps: in.LagOps, WriteRate: in.WriteRate, Throughput: in.Throughput,
		}
		c.Decisions = append(c.Decisions, d)
		made = append(made, d)
	}

	// Raise concurrency when lag grows and memory allows.
	if in.LagOps > 500 && in.MemBusy < 0.8 && c.Params.Concurrency < 8 {
		old := c.Params.Concurrency
		c.Params.Concurrency++
		d := types.TuningDecision{
			At: at, Parameter: "concurrency",
			OldValue: fmt.Sprintf("%d", old), NewValue: fmt.Sprintf("%d", c.Params.Concurrency),
			Reason: "backlog pressure with available memory", LagOps: in.LagOps, WriteRate: in.WriteRate, Throughput: in.Throughput,
		}
		c.Decisions = append(c.Decisions, d)
		made = append(made, d)
	}

	// More frequent verify when storage lag indicates risk.
	if in.StorageLag > 0.5 && c.Params.VerifyEveryN > 100 {
		old := c.Params.VerifyEveryN
		c.Params.VerifyEveryN = c.Params.VerifyEveryN / 2
		d := types.TuningDecision{
			At: at, Parameter: "verify_every_n",
			OldValue: fmt.Sprintf("%d", old), NewValue: fmt.Sprintf("%d", c.Params.VerifyEveryN),
			Reason: "storage lag elevated; tighten verification cadence", LagOps: in.LagOps, WriteRate: in.WriteRate, Throughput: in.Throughput,
		}
		c.Decisions = append(c.Decisions, d)
		made = append(made, d)
	}

	return made
}
