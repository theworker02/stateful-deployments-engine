package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/local"
	"github.com/theworker02/stateful-deployments-engine/internal/agent"
	"github.com/theworker02/stateful-deployments-engine/internal/chaos"
	"github.com/theworker02/stateful-deployments-engine/internal/chunk"
	"github.com/theworker02/stateful-deployments-engine/internal/cli"
	"github.com/theworker02/stateful-deployments-engine/internal/coordinator"
	"github.com/theworker02/stateful-deployments-engine/internal/dr"
	"github.com/theworker02/stateful-deployments-engine/internal/escrow"
	"github.com/theworker02/stateful-deployments-engine/internal/fence"
	"github.com/theworker02/stateful-deployments-engine/internal/journal"
	"github.com/theworker02/stateful-deployments-engine/internal/planner"
	"github.com/theworker02/stateful-deployments-engine/internal/recoverypoint"
	"github.com/theworker02/stateful-deployments-engine/internal/rollback"
	"github.com/theworker02/stateful-deployments-engine/internal/store"
	"github.com/theworker02/stateful-deployments-engine/internal/types"
	"github.com/theworker02/stateful-deployments-engine/internal/verifier"
	"github.com/theworker02/stateful-deployments-engine/internal/version"
)

func main() {
	root := &cobra.Command{
		Use:   "sde",
		Short: "Stateful Deployments Engine — transactional blue/green for persistent workloads",
		Long: `SDE coordinates near-zero-downtime deployments and live state mobility.

Core: journal → sync → verify → barrier → cutover → observe → commit | recover.
Adapters: local (demo), railway (production target). Networking is out of scope.`,
	}

	root.PersistentFlags().String("root", ".sde", "SDE working directory (journal + local adapter state)")
	root.AddCommand(
		initCmd(), planCmd(), planMigrationCmd(), deployCmd(), statusCmd(), readinessCmd(), verifyCmd(),
		cutoverCmd(), rollbackCmd(), recoverCmd(), benchmarkCmd(), chaosCmd(), demoCmd(),
		modulesCmd(), versionCmd(), completionCmd(), doctorCmd(),
		exportCmd(), inspectArchiveCmd(), verifyArchiveCmd(), importArchiveCmd(), restoreCmd(),
		escrowCmd(), catalogCmd(), recoveryPointsCmd(), drPlanCmd(), fireDrillCmd(), cloneArchiveCmd(),
		railwayCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootDir(cmd *cobra.Command) string {
	r, _ := cmd.Flags().GetString("root")
	return r
}

func openStack(root string) (*local.Adapter, *journal.Journal, *fence.Manager, *store.Store, error) {
	plat, err := local.Open(root)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	j, err := journal.Open(filepath.Join(root, "journal"), 1)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	plat.BindJournal(j)
	fm, err := fence.Open(root, "cli")
	if err != nil {
		j.Close()
		return nil, nil, nil, nil, err
	}
	return plat, j, fm, store.New(root), nil
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize local SDE workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, fm, _, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()
			fmt.Printf("Initialized SDE workspace at %s (adapter=%s, epoch=%d, fence=%d)\n",
				root, plat.Name(), j.Epoch(), fm.Current().Token)
			return nil
		},
	}
}

func planCmd() *cobra.Command {
	return planMigrationCmdAlias("plan", "Alias for plan-migration")
}

func planMigrationCmd() *cobra.Command {
	return planMigrationCmdAlias("plan-migration", "Inspect source and produce a MigrationPlan")
}

