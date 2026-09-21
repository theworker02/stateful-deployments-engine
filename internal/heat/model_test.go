package heat

import (
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestClassifyPartition(t *testing.T) {
	now := time.Now().UTC()
	th := DefaultThresholds()
	hot := Classify("/db/wal", 20, 1000, now.Add(-time.Second), now, th)
	if hot.Temperature != types.TempHot {
		t.Fatalf("got %s", hot.Temperature)
	}
	cold := Classify("/assets/logo.png", 1, 10, now.Add(-30*time.Second), now, th)
	if cold.Temperature != types.TempCold {
		t.Fatalf("got %s", cold.Temperature)
	}
	base, cont := Partition([]Score{hot, cold})
	if len(cont) != 1 || len(base) != 1 {
		t.Fatalf("base=%v cont=%v", base, cont)
	}
}
