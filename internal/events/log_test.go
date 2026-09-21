package events

import (
	"strings"
	"testing"
)

func TestLogRender(t *testing.T) {
	var l Log
	l.OK("d1", KindPhaseEnter, "Captured state checkpoint", nil)
	l.Info("d1", KindMutationSynced, "14,821 mutations synchronized", map[string]interface{}{"n": 14821})
	lines := RenderHuman(l.Snapshot())
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "✓") {
		t.Fatalf("%v", lines)
	}
	if _, err := l.JSON(); err != nil {
		t.Fatal(err)
	}
}
