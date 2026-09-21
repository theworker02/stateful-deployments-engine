package kubernetes

import "testing"

func TestK8sStub(t *testing.T) {
	a := New(Config{Namespace: "default", StatefulSet: "app"})
	if a.Name() != "kubernetes" {
		t.Fatal(a.Name())
	}
	s := DefaultCutoverStrategy()
	if len(s.Steps) < 3 {
		t.Fatal(s)
	}
}
