package converge

import (
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Analyzer tracks backlog velocity across samples.
type Analyzer struct {
	priorBacklog uint64
	priorAt      time.Time
	history      []Sample
}

// NewAnalyzer constructs an empty analyzer.
func NewAnalyzer() *Analyzer { return &Analyzer{} }

// Push observes a sample and returns the latest snapshot.
func (a *Analyzer) Push(s Sample) types.ConvergenceSnapshot {
	snap := Observe(s, a.priorBacklog, a.priorAt)
	a.priorBacklog = s.BacklogOps
	if s.At.IsZero() {
		a.priorAt = snap.ObservedAt
	} else {
		a.priorAt = s.At
	}
	a.history = append(a.history, s)
	return snap
}

// HistoryLen returns observed sample count.
func (a *Analyzer) HistoryLen() int { return len(a.history) }

// TrendImproving reports whether the last two backlog points are decreasing.
func (a *Analyzer) TrendImproving() bool {
	if len(a.history) < 2 {
		return false
	}
	a1 := a.history[len(a.history)-2]
	a2 := a.history[len(a.history)-1]
	return a2.BacklogOps < a1.BacklogOps
}
