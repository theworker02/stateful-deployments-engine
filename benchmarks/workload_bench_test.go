package bench

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/chunk"
	"github.com/theworker02/stateful-deployments-engine/workloads"
)

func BenchmarkWorkloadGenerateSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root := filepath.Join(b.TempDir(), "w")
		if _, err := workloads.Generate(root, workloads.ProfileSmall); err != nil {
			b.Fatal(err)
		}
		_ = os.RemoveAll(root)
	}
}

func BenchmarkChunkIndexSmall(b *testing.B) {
	root := b.TempDir()
	if _, err := workloads.Generate(root, workloads.ProfileSmall); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := chunk.BuildIndex(root, chunk.DefaultBlockSize)
		if err != nil {
			b.Fatal(err)
		}
	}
}
