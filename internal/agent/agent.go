// Package agent records application filesystem mutations into the journal.
//
// Production deployments would typically run this as a sidecar or link
// applications against a small SDK that journals writes before they hit disk.
// The local agent is used by demos and the bench harness.
package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Writer applies mutations to an active state directory and journals them.
type Writer struct {
	Root    string
	Journal *journal.Journal
	Local   *local.Adapter // optional; if set, respects write barriers
}

// WriteFile creates/overwrites a file under Root and appends a journal entry.
func (w *Writer) WriteFile(ctx context.Context, rel string, data []byte, mode os.FileMode) (uint64, error) {
	if w.Local != nil {
		if err := w.Local.BeginWrite(ctx); err != nil {
			return 0, err
		}
		defer w.Local.EndWrite()
	}
	path := filepath.Join(w.Root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}
	if mode == 0 {
		mode = 0o644
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		return 0, err
	}
	sum := sha256.Sum256(data)
	return w.Journal.Append(types.Mutation{
		Kind:        types.OpWrite,
		Path:        filepath.ToSlash(rel),
		Mode:        uint32(mode),
		Size:        int64(len(data)),
		ContentHash: hex.EncodeToString(sum[:]),
		Payload:     data,
		Timestamp:   time.Now().UTC(),
	})
}

// Delete removes a path and journals the delete.
func (w *Writer) Delete(ctx context.Context, rel string) (uint64, error) {
	if w.Local != nil {
		if err := w.Local.BeginWrite(ctx); err != nil {
			return 0, err
		}
		defer w.Local.EndWrite()
	}
	path := filepath.Join(w.Root, filepath.FromSlash(rel))
	if err := os.RemoveAll(path); err != nil {
		return 0, err
	}
	return w.Journal.Append(types.Mutation{
		Kind:      types.OpDelete,
		Path:      filepath.ToSlash(rel),
		Timestamp: time.Now().UTC(),
	})
}

// Rename journals and performs a rename within Root.
func (w *Writer) Rename(ctx context.Context, from, to string) (uint64, error) {
	if w.Local != nil {
		if err := w.Local.BeginWrite(ctx); err != nil {
			return 0, err
		}
		defer w.Local.EndWrite()
	}
	src := filepath.Join(w.Root, filepath.FromSlash(from))
	dst := filepath.Join(w.Root, filepath.FromSlash(to))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, err
	}
	if err := os.Rename(src, dst); err != nil {
		return 0, err
	}
	return w.Journal.Append(types.Mutation{
		Kind:      types.OpRename,
		Path:      filepath.ToSlash(from),
		DestPath:  filepath.ToSlash(to),
		Timestamp: time.Now().UTC(),
	})
}
