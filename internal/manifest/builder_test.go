package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildEqualDiff(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	if err := os.WriteFile(filepath.Join(a, "f.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(b, "f.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	ma, err := (Builder{}).Build(a, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	mb, err := (Builder{}).Build(b, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !Equal(ma, mb) {
		t.Fatalf("expected equal merkle %s vs %s", ma.MerkleRoot, mb.MerkleRoot)
	}
	if err := os.WriteFile(filepath.Join(b, "f.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}
	mb2, err := (Builder{}).Build(b, 1, 11)
	if err != nil {
		t.Fatal(err)
	}
	_, _, mismatch := DiffPaths(ma, mb2)
	if len(mismatch) != 1 || mismatch[0] != "f.txt" {
		t.Fatalf("mismatch=%v", mismatch)
	}
}

func TestWriteJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "x"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := (Builder{}).Build(root, 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "m.json")
	if err := WriteJSON(out, m); err != nil {
		t.Fatal(err)
	}
}