func planMigrationCmdAlias(use, short string) *cobra.Command {
	var image string
	var writeRate float64
	var capacity int64
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, _, _, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()
			p := &planner.Planner{Platform: plat, Storage: plat, Journal: j}
			plan, err := p.Plan(context.Background(), planner.Input{
				TargetImage:         image,
				WriteRateOpsPerSec:  writeRate,
				TargetCapacityBytes: capacity,
				HealthRequired:      true,
			})
			if err != nil {
				return err
			}
			if !planner.IsReady(plan.Conclusion) && plan.Conclusion == types.MigrateUnknown {
				fmt.Println("NOTE: UNKNOWN must never be treated as READY")
			}
			b, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&image, "image", "local/app:next", "target image")
	cmd.Flags().Float64Var(&writeRate, "write-rate", 0, "observed write rate (ops/s)")
	cmd.Flags().Int64Var(&capacity, "target-capacity", 0, "target capacity bytes (0=unlimited)")
	return cmd
}

func deployCmd() *cobra.Command {
	var image string
	var dryRun bool
	var maxPause time.Duration
	var override bool
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Run a transactional stateful deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, fm, st, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()

			c := coordinator.NewFull(plat, plat, j, fm, st, func(phase types.DeployPhase, msg string) {
				fmt.Printf("✓ %s\n", msg)
			})
			report, err := c.Deploy(context.Background(), coordinator.Config{
				ImageRef:      image,
				DryRun:        dryRun,
				MaxWritePause: maxPause,
				OverrideCutover: override,
				SDERoot:       root,
			})
			if err != nil {
				printReport(report)
				return err
			}
			printReport(report)
			return saveLastReport(root, report)
		},
	}
	cmd.Flags().StringVar(&image, "image", "local/app:next", "candidate image reference")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "plan/validate/simulate only")
	cmd.Flags().DurationVar(&maxPause, "max-write-pause", 250*time.Millisecond, "maximum allowed write barrier")
	cmd.Flags().BoolVar(&override, "override-cutover", false, "explicit override for soft safety-gate findings")
	return cmd
}

func statusCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show active slot, journal epoch, fence, and in-flight deploys",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, fm, st, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()
			active, err := plat.ActiveSlot(context.Background())
			if err != nil {
				return err
			}
			out := map[string]interface{}{
				"adapter":  plat.Name(),
				"active":   active.ID,
				"image":    active.ImageRef,
				"state":    active.StatePath,
				"epoch":    j.Epoch(),
				"fence":    fm.Current().Token,
				"next_seq": j.NextSeq(),
			}
			if mid, _ := st.LatestNonTerminal(); mid != nil {
				out["inflight"] = map[string]interface{}{
					"deploy_id": mid.DeployID, "phase": mid.Phase, "dry_run": mid.DryRun,
				}
			}
			if asJSON {
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			fmt.Printf("adapter:  %s\nactive:   %s\nimage:    %s\nstate:    %s\nepoch:    %d\nfence:    %d\nnext_seq: %d\n",
				out["adapter"], out["active"], out["image"], out["state"], out["epoch"], out["fence"], out["next_seq"])
			if mid, _ := st.LatestNonTerminal(); mid != nil {
				fmt.Printf("inflight: %s phase=%s dry_run=%v\n", mid.DeployID, mid.Phase, mid.DryRun)
			} else {
				fmt.Println("inflight: (none)")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON output")
	return cmd
}

func readinessCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "readiness",
		Short: "Factual DR + deploy readiness scorecard (UNKNOWN ≠ ready)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, fm, _, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()
			p := &dr.Planner{
				Escrow: &escrow.Store{CatalogDir: filepath.Join(root, "escrow")},
				Points: &recoverypoint.Store{Root: filepath.Join(root, "recovery")},
			}
			plan, err := p.Plan()
			if err != nil {
				return err
			}
			active, _ := plat.ActiveSlot(context.Background())
			card := map[string]interface{}{
				"version":        version.Version,
				"adapter":        plat.Name(),
				"active_slot":    "",
				"epoch":          j.Epoch(),
				"fence":          fm.Current().Token,
				"dr_conclusion":  plan.Conclusion,
				"dr_scorecard":   plan.Scorecard,
				"dr_evidence":    plan.Evidence,
				"note":           "UNKNOWN must never be treated as RECOVERABLE / VERIFIED",
			}
			if active != nil {
				card["active_slot"] = active.ID
			}
			if asJSON {
				b, _ := json.MarshalIndent(card, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			fmt.Println("=== SDE readiness scorecard ===")
			fmt.Printf("adapter:       %s\n", card["adapter"])
			fmt.Printf("epoch/fence:   %d / %d\n", j.Epoch(), fm.Current().Token)
			fmt.Printf("DR conclusion: %s\n", plan.Conclusion)
			if plan.Conclusion == dr.Unknown {
				fmt.Println("NOTE: UNKNOWN must never be treated as RECOVERABLE")
			}
			sc := plan.Scorecard
			fmt.Printf("  archive_present:        %v\n", sc.ArchivePresent)
			fmt.Printf("  archive_verified:       %v\n", sc.ArchiveVerified)
			fmt.Printf("  escrow_present:         %v\n", sc.EscrowPresent)
			fmt.Printf("  recovery_point_present: %v\n", sc.RecoveryPointPresent)
			fmt.Printf("  fire_drill_recorded:    %v\n", sc.FireDrillRecorded)
			fmt.Printf("  last_fire_drill_ok:     %v\n", sc.LastFireDrillOK)
			fmt.Printf("  source_independent:     %v\n", sc.SourceIndependent)
			for _, e := range plan.Evidence {
				fmt.Printf("  - %s\n", e)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON output")
	return cmd
}

func verifyCmd() *cobra.Command {
	var level string
	var other string
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify active vs candidate (or --path) consistency",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, _, _, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()
			active, err := plat.ActiveSlot(context.Background())
			if err != nil {
				return err
			}
			cand := other
			if cand == "" {
				cand = active.StatePath
			}
			vr, err := verifier.Verify(active.StatePath, cand, j.Epoch(), verifier.Options{
				Level: types.VerifyLevel(level),
			})
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(vr, "", "  ")
			fmt.Println(string(b))
			if vr.Outcome == types.VerifyUnknown {
				fmt.Println("NOTE: UNKNOWN must never be treated as VERIFIED")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&level, "level", string(types.VerifyChecksum), "METADATA|CHECKSUM|STRUCTURAL|APPLICATION_DEFINED")
	cmd.Flags().StringVar(&other, "path", "", "candidate root (default: active self-check)")
	return cmd
}

func cutoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cutover",
		Short: "Alias for deploy (full transactional cutover)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployCmd().RunE(cmd, args)
		},
	}
}

func rollbackCmd() *cobra.Command {
	var epoch uint64
	var image string
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Restore previous application + state epoch (safety-classified)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, _, _, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()

			var writes uint64
			var phase types.DeployPhase
			if last, err := loadLastReport(root); err == nil {
				if image == "" {
					image = last.FromImage
				}
				if epoch == 0 {
					epoch = last.FromEpoch
				}
				phase = last.Phase
			}
			eng := rollback.New(plat, j, func(msg string) { fmt.Println(msg) })
			report, err := eng.Execute(context.Background(), rollback.Request{
				StandbySlotID:      "",
				TargetEpoch:        epoch,
				TargetImage:        image,
				WritesAfterCutover: writes,
				PhaseAtFailure:     phase,
			})
			if report != nil {
				b, _ := json.MarshalIndent(report, "", "  ")
				fmt.Println(string(b))
			}
			return err
		},
	}
	cmd.Flags().Uint64Var(&epoch, "epoch", 0, "target state epoch")
	cmd.Flags().StringVar(&image, "image", "", "target image")
	return cmd
}

func recoverCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Resume / retry / abort / rollback after coordinator crash",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			plat, j, fm, st, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()
			c := coordinator.NewFull(plat, plat, j, fm, st, func(_ types.DeployPhase, msg string) {
				fmt.Printf("✓ %s\n", msg)
			})
			dec, report, err := c.Recover(context.Background(), id)
			if dec != nil {
				b, _ := json.MarshalIndent(dec, "", "  ")
				fmt.Println(string(b))
			}
			if report != nil {
				printReport(report)
			}
			return err
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "deployment id (default: latest non-terminal)")
	return cmd
}

func benchmarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "benchmark",
		Short: "Run realistic sync/cutover/differential benchmarks; write results/",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			out := filepath.Join("benchmarks", "results")
			_ = os.MkdirAll(out, 0o755)
			results := runBenchSuite(root, out)
			b, _ := json.MarshalIndent(results, "", "  ")
			path := filepath.Join(out, "latest.json")
			if err := os.WriteFile(path, b, 0o644); err != nil {
				return err
			}
			fmt.Printf("Wrote %s\n", path)
			fmt.Println(string(b))
			return nil
		},
	}
}

func chaosCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chaos",
		Short: "Run fault-injection matrix; write machine-readable receipts",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := filepath.Join("benchmarks", "results", "chaos")
			receipts, err := chaos.RunAll(out)
			if err != nil {
				return err
			}
			pass := 0
			for _, r := range receipts {
				if r.Detected {
					pass++
				}
			}
			fmt.Printf("Chaos matrix: %d/%d detected faults under %s\n", pass, len(receipts), out)
			return nil
		},
	}
}

func demoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "demo",
		Short: "Seed data, generate live writes, and run a stateful deploy",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			_ = os.RemoveAll(root)
			plat, j, fm, st, err := openStack(root)
			if err != nil {
				return err
			}
			defer j.Close()

			active, err := plat.ActiveSlot(context.Background())
			if err != nil {
				return err
			}
			w := &agent.Writer{Root: active.StatePath, Journal: j, Local: plat}

			fmt.Println("Seeding initial state…")
			for i := 0; i < 20; i++ {
				_, err := w.WriteFile(context.Background(), fmt.Sprintf("data/item-%02d.json", i),
					[]byte(fmt.Sprintf(`{"n":%d,"v":"seed"}`, i)), 0o644)
				if err != nil {
					return err
				}
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				n := 0
				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(15 * time.Millisecond):
						n++
						_, _ = w.WriteFile(ctx, fmt.Sprintf("data/live-%04d.txt", n),
							[]byte(fmt.Sprintf("mutation-%d-%d", n, time.Now().UnixNano())), 0o644)
					}
				}
			}()

			time.Sleep(100 * time.Millisecond)

			fmt.Println("\nDeploying myapp")
			fmt.Println()
			c := coordinator.NewFull(plat, plat, j, fm, st, func(_ types.DeployPhase, msg string) {
				fmt.Printf("✓ %s\n", msg)
			})
			report, err := c.Deploy(context.Background(), coordinator.Config{
				ImageRef:      "local/app:v2",
				MaxWritePause: time.Second,
			})
			cancel()
			if err != nil {
				return err
			}
			_ = saveLastReport(root, report)
			fmt.Printf("\nBarrier: %s | Pause: %.2fms | Synced: %d | Final delta: %d | Total: %s\n",
				report.BarrierDuration, report.CutoverWritePauseMs, report.MutationsSynced, report.FinalDeltaWrites, report.TotalDuration.Round(time.Millisecond))
			return nil
		},
	}
}

func modulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "modules",
		Short: "Print package / module map (acquisition surface area)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cli.ModuleMap())
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run:   func(cmd *cobra.Command, args []string) { fmt.Println(version.String()) },
	}
}

func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate completion script for your shell.

  sde completion bash > /etc/bash_completion.d/sde
  sde completion zsh > "${fpath[1]}/_sde"
  sde completion powershell | Out-String | Invoke-Expression
