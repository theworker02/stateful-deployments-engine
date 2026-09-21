// Package manifest builds cryptographically verifiable state manifests (Merkle-ish digests).
package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

const AlgoSHA256Sorted = "sha256-sorted-path-digest-v1"

// Builder walks a state root and produces a StateManifest.
type Builder struct {
	SkipPrefixes []string // default: .sde/
}

// Build constructs a StateManifest for root at the given epoch/journal position.
func (b Builder) Build(root string, epoch, journalPos uint64) (*types.StateManifest, error) {
	skips := b.SkipPrefixes
	if len(skips) == 0 {
		skips = []string{".sde/"}
	}
	digests := make(map[string]string)
	var total int64
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		for _, p := range skips {
			if strings.HasPrefix(rel, p) {
				return nil
			}
		}
		sum, n, err := hashFile(path)
		if err != nil {
			return err
		}
		digests[rel] = sum
		total += n
		return nil
	})
	if err != nil {
		return nil, err
	}
	merkle := merkleRoot(digests)
	rootDigest := sha256.Sum256([]byte(merkle))
	return &types.StateManifest{
		Epoch:            epoch,
		ObjectCount:      len(digests),
		TotalBytes:       total,
		RootDigest:       hex.EncodeToString(rootDigest[:]),
		JournalPosition:  journalPos,
		VerificationAlgo: AlgoSHA256Sorted,
		MerkleRoot:       merkle,
		ObjectDigests:    digests,
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func hashFile(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// merkleRoot sorts path:digest pairs and hashes the concatenation.
func merkleRoot(digests map[string]string) string {
	keys := make([]string, 0, len(digests))
	for k := range digests {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write([]byte(digests[k]))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Equal compares two manifests by Merkle root (and object count).
func Equal(a, b *types.StateManifest) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.MerkleRoot == b.MerkleRoot && a.ObjectCount == b.ObjectCount && a.TotalBytes == b.TotalBytes
}

// DiffPaths returns paths whose digests differ (or are missing).
func DiffPaths(a, b *types.StateManifest) (missing, extra, mismatch []string) {
	if a == nil || b == nil {
		return nil, nil, nil
	}
	for p, ha := range a.ObjectDigests {
		hb, ok := b.ObjectDigests[p]
		if !ok {
			missing = append(missing, p)
			continue
		}
		if ha != hb {
			mismatch = append(mismatch, p)
		}
	}
	for p := range b.ObjectDigests {
		if _, ok := a.ObjectDigests[p]; !ok {
			extra = append(extra, p)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	sort.Strings(mismatch)
	return missing, extra, mismatch
}

// WriteJSON persists a manifest.
func WriteJSON(path string, m *types.StateManifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
