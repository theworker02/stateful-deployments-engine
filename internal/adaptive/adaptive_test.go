package adaptive_test

import (
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/adaptive"
)

func TestTuneIncreasesBatchUnderLag(t *testing.T) {
	c := adaptive.New()
	old := c.Params.BatchSize
	made := c.Tune(adaptive.Input{LagOps: 200, CPUBusy: 0.2, WriteRate: 10, Throughput: 100})
	if len(made) == 0 {
		t.Fatal("expected tuning decisions")
	}
	if c.Params.BatchSize <= old {
		t.Fatalf("batch_size %d want > %d", c.Params.BatchSize, old)
	}
	if len(c.Decisions) == 0 {
		t.Fatal("decisions not recorded")
	}
}

func TestTuneEnablesCompression(t *testing.T) {
	c := adaptive.New()
	c.Tune(adaptive.Input{WriteRate: 150, Throughput: 50})
	if !c.Params.Compression {
		t.Fatal("expected compression enabled")
	}
}

func TestDefaultBounds(t *testing.T) {
	p := adaptive.Default()
	if p.Concurrency < 1 || p.ChunkSize < 1024 {
		t.Fatalf("unexpected defaults: %+v", p)
	}
}
