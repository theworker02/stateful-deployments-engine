package docker

import "testing"

func TestDockerCaps(t *testing.T) {
	a := New(Config{ComposeProject: "sde"})
	if a.Name() != "docker" || len(a.Capabilities()) < 3 {
		t.Fatal(a.Name(), a.Capabilities())
	}
}
