// Package hotstate classifies objects by observed mutation heat for sync scheduling.
package hotstate

import (
	"sort"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Tracker accumulates per-path mutation observations.
type Tracker struct {
	counts map[string]uint64
	bytes  map[string]int64
	last   map[string]time.Time
}

// New creates an empty tracker.
func New() *Tracker {
	return &Tracker{
		counts: make(map[string]uint64),
		bytes:  make(map[string]int64),
		last:   make(map[string]time.Time),
	}
}

// Observe records a mutation against a path.
func (t *Tracker) Observe(path string, size int64, at time.Time) {
	t.counts[path]++
	t.bytes[path] += size
	if at.IsZero() {
		at = time.Now().UTC()
	}
	t.last[path] = at
}

// Classify returns HOT/WARM/COLD/STATIC for all observed paths.
func (t *Tracker) Classify(now time.Time) []types.HotObject {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var out []types.HotObject
	for path, n := range t.counts {
		temp := types.TempStatic
		age := now.Sub(t.last[path])
		switch {
		case n >= 10 && age < 2*time.Second:
			temp = types.TempHot
		case n >= 3 && age < 10*time.Second:
			temp = types.TempWarm
		case n >= 1 && age < 60*time.Second:
			temp = types.TempCold
		default:
			temp = types.TempStatic
		}
		out = append(out, types.HotObject{
			Path:          path,
			Temperature:   temp,
			MutationCount: n,
			LastMutation:  t.last[path],
			BytesTouched:  t.bytes[path],
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return rank(out[i].Temperature) > rank(out[j].Temperature)
	})
	return out
}

func rank(t types.ObjectTemperature) int {
	switch t {
	case types.TempHot:
		return 3
	case types.TempWarm:
		return 2
	case types.TempCold:
		return 1
	default:
		return 0
	}
}

// ScheduleOrder returns cold/static first (baseline), hot last (continuous sync).
func ScheduleOrder(objs []types.HotObject) (early []string, continuous []string) {
	for _, o := range objs {
		switch o.Temperature {
		case types.TempHot, types.TempWarm:
			continuous = append(continuous, o.Path)
		default:
			early = append(early, o.Path)
		}
	}
	return early, continuous
}
