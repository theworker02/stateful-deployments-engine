// Package escrow stores Portable State Archives on external targets.
// Core never hardcodes AWS SDKs — S3-compatible is a generic HTTP/path contract.
package escrow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
)

// TargetKind enumerates escrow backends.
type TargetKind string

const (
	LocalFilesystem    TargetKind = "LOCAL_FILESYSTEM"
	S3Compatible       TargetKind = "S3_COMPATIBLE"
	GenericObjectStore TargetKind = "GENERIC_OBJECT_STORE"
)

// Target is a negotiated escrow destination.
type Target struct {
	Kind     TargetKind `json:"kind"`
	Name     string     `json:"name"`
	RootPath string     `json:"root_path"`          // local path or mount
	Endpoint string     `json:"endpoint,omitempty"` // S3-compatible endpoint URL (opaque)
	Bucket   string     `json:"bucket,omitempty"`
	Prefix   string     `json:"prefix,omitempty"`
	// AccessKey / SecretKey are optional for path-style MinIO-compatible PUT/GET.
	// Prefer injecting via ObjectStore rather than persisting secrets in catalog.
	AccessKey string `json:"-"`
	SecretKey string `json:"-"`
}

// CatalogEntry indexes one escrowed archive.
type CatalogEntry struct {
	ArchiveID      string     `json:"archive_id"`
	StoredAt       time.Time  `json:"stored_at"`
	TargetName     string     `json:"target_name"`
	LocalPath      string     `json:"local_path"`
	RootDigest     string     `json:"root_digest"`
	Epoch          uint64     `json:"epoch"`
	LogicalBytes   int64      `json:"logical_bytes"`
	Verified       bool       `json:"verified"`
	LastFireDrill  *FireDrillMeta `json:"last_fire_drill,omitempty"`
	ObjectURI      string     `json:"object_uri,omitempty"` // for S3-compatible remote keys
}

// FireDrillMeta records the most recent successful/failed isolated restore drill.
type FireDrillMeta struct {
	ReceiptID    string    `json:"receipt_id"`
	RanAt        time.Time `json:"ran_at"`
	DurationMs   float64   `json:"duration_ms"`
	RootDigestOK bool      `json:"root_digest_ok"`
	Error        string    `json:"error,omitempty"`
}

// ObjectStore is the minio-compatible API surface used by S3_COMPATIBLE escrow.
// Implementations must not require an AWS SDK in core.
type ObjectStore interface {
	PutObject(key string, body io.Reader, size int64, contentType string) error
	GetObject(key string) (io.ReadCloser, error)
	DeleteObject(key string) error
}

// HTTPObjectStore talks to path-style S3-compatible endpoints (MinIO, Garage, etc.)
// using optional static bearer/basic via AccessKey as username when SecretKey set
// (HTTP Basic). Signature V4 is intentionally out of core — operators may front
// with a signed proxy or use LOCAL_FILESYSTEM mounts.
type HTTPObjectStore struct {
	Endpoint  string
	Bucket    string
	Prefix    string
	AccessKey string
	SecretKey string
	Client    *http.Client
}

func (h *HTTPObjectStore) client() *http.Client {
	if h.Client != nil {
		return h.Client
	}
	return http.DefaultClient
}

func (h *HTTPObjectStore) objectURL(key string) (string, error) {
	base := strings.TrimRight(h.Endpoint, "/")
	if base == "" || h.Bucket == "" {
		return "", fmt.Errorf("endpoint and bucket required for S3_COMPATIBLE")
	}
	fullKey := path.Join(strings.Trim(h.Prefix, "/"), key)
	u, err := url.Parse(base + "/" + h.Bucket + "/" + fullKey)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (h *HTTPObjectStore) PutObject(key string, body io.Reader, size int64, contentType string) error {
	uri, err := h.objectURL(key)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, uri, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if size >= 0 {
		req.ContentLength = size
	}
	if h.AccessKey != "" {
		req.SetBasicAuth(h.AccessKey, h.SecretKey)
	}
	resp, err := h.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("S3_COMPATIBLE PUT %s: status %d: %s", key, resp.StatusCode, string(b))
	}
	return nil
}

