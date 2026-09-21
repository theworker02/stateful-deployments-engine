package metrics

import "testing"

func TestFromPause(t *testing.T) {
	b := FromPause("d1", 43.2, 50)
	s, ok := b.Find(CutoverWritePauseMs)
	if !ok || s.Value != 43.2 {
		t.Fatalf("%+v", b)
	}
}
