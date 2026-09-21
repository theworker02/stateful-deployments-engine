// Package verifier builds content-addressed manifests and verification receipts.
package verifier

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// BuildCheckpoint walks root and produces a content-addressed checkpoint.
func BuildCheckpoint(root string, epoch uint64) (*types.Checkpoint, error) {
	manifest := make(map[string]string)
	var totalBytes int64

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
		if strings.HasPrefix(rel, ".sde/") {
			return nil
		}
		h, size, err := hashFile(path)
		if err != nil {
			return err
		}
		manifest[rel] = h
		totalBytes += size
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.Checkpoint{
		Epoch:     epoch,
		RootHash:  rootHash(manifest),
		FileCount: len(manifest),
		ByteSize:  totalBytes,
		CreatedAt: time.Now().UTC(),
		Manifest:  manifest,
	}, nil
}

// BuildManifest produces a cryptographically verifiable state manifest (Merkle).
func BuildManifest(root string, epoch, journalPos uint64) (*types.StateManifest, error) {
	cp, err := BuildCheckpoint(root, epoch)
	if err != nil {
		return nil, err
	}
	return &types.StateManifest{
		Epoch:            epoch,
		ObjectCount:      cp.FileCount,
		TotalBytes:       cp.ByteSize,
		RootDigest:       cp.RootHash,
		JournalPosition:  journalPos,
		VerificationAlgo: "sha256-manifest-v1",
		MerkleRoot:       merkleRoot(cp.Manifest),
		ObjectDigests:    cp.Manifest,
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

func rootHash(manifest map[string]string) string {
	keys := make([]string, 0, len(manifest))
	for k := range manifest {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s:%s\n", k, manifest[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func merkleRoot(manifest map[string]string) string {
	keys := make([]string, 0, len(manifest))
	for k := range manifest {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		sum := sha256.Sum256(nil)
		return hex.EncodeToString(sum[:])
	}
	level := make([][]byte, len(keys))
	for i, k := range keys {
		leaf := sha256.Sum256([]byte(k + ":" + manifest[k]))
		level[i] = leaf[:]
	}
	for len(level) > 1 {
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, level[i])
				continue
			}
			h := sha256.New()
			h.Write(level[i])
			h.Write(level[i+1])
			next = append(next, h.Sum(nil))
		}
		level = next
	}
	return hex.EncodeToString(level[0])
}

// Result is the outcome of comparing two state trees.
type Result struct {
	OK              bool
	ActiveHash      string
	ShadowHash      string
	MissingInShadow []string
	ExtraInShadow   []string
	HashMismatch    []string
}

// CompareManifests verifies shadow is a consistent replica of active.
func CompareManifests(active, shadow *types.Checkpoint) Result {
	r := Result{
		ActiveHash: active.RootHash,
		ShadowHash: shadow.RootHash,
		OK:         active.RootHash == shadow.RootHash,
	}
	for path, ah := range active.Manifest {
		sh, ok := shadow.Manifest[path]
		if !ok {
			r.MissingInShadow = append(r.MissingInShadow, path)
			r.OK = false
			continue
		}
		if ah != sh {
			r.HashMismatch = append(r.HashMismatch, path)
			r.OK = false
		}
	}
	for path := range shadow.Manifest {
		if _, ok := active.Manifest[path]; !ok {
			r.ExtraInShadow = append(r.ExtraInShadow, path)
			r.OK = false
		}
	}
	sort.Strings(r.MissingInShadow)
	sort.Strings(r.ExtraInShadow)
	sort.Strings(r.HashMismatch)
	return r
}

// CompareRoots builds checkpoints for both roots and compares them.
func CompareRoots(activeRoot, shadowRoot string, epoch uint64) (Result, *types.Checkpoint, *types.Checkpoint, error) {
	a, err := BuildCheckpoint(activeRoot, epoch)
	if err != nil {
		return Result{}, nil, nil, err
	}
	s, err := BuildCheckpoint(shadowRoot, epoch)
	if err != nil {
		return Result{}, nil, nil, err
	}
	return CompareManifests(a, s), a, s, nil
}

// Options configure a leveled verification.
type Options struct {
	DeployID   string
	Level      types.VerifyLevel
	AppCheck   func(activeRoot, candidateRoot string) (bool, string, error) // APPLICATION_DEFINED
}