func (h *HTTPObjectStore) GetObject(key string) (io.ReadCloser, error) {
	uri, err := h.objectURL(key)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	if h.AccessKey != "" {
		req.SetBasicAuth(h.AccessKey, h.SecretKey)
	}
	resp, err := h.client().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("S3_COMPATIBLE GET %s: status %d: %s", key, resp.StatusCode, string(b))
	}
	return resp.Body, nil
}

func (h *HTTPObjectStore) DeleteObject(key string) error {
	uri, err := h.objectURL(key)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	if h.AccessKey != "" {
		req.SetBasicAuth(h.AccessKey, h.SecretKey)
	}
	resp, err := h.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 && resp.StatusCode != 404 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("S3_COMPATIBLE DELETE %s: status %d: %s", key, resp.StatusCode, string(b))
	}
	return nil
}

// FilesystemObjectStore maps object keys onto a local directory (mount / rclone target).
type FilesystemObjectStore struct {
	Root string
}

func (f *FilesystemObjectStore) PutObject(key string, body io.Reader, size int64, contentType string) error {
	_ = contentType
	_ = size
	dst := filepath.Join(f.Root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, body)
	return err
}

func (f *FilesystemObjectStore) GetObject(key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(f.Root, filepath.FromSlash(key)))
}

func (f *FilesystemObjectStore) DeleteObject(key string) error {
	return os.Remove(filepath.Join(f.Root, filepath.FromSlash(key)))
}

// Store manages catalog + put/get for LOCAL_FILESYSTEM (and path-mapped S3 mounts).
type Store struct {
	CatalogDir string
	// Objects, when set, is used for S3_COMPATIBLE / GENERIC_OBJECT_STORE puts
	// in addition to (or instead of) RootPath staging.
	Objects ObjectStore
}

func (s *Store) catalogPath() string {
	return filepath.Join(s.CatalogDir, "catalog.json")
}

func (s *Store) load() ([]CatalogEntry, error) {
	b, err := os.ReadFile(s.catalogPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []CatalogEntry
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Store) save(entries []CatalogEntry) error {
	if err := os.MkdirAll(s.CatalogDir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.catalogPath(), b, 0o644)
}

// Put copies an archive directory into the target and catalogs it.
func (s *Store) Put(t Target, archiveDir string) (*CatalogEntry, error) {
	m, err := archive.Inspect(archiveDir)
	if err != nil {
		return nil, err
	}
	destRoot := t.RootPath
	if destRoot == "" && s.Objects == nil {
		return nil, fmt.Errorf("target root_path required (or inject ObjectStore)")
	}
	switch t.Kind {
	case LocalFilesystem, S3Compatible, GenericObjectStore:
		// ok
	default:
		return nil, fmt.Errorf("unsupported target kind %q", t.Kind)
	}

	var dest string
	var objectURI string
	if destRoot != "" {
		dest = filepath.Join(destRoot, m.ArchiveID)
		if err := copyTree(archiveDir, dest); err != nil {
			return nil, err
		}
		if _, err := archive.Verify(dest); err != nil {
			return nil, fmt.Errorf("post-put verify failed: %w", err)
		}
	}

	objs := s.Objects
	if objs == nil && t.Kind == S3Compatible && t.Endpoint != "" {
		objs = &HTTPObjectStore{
			Endpoint: t.Endpoint, Bucket: t.Bucket, Prefix: t.Prefix,
			AccessKey: t.AccessKey, SecretKey: t.SecretKey,
		}
	}
	if objs == nil && t.Kind != LocalFilesystem && destRoot != "" {
		// Path-mapped mount: also expose as ObjectStore for symmetry.
		objs = &FilesystemObjectStore{Root: destRoot}
	}
	if objs != nil && dest != "" {
		// Upload archive.json as a manifest marker (full tree already on mount / local).
		manifestPath := filepath.Join(dest, "archive.json")
		b, err := os.ReadFile(manifestPath)
		if err == nil {
			key := path.Join(m.ArchiveID, "archive.json")
			if err := objs.PutObject(key, bytes.NewReader(b), int64(len(b)), "application/json"); err != nil {
				return nil, fmt.Errorf("object store put: %w", err)
			}
			objectURI = key
		}
	} else if objs != nil && dest == "" {
		// Remote-only: walk archive and put each file.
		err := filepath.WalkDir(archiveDir, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(archiveDir, p)
			if err != nil {
				return err
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			key := path.Join(m.ArchiveID, filepath.ToSlash(rel))
			return objs.PutObject(key, bytes.NewReader(b), int64(len(b)), "application/octet-stream")
		})
		if err != nil {
			return nil, err
		}
		objectURI = m.ArchiveID + "/"
		// Stage a local verified copy under catalog for restore tooling.
		dest = filepath.Join(s.CatalogDir, "staging", m.ArchiveID)
		if err := copyTree(archiveDir, dest); err != nil {
			return nil, err
		}
		if _, err := archive.Verify(dest); err != nil {
			return nil, fmt.Errorf("post-put verify failed: %w", err)
		}
	}

	entry := CatalogEntry{
		ArchiveID:    m.ArchiveID,
		StoredAt:     time.Now().UTC(),
		TargetName:   t.Name,
		LocalPath:    dest,
		RootDigest:   m.RootDigest,
		Epoch:        m.SourceEpoch,
		LogicalBytes: m.LogicalBytes,
		Verified:     true,
		ObjectURI:    objectURI,
	}
	entries, _ := s.load()
	entries = append(entries, entry)
	if err := s.save(entries); err != nil {
		return nil, err
	}
	return &entry, nil
}

// RecordFireDrill attaches last-drill metadata to a catalog entry by archive id.
func (s *Store) RecordFireDrill(archiveID string, meta FireDrillMeta) error {
	entries, err := s.load()
	if err != nil {
		return err
	}
	found := false
	for i := range entries {
		if entries[i].ArchiveID == archiveID {
			cp := meta
			entries[i].LastFireDrill = &cp
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("archive %q not in catalog", archiveID)
	}
	return s.save(entries)
}

// List returns catalog entries sorted by StoredAt desc.
func (s *Store) List() ([]CatalogEntry, error) {
	entries, err := s.load()
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StoredAt.After(entries[j].StoredAt)
	})
	return entries, nil
}

