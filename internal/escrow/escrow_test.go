package escrow_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/escrow"
)

type memStore struct {
	objects map[string][]byte
}

func (m *memStore) PutObject(key string, body io.Reader, size int64, contentType string) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if m.objects == nil {
		m.objects = map[string][]byte{}
	}
	m.objects[key] = b
	return nil
}
func (m *memStore) GetObject(key string) (io.ReadCloser, error) {
	b, ok := m.objects[key]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func (m *memStore) DeleteObject(key string) error {
	delete(m.objects, key)
	return nil
}

func TestPutListFireDrillMeta(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "src")
	_ = os.MkdirAll(filepath.Join(src, "data"), 0o755)
	_ = os.WriteFile(filepath.Join(src, "data", "a.json"), []byte(`{"x":1}`), 0o644)
	arch := filepath.Join(base, "arch")
	m, _, err := archive.Export(src, arch, 1, 1, "")
	if err != nil {
		t.Fatal(err)
	}

	mem := &memStore{}
	st := &escrow.Store{CatalogDir: filepath.Join(base, "escrow"), Objects: mem}
	entry, err := st.Put(escrow.Target{
		Kind: escrow.S3Compatible, Name: "mem", RootPath: filepath.Join(base, "store"),
	}, arch)
	if err != nil {
		t.Fatal(err)
	}
	if entry.ArchiveID != m.ArchiveID {
		t.Fatal("id mismatch")
	}
	if len(mem.objects) == 0 {
		t.Fatal("expected object store put")
	}

	if err := st.RecordFireDrill(m.ArchiveID, escrow.FireDrillMeta{
		ReceiptID: "drill-1", RootDigestOK: true,
	}); err != nil {
		t.Fatal(err)
	}
	show, err := st.Show(m.ArchiveID)
	if err != nil {
		t.Fatal(err)
	}
	if show.LastFireDrill == nil || !show.LastFireDrill.RootDigestOK {
		t.Fatal("expected last fire drill")
	}
	table := escrow.FormatTable([]escrow.CatalogEntry{*show})
	if table == "" {
		t.Fatal("empty table")
	}
}
