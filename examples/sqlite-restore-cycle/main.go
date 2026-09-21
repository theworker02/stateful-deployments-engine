// SQLite-like workload: seed → archive → destroy source → restore → verify digest.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
)

func main() {
	base := filepath.Join(os.TempDir(), "sde-sqlite-restore")
	_ = os.RemoveAll(base)
	src := filepath.Join(base, "db")
	_ = os.MkdirAll(src, 0o755)
	// Simulate a sqlite file + wal-ish companion (opaque bytes).
	db := make([]byte, 64*1024)
	for i := range db {
		db[i] = byte(i % 251)
	}
	_ = os.WriteFile(filepath.Join(src, "app.db"), db, 0o644)
	_ = os.WriteFile(filepath.Join(src, "app.db-wal"), []byte("wal-fragment"), 0o644)
	_ = os.WriteFile(filepath.Join(src, "meta.json"), []byte(`{"engine":"sqlite-like","pages":256}`), 0o644)

	arch := filepath.Join(base, "archive")
	m, stats, err := archive.Export(src, arch, 3, 10, "")
	must(err)
	fmt.Printf("exported %s logical=%d physical=%d\n", m.ArchiveID, stats.LogicalBytes, stats.PhysicalBytes)

	vr, err := archive.Verify(arch)
	must(err)
	fmt.Printf("verify %s\n", vr.Outcome)

	must(os.RemoveAll(src))
	fmt.Println("source destroyed")

	restored := filepath.Join(base, "restored")
	m2, err := archive.Import(arch, restored)
	must(err)
	if m2.RootDigest != m.RootDigest {
		panic("digest drift")
	}
	// Re-export restored tree and compare root digest.
	re := filepath.Join(base, "reexport")
	m3, _, err := archive.Export(restored, re, 3, 10, "")
	must(err)
	if m3.RootDigest != m.RootDigest {
		panic("round-trip digest mismatch")
	}
	fmt.Printf("restored digest ok=%s\n", m3.RootDigest[:16])
	fmt.Println("SQLITE RESTORE CYCLE COMPLETE (ENGINE_PROVIDED)")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
