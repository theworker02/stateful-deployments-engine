// Package archive implements the Portable State Archive (PSA) format —
// deterministic, checksummed, streamable, resumable, chunked, platform-neutral.
//
// This is ENGINE_PROVIDED portability. It is NOT Railway-native backup and
// must never be described as endorsed by Railway.
package archive

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/chunk"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

const (
	FormatMagic   = "SDEPSA"
	FormatVersion = 1
)

// PortabilityKind distinguishes engine vs platform-native recovery.
type PortabilityKind string

const (
	EngineProvided PortabilityKind = "ENGINE_PROVIDED"
	RailwayNative  PortabilityKind = "RAILWAY_NATIVE"
)

// Manifest is the archive root descriptor (archive.json).
type Manifest struct {
	Magic              string            `json:"magic"`
	FormatVersion      int               `json:"format_version"`
	MinReaderVersion   int               `json:"min_reader_version"`
	ArchiveID          string            `json:"archive_id"`
	CreatedAt          time.Time         `json:"created_at"`
	Portability        PortabilityKind   `json:"portability"`
	SourceEpoch        uint64            `json:"source_epoch"`
	JournalPosition    uint64            `json:"journal_position"`
	ObjectCount        int               `json:"object_count"`
	LogicalBytes       int64             `json:"logical_bytes"`
	PhysicalBytes      int64             `json:"physical_bytes"`
	DedupBytesSaved    int64             `json:"dedup_bytes_saved"`
	Compression        string            `json:"compression"` // none|stored (no fabricated ratios)
	RootDigest         string            `json:"root_digest"`
	ChunkDigests       map[string]string `json:"chunk_digests"` // chunkID -> sha256
	ObjectIndex        []ObjectMeta      `json:"object_index"`
	JournalIncluded    bool              `json:"journal_included"`
	VerificationAlgo   string            `json:"verification_algorithm"`
	ForwardVersionOK   bool              `json:"forward_version_aware"`
	Notes              string            `json:"notes,omitempty"`
}

// ObjectMeta describes one logical object in the archive.
type ObjectMeta struct {
	Path       string   `json:"path"`
	Size       int64    `json:"size"`
	Mode       uint32   `json:"mode"`
	ContentHash string  `json:"content_hash"`
	ChunkIDs   []string `json:"chunk_ids"`
}

// ExportStats are measured (never fabricated).
type ExportStats struct {
	LogicalBytes  int64
	PhysicalBytes int64
	DedupSaved    int64
	DurationMs    float64
	Objects       int
	Chunks        int
}

// ProgressFunc reports export/import milestones (phase, completed units, total, detail).
// Callers may pass nil.
type ProgressFunc func(phase string, done, total int, detail string)

// Export builds a PSA directory at destDir from stateRoot (+ optional journal file).
func Export(stateRoot, destDir string, epoch, journalPos uint64, journalPath string) (*Manifest, *ExportStats, error) {
	return ExportProgress(stateRoot, destDir, epoch, journalPos, journalPath, nil)
}

