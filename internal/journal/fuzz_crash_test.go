package journal_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

// FuzzJournalCrashConsistency hammers append/reopen boundaries with persisted seeds.
func FuzzJournalCrashConsistency(f *testing.F) {
	seedDir := filepath.Join("..", "..", "testdata", "fuzz", "journal")
	_ = os.MkdirAll(seedDir, 0o755)
	f.Add([]byte(`{"kind":"write","path":"a","payload":"x"}`))
	f.Add([]byte(`{"kind":"delete","path":"b"}`))
	f.Add([]byte(`not-json`))

	f.Fuzz(func(t *testing.T, raw []byte) {
		dir := t.TempDir()
		j, err := journal.Open(dir, 1)
		if err != nil {
			t.Fatal(err)
		}
		var m types.Mutation
		if json.Unmarshal(raw, &m) == nil && m.Kind != "" {
			m.Path = "fuzz.txt"
			if m.Kind == types.OpWrite || m.Kind == types.OpCreate {
				if len(m.Payload) == 0 {
					m.Payload = []byte("x")
				}
			}
			_, _ = j.Append(m)
		} else {
			// Simulate crash mid-write: append garbage then reopen must fail or recover cleanly.
			p := filepath.Join(dir, "mutations.jsonl")
			f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err == nil {
				_, _ = f.Write(append(raw, '\n'))
				_ = f.Close()
			}
		}
		_ = j.Close()
		j2, err := journal.Open(dir, 1)
		if err != nil {
			// Corruption detected — invariant: never silent success on corrupt trailers/lines.
			_ = os.WriteFile(filepath.Join(seedDir, "last-corrupt.seed"), raw, 0o644)
			return
		}
		defer j2.Close()
		// Invariant: NextSeq monotonic
		if j2.NextSeq() == 0 {
			t.Fatal("next seq must be >= 1")
		}
		all, err := j2.ReadFrom(1)
		if err != nil {
			t.Fatal(err)
		}
		var prev uint64
		for _, e := range all {
			if prev != 0 && e.Seq < prev {
				t.Fatalf("seq regression %d < %d", e.Seq, prev)
			}
			prev = e.Seq
		}
	})
}
