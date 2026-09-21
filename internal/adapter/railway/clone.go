package railway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/sensitive"
)

// CloneMode selects read-oriented vs writable clone materialization.
type CloneMode string

const (
	CloneReadOnly CloneMode = "read-only"
	CloneWritable CloneMode = "writable"
)

// CloneRequest orchestrates ENGINE_PROVIDED volume clone semantics (local or staged).
// Live GraphQL orchestration is available when SDE_RAILWAY_LIVE=1; clone itself uses PSA mechanics.
type CloneRequest struct {
	SourceState string
	DestDir     string // clone output (archive + optional materialization)
	Mode        CloneMode
	Sanitize    bool
	Epoch       uint64
	JournalPos  uint64
	JournalPath string
}

// CloneReceipt is written as CLONE_RECEIPT.json.
type CloneReceipt struct {
	ReceiptID       string    `json:"receipt_id"`
	CreatedAt       time.Time `json:"created_at"`
	Mode            CloneMode `json:"mode"`
	Portability     string    `json:"portability"` // ENGINE_PROVIDED
	SourcePath      string    `json:"source_path"`
	ArchivePath     string    `json:"archive_path"`
	MaterializedPath string   `json:"materialized_path,omitempty"`
	ArchiveID       string    `json:"archive_id"`
	RootDigest      string    `json:"root_digest"`
	Sanitized       bool      `json:"sanitized"`
	Verified        bool      `json:"verified"`
	NewIdentity     string    `json:"new_identity"`
	EvidenceHash    string    `json:"evidence_hash"`
	Note            string    `json:"note"`
	Error           string    `json:"error,omitempty"`
}

// CloneVolume exports source state to a new-identity archive (and optional materialization).
// This is ENGINE_PROVIDED orchestration around the generic core — not RAILWAY_NATIVE backup clone.
func (a *Adapter) CloneVolume(ctx context.Context, req CloneRequest) (*CloneReceipt, error) {
	_ = ctx
	if req.Mode == "" {
		req.Mode = CloneReadOnly
	}
	r := &CloneReceipt{
		ReceiptID:   fmt.Sprintf("clone-%d", time.Now().UnixNano()),
		CreatedAt:   time.Now().UTC(),
		Mode:        req.Mode,
		Portability: string(ModeEngineProvided),
		SourcePath:  req.SourceState,
		Sanitized:   req.Sanitize,
		Note:        "ENGINE_PROVIDED clone; not RAILWAY_NATIVE; not Railway-endorsed",
	}
	src := req.SourceState
	if req.Sanitize {
		san := filepath.Join(req.DestDir, ".sanitize-src")
		if err := copyTreeFiltered(src, san, sensitive.Pipeline{Hooks: []sensitive.Hook{sensitive.RedactEmails{}, sensitive.TokenizeSecrets{}}}); err != nil {
			r.Error = err.Error()
			return sealClone(r, req.DestDir), err
		}
		src = san
	}
	archDir := filepath.Join(req.DestDir, "archive")
	m, _, err := archive.Export(src, archDir, req.Epoch, req.JournalPos, req.JournalPath)
	if err != nil {
		r.Error = err.Error()
		return sealClone(r, req.DestDir), err
	}
	r.ArchivePath = archDir
	r.ArchiveID = m.ArchiveID
	r.RootDigest = m.RootDigest
	r.NewIdentity = m.ArchiveID

	if _, err := archive.Verify(archDir); err != nil {
		r.Error = err.Error()
		return sealClone(r, req.DestDir), err
	}
	r.Verified = true

	if req.Mode == CloneWritable || req.Mode == CloneReadOnly {
		mat := filepath.Join(req.DestDir, "materialized")
		if _, err := archive.Import(archDir, mat); err != nil {
			r.Error = err.Error()
			return sealClone(r, req.DestDir), err
		}
		r.MaterializedPath = mat
		if req.Mode == CloneReadOnly {
			_ = filepath.WalkDir(mat, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				return os.Chmod(path, 0o444)
			})
		}
	}
	return sealClone(r, req.DestDir), nil
}

func sealClone(r *CloneReceipt, dir string) *CloneReceipt {
	r.EvidenceHash = ""
	b, _ := json.Marshal(r)
	sum := sha256.Sum256(b)
	r.EvidenceHash = hex.EncodeToString(sum[:])
	if dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		out, _ := json.MarshalIndent(r, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "CLONE_RECEIPT.json"), out, 0o644)
	}
	return r
}

func copyTreeFiltered(src, dst string, pipe sensitive.Pipeline) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		b, err = pipe.Apply(rel, b)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}
