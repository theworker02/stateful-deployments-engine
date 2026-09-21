package dr_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/dr"
	"github.com/theworker02/stateful-deployments-engine/internal/escrow"
	"github.com/theworker02/stateful-deployments-engine/internal/recoverypoint"
)

func TestPlanUnknownNeverRecoverableEmpty(t *testing.T) {
	base := t.TempDir()
	p := &dr.Planner{
		Escrow: &escrow.Store{CatalogDir: filepath.Join(base, "escrow")},
		Points: &recoverypoint.Store{Root: filepath.Join(base, "recovery")},
	}
	plan, err := p.Plan()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Conclusion == dr.Recoverable {
		t.Fatal("empty catalog must not be RECOVERABLE")
	}
	if dr.IsRecoverable(dr.Unknown) {
		t.Fatal("UNKNOWN must never be treated as recoverable")
	}
}

func TestPlanWithFireDrillScorecard(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "src")
	_ = os.MkdirAll(src, 0o755)
	_ = os.WriteFile(filepath.Join(src, "f.txt"), []byte("hi"), 0o644)
	arch := filepath.Join(base, "arch")
	m, _, err := archive.Export(src, arch, 2, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	esc := &escrow.Store{CatalogDir: filepath.Join(base, "escrow")}
	_, err = esc.Put(escrow.Target{Kind: escrow.LocalFilesystem, Name: "local", RootPath: filepath.Join(base, "store")}, arch)
	if err != nil {
		t.Fatal(err)
	}
	_ = esc.RecordFireDrill(m.ArchiveID, escrow.FireDrillMeta{
		ReceiptID: "d1", RanAt: time.Now().UTC(), RootDigestOK: true, DurationMs: 1.5,
	})
	rp := &recoverypoint.Store{Root: filepath.Join(base, "recovery")}
	_, _ = rp.CreateFromArchive(arch)

	plan, err := (&dr.Planner{Escrow: esc, Points: rp}).Plan()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Conclusion != dr.Recoverable {
		t.Fatalf("conclusion=%s", plan.Conclusion)
	}
	if !plan.Scorecard.FireDrillRecorded || !plan.Scorecard.LastFireDrillOK {
		t.Fatalf("scorecard=%+v", plan.Scorecard)
	}
}
