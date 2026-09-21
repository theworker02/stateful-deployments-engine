package recoverypoint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/recoverypoint"
)

func TestCreateListFromArchive(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	_ = os.MkdirAll(src, 0o755)
	_ = os.WriteFile(filepath.Join(src, "f.txt"), []byte("hi"), 0o644)
	arch := filepath.Join(root, "arch")
	if _, _, err := archive.Export(src, arch, 3, 9, ""); err != nil {
		t.Fatal(err)
	}
	st := &recoverypoint.Store{Root: root}
	p, err := st.CreateFromArchive(arch)
	if err != nil {
		t.Fatal(err)
	}
	if p.Epoch != 3 || p.Kind != "SNAPSHOT" {
		t.Fatalf("%+v", p)
	}
	pts, err := st.List()
	if err != nil || len(pts) != 1 {
		t.Fatalf("list=%v err=%v", pts, err)
	}
}
