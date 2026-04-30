package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/baseline"
	"github.com/driftwatch/internal/quota"
	"github.com/spf13/cobra"
)

var (
	quotaConfigPath   string
	quotaBaselinePath string
	quotaFormat       string
)

func init() {
	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Check drift field counts against configured quotas",
		RunE:  runQuota,
	}
	cmd.Flags().StringVar(&quotaConfigPath, "config", "quota.json", "path to quota config file")
	cmd.Flags().StringVar(&quotaBaselinePath, "baseline", "baseline.json", "path to baseline file")
	cmd.Flags().StringVar(&quotaFormat, "format", "text", "output format: text|json")
	rootCmd.AddCommand(cmd)
}

func runQuota(cmd *cobra.Command, _ []string) error {
	cfg, err := quota.LoadConfig(quotaConfigPath)
	if err != nil {
		return fmt.Errorf("quota: load config: %w", err)
	}

	bl, err := baseline.Load(quotaBaselinePath)
	if err != nil {
		return fmt.Errorf("quota: load baseline: %w", err)
	}

	violations := quota.Evaluate(bl.Results, cfg)

	if err := quota.Write(cmd.OutOrStdout(), violations, quotaFormat); err != nil {
		return fmt.Errorf("quota: write output: %w", err)
	}

	if len(violations) > 0 {
		os.Exit(1)
	}
	return nil
}
