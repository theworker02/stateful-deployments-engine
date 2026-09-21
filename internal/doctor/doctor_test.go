package doctor_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/doctor"
	"github.com/theworker02/stateful-deployments-engine/internal/version"
)

func TestRunOfflineReady(t *testing.T) {
	root := t.TempDir()
	_ = os.Unsetenv("SDE_RAILWAY_LIVE")
	r := doctor.Run(context.Background(), doctor.Options{Root: root})
	if r.Version != version.Version {
		t.Fatalf("version %q", r.Version)
	}
	if !r.Ready {
		t.Fatalf("expected ready offline: %+v", r.Checks)
	}
	human := doctor.FormatHuman(r)
	if !strings.Contains(human, version.Version) {
		t.Fatalf("human missing version: %s", human)
	}
	b, err := json.Marshal(r)
	if err != nil || !strings.Contains(string(b), `"ready":true`) {
		t.Fatalf("json: %v %s", err, b)
	}
}

func TestRunWarnsMissingWorkspace(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-root")
	r := doctor.Run(context.Background(), doctor.Options{Root: missing})
	found := false
	for _, c := range r.Checks {
		if c.Name == "workspace" && c.Severity == doctor.Warn {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected workspace warn: %+v", r.Checks)
	}
}

func TestLiveMissingTokenFails(t *testing.T) {
	t.Setenv("SDE_RAILWAY_LIVE", "1")
	t.Setenv("RAILWAY_TOKEN", "")
	t.Setenv("RAILWAY_API_TOKEN", "")
	t.Setenv("RAILWAY_PROJECT_TOKEN", "")
	r := doctor.Run(context.Background(), doctor.Options{Root: t.TempDir()})
	if r.Ready {
		t.Fatal("expected not ready without token")
	}
}
