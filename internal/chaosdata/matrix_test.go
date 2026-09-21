package chaosdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type matrix struct {
	Version    int `json:"version"`
	Scenarios  []struct {
		ID string `json:"id"`
	} `json:"scenarios"`
}

func TestMatrixPresent(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	// testdata/chaos/matrix.json relative to repo root: walk up from this test file
	// This test lives in internal/chaosdata — find repo root via go.mod walk.
	dir := filepath.Dir(file)
	var root string
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			root = dir
			break
		}
		dir = filepath.Dir(dir)
	}
	if root == "" {
		t.Fatal("go.mod not found")
	}
	b, err := os.ReadFile(filepath.Join(root, "testdata", "chaos", "matrix.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m matrix
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Scenarios) < 5 {
		t.Fatalf("scenarios=%d", len(m.Scenarios))
	}
}
