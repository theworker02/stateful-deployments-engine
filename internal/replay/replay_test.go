package replay_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/replay"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestApplyWriteAndDelete(t *testing.T) {
	root := t.TempDir()
	n, err := replay.ApplyMutations(root, []types.Mutation{
		{Seq: 1, Kind: types.OpWrite, Path: "a.txt", Payload: []byte("hello")},
		{Seq: 2, Kind: types.OpDelete, Path: "a.txt"},
	})
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); !os.IsNotExist(err) {
		t.Fatal("file should be gone")
	}
}

func TestUnknownOpFails(t *testing.T) {
	_, err := replay.ApplyMutations(t.TempDir(), []types.Mutation{
		{Seq: 1, Kind: types.OpKind("nope"), Path: "x"},
	})
	if err == nil {
		t.Fatal("expected error for unknown op")
	}
}