`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
	return cmd
}

func printReport(r *types.DeployReport) {
	if r == nil {
		return
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(b))
}

func saveLastReport(root string, r *types.DeployReport) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "last-deploy.json"), b, 0o644)
}

func loadLastReport(root string) (*types.DeployReport, error) {
	b, err := os.ReadFile(filepath.Join(root, "last-deploy.json"))
	if err != nil {
		return nil, err
	}
	var r types.DeployReport
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func runBenchSuite(root, out string) map[string]interface{} {
	_ = os.RemoveAll(root)
	plat, j, fm, st, err := openStack(root)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	defer j.Close()
	active, _ := plat.ActiveSlot(context.Background())
	w := &agent.Writer{Root: active.StatePath, Journal: j, Local: plat}

	// SMALL dataset
	for i := 0; i < 50; i++ {
		_, _ = w.WriteFile(context.Background(), fmt.Sprintf("small/%d.txt", i), make([]byte, 256), 0o644)
	}
	c := coordinator.NewFull(plat, plat, j, fm, st, nil)
	t0 := time.Now()
	rep, err := c.Deploy(context.Background(), coordinator.Config{ImageRef: "bench:small", MaxWritePause: time.Second, SyncPollInterval: 2 * time.Millisecond})
	small := map[string]interface{}{
		"ok": err == nil, "cutover_pause_ms": 0.0, "total_ms": float64(time.Since(t0).Microseconds()) / 1000,
	}
	if rep != nil {
		small["cutover_pause_ms"] = rep.CutoverWritePauseMs
		small["mutations_synced"] = rep.MutationsSynced
		small["rto_seconds"] = rep.TotalDuration.Seconds()
	}

	// Differential vs full copy on MEDIUM-like tree
	src := filepath.Join(root, "bench-src")
	dstFull := filepath.Join(root, "bench-dst-full")
	dstDiff := filepath.Join(root, "bench-dst-diff")
	_ = os.MkdirAll(src, 0o755)
	for i := 0; i < 40; i++ {
		_ = os.WriteFile(filepath.Join(src, fmt.Sprintf("m-%d.bin", i)), make([]byte, 8*1024), 0o644)
	}
	tFull := time.Now()
	fullBytes, _ := chunk.FullCopy(src, dstFull)
	fullMs := float64(time.Since(tFull).Microseconds()) / 1000
	// mutate a few files
	for i := 0; i < 5; i++ {
		_ = os.WriteFile(filepath.Join(src, fmt.Sprintf("m-%d.bin", i)), make([]byte, 8*1024), 0o644)
	}
	_ = os.MkdirAll(dstDiff, 0o755)
	_, _ = chunk.FullCopy(src, dstDiff) // seed
	for i := 0; i < 5; i++ {
		_ = os.WriteFile(filepath.Join(src, fmt.Sprintf("m-%d.bin", i)), append(make([]byte, 8*1024), byte(i)), 0o644)
	}
	si, _, _ := chunk.BuildIndex(src, chunk.DefaultBlockSize)
	di, _, _ := chunk.BuildIndex(dstDiff, chunk.DefaultBlockSize)
	changed := chunk.Diff(si, di)
	tDiff := time.Now()
	diffBytes, _ := chunk.TransferChanged(src, dstDiff, changed)
	diffMs := float64(time.Since(tDiff).Microseconds()) / 1000

	return map[string]interface{}{
		"generated_at": time.Now().UTC(),
		"profiles": map[string]interface{}{
			"SMALL":        small,
			"MEDIUM_DIFF":  map[string]interface{}{"full_copy_ms": fullMs, "full_bytes": fullBytes, "diff_ms": diffMs, "diff_bytes": diffBytes, "changed_blocks": len(changed)},
			"DATABASE_LIKE": map[string]interface{}{"note": "seeded via small deploy path"},
			"MEDIA_LIKE":    map[string]interface{}{"block_size": chunk.DefaultBlockSize},
			"LOG_LIKE":      map[string]interface{}{"append_heavy": true},
			"LARGE_SIMULATED": map[string]interface{}{"objects": 50, "note": "scaled-down simulation"},
		},
		"metrics": map[string]interface{}{
			"baseline_sync_included_in_small": true,
			"journal_replay_via_deploy":       true,
			"cutover_pause_ms":                small["cutover_pause_ms"],
		},
	}
}
