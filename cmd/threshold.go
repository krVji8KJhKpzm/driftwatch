package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/threshold"
	"github.com/spf13/cobra"
)

func init() {
	var (
		manifestPath  string
		policyPath    string
		thresholdPath string
		outputFormat  string
		failOnBreach  bool
	)

	cmd := &cobra.Command{
		Use:   "threshold",
		Short: "Evaluate drift results against configured thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runThreshold(manifestPath, policyPath, thresholdPath, outputFormat, failOnBreach)
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", "manifest.yaml", "path to manifest file")
	cmd.Flags().StringVarP(&policyPath, "policy", "p", "", "path to policy file (optional)")
	cmd.Flags().StringVarP(&thresholdPath, "config", "c", "threshold.json", "path to threshold config")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text|json")
	cmd.Flags().BoolVar(&failOnBreach, "fail", false, "exit with non-zero status if violations found")

	rootCmd.AddCommand(cmd)
}

func runThreshold(manifestPath, policyPath, thresholdPath, format string, failOnBreach bool) error {
	results, err := collectResults(manifestPath, policyPath)
	if err != nil {
		return fmt.Errorf("threshold: collect results: %w", err)
	}

	cfg, err := threshold.LoadConfig(thresholdPath)
	if err != nil {
		return fmt.Errorf("threshold: load config: %w", err)
	}

	violations := threshold.Evaluate(cfg, results)

	if err := threshold.Write(os.Stdout, violations, format); err != nil {
		return fmt.Errorf("threshold: write output: %w", err)
	}

	if failOnBreach && len(violations) > 0 {
		os.Exit(1)
	}
	return nil
}
