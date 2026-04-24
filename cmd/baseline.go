package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/baseline"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/report"
	"github.com/spf13/cobra"
)

var (
	baselinePath   string
	baselineRecord bool
)

func init() {
	baselineCmd := &cobra.Command{
		Use:   "baseline",
		Short: "Record or compare against a configuration baseline",
		RunE:  runBaseline,
	}

	baselineCmd.Flags().StringVarP(&baselinePath, "file", "f", "baseline.json", "path to baseline file")
	baselineCmd.Flags().BoolVarP(&baselineRecord, "record", "r", false, "record current state as new baseline")

	rootCmd.AddCommand(baselineCmd)
}

func runBaseline(cmd *cobra.Command, args []string) error {
	results, err := collectResults()
	if err != nil {
		return fmt.Errorf("baseline: collect results: %w", err)
	}

	if baselineRecord {
		if err := baseline.Save(baselinePath, results); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Baseline recorded to %s (%d containers)\n", baselinePath, len(results))
		return nil
	}

	b, err := baseline.Load(baselinePath)
	if err != nil {
		return err
	}

	drifted := baseline.Compare(b, results)
	if len(drifted) == 0 {
		fmt.Fprintln(os.Stdout, "No drift detected against baseline.")
		return nil
	}

	return report.Write(os.Stdout, drifted, outputFormat)
}

// collectResults is a thin shim that wires manifest + inspector + detector.
// Shared logic lives here so commands stay small.
func collectResults() ([]drift.Result, error) {
	// Reuse flags already registered by the root / run command.
	manifests, err := loadManifests()
	if err != nil {
		return nil, err
	}
	inspector := buildInspector()
	return runDetector(inspector, manifests)
}
