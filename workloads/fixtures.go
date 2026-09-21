package workloads

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FixtureKind names universal volume workload fixtures (no DB semantics required).
type FixtureKind string

const (
	FixtureSQLiteLike FixtureKind = "sqlite-like"
	FixtureMedia      FixtureKind = "media"
	FixtureAppendLog  FixtureKind = "append-log"
	FixtureKVFS       FixtureKind = "kv-fs"
	FixtureGenericDir FixtureKind = "generic-dir"
)

// FixtureMeta is written beside generated trees for evaluate harnesses.
type FixtureMeta struct {
	Kind      FixtureKind `json:"kind"`
	CreatedAt time.Time   `json:"created_at"`
	Files     int         `json:"files"`
	Bytes     int64       `json:"bytes"`
	Note      string      `json:"note"`
}

// MaterializeFixture builds a realistic-ish volume tree for the kind.
func MaterializeFixture(root string, kind FixtureKind) (*FixtureMeta, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	meta := &FixtureMeta{Kind: kind, CreatedAt: time.Now().UTC(), Note: "universal volume fixture; no DB engine required"}
	switch kind {
	case FixtureSQLiteLike:
		// Single large "db" file + wal/shm companions (opaque bytes).
		n := writeFile(filepath.Join(root, "app.db"), sized(256*1024, 0x11))
		n += writeFile(filepath.Join(root, "app.db-wal"), sized(64*1024, 0x22))
		n += writeFile(filepath.Join(root, "app.db-shm"), sized(32*1024, 0x33))
		meta.Files, meta.Bytes = 3, n
	case FixtureMedia:
		_ = os.MkdirAll(filepath.Join(root, "media"), 0o755)
		var n int64
		for i := 0; i < 8; i++ {
			n += writeFile(filepath.Join(root, "media", fmt.Sprintf("clip-%02d.bin", i)), sized(128*1024, byte(i)))
		}
		meta.Files, meta.Bytes = 8, n
	case FixtureAppendLog:
		_ = os.MkdirAll(filepath.Join(root, "logs"), 0o755)
		var n int64
		for i := 0; i < 3; i++ {
			n += writeFile(filepath.Join(root, "logs", fmt.Sprintf("app-%d.log", i)), sized(96*1024, 0xab))
		}
		meta.Files, meta.Bytes = 3, n
	case FixtureKVFS:
		_ = os.MkdirAll(filepath.Join(root, "kv"), 0o755)
		var n int64
		for i := 0; i < 40; i++ {
			n += writeFile(filepath.Join(root, "kv", fmt.Sprintf("k-%04d", i)), []byte(fmt.Sprintf(`{"v":%d}`, i)))
		}
		meta.Files, meta.Bytes = 40, n
	case FixtureGenericDir:
		_ = os.MkdirAll(filepath.Join(root, "data", "nested"), 0o755)
		n := writeFile(filepath.Join(root, "data", "readme.txt"), []byte("generic volume tree\n"))
		n += writeFile(filepath.Join(root, "data", "nested", "blob.bin"), sized(16*1024, 0x5a))
		meta.Files, meta.Bytes = 2, n
	default:
		return nil, fmt.Errorf("unknown fixture %q", kind)
	}
	b, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(root, "FIXTURE.json"), b, 0o644)
	return meta, nil
}

func sized(n int, fill byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = fill
	}
	return b
}

func writeFile(path string, data []byte) int64 {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o644)
	return int64(len(data))
}

// AllFixtureKinds lists fixtures for evaluate loops.
func AllFixtureKinds() []FixtureKind {
	return []FixtureKind{FixtureSQLiteLike, FixtureMedia, FixtureAppendLog, FixtureKVFS, FixtureGenericDir}
}
