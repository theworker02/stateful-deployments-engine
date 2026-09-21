package nomad

import "testing"

func TestNomadStub(t *testing.T) {
	a := New(Config{JobID: "app"})
	if a.Name() != "nomad" {
		t.Fatal(a.Name())
	}
	p := DefaultHostVolumePlan("app")
	if p.ActiveVolume == "" {
		t.Fatal(p)
	}
}
