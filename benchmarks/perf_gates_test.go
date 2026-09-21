package bench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/chunk"
)

// TestPerfRegressionGates checks coarse throughput/latency gates (not micro-flaky timings).
func TestPerfRegressionGates(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	for i := 0; i < 20; i++ {
		_ = os.WriteFile(filepath.Join(src, filepath.Join("f"+string(rune('a'+i%26))+".bin")), make([]byte, 8192), 0o644)
	}
	// fix filenames properly
	_ = os.RemoveAll(src)
	_ = os.MkdirAll(src, 0o755)
	for i := 0; i < 20; i++ {
		name := filepath.Join(src, "f-"+itoa(i)+".bin")
		_ = os.WriteFile(name, make([]byte, 8192), 0o644)
	}
	start := time.Now()
	n, err := chunk.FullCopy(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	fullMs := float64(time.Since(start).Milliseconds())
	if n < 20*8192 {
		t.Fatalf("copied %d", n)
	}
	// Gate: full copy of ~160KiB should finish well under 5s on CI hardware.
	if fullMs > 5000 {
		t.Fatalf("full copy regression: %.0fms > 5000ms", fullMs)
	}
	// Differential after one file change should touch fewer bytes than full logical size.
	_ = os.WriteFile(filepath.Join(src, "f-0.bin"), append(make([]byte, 8192), 1), 0o644)
	si, _, _ := chunk.BuildIndex(src, chunk.DefaultBlockSize)
	di, _, _ := chunk.BuildIndex(dst, chunk.DefaultBlockSize)
	ch := chunk.Diff(si, di)
	dstart := time.Now()
	db, err := chunk.TransferChanged(src, dst, ch)
	if err != nil {
		t.Fatal(err)
	}
	diffMs := float64(time.Since(dstart).Milliseconds())
	if db >= n && len(ch) > 0 {
		t.Fatalf("diff should copy less than full when few blocks change: diff=%d full=%d", db, n)
	}
	if diffMs > 5000 {
		t.Fatalf("diff regression: %.0fms", diffMs)
	}
	out := map[string]interface{}{
		"full_ms": fullMs, "full_bytes": n, "diff_ms": diffMs, "diff_bytes": db, "changed_blocks": len(ch),
	}
	_ = os.MkdirAll(filepath.Join("..", "benchmarks", "results"), 0o755)
	b, _ := json.MarshalIndent(out, "", "  ")
	_ = os.WriteFile(filepath.Join("..", "benchmarks", "results", "perf_gates.json"), b, 0o644)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var d []byte
	for i > 0 {
		d = append([]byte{byte('0' + i%10)}, d...)
		i /= 10
	}
	return string(d)
}
