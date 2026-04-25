package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/driftwatch/internal/container"
	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/filter"
	"github.com/yourorg/driftwatch/internal/manifest"
	"github.com/yourorg/driftwatch/internal/scorecard"
)

var (
	scorecardManifest string
	scorecardFormat   string
	scorecardOnlyDrift bool
	scorecardNames    []string
)

func init() {
	scorecardCmd := &cobra.Command{
		Use:   "scorecard",
		Short: "Generate a drift scorecard graded by container health",
		Long: `Inspect running containers against their manifests and produce a
scorecard that grades each container from A to F based on the number
and severity of configuration drifts detected.`,
		RunE: runScorecard,
	}

	scorecardCmd.Flags().StringVarP(&scorecardManifest, "manifest", "m", "manifest.yaml", "Path to the manifest file")
	scorecardCmd.Flags().StringVarP(&scorecardFormat, "format", "f", "text", "Output format: text or json")
	scorecardCmd.Flags().BoolVar(&scorecardOnlyDrift, "only-drift", false, "Include only drifted containers in the scorecard")
	scorecardCmd.Flags().StringSliceVarP(&scorecardNames, "name", "n", nil, "Filter by container name (repeatable)")

	rootCmd.AddCommand(scorecardCmd)
}

func runScorecard(cmd *cobra.Command, args []string) error {
	entries, err := manifest.LoadFromFile(scorecardManifest)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	inspector, err := container.NewInspector()
	if err != nil {
		return fmt.Errorf("creating inspector: %w", err)
	}

	results, err := drift.Detect(entries, inspector)
	if err != nil {
		return fmt.Errorf("detecting drift: %w", err)
	}

	// Apply name and drift filters before scoring.
	results = filter.Apply(results, filter.Options{
		OnlyDrifted: scorecardOnlyDrift,
		Names:       scorecardNames,
	})

	card := scorecard.Build(results)

	if err := scorecard.Write(os.Stdout, card, scorecardFormat); err != nil {
		return fmt.Errorf("writing scorecard: %w", err)
	}

	return nil
}
