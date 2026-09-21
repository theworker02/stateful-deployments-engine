package journal_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func TestAppendAndRead(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()

	for i := 0; i < 5; i++ {
		_, err := j.Append(types.Mutation{Kind: types.OpWrite, Path: "a.txt", Payload: []byte("x")})
		if err != nil {
			t.Fatal(err)
		}
	}
	all, err := j.ReadFrom(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 5 {
		t.Fatalf("got %d want 5", len(all))
	}
	if all[4].Seq != 5 {
		t.Fatalf("seq=%d", all[4].Seq)
	}
	if all[0].EntryChecksum == "" {
		t.Fatal("expected entry checksum")
	}

	j.Close()
	j2, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j2.Close()
	if j2.NextSeq() != 6 {
		t.Fatalf("nextSeq=%d", j2.NextSeq())
	}
	_ = filepath.Join(dir, "mutations.jsonl")
}

func TestDuplicateDetection(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	s1, _ := j.Append(types.Mutation{Kind: types.OpWrite, Path: "a", Payload: []byte("1"), ClientOpID: "x"})
	s2, err := j.Append(types.Mutation{Kind: types.OpWrite, Path: "a", Payload: []byte("2"), ClientOpID: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if s1 != s2 {
		t.Fatalf("%d != %d", s1, s2)
	}
}

func TestStaleFenceRejected(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	_ = j.SetFenceToken(5)
	_, err = j.Append(types.Mutation{Kind: types.OpWrite, Path: "a", Payload: []byte("1"), FenceToken: 3})
	if err == nil {
		t.Fatal("expected stale fence error")
	}
}

func TestCorruptJournalDetected(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = j.Append(types.Mutation{Kind: types.OpWrite, Path: "a", Payload: []byte("1")})
	j.Close()
	p := filepath.Join(dir, "mutations.jsonl")
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f.WriteString("{bad\n")
	_ = f.Close()
	_, err = journal.Open(dir, 1)
	if err == nil {
		t.Fatal("expected corrupt detection")
	}
}

func TestCheckpointResume(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	_, _ = j.Append(types.Mutation{Kind: types.OpWrite, Path: "a", Payload: []byte("1")})
	if err := j.SaveCheckpoint(1, "test"); err != nil {
		t.Fatal(err)
	}
	cp := j.Checkpoint()
	if cp.Seq != 1 {
		t.Fatalf("seq=%d", cp.Seq)
	}
}

func TestUnknownOpPersisted(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	_, err = j.Append(types.Mutation{Kind: types.OpKind("weird_op"), Path: "a", Payload: []byte("1")})
	if err != nil {
		t.Fatal(err)
	}
	all, _ := j.ReadFrom(1)
	if all[0].Kind != "weird_op" {
		t.Fatal("unrecognized op must not be discarded")
	}
}

func TestTruncationRecovery(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = j.Append(types.Mutation{Kind: types.OpWrite, Path: "a", Payload: []byte("1")})
	_, _ = j.Append(types.Mutation{Kind: types.OpWrite, Path: "b", Payload: []byte("2")})
	j.Close()

	p := filepath.Join(dir, "mutations.jsonl")
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"seq":3,"partial`) // no trailing newline — crash mid-write
	_ = f.Close()

	j2, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j2.Close()
	if j2.NextSeq() != 3 {
		t.Fatalf("nextSeq=%d want 3 after truncating partial line", j2.NextSeq())
	}
	all, err := j2.ReadFrom(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d records", len(all))
	}
}

func TestCRCTrailerTamperDetected(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = j.Append(types.Mutation{Kind: types.OpWrite, Path: "a", Payload: []byte("1")})
	j.Close()

	p := filepath.Join(dir, "mutations.jsonl")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	// Flip a CRC hex nibble while leaving JSON structurally valid.
	s := string(b)
	idx := strings.LastIndex(s, "#crc32=")
	if idx < 0 {
		t.Fatal("expected crc trailer")
	}
	raw := []byte(s)
	// Flip last hex char of CRC.
	pos := len(raw) - 2 // before newline
	if raw[pos] == '0' {
		raw[pos] = '1'
	} else {
		raw[pos] = '0'
	}
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = journal.Open(dir, 1)
	if err == nil {
		t.Fatal("expected crc mismatch")
	}
}

func TestConcurrentAppendSafe(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()

	const n = 50
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_, err := j.Append(types.Mutation{
				Kind: types.OpWrite, Path: fmt.Sprintf("f-%d", i), Payload: []byte("x"),
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
	all, err := j.ReadFrom(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != n {
		t.Fatalf("got %d want %d", len(all), n)
	}
	seen := map[uint64]bool{}
	for _, m := range all {
		if seen[m.Seq] {
			t.Fatalf("duplicate seq %d", m.Seq)
		}
		seen[m.Seq] = true
	}
}
