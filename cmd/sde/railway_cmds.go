package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/railway"
)

func railwayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "railway",
		Short: "Railway adapter helpers (ENGINE_PROVIDED + optional live GraphQL)",
	}
	cmd.AddCommand(railwayCloneCmd(), railwayStatusCmd())
	return cmd
}

func railwayStatusCmd() *cobra.Command {
	var probe bool
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Railway live-mode configuration (optionally probe GraphQL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			live := railway.LiveEnabled()
			out := map[string]interface{}{
				"live_enabled":   live,
				"graphql_url":    envOr("RAILWAY_GRAPHQL_URL", "https://backboard.railway.com/graphql/v2"),
				"project_id":     os.Getenv("RAILWAY_PROJECT_ID"),
				"environment_id": os.Getenv("RAILWAY_ENVIRONMENT_ID"),
				"service_id":     os.Getenv("RAILWAY_SERVICE_ID"),
				"shadow_service": os.Getenv("RAILWAY_SHADOW_SERVICE_ID"),
				"token_set":      tokenPresent(),
				"err_not_wired":  !live,
			}
			if probe && live {
				ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
				defer cancel()
				a, err := railway.NewFromEnv()
				if err != nil {
					out["probe_ok"] = false
					out["probe_error"] = err.Error()
				} else if err := a.ValidateAPI(ctx); err != nil {
					out["probe_ok"] = false
					out["probe_error"] = err.Error()
				} else {
					out["probe_ok"] = true
				}
			}
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			fmt.Printf("railway live:     %v\n", live)
			fmt.Printf("token set:       %v\n", out["token_set"])
			fmt.Printf("project:         %v\n", out["project_id"])
			fmt.Printf("environment:     %v\n", out["environment_id"])
			fmt.Printf("service:         %v\n", out["service_id"])
			fmt.Printf("graphql:         %v\n", out["graphql_url"])
			if !live {
				fmt.Println("\nTip: copy .env.example → .env, set SDE_RAILWAY_LIVE=1 and RAILWAY_TOKEN, then:")
				fmt.Println("  sde railway status --probe")
				fmt.Println("See docs/RAILWAY_LIVE.md")
			} else if probe {
				if ok, _ := out["probe_ok"].(bool); ok {
					fmt.Println("\nprobe: OK — GraphQL token validated")
				} else {
					fmt.Printf("\nprobe: FAIL — %v\n", out["probe_error"])
					os.Exit(1)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&probe, "probe", false, "validate token against Railway GraphQL")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON output")
	return cmd
}

func railwayCloneCmd() *cobra.Command {
	var src, dest, mode string
	var sanitize bool
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone volume state to new identity (ENGINE_PROVIDED PSA); writes CLONE_RECEIPT.json",
		Long: `Exports source state to a Portable State Archive with a new archive identity,
optionally sanitizes text secrets, verifies, and materializes read-only or writable trees.

This is ENGINE_PROVIDED orchestration around the generic core — not a Railway-native
volume backup clone, and not affiliated with / endorsed by Railway.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootDir(cmd)
			if src == "" {
				plat, j, _, _, err := openStack(root)
				if err != nil {
					return err
				}
				defer j.Close()
				active, err := plat.ActiveSlot(context.Background())
				if err != nil {
					return err
				}
				src = active.StatePath
			}
			if dest == "" {
				dest = filepath.Join(root, "clones", "latest")
			}
			a := railway.NewWithConfig(railway.Config{})
			rcpt, err := a.CloneVolume(context.Background(), railway.CloneRequest{
				SourceState: src,
				DestDir:     dest,
				Mode:        railway.CloneMode(mode),
				Sanitize:    sanitize,
			})
			if rcpt != nil {
				b, _ := json.MarshalIndent(rcpt, "", "  ")
				fmt.Println(string(b))
			}
			return err
		},
	}
	cmd.Flags().StringVar(&src, "source", "", "source state directory (default: active slot)")
	cmd.Flags().StringVar(&dest, "dest", "", "clone output directory")
	cmd.Flags().StringVar(&mode, "mode", string(railway.CloneReadOnly), "read-only|writable")
	cmd.Flags().BoolVar(&sanitize, "sanitize", false, "apply text redaction/tokenization hooks")
	return cmd
}

func tokenPresent() bool {
	return os.Getenv("RAILWAY_TOKEN") != "" ||
		os.Getenv("RAILWAY_API_TOKEN") != "" ||
		os.Getenv("RAILWAY_PROJECT_TOKEN") != ""
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
