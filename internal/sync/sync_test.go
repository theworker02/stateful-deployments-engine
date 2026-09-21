package sync_test

import (
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/sync"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestMeasureReadyOnlyWhenLagZeroAndIntegrityOK(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(filepath.Join(dir, "j"), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	eng := &sync.Engine{Journal: j}
	m, err := eng.Measure(dir, j.NextSeq(), 100, "ok")
	if err != nil {
		t.Fatal(err)
	}
	if !m.ReadyForCutover {
		t.Fatal("expected ready with empty lag + ok integrity")
	}
	_, _ = j.Append(types.Mutation{Kind: types.OpWrite, Path: "x", Payload: []byte("1"), Size: 1})
	m2, err := eng.Measure(dir, 1, 100, "ok")
	if err != nil {
		t.Fatal(err)
	}
	if m2.ReadyForCutover {
		t.Fatal("must not be ready with journal lag")
	}
	m3, err := eng.Measure(dir, j.NextSeq(), 100, "unknown")
	if err != nil {
		t.Fatal(err)
	}
	if m3.ReadyForCutover {
		t.Fatal("unknown integrity must not be ready")
	}
}
