package chunk

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestBuildIndexDiffTransfer(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.bin"), []byte("abcdefghijklmnopqrstuvwxyz"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "a.bin"), []byte("abcdefghijklmnopqrstuvwxyz"), 0o644); err != nil {
		t.Fatal(err)
	}
	si, _, err := BuildIndex(src, 8)
	if err != nil {
		t.Fatal(err)
	}
	di, _, err := BuildIndex(dst, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(Diff(si, di)) != 0 {
		t.Fatal("expected identical")
	}
	if err := os.WriteFile(filepath.Join(src, "a.bin"), []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"), 0o644); err != nil {
		t.Fatal(err)
	}
	si2, _, err := BuildIndex(src, 8)
	if err != nil {
		t.Fatal(err)
	}
	changed := Diff(si2, di)
	if len(changed) == 0 {
		t.Fatal("expected changes")
	}
	n, err := TransferChanged(src, dst, changed)
	if err != nil || n == 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

// TestWriteDifferentialBenchResults measures full vs differential transfer on a
// synthetic tree and writes measured numbers under benchmarks/results/.
func TestWriteDifferentialBenchResults(t *testing.T) {
	src := t.TempDir()
	_ = os.MkdirAll(filepath.Join(src, "obj"), 0o755)
	const files = 80
	const fileBytes = 16 * 1024
	for i := 0; i < files; i++ {
		buf := make([]byte, fileBytes)
		for j := range buf {
			buf[j] = byte((i + j) % 251)
		}
		path := filepath.Join(src, "obj", "file-"+strconv.Itoa(i)+".bin")
		if err := os.WriteFile(path, buf, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	dstFull := t.TempDir()
	tFull := time.Now()
	fullBytes, err := FullCopy(src, dstFull)
	if err != nil {
		t.Fatal(err)
	}
	fullMs := float64(time.Since(tFull).Microseconds()) / 1000

	dstDiff := t.TempDir()
	if _, err := FullCopy(src, dstDiff); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		buf := make([]byte, fileBytes)
		for j := range buf {
			buf[j] = byte(255 - ((i + j) % 251))
		}
		if err := os.WriteFile(filepath.Join(src, "obj", "file-"+strconv.Itoa(i)+".bin"), buf, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	si, _, err := BuildIndex(src, DefaultBlockSize)
	if err != nil {
		t.Fatal(err)
	}
	di, _, err := BuildIndex(dstDiff, DefaultBlockSize)
	if err != nil {
		t.Fatal(err)
	}
	changed := Diff(si, di)
	tDiff := time.Now()
	diffBytes, err := TransferChanged(src, dstDiff, changed)
	if err != nil {
		t.Fatal(err)
	}
	diffMs := float64(time.Since(tDiff).Microseconds()) / 1000

	outDir := filepath.Join("..", "..", "benchmarks", "results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := map[string]interface{}{
		"generated_at":   time.Now().UTC(),
		"suite":          "chunk_differential_v041",
		"block_size":     DefaultBlockSize,
		"files":          files,
		"file_bytes":     fileBytes,
		"full_copy_ms":   fullMs,
		"full_bytes":     fullBytes,
		"diff_ms":        diffMs,
		"diff_bytes":     diffBytes,
		"changed_blocks": len(changed),
		"note":           "measured locally by TestWriteDifferentialBenchResults; not fabricated",
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	path := filepath.Join(outDir, "chunk_differential.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s full_ms=%.3f diff_ms=%.3f changed=%d", path, fullMs, diffMs, len(changed))
}