// ExportProgress is Export with optional progress events.
func ExportProgress(stateRoot, destDir string, epoch, journalPos uint64, journalPath string, onProgress ProgressFunc) (*Manifest, *ExportStats, error) {
	start := time.Now()
	report := func(phase string, done, total int, detail string) {
		if onProgress != nil {
			onProgress(phase, done, total, detail)
		}
	}
	report("export_start", 0, 0, stateRoot)
	if err := os.MkdirAll(filepath.Join(destDir, "chunks"), 0o755); err != nil {
		return nil, nil, err
	}
	idx, logical, err := chunk.BuildIndex(stateRoot, chunk.DefaultBlockSize)
	if err != nil {
		return nil, nil, err
	}
	report("indexed", len(idx), len(idx), "blocks")

	// Dedup: store unique chunk hashes once.
	stored := make(map[string]string) // hash -> chunk filename
	chunkDigests := make(map[string]string)
	var physical int64
	objects := map[string]*ObjectMeta{}

	keys := make([]string, 0, len(idx))
	for k := range idx {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		fp := idx[k]
		om := objects[fp.Path]
		if om == nil {
			om = &ObjectMeta{Path: fp.Path, Mode: 0o644}
			objects[fp.Path] = om
		}
		om.Size += int64(fp.Length)
		cid := fp.Hash
		om.ChunkIDs = append(om.ChunkIDs, cid)
		if _, ok := stored[cid]; !ok {
			data, err := readBlock(stateRoot, fp.Path, fp.Offset, fp.Length)
			if err != nil {
				return nil, nil, err
			}
			cname := cid + ".chk"
			if err := os.WriteFile(filepath.Join(destDir, "chunks", cname), data, 0o644); err != nil {
				return nil, nil, err
			}
			stored[cid] = cname
			chunkDigests[cid] = cid
			physical += int64(len(data))
		}
		if i == len(keys)-1 || i%64 == 0 {
			report("chunking", i+1, len(keys), fp.Path)
		}
	}

	var objList []ObjectMeta
	paths := make([]string, 0, len(objects))
	for p := range objects {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		om := objects[p]
		// content hash = hash of concatenated chunk ids for determinism
		h := sha256.New()
		for _, c := range om.ChunkIDs {
			h.Write([]byte(c))
		}
		om.ContentHash = hex.EncodeToString(h.Sum(nil))
		objList = append(objList, *om)
	}

	journalIncluded := false
	if journalPath != "" {
		if b, err := os.ReadFile(journalPath); err == nil {
			jp := filepath.Join(destDir, "journal.jsonl")
			if err := os.WriteFile(jp, b, 0o644); err == nil {
				journalIncluded = true
				physical += int64(len(b))
			}
		}
	}

	root := rootDigest(objList)
	m := &Manifest{
		Magic:            FormatMagic,
		FormatVersion:    FormatVersion,
		MinReaderVersion: 1,
		ArchiveID:        fmt.Sprintf("psa-%d", time.Now().UnixNano()),
		CreatedAt:        time.Now().UTC(),
		Portability:      EngineProvided,
		SourceEpoch:      epoch,
		JournalPosition:  journalPos,
		ObjectCount:      len(objList),
		LogicalBytes:     logical,
		PhysicalBytes:    physical,
		DedupBytesSaved:  logical - physical,
		Compression:      "none",
		RootDigest:       root,
		ChunkDigests:     chunkDigests,
		ObjectIndex:      objList,
		JournalIncluded:  journalIncluded,
		VerificationAlgo: "sha256-chunk-v1",
		ForwardVersionOK: true,
		Notes:            "ENGINE_PROVIDED portable archive; not RAILWAY_NATIVE",
	}
	if m.DedupBytesSaved < 0 {
		m.DedupBytesSaved = 0
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(filepath.Join(destDir, "archive.json"), b, 0o644); err != nil {
		return nil, nil, err
	}
	stats := &ExportStats{
		LogicalBytes:  logical,
		PhysicalBytes: physical,
		DedupSaved:    m.DedupBytesSaved,
		DurationMs:    float64(time.Since(start).Microseconds()) / 1000,
		Objects:       len(objList),
		Chunks:        len(stored),
	}
	report("export_done", stats.Objects, stats.Objects, m.ArchiveID)
	return m, stats, nil
}

func readBlock(root, rel string, off int64, n int) ([]byte, error) {
	f, err := os.Open(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, n)
	_, err = f.ReadAt(buf, off)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf, nil
}

func rootDigest(objs []ObjectMeta) string {
	h := sha256.New()
	for _, o := range objs {
		fmt.Fprintf(h, "%s:%s:%d\n", o.Path, o.ContentHash, o.Size)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Inspect loads archive.json.
func Inspect(archiveDir string) (*Manifest, error) {
	b, err := os.ReadFile(filepath.Join(archiveDir, "archive.json"))
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Magic != FormatMagic {
		return nil, fmt.Errorf("not an SDE PSA archive (magic=%q)", m.Magic)
	}
	if m.FormatVersion > FormatVersion && m.MinReaderVersion > FormatVersion {
		return nil, fmt.Errorf("archive requires reader >= %d (this reader %d)", m.MinReaderVersion, FormatVersion)
	}
	return &m, nil
}

// Verify checks chunk digests and recomputes root digest.
func Verify(archiveDir string) (*types.VerificationReceipt, error) {
	m, err := Inspect(archiveDir)
	if err != nil {
		return &types.VerificationReceipt{Outcome: types.VerifyUnknown, EvidenceSummary: err.Error()}, err
	}
	start := time.Now()
	var mismatches []string
	for cid := range m.ChunkDigests {
		p := filepath.Join(archiveDir, "chunks", cid+".chk")
		b, err := os.ReadFile(p)
		if err != nil {
			mismatches = append(mismatches, cid+": missing")
			continue
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != cid {
			mismatches = append(mismatches, cid+": content mismatch")
		}
	}
	got := rootDigest(m.ObjectIndex)
	receipt := &types.VerificationReceipt{
		Level:           types.VerifyChecksum,
		Epoch:           m.SourceEpoch,
		ActiveRootHash:  m.RootDigest,
		CandidateHash:   got,
		FileCountActive: m.ObjectCount,
		FileCountCand:   m.ObjectCount,
		BytesCompared:   m.LogicalBytes,
		CheckedAt:       time.Now().UTC(),
		DurationMs:      float64(time.Since(start).Microseconds()) / 1000,
	}
	if len(mismatches) > 0 || got != m.RootDigest {
		receipt.Outcome = types.VerifyFailed
		receipt.HashMismatches = mismatches
		receipt.EvidenceSummary = "archive integrity failed"
		return receipt, fmt.Errorf("%s", receipt.EvidenceSummary)
	}
	receipt.Outcome = types.VerifyVerified
	receipt.EvidenceSummary = "archive chunk digests and root digest OK"
	return receipt, nil
}

// Import reconstructs state into destRoot from an archive. Does not mutate the archive.
func Import(archiveDir, destRoot string) (*Manifest, error) {
	m, err := Inspect(archiveDir)
	if err != nil {
		return nil, err
	}
	if _, err := Verify(archiveDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return nil, err
	}
	for _, obj := range m.ObjectIndex {
		outPath := filepath.Join(destRoot, filepath.FromSlash(obj.Path))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return nil, err
		}
		var data []byte
		for _, cid := range obj.ChunkIDs {
			b, err := os.ReadFile(filepath.Join(archiveDir, "chunks", cid+".chk"))
			if err != nil {
				return nil, err
			}
			data = append(data, b...)
		}
		mode := os.FileMode(obj.Mode)
		if mode == 0 {
			mode = 0o644
		}
		if err := os.WriteFile(outPath, data, mode); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// CloneReadOnly copies archive metadata+chunks to a new identity directory.
// Writes to the clone must not mutate the original (separate tree).
func CloneReadOnly(srcArchive, dstArchive string) (*Manifest, error) {
	if _, err := Inspect(srcArchive); err != nil {
		return nil, err
	}
	if err := copyDir(srcArchive, dstArchive); err != nil {
		return nil, err
	}
	m2, err := Inspect(dstArchive)
	if err != nil {
		return nil, err
	}
	m2.ArchiveID = fmt.Sprintf("psa-clone-%d", time.Now().UnixNano())
	m2.Notes = "read-oriented clone; original archive untouched; ENGINE_PROVIDED"
	b, _ := json.MarshalIndent(m2, "", "  ")
	if err := os.WriteFile(filepath.Join(dstArchive, "archive.json"), b, 0o644); err != nil {
		return nil, err
	}
	// Best-effort mark chunks read-only on platforms that support it.
	_ = filepath.WalkDir(filepath.Join(dstArchive, "chunks"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		return os.Chmod(path, 0o444)
	})
	return m2, nil
}

// ApplyHooksToState walks a restored/cloned state tree and applies text redaction hooks.
// Binary files are left unchanged (hooks refuse invented anonymization).
func ApplyHooksToState(stateRoot string, apply func(path string, data []byte) ([]byte, error)) (int, error) {
	changed := 0
	err := filepath.WalkDir(stateRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(stateRoot, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out, err := apply(filepath.ToSlash(rel), b)
		if err != nil {
			return err
		}
		if !bytesEqual(b, out) {
			if err := os.WriteFile(path, out, 0o644); err != nil {
				return err
			}
			changed++
		}
		return nil
	})
	return changed, err
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

// IsPSADir reports whether path looks like a PSA.
func IsPSADir(path string) bool {
	m, err := Inspect(path)
	return err == nil && m != nil && strings.HasPrefix(m.ArchiveID, "psa-")
}
