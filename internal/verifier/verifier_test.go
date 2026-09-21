package verifier_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
)

func TestCompareRootsMismatch(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	_ = os.WriteFile(filepath.Join(a, "x.txt"), []byte("1"), 0o644)
	_ = os.WriteFile(filepath.Join(b, "x.txt"), []byte("2"), 0o644)
	r, _, _, err := verifier.CompareRoots(a, b, 1)
	if err != nil {
		t.Fatal(err)
	}
	if r.OK {
		t.Fatal("expected mismatch")
	}
}

func TestCompareRootsEqual(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	_ = os.WriteFile(filepath.Join(a, "x.txt"), []byte("same"), 0o644)
	_ = os.WriteFile(filepath.Join(b, "x.txt"), []byte("same"), 0o644)
	r, ca, cb, err := verifier.CompareRoots(a, b, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK || ca.RootHash != cb.RootHash {
		t.Fatalf("expected equal: %+v", r)
	}
}
