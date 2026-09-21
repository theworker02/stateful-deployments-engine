package bench

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/agent"
	"github.com/theworker02/stateful-deployments-engine/internal/chunk"
	"github.com/theworker02/stateful-deployments-engine/internal/coordinator"
	"github.com/theworker02/stateful-deployments-engine/internal/fence"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/store"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

func BenchmarkCutoverPause(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root := b.TempDir()
		plat, err := local.Open(root)
		if err != nil {
			b.Fatal(err)
		}
		j, err := journal.Open(filepath.Join(root, "journal"), 1)
		if err != nil {
			b.Fatal(err)
		}
		plat.BindJournal(j)
		fm, _ := fence.Open(root, "bench")
		st := store.New(root)

		active, _ := plat.ActiveSlot(context.Background())
		w := &agent.Writer{Root: active.StatePath, Journal: j, Local: plat}
		for n := 0; n < 100; n++ {
			_, _ = w.WriteFile(context.Background(), fmt.Sprintf("seed-%d.bin", n), make([]byte, 1024), 0o644)
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			n := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n++
					_, _ = w.WriteFile(ctx, fmt.Sprintf("w-%d.bin", n), make([]byte, 256), 0o644)
				}
			}
		}()
		time.Sleep(20 * time.Millisecond)

		c := coordinator.NewFull(plat, plat, j, fm, st, func(types.DeployPhase, string) {})
		report, err := c.Deploy(context.Background(), coordinator.Config{
			ImageRef:         fmt.Sprintf("app:bench-%d", i),
			SyncPollInterval: 2 * time.Millisecond,
			MaxWritePause:    time.Second,
		})
		cancel()
		j.Close()
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(report.CutoverWritePauseMs, "cutover_ms")
		b.ReportMetric(float64(report.MutationsSynced), "mutations_synced")
		b.ReportMetric(float64(report.FinalDeltaWrites), "final_delta")
	}
}

func BenchmarkDifferentialTransfer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		src := b.TempDir()
		dst := b.TempDir()
		for n := 0; n < 30; n++ {
			if err := os.WriteFile(filepath.Join(src, fmt.Sprintf("%d.bin", n)), make([]byte, 4096), 0o644); err != nil {
				b.Fatal(err)
			}
		}
		if _, err := chunk.FullCopy(src, dst); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(src, "0.bin"), append(make([]byte, 4096), 1), 0o644); err != nil {
			b.Fatal(err)
		}
		si, _, _ := chunk.BuildIndex(src, chunk.DefaultBlockSize)
		di, _, _ := chunk.BuildIndex(dst, chunk.DefaultBlockSize)
		ch := chunk.Diff(si, di)
		start := time.Now()
		n, err := chunk.TransferChanged(src, dst, ch)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(time.Since(start).Microseconds())/1000, "diff_ms")
		b.ReportMetric(float64(n), "diff_bytes")
	}
}
