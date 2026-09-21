// Package embed exposes a headless library entry for control-plane hosts.
package embed

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/coordinator"
	"github.com/theworker02/stateful-deployments-engine/internal/fence"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/store"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// Mode documents how the host process is using SDE.
type Mode string

const (
	ModeCLI          Mode = "CLI"
	ModeLibrary      Mode = "LIBRARY"
	ModeControlPlane Mode = "CONTROL-PLANE"
)

// Engine is a headless transactional deploy handle (no TTY).
type Engine struct {
	Mode   Mode
	Root   string
	Coord  *coordinator.Coordinator
	Journal *journal.Journal
}

// OpenLocal constructs a LIBRARY/CONTROL-PLANE engine on the local adapter.
func OpenLocal(root string, mode Mode) (*Engine, error) {
	if mode == "" {
		mode = ModeLibrary
	}
	plat, err := local.Open(root)
	if err != nil {
		return nil, err
	}
	j, err := journal.Open(filepath.Join(root, "journal"), 1)
	if err != nil {
		return nil, err
	}
	plat.BindJournal(j)
	fm, err := fence.Open(root, "embed")
	if err != nil {
		_ = j.Close()
		return nil, err
	}
	st := store.New(root)
	c := coordinator.NewFull(plat, plat, j, fm, st, nil)
	return &Engine{Mode: mode, Root: root, Coord: c, Journal: j}, nil
}

// Deploy runs a transactional deploy (or dry-run).
func (e *Engine) Deploy(ctx context.Context, image string, dryRun bool, maxPause time.Duration) (*types.DeployReport, error) {
	if e == nil || e.Coord == nil {
		return nil, fmt.Errorf("engine not open")
	}
	if maxPause <= 0 {
		maxPause = 250 * time.Millisecond
	}
	return e.Coord.Deploy(ctx, coordinator.Config{
		ImageRef:      image,
		DryRun:        dryRun,
		MaxWritePause: maxPause,
		SDERoot:       e.Root,
	})
}

// Close releases journal resources.
func (e *Engine) Close() error {
	if e.Journal != nil {
		return e.Journal.Close()
	}
	return nil
}
