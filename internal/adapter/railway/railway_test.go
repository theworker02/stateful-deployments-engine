package railway

import "testing"

func TestVolumeCapabilities(t *testing.T) {
	m := DefaultVolumeModel()
	if m.MaxVolumesPerService != 1 || m.SupportsReplicas {
		t.Fatalf("%+v", m)
	}
	caps := StorageCapabilities()
	if len(caps) < 2 {
		t.Fatal(caps)
	}
}

func TestCutoverPlan(t *testing.T) {
	a := NewWithConfig(Config{ActiveServiceID: "a", ShadowServiceID: "b", PublicDomain: "app.example"})
	p := a.BuildCutoverPlan()
	if p.PublicDomain == "" || p.ActiveServiceID != "a" {
		t.Fatalf("%+v", p)
	}
}
