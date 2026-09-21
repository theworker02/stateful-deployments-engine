// Package replay applies journaled mutations onto a shadow state tree.
package replay

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// ApplyMutations replays an ordered mutation list onto root.
func ApplyMutations(root string, mutations []types.Mutation) (applied uint64, err error) {
	for _, m := range mutations {
		if err := applyOne(root, m); err != nil {
			return applied, fmt.Errorf("seq %d %s %s: %w", m.Seq, m.Kind, m.Path, err)
		}
		applied++
	}
	return applied, nil
}

func applyOne(root string, m types.Mutation) error {
	path := filepath.Join(root, filepath.FromSlash(m.Path))
	switch m.Kind {
	case types.OpCreate, types.OpWrite, types.OpTrunc:
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if m.Mode != 0 {
			mode = os.FileMode(m.Mode)
		}
		return os.WriteFile(path, m.Payload, mode)
	case types.OpDelete:
		return os.RemoveAll(path)
	case types.OpRename:
		dest := filepath.Join(root, filepath.FromSlash(m.DestPath))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.Rename(path, dest)
	case types.OpChmod:
		mode := os.FileMode(m.Mode)
		if mode == 0 {
			mode = 0o644
		}
		return os.Chmod(path, mode)
	default:
		return fmt.Errorf("unknown op kind %q", m.Kind)
	}
}
