package epoch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Record is a durable epoch assignment for a state resource.
type Record struct {
	Epoch      uint64    `json:"epoch"`
	ResourceID string    `json:"resource_id"`
	HolderID   string    `json:"holder_id"`
	UpdatedAt  time.Time `json:"updated_at"`
	Note       string    `json:"note,omitempty"`
}

// Store persists epoch records under a directory (one JSON file per resource).
type Store struct {
	mu  sync.Mutex
	dir string
}

// Open creates or opens an epoch store at dir.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) path(resourceID string) string {
	safe := filepath.Base(resourceID)
	if safe == "" || safe == "." {
		safe = "default"
	}
	return filepath.Join(s.dir, safe+".json")
}

// Get loads the epoch record for a resource.
func (s *Store) Get(resourceID string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.path(resourceID))
	if err != nil {
		if os.IsNotExist(err) {
			return Record{ResourceID: resourceID, Epoch: 0}, nil
		}
		return Record{}, err
	}
	var r Record
	if err := json.Unmarshal(b, &r); err != nil {
		return Record{}, err
	}
	return r, nil
}

// Advance increments the epoch for resourceID and returns the new record.
func (s *Store) Advance(resourceID, holderID, note string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, err := s.getUnlocked(resourceID)
	if err != nil {
		return Record{}, err
	}
	cur.Epoch++
	cur.ResourceID = resourceID
	cur.HolderID = holderID
	cur.UpdatedAt = time.Now().UTC()
	cur.Note = note
	if err := s.saveUnlocked(cur); err != nil {
		return Record{}, err
	}
	return cur, nil
}

// Pin sets epoch exactly (used by rollback). Refuses to lower silently unless force.
func (s *Store) Pin(resourceID string, epoch uint64, holderID string, force bool) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, err := s.getUnlocked(resourceID)
	if err != nil {
		return Record{}, err
	}
	if epoch < cur.Epoch && !force {
		return Record{}, fmt.Errorf("epoch: refuse pin %d < current %d without force", epoch, cur.Epoch)
	}
	cur.Epoch = epoch
	cur.ResourceID = resourceID
	cur.HolderID = holderID
	cur.UpdatedAt = time.Now().UTC()
	cur.Note = "pin"
	if err := s.saveUnlocked(cur); err != nil {
		return Record{}, err
	}
	return cur, nil
}

func (s *Store) getUnlocked(resourceID string) (Record, error) {
	b, err := os.ReadFile(s.path(resourceID))
	if err != nil {
		if os.IsNotExist(err) {
			return Record{ResourceID: resourceID, Epoch: 0}, nil
		}
		return Record{}, err
	}
	var r Record
	return r, json.Unmarshal(b, &r)
}

func (s *Store) saveUnlocked(r Record) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(r.ResourceID), b, 0o644)
}

// Compare returns -1 if a<b, 0 if equal, 1 if a>b.
func Compare(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// IsStale reports whether observed is behind required.
func IsStale(observed, required uint64) bool {
	return observed < required
}
