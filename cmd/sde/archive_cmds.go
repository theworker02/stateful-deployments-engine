package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/dr"
	"github.com/theworker02/stateful-deployments-engine/internal/escrow"
	"github.com/theworker02/stateful-deployments-engine/internal/recovery"
	"github.com/theworker02/stateful-deployments-engine/internal/recoverypoint"
)
func exportCmd() *cobra.Command {
	var out string
	var state string
	var progress bool
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export active state to a Portable State Archive (ENGINE_PROVIDED)",
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
			src := state
			if src == "" {
				src = active.StatePath
			}
			if out == "" {
				out = filepath.Join(root, "archives", fmt.Sprintf("psa-%d", j.NextSeq()))
			}
			var onProg archive.ProgressFunc
			if progress {
				onProg = func(phase string, done, total int, detail string) {
					fmt.Fprintf(os.Stderr, "… %s %d/%d %s\n", phase, done, total, detail)
				}
			}
			m, stats, err := archive.ExportProgress(src, out, j.Epoch(), j.NextSeq()-1, j.Path(), onProg)
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(map[string]interface{}{"manifest": m, "stats": stats}, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "archive output directory")
	cmd.Flags().StringVar(&state, "state", "", "state root (default: active slot)")
	cmd.Flags().BoolVar(&progress, "progress", false, "print progress events to stderr")
	return cmd
}

func inspectArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect-archive [dir]",
		Short: "Inspect a Portable State Archive manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := archive.Inspect(args[0])
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(m, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
}

func verifyArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify-archive [dir]",
		Short: "Verify archive chunk digests and root digest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vr, err := archive.Verify(args[0])
			if vr != nil {
				b, _ := json.MarshalIndent(vr, "", "  ")
				fmt.Println(string(b))
			}
			return err
		},
	}
}

func importArchiveCmd() *cobra.Command {
	var dest string
	cmd := &cobra.Command{
		Use:   "import [archive-dir]",
		Short: "Import/restore a PSA into a destination state directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dest == "" {
				return fmt.Errorf("--dest required")
			}
			m, err := archive.Import(args[0], dest)
			if err != nil {
				return err
			}
			fmt.Printf("imported archive=%s digest=%s into %s\n", m.ArchiveID, m.RootDigest, dest)
			return nil
		},
	}
	cmd.Flags().StringVar(&dest, "dest", "", "destination state root")
	return cmd
}

func restoreCmd() *cobra.Command {
	var archivePath string
	var epoch uint64
	var dest string
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore from --archive (optional --epoch recovery point) via ENGINE_PROVIDED path",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			path := archivePath
			if path == "" {
				return fmt.Errorf("--archive required")
			}
			if epoch > 0 {
				rp := &recoverypoint.Store{Root: filepath.Join(root, "recovery")}
				pt, err := rp.FindByEpoch(epoch)
				if err != nil {
					return err
				}
				path = pt.ArchivePath
				fmt.Printf("using recovery point %s (epoch %d)\n", pt.ID, pt.Epoch)
			}
			if dest == "" {
				dest = filepath.Join(root, "restored")
			}
			tgt := &recovery.LocalTarget{Root: dest}
			rcpt, err := recovery.Restore(context.Background(), path, tgt, filepath.Join(dest, "receipts"), false)
			if rcpt != nil {
				b, _ := json.MarshalIndent(rcpt, "", "  ")
				fmt.Println(string(b))
			}
			return err
		},
	}
	cmd.Flags().StringVar(&archivePath, "archive", "", "PSA directory")
	cmd.Flags().Uint64Var(&epoch, "epoch", 0, "restore using recovery point at/before epoch")
	cmd.Flags().StringVar(&dest, "dest", "", "restore target root")
	return cmd
}

