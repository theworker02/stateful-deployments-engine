package fly

import "testing"

func TestNameCapabilities(t *testing.T) {
	a := New(Config{AppName: "demo"})
	if a.Name() != "fly" {
		t.Fatal(a.Name())
	}
	if len(a.Capabilities()) < 2 {
		t.Fatal(a.Capabilities())
	}
}