// Show finds one archive by id.
func (s *Store) Show(id string) (*CatalogEntry, error) {
	entries, err := s.load()
	if err != nil {
		return nil, err
	}
	for i := range entries {
		if entries[i].ArchiveID == id {
			return &entries[i], nil
		}
	}
	return nil, fmt.Errorf("archive %q not in catalog", id)
}

// Verify re-verifies a cataloged archive.
func (s *Store) Verify(id string) error {
	e, err := s.Show(id)
	if err != nil {
		return err
	}
	_, err = archive.Verify(e.LocalPath)
	return err
}

// Prune removes a catalog entry and optionally deletes storage (explicit destructive).
func (s *Store) Prune(id string, deleteData bool) error {
	entries, err := s.load()
	if err != nil {
		return err
	}
	var keep []CatalogEntry
	var removed *CatalogEntry
	for i := range entries {
		if entries[i].ArchiveID == id {
			removed = &entries[i]
			continue
		}
		keep = append(keep, entries[i])
	}
	if removed == nil {
		return fmt.Errorf("archive %q not found", id)
	}
	if deleteData {
		_ = os.RemoveAll(removed.LocalPath)
		if s.Objects != nil && removed.ObjectURI != "" {
			_ = s.Objects.DeleteObject(removed.ObjectURI)
		}
	}
	return s.save(keep)
}

// FormatTable renders a human-readable catalog table.
func FormatTable(entries []CatalogEntry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-28s %-8s %12s %-10s %-20s %s\n",
		"ARCHIVE_ID", "EPOCH", "BYTES", "VERIFIED", "STORED_AT", "LAST_DRILL")
	for _, e := range entries {
		id := e.ArchiveID
		if len(id) > 28 {
			id = id[:25] + "…"
		}
		drill := "-"
		if e.LastFireDrill != nil {
			if e.LastFireDrill.RootDigestOK {
				drill = e.LastFireDrill.RanAt.Format("2006-01-02") + " ok"
			} else {
				drill = e.LastFireDrill.RanAt.Format("2006-01-02") + " fail"
			}
		}
		fmt.Fprintf(&b, "%-28s %-8d %12d %-10v %-20s %s\n",
			id, e.Epoch, e.LogicalBytes, e.Verified, e.StoredAt.Format(time.RFC3339), drill)
	}
	return b.String()
}

func copyTree(src, dst string) error {
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
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}
