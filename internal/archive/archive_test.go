package archive_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/recovery"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestExportVerifyImport(t *testing.T) {
	src := t.TempDir()
	_ = os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0o644)
	_ = os.WriteFile(filepath.Join(src, "b.bin"), make([]byte, 3000), 0o644)

	arch := filepath.Join(t.TempDir(), "psa")
	m, stats, err := archive.Export(src, arch, 1, 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if m.Portability != archive.EngineProvided {
		t.Fatalf("portability=%s", m.Portability)
	}
	if stats.LogicalBytes <= 0 || stats.PhysicalBytes <= 0 {
		t.Fatalf("stats=%+v", stats)
	}
	vr, err := archive.Verify(arch)
	if err != nil || vr.Outcome != types.VerifyVerified {
		t.Fatalf("verify: %v %+v", err, vr)
	}
	dst := t.TempDir()
	if _, err := archive.Import(arch, dst); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(b) != "hello" {
		t.Fatalf("got %q", b)
	}
}

func TestCloneDoesNotMutateOriginal(t *testing.T) {
	src := t.TempDir()
	_ = os.WriteFile(filepath.Join(src, "x.txt"), []byte("x"), 0o644)
	arch := filepath.Join(t.TempDir(), "a")
	_, _, err := archive.Export(src, arch, 1, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	orig, _ := os.ReadFile(filepath.Join(arch, "archive.json"))
	cloneDir := filepath.Join(t.TempDir(), "clone")
	m2, err := archive.CloneReadOnly(arch, cloneDir)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(cloneDir, "archive.json"), []byte(string(orig)+"\n"), 0o644)
	after, _ := os.ReadFile(filepath.Join(arch, "archive.json"))
	if string(after) != string(orig) {
		t.Fatal("original archive mutated by clone writes")
	}
	if m2.ArchiveID == "" {
		t.Fatal("clone needs new identity")
	}
}

// Headline recovery independence test:
// create → archive → verify → destroy source → restore to new target → verify digest → receipt.
func TestRecoveryIndependence(t *testing.T) {
	src := t.TempDir()
	_ = os.MkdirAll(filepath.Join(src, "data"), 0o755)
	_ = os.WriteFile(filepath.Join(src, "data", "record.json"), []byte(`{"v":1}`), 0o644)
	_ = os.WriteFile(filepath.Join(src, "data", "blob.bin"), make([]byte, 8192), 0o644)

	arch := filepath.Join(t.TempDir(), "archive")
	m, _, err := archive.Export(src, arch, 7, 42, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := archive.Verify(arch); err != nil {
		t.Fatal(err)
	}

	// Destroy source
	if err := os.RemoveAll(src); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("source should be gone")
	}

	targetRoot := t.TempDir()
	tgt := &recovery.LocalTarget{Root: targetRoot}
	rcptDir := filepath.Join(targetRoot, "receipts")
	rcpt, err := recovery.Restore(context.Background(), arch, tgt, rcptDir, true)
	if err != nil {
		t.Fatal(err)
	}
	if rcpt.RestoredDigest != m.RootDigest {
		t.Fatalf("digest %s != %s", rcpt.RestoredDigest, m.RootDigest)
	}
	if rcpt.VerifyOutcome != string(types.VerifyVerified) {
		t.Fatalf("outcome=%s", rcpt.VerifyOutcome)
	}
	if rcpt.Mode != recovery.ModeEngineProvided {
		t.Fatalf("mode=%s", rcpt.Mode)
	}
	if _, err := os.Stat(filepath.Join(rcptDir, "RECOVERY_RECEIPT.json")); err != nil {
		t.Fatal("missing RECOVERY_RECEIPT.json")
	}
	restored := filepath.Join(targetRoot, "restored-state", "data", "record.json")
	b, err := os.ReadFile(restored)
	if err != nil || string(b) != `{"v":1}` {
		t.Fatalf("restored payload: %v %q", err, b)
	}
}
