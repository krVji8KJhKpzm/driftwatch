package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/alert"
	"github.com/driftwatch/internal/container"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/manifest"
	"github.com/spf13/cobra"
)

var (
	alertManifest string
	alertPolicy   string
	alertConfig   string
)

func init() {
	alertCmd := &cobra.Command{
		Use:   "alert",
		Short: "Evaluate drift against alert rules and exit non-zero on error-level alerts",
		RunE:  runAlert,
	}
	alertCmd.Flags().StringVarP(&alertManifest, "manifest", "m", "manifest.yaml", "path to manifest file")
	alertCmd.Flags().StringVarP(&alertConfig, "alerts", "a", "alerts.yaml", "path to alert rules file")
	rootCmd.AddCommand(alertCmd)
}

func runAlert(cmd *cobra.Command, args []string) error {
	entries, err := manifest.LoadFromFile(alertManifest)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	inspector := container.NewInspector(nil)
	var results []drift.Result
	for _, e := range entries {
		info, err := inspector.Inspect(e.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: inspect %s: %v\n", e.Name, err)
			continue
		}
		results = append(results, drift.Detect(e, *info))
	}

	cfg, err := alert.LoadConfig(alertConfig)
	if err != nil {
		return fmt.Errorf("load alert config: %w", err)
	}

	alerts := alert.Evaluate(results, cfg.Rules)
	if err := alert.Write(cmd.OutOrStdout(), alerts); err != nil {
		os.Exit(1)
	}
	return nil
}
