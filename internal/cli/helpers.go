// Package cli holds shared helpers for cmd/sde without owning the cobra tree.
// Agent-owned command wiring stays in cmd/sde; this package is additive.
package cli

import (
	"fmt"
	"strings"

	"github.com/theworker02/stateful-deployments-engine/internal/events"
)

// NotWired prints a consistent message for stub subcommands.
func NotWired(feature string) error {
	return fmt.Errorf("%s: not yet wired (contract exists; see docs and internal packages)", feature)
}

// PrintEventTimeline renders structured events for operator UX.
func PrintEventTimeline(evts []events.Event) {
	for _, line := range events.RenderHuman(evts) {
		fmt.Println(line)
	}
}

// Banner returns the transactional deploy banner used in demos.
func Banner(ok bool) string {
	if ok {
		return "STATEFUL ZERO-DOWNTIME DEPLOYMENT"
	}
	return "STATEFUL DEPLOYMENT FAILED"
}

// ModuleMap is a short package index for `sde modules` style output.
func ModuleMap() string {
	lines := []string{
		"internal/epoch       fencing epoch store",
		"internal/fsm         deploy state machine + persistence",
		"internal/receipt     sealed deployment receipts",
		"internal/manifest    Merkle state manifests",
		"internal/heat        hot/warm/cold classification",
		"internal/converge    catch-up / non-convergence",
		"internal/chunk       block fingerprinting",
		"internal/plan        migration plan helpers",
		"internal/safety      cutover gate policies",
		"internal/observe     post-cutover observation window",
		"internal/recover     forward recovery classification",
		"internal/capability  storage capability negotiation",
		"internal/metrics     CUTOVER_WRITE_PAUSE_MS + KPI names",
		"internal/events      transactional deploy event log",
		"internal/doctor       install / env readiness checks",
		"internal/archive      Portable State Archive (ENGINE_PROVIDED)",
		"internal/escrow       escrow catalog + ObjectStore",
		"internal/dr           DR planner + fire drills",
		"internal/invariants   safety invariant tests",
		"sdk/go/sdesdk        public mutation journal SDK",
		"internal/adapter/*   local, railway, fly, docker, k8s, nomad",
	}
	return strings.Join(lines, "\n")
}