func escrowCmd() *cobra.Command {
	var targetKind, targetPath, name string
	cmd := &cobra.Command{
		Use:   "escrow [archive-dir]",
		Short: "Copy archive to escrow target and catalog it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			st := &escrow.Store{CatalogDir: filepath.Join(root, "escrow")}
			t := escrow.Target{Kind: escrow.TargetKind(targetKind), Name: name, RootPath: targetPath}
			if t.Name == "" {
				t.Name = "default"
			}
			if t.RootPath == "" {
				t.RootPath = filepath.Join(root, "escrow", "store")
			}
			e, err := st.Put(t, args[0])
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(e, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&targetKind, "target", string(escrow.LocalFilesystem), "LOCAL_FILESYSTEM|S3_COMPATIBLE|GENERIC_OBJECT_STORE")
	cmd.Flags().StringVar(&targetPath, "path", "", "target root path / mount")
	cmd.Flags().StringVar(&name, "name", "default", "target name")
	return cmd
}

func catalogCmd() *cobra.Command {
	rootCmd := &cobra.Command{Use: "catalog", Short: "Recovery catalog list/show/verify/prune"}
	var asJSON bool
	list := &cobra.Command{
		Use: "list", Short: "List cataloged archives",
		RunE: func(cmd *cobra.Command, args []string) error {
			st := &escrow.Store{CatalogDir: filepath.Join(rootDir(cmd), "escrow")}
			entries, err := st.List()
			if err != nil {
				return err
			}
			if asJSON {
				b, _ := json.MarshalIndent(entries, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			if len(entries) == 0 {
				fmt.Println("(catalog empty)")
				return nil
			}
			fmt.Print(escrow.FormatTable(entries))
			return nil
		},
	}
	list.Flags().BoolVar(&asJSON, "json", false, "JSON output")
	rootCmd.AddCommand(list)
	rootCmd.AddCommand(&cobra.Command{
		Use: "show [id]", Short: "Show one catalog entry", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st := &escrow.Store{CatalogDir: filepath.Join(rootDir(cmd), "escrow")}
			e, err := st.Show(args[0])
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(e, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use: "verify [id]", Short: "Re-verify cataloged archive", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st := &escrow.Store{CatalogDir: filepath.Join(rootDir(cmd), "escrow")}
			return st.Verify(args[0])
		},
	})
	var deleteData bool
	prune := &cobra.Command{
		Use: "prune [id]", Short: "Explicit destructive prune of catalog entry", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st := &escrow.Store{CatalogDir: filepath.Join(rootDir(cmd), "escrow")}
			return st.Prune(args[0], deleteData)
		},
	}
	prune.Flags().BoolVar(&deleteData, "delete-data", false, "also delete stored archive files")
	rootCmd.AddCommand(prune)
	return rootCmd
}

func recoveryPointsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "recovery-points", Short: "List or create SNAPSHOT recovery points"}
	cmd.AddCommand(&cobra.Command{
		Use: "list", Short: "List recovery points",
		RunE: func(c *cobra.Command, args []string) error {
			rp := &recoverypoint.Store{Root: filepath.Join(rootDir(c), "recovery")}
			pts, err := rp.List()
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(pts, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	})
	var archivePath string
	create := &cobra.Command{
		Use: "create", Short: "Create SNAPSHOT recovery point from --archive",
		RunE: func(c *cobra.Command, args []string) error {
			if archivePath == "" {
				return fmt.Errorf("--archive required")
			}
			rp := &recoverypoint.Store{Root: filepath.Join(rootDir(c), "recovery")}
			pt, err := rp.CreateFromArchive(archivePath)
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(pt, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	create.Flags().StringVar(&archivePath, "archive", "", "PSA directory")
	cmd.AddCommand(create)
	return cmd
}

func drPlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dr-plan",
		Short: "Disaster recovery planner with evidence and scorecard",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			p := &dr.Planner{
				Escrow: &escrow.Store{CatalogDir: filepath.Join(root, "escrow")},
				Points: &recoverypoint.Store{Root: filepath.Join(root, "recovery")},
			}
			plan, err := p.Plan()
			if err != nil {
				return err
			}
			if plan.Conclusion == dr.Unknown {
				fmt.Println("NOTE: UNKNOWN must never be treated as RECOVERABLE")
			}
			b, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
}

func fireDrillCmd() *cobra.Command {
	var archivePath string
	cmd := &cobra.Command{
		Use:   "fire-drill",
		Short: "Isolated restore drill; never modifies production",
		RunE: func(cmd *cobra.Command, args []string) error {
			if archivePath == "" {
				return fmt.Errorf("--archive required")
			}
			root := rootDir(cmd)
			res, err := dr.RunFireDrill(archivePath, filepath.Join(root, "fire-drill-tmp"))
			if res != nil {
				b, _ := json.MarshalIndent(res, "", "  ")
				fmt.Println(string(b))
				_ = os.MkdirAll(filepath.Join(root, "receipts"), 0o755)
				_ = os.WriteFile(filepath.Join(root, "receipts", "DR_FIRE_DRILL.json"), b, 0o644)
				// Best-effort: attach last-drill metadata to catalog entry when present.
				st := &escrow.Store{CatalogDir: filepath.Join(root, "escrow")}
				_ = st.RecordFireDrill(res.ArchiveID, escrow.FireDrillMeta{
					ReceiptID: res.ReceiptID, RanAt: res.StartedAt,
					DurationMs: res.DurationMs, RootDigestOK: res.RootDigestOK, Error: res.Error,
				})
			}
			return err
		},
	}
	cmd.Flags().StringVar(&archivePath, "archive", "", "PSA directory")
	return cmd
}

func cloneArchiveCmd() *cobra.Command {
	var dest string
	cmd := &cobra.Command{
		Use:   "clone-archive [archive-dir]",
		Short: "Safe state clone with new identity (does not mutate original)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dest == "" {
				return fmt.Errorf("--dest required")
			}
			m, err := archive.CloneReadOnly(args[0], dest)
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(m, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&dest, "dest", "", "clone output directory")
	return cmd
}
