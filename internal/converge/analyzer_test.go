package converge

import (
	"testing"
	"time"
)

func TestAnalyzerTrend(t *testing.T) {
	a := NewAnalyzer()
	t0 := time.Now().UTC()
	s1 := a.Push(Sample{WriteRateOpsPerSec: 10, ReplicationRateOpsPerSec: 50, BacklogOps: 1000, At: t0})
	if s1.NonConvergent {
		t.Fatal("should be converging")
	}
	s2 := a.Push(Sample{WriteRateOpsPerSec: 10, ReplicationRateOpsPerSec: 50, BacklogOps: 500, At: t0.Add(time.Second)})
	if s2.BacklogVelocityOpsPerSec >= 0 {
		t.Fatalf("velocity=%v", s2.BacklogVelocityOpsPerSec)
	}
	if !a.TrendImproving() {
		t.Fatal("expected improving trend")
	}
}

func TestNonConvergent(t *testing.T) {
	snap := Observe(Sample{WriteRateOpsPerSec: 100, ReplicationRateOpsPerSec: 50, BacklogOps: 10}, 0, time.Time{})
	if !snap.NonConvergent {
		t.Fatal("expected non-convergent")
	}
}
