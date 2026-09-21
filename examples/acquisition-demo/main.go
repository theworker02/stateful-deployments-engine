// Acquisition demo: create → archive → verify → destroy source → restore → receipt.
// ENGINE_PROVIDED portability — not Railway-native backup.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/dr"
	"github.com/theworker02/stateful-deployments-engine/internal/escrow"
	"github.com/theworker02/stateful-deployments-engine/internal/recovery"
	"github.com/theworker02/stateful-deployments-engine/internal/recoverypoint"
)

func main() {
	base := filepath.Join(os.TempDir(), "sde-acq-demo")
	_ = os.RemoveAll(base)
	src := filepath.Join(base, "src")
	_ = os.MkdirAll(filepath.Join(src, "data"), 0o755)
	_ = os.WriteFile(filepath.Join(src, "data", "record.json"), []byte(`{"acq":true}`), 0o644)
	_ = os.WriteFile(filepath.Join(src, "data", "blob.bin"), make([]byte, 8192), 0o644)

	arch := filepath.Join(base, "archive")
	m, stats, err := archive.Export(src, arch, 1, 1, "")
	must(err)
	fmt.Printf("exported %s logical=%d physical=%d dedup_saved=%d\n", m.ArchiveID, stats.LogicalBytes, stats.PhysicalBytes, stats.DedupSaved)

	vr, err := archive.Verify(arch)
	must(err)
	fmt.Printf("verify: %s\n", vr.Outcome)

	esc := &escrow.Store{CatalogDir: filepath.Join(base, "escrow")}
	_, err = esc.Put(escrow.Target{Kind: escrow.LocalFilesystem, Name: "demo", RootPath: filepath.Join(base, "escrow", "store")}, arch)
	must(err)

	rp := &recoverypoint.Store{Root: filepath.Join(base, "recovery")}
	_, err = rp.CreateFromArchive(arch)
	must(err)

	must(os.RemoveAll(src))
	fmt.Println("source destroyed")

	tgt := &recovery.LocalTarget{Root: filepath.Join(base, "restored")}
	rcpt, err := recovery.Restore(context.Background(), arch, tgt, filepath.Join(base, "receipts"), true)
	must(err)
	b, _ := json.MarshalIndent(rcpt, "", "  ")
	fmt.Println(string(b))

	drill, err := dr.RunFireDrill(arch, filepath.Join(base, "drill"))
	must(err)
	fmt.Printf("fire-drill ok=%v duration_ms=%.2f\n", drill.RootDigestOK, drill.DurationMs)

	plan, err := (&dr.Planner{Escrow: esc, Points: rp}).Plan()
	must(err)
	fmt.Printf("dr-plan: %s\n", plan.Conclusion)
	fmt.Println("ACQUISITION DEMO COMPLETE (ENGINE_PROVIDED)")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
