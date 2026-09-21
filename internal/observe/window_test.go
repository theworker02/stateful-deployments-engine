package observe

import (
	"context"
	"testing"
	"time"
)

func TestRunPasses(t *testing.T) {
	cfg := Config{Duration: 50 * time.Millisecond, PollInterval: 10 * time.Millisecond, MaxErrorEvents: 0}
	res, err := Run(context.Background(), cfg, func(ctx context.Context) (bool, string, error) {
		return true, "ok", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Polls < 1 {
		t.Fatalf("%+v", res)
	}
}

func TestRunFailsOnUnhealthy(t *testing.T) {
	cfg := Config{Duration: time.Second, PollInterval: 5 * time.Millisecond, MaxErrorEvents: 0}
	res, err := Run(context.Background(), cfg, func(ctx context.Context) (bool, string, error) {
		return false, "5xx", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || !res.AbortedEarly {
		t.Fatalf("%+v", res)
	}
}
