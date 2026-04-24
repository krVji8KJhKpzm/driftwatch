package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/container"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/manifest"
	"github.com/driftwatch/internal/remediation"
	"github.com/spf13/cobra"
)

var (
	remediationFormat     string
	remediationManifest  string
)

func init() {
	cmd := &cobra.Command{
		Use:   "remediation",
		Short: "Show remediation suggestions for detected drift",
		RunE:  runRemediation,
	}
	cmd.Flags().StringVarP(&remediationFormat, "format", "f", "text", "Output format: text or json")
	cmd.Flags().StringVarP(&remediationManifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	rootCmd.AddCommand(cmd)
}

func runRemediation(cmd *cobra.Command, args []string) error {
	entries, err := manifest.LoadFromFile(remediationManifest)
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
		results = append(results, drift.Detect(e, info))
	}

	suggestions := remediation.Generate(results)
	return remediation.Write(os.Stdout, suggestions, remediationFormat)
}