// Verify produces a VerificationReceipt. UNKNOWN is never promoted to VERIFIED.
func Verify(activeRoot, candidateRoot string, epoch uint64, opt Options) (*types.VerificationReceipt, error) {
	start := time.Now()
	if opt.Level == "" {
		opt.Level = types.VerifyChecksum
	}
	receipt := &types.VerificationReceipt{
		DeployID:  opt.DeployID,
		Level:     opt.Level,
		Epoch:     epoch,
		CheckedAt: time.Now().UTC(),
		Outcome:   types.VerifyUnknown,
	}

	aInfo, err := os.Stat(activeRoot)
	cInfo, err2 := os.Stat(candidateRoot)
	if err != nil || err2 != nil || !aInfo.IsDir() || !cInfo.IsDir() {
		receipt.Outcome = types.VerifyUnknown
		receipt.EvidenceSummary = "roots inaccessible; UNKNOWN must not be treated as VERIFIED"
		receipt.DurationMs = float64(time.Since(start).Microseconds()) / 1000
		return receipt, nil
	}

	switch opt.Level {
	case types.VerifyMetadata:
		a, e1 := BuildCheckpoint(activeRoot, epoch)
		c, e2 := BuildCheckpoint(candidateRoot, epoch)
		if e1 != nil || e2 != nil {
			receipt.Outcome = types.VerifyUnknown
			receipt.EvidenceSummary = fmt.Sprintf("metadata walk failed: %v %v", e1, e2)
			break
		}
		receipt.FileCountActive = a.FileCount
		receipt.FileCountCand = c.FileCount
		receipt.BytesCompared = a.ByteSize
		if a.FileCount == c.FileCount && a.ByteSize == c.ByteSize {
			receipt.Outcome = types.VerifyPartial // size match alone is never full VERIFIED
			receipt.EvidenceSummary = "metadata counts/bytes match; checksum not performed (PARTIAL)"
		} else {
			receipt.Outcome = types.VerifyFailed
			receipt.EvidenceSummary = "metadata counts or bytes differ"
		}

	case types.VerifyChecksum, types.VerifyStructural:
		res, a, c, err := CompareRoots(activeRoot, candidateRoot, epoch)
		if err != nil {
			receipt.Outcome = types.VerifyUnknown
			receipt.EvidenceSummary = err.Error()
			break
		}
		receipt.ActiveRootHash = a.RootHash
		receipt.CandidateHash = c.RootHash
		receipt.FileCountActive = a.FileCount
		receipt.FileCountCand = c.FileCount
		receipt.BytesCompared = a.ByteSize
		receipt.MissingPaths = res.MissingInShadow
		receipt.ExtraPaths = res.ExtraInShadow
		receipt.HashMismatches = res.HashMismatch
		if opt.Level == types.VerifyStructural {
			notes := structuralNotes(a, c)
			receipt.StructuralNotes = notes
		}
		if res.OK {
			receipt.Outcome = types.VerifyVerified
			receipt.EvidenceSummary = fmt.Sprintf("checksum match root=%s", short(a.RootHash))
		} else if len(res.HashMismatch)+len(res.MissingInShadow) > 0 && len(res.HashMismatch) < a.FileCount {
			receipt.Outcome = types.VerifyPartial
			receipt.EvidenceSummary = "partial overlap with mismatches"
		} else {
			receipt.Outcome = types.VerifyFailed
			receipt.EvidenceSummary = "checksum verification failed"
		}

	case types.VerifyApplicationDefined:
		res, a, c, err := CompareRoots(activeRoot, candidateRoot, epoch)
		if err != nil {
			receipt.Outcome = types.VerifyUnknown
			receipt.EvidenceSummary = err.Error()
			break
		}
		receipt.ActiveRootHash = a.RootHash
		receipt.CandidateHash = c.RootHash
		receipt.FileCountActive = a.FileCount
		receipt.FileCountCand = c.FileCount
		receipt.BytesCompared = a.ByteSize
		if opt.AppCheck == nil {
			receipt.Outcome = types.VerifyUnknown
			receipt.EvidenceSummary = "APPLICATION_DEFINED check not provided; UNKNOWN"
			break
		}
		ok, note, err := opt.AppCheck(activeRoot, candidateRoot)
		if err != nil {
			receipt.Outcome = types.VerifyUnknown
			receipt.EvidenceSummary = err.Error()
			break
		}
		receipt.AppDefinedOK = &ok
		if ok && res.OK {
			receipt.Outcome = types.VerifyVerified
			receipt.EvidenceSummary = "app-defined + checksum OK: " + note
		} else if ok {
			receipt.Outcome = types.VerifyPartial
			receipt.EvidenceSummary = "app-defined OK but checksum drift: " + note
			receipt.HashMismatches = res.HashMismatch
		} else {
			receipt.Outcome = types.VerifyFailed
			receipt.EvidenceSummary = "app-defined check failed: " + note
		}

	default:
		receipt.Outcome = types.VerifyUnknown
		receipt.EvidenceSummary = fmt.Sprintf("unrecognized verify level %q", opt.Level)
	}

	receipt.DurationMs = float64(time.Since(start).Microseconds()) / 1000
	return receipt, nil
}

func structuralNotes(a, c *types.Checkpoint) []string {
	var notes []string
	if a.FileCount != c.FileCount {
		notes = append(notes, fmt.Sprintf("file_count active=%d candidate=%d", a.FileCount, c.FileCount))
	}
	if a.ByteSize != c.ByteSize {
		notes = append(notes, fmt.Sprintf("byte_size active=%d candidate=%d", a.ByteSize, c.ByteSize))
	}
	return notes
}

func short(h string) string {
	if len(h) <= 12 {
		return h
	}
	return h[:12]
}

// IsVerified is the only safe promotion path; UNKNOWN/PARTIAL/FAILED are not VERIFIED.
func IsVerified(r *types.VerificationReceipt) bool {
	return r != nil && r.Outcome == types.VerifyVerified
}
