package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/container"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/manifest"
	"github.com/driftwatch/internal/metrics"
	"github.com/spf13/cobra"
)

func init() {
	var (
		manifestFile string
		format       string
	)

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show aggregated drift metrics across all containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetrics(manifestFile, format)
		},
	}

	cmd.Flags().StringVarP(&manifestFile, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text or json")

	rootCmd.AddCommand(cmd)
}

func runMetrics(manifestFile, format string) error {
	entries, err := manifest.LoadFromFile(manifestFile)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	insp, err := container.NewInspector()
	if err != nil {
		return fmt.Errorf("create inspector: %w", err)
	}

	var results []drift.Result
	for _, entry := range entries {
		info, err := insp.Inspect(entry.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: inspect %s: %v\n", entry.Name, err)
			continue
		}
		results = append(results, drift.Detect(entry, info))
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "no containers inspected")
		return nil
	}

	s := metrics.Compute(results)
	return metrics.Write(os.Stdout, s, format)
}
