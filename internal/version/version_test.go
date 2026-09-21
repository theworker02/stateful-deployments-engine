package version_test

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/version"
)

func TestVersionIsFrozen(t *testing.T) {
	if version.Version != "1.1.1" {
		t.Fatalf("version=%q want 1.1.1", version.Version)
	}
	if version.Codename != "FROZEN" {
		t.Fatalf("codename=%q want FROZEN", version.Codename)
	}
	if version.String() != "sde 1.1.1" {
		t.Fatalf("String=%q", version.String())
	}
}
