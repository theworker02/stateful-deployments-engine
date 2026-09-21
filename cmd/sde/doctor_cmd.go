package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/theworker02/stateful-deployments-engine/internal/doctor"
)

func doctorCmd() *cobra.Command {
	var jsonOut bool
	var probe bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check local install, workspace, and Railway env readiness",
		Long: `Reports whether SDE is ready for local demos or live Railway GraphQL.

Examples:
  sde doctor
  sde doctor --root .sde --probe   # validate Railway token when SDE_RAILWAY_LIVE=1
  sde doctor --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := doctor.Run(context.Background(), doctor.Options{
				Root:      rootDir(cmd),
				ProbeLive: probe,
				Timeout:   10 * time.Second,
			})
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(r); err != nil {
					return err
				}
			} else {
				fmt.Print(doctor.FormatHuman(r))
			}
			if !r.Ready {
				return fmt.Errorf("doctor: not ready — fix failed checks")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().BoolVar(&probe, "probe", false, "when live mode is on, call Railway GraphQL ValidateToken")
	return cmd
}
