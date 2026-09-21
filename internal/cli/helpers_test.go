package cli

import "testing"

func TestModuleMap(t *testing.T) {
	m := ModuleMap()
	if !contains(m, "internal/fsm") || !contains(m, "CUTOVER_WRITE_PAUSE_MS") {
		t.Fatal(m)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
