package epoch

import (
	"path/filepath"
	"testing"
)

func TestStoreAdvanceAndPin(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "epochs"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Advance("vol-a", "coord-1", "cutover")
	if err != nil {
		t.Fatal(err)
	}
	if r.Epoch != 1 {
		t.Fatalf("epoch=%d want 1", r.Epoch)
	}
	got, err := s.Get("vol-a")
	if err != nil {
		t.Fatal(err)
	}
	if got.Epoch != 1 || got.HolderID != "coord-1" {
		t.Fatalf("got=%+v", got)
	}
	if _, err := s.Pin("vol-a", 0, "rollback", false); err == nil {
		t.Fatal("expected refuse lower pin")
	}
	pinned, err := s.Pin("vol-a", 0, "rollback", true)
	if err != nil {
		t.Fatal(err)
	}
	if pinned.Epoch != 0 {
		t.Fatalf("pinned=%d", pinned.Epoch)
	}
}

func TestCompareIsStale(t *testing.T) {
	if Compare(1, 2) != -1 || Compare(2, 2) != 0 || Compare(3, 2) != 1 {
		t.Fatal("compare broken")
	}
	if !IsStale(1, 2) || IsStale(2, 2) {
		t.Fatal("stale broken")
	}
}
