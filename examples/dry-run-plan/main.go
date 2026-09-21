package main

import (
	"fmt"
	"os"

	"github.com/theworker02/stateful-deployments-engine/internal/plan"
)

func main() {
	p := plan.SampleDryRun("app:candidate", 10<<20)
	path := "plan-dry-run.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if err := plan.WriteJSON(path, p); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s conclusion=%s pause_ms=%.1f\n", path, p.Conclusion, p.EstimatedPauseMs)
}
