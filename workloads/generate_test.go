package workloads

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSmall(t *testing.T) {
	root := t.TempDir()
	spec, err := Generate(root, ProfileSmall)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			n++
		}
		return nil
	})
	if n != spec.FileCount {
		t.Fatalf("files=%d want %d", n, spec.FileCount)
	}
}
