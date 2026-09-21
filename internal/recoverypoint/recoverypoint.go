// Package recoverypoint tracks SNAPSHOT + journal-chain continuous recovery points.
// Granularity is epoch/journal position — never advertise finer PITR than supported.
package recoverypoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
)

// Point is a recoverable snapshot reference.
type Point struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Epoch           uint64    `json:"epoch"`
	JournalPosition uint64    `json:"journal_position"`
	ArchivePath     string    `json:"archive_path"`
	RootDigest      string    `json:"root_digest"`
	Kind            string    `json:"kind"` // SNAPSHOT|JOURNAL_CHAIN
	Note            string    `json:"note"`
}

// Store persists recovery points under root/recovery-points.json.
type Store struct {
	Root string
}

func (s *Store) path() string {
	return filepath.Join(s.Root, "recovery-points.json")
}

func (s *Store) load() ([]Point, error) {
	b, err := os.ReadFile(s.path())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var pts []Point
	return pts, json.Unmarshal(b, &pts)
}

func (s *Store) save(pts []Point) error {
	if err := os.MkdirAll(s.Root, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(pts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), b, 0o644)
}

// CreateFromArchive registers a SNAPSHOT recovery point.
func (s *Store) CreateFromArchive(archiveDir string) (*Point, error) {
	m, err := archive.Inspect(archiveDir)
	if err != nil {
		return nil, err
	}
	p := Point{
		ID:              fmt.Sprintf("rp-%d", time.Now().UnixNano()),
		CreatedAt:       time.Now().UTC(),
		Epoch:           m.SourceEpoch,
		JournalPosition: m.JournalPosition,
		ArchivePath:     archiveDir,
		RootDigest:      m.RootDigest,
		Kind:            "SNAPSHOT",
		Note:            "Recoverable to archive epoch/journal position only — not sub-mutation PITR",
	}
	pts, _ := s.load()
	pts = append(pts, p)
	if err := s.save(pts); err != nil {
		return nil, err
	}
	return &p, nil
}

// List returns points newest first.
func (s *Store) List() ([]Point, error) {
	pts, err := s.load()
	if err != nil {
		return nil, err
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].CreatedAt.After(pts[j].CreatedAt) })
	return pts, nil
}

// FindByEpoch returns the newest point at or before epoch.
func (s *Store) FindByEpoch(epoch uint64) (*Point, error) {
	pts, err := s.List()
	if err != nil {
		return nil, err
	}
	for i := range pts {
		if pts[i].Epoch <= epoch {
			return &pts[i], nil
		}
	}
	return nil, fmt.Errorf("no recovery point at or before epoch %d", epoch)
}
