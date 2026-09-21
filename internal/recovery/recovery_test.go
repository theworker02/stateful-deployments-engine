package recovery_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/recovery"
)

func TestLocalTargetPrepareFinalize(t *testing.T) {
	root := t.TempDir()
	lt := &recovery.LocalTarget{Root: root}
	if lt.Name() == "" {
		t.Fatal("empty name")
	}
	if len(lt.Capabilities()) == 0 {
		t.Fatal("no capabilities")
	}
	sp, err := lt.PrepareRestore(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sp, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := lt.FinalizeRestore(context.Background(), sp, 7); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "restore-meta.json")); err != nil {
		t.Fatal(err)
	}
}

func TestSelectRestoreStrategies(t *testing.T) {
	_ = recovery.SelectRestoreStrategies(nil)
}
