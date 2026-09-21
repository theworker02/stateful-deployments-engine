package local

import "testing"

func TestStorageCapabilities(t *testing.T) {
	caps := StorageCapabilities()
	if len(caps) < 5 {
		t.Fatal(caps)
	}
}

func TestLayout(t *testing.T) {
	a, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := a.Layout()
	if p.Slots == "" || p.Journal == "" {
		t.Fatalf("%+v", p)
	}
}
