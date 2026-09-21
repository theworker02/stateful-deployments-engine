package hotstate

import (
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestClassifySchedule(t *testing.T) {
	tr := New()
	now := time.Now().UTC()
	for i := 0; i < 12; i++ {
		tr.Observe("/hot", 10, now)
	}
	tr.Observe("/cold", 1, now.Add(-30*time.Second))
	objs := tr.Classify(now)
	if len(objs) != 2 {
		t.Fatal(objs)
	}
	early, cont := ScheduleOrder(objs)
	if len(cont) < 1 || objs[0].Temperature != types.TempHot {
		t.Fatalf("early=%v cont=%v objs=%+v", early, cont, objs)
	}
}
