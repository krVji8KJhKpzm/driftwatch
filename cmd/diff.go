package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/container"
	"github.com/driftwatch/internal/diff"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/manifest"
	"github.com/spf13/cobra"
)

var (
	diffFormat string
	diffNames  []string
)

func init() {
	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Show field-level drift highlights for running containers",
		RunE:  runDiff,
	}
	diffCmd.Flags().StringVarP(&diffFormat, "format", "f", "text", "Output format: text or json")
	diffCmd.Flags().StringArrayVarP(&diffNames, "name", "n", nil, "Filter by container name")
	diffCmd.Flags().StringVarP(&manifestFile, "manifest", "m", "manifest.yaml", "Path to manifest file")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	entries, err := manifest.LoadFromFile(manifestFile)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	inspector := container.NewInspector(nil)
	var results []drift.Result

	for _, entry := range entries {
		if len(diffNames) > 0 && !containsStr(diffNames, entry.Name) {
			continue
		}
		info, err := inspector.Inspect(entry.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: inspect %s: %v\n", entry.Name, err)
			continue
		}
		results = append(results, drift.Detect(entry, info))
	}

	if len(results) == 0 {
		fmt.Println("no containers matched")
		return nil
	}

	return diff.Write(os.Stdout, results, diffFormat)
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
