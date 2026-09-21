// Package heat models hot/warm/cold/static state for sync scheduling.
//
// Complements internal/hotstate with thresholds, windowed scoring, and
// JSON-serializable heat maps used by planners and adaptive sync.
package heat

import (
	"encoding/json"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Thresholds control temperature classification.
type Thresholds struct {
	HotMinCount   uint64
	HotMaxAge     time.Duration
	WarmMinCount  uint64
	WarmMaxAge    time.Duration
	ColdMaxAge    time.Duration
}

// DefaultThresholds mirrors the product defaults used by hotstate.
func DefaultThresholds() Thresholds {
	return Thresholds{
		HotMinCount:  10,
		HotMaxAge:    2 * time.Second,
		WarmMinCount: 3,
		WarmMaxAge:   10 * time.Second,
		ColdMaxAge:   60 * time.Second,
	}
}

// Score is a single path heat observation.
type Score struct {
	Path          string                   `json:"path"`
	Temperature   types.ObjectTemperature  `json:"temperature"`
	MutationCount uint64                   `json:"mutation_count"`
	BytesTouched  int64                    `json:"bytes_touched"`
	LastMutation  time.Time                `json:"last_mutation"`
	AgeMs         float64                  `json:"age_ms"`
}

// Map is a serializable heat map for a deploy cycle.
type Map struct {
	Epoch     uint64    `json:"epoch"`
	Generated time.Time `json:"generated_at"`
	Scores    []Score   `json:"scores"`
}

// Classify applies thresholds to counts/ages.
func Classify(path string, count uint64, bytes int64, last, now time.Time, th Thresholds) Score {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if last.IsZero() {
		last = now
	}
	age := now.Sub(last)
	temp := types.TempStatic
	switch {
	case count >= th.HotMinCount && age <= th.HotMaxAge:
		temp = types.TempHot
	case count >= th.WarmMinCount && age <= th.WarmMaxAge:
		temp = types.TempWarm
	case count >= 1 && age <= th.ColdMaxAge:
		temp = types.TempCold
	default:
		temp = types.TempStatic
	}
	return Score{
		Path:          path,
		Temperature:   temp,
		MutationCount: count,
		BytesTouched:  bytes,
		LastMutation:  last,
		AgeMs:         float64(age.Milliseconds()),
	}
}

// FromHotObjects converts types.HotObject slices into a Map.
func FromHotObjects(epoch uint64, objs []types.HotObject) Map {
	m := Map{Epoch: epoch, Generated: time.Now().UTC()}
	now := m.Generated
	th := DefaultThresholds()
	for _, o := range objs {
		m.Scores = append(m.Scores, Classify(o.Path, o.MutationCount, o.BytesTouched, o.LastMutation, now, th))
	}
	return m
}

// MarshalJSON is a convenience wrapper.
func (m Map) Bytes() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// Partition splits scores into baseline (cold/static) vs continuous (hot/warm).
func Partition(scores []Score) (baseline, continuous []string) {
	for _, s := range scores {
		switch s.Temperature {
		case types.TempHot, types.TempWarm:
			continuous = append(continuous, s.Path)
		default:
			baseline = append(baseline, s.Path)
		}
	}
	return baseline, continuous
}
