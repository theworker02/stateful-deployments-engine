package railway_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/railway"
)

func TestCloneVolumeReceipt(t *testing.T) {
	src := t.TempDir()
	_ = os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0o644)
	dest := t.TempDir()
	a := railway.NewWithConfig(railway.Config{})
	rcpt, err := a.CloneVolume(context.Background(), railway.CloneRequest{
		SourceState: src,
		DestDir:     dest,
		Mode:        railway.CloneReadOnly,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !rcpt.Verified || rcpt.Portability != string(railway.ModeEngineProvided) {
		t.Fatalf("%+v", rcpt)
	}
	if _, err := os.Stat(filepath.Join(dest, "CLONE_RECEIPT.json")); err != nil {
		t.Fatal(err)
	}
}
