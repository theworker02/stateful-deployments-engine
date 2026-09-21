package version_test

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/version"
)

func TestVersionIsReady(t *testing.T) {
	if version.Version != "1.1.0" {
		t.Fatalf("version=%q want 1.1.0", version.Version)
	}
	if version.Codename != "READY" {
		t.Fatalf("codename=%q want READY", version.Codename)
	}
	if version.String() != "sde 1.1.0" {
		t.Fatalf("String=%q", version.String())
	}
}
