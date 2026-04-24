package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourorg/driftwatch/internal/policy"
)

var policyFile string

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Validate a policy file",
	Long:  "Load and validate a driftwatch policy YAML file, reporting any parse errors.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if policyFile == "" {
			return fmt.Errorf("--file is required")
		}
		p, err := policy.LoadPolicy(policyFile)
		if err != nil {
			return fmt.Errorf("failed to load policy: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Policy loaded successfully: %d rule(s) defined\n", len(p.Rules))
		for name, r := range p.Rules {
			fmt.Fprintf(os.Stdout, "  container=%q ignore_image=%v ignored_envs=%v\n",
				name, r.IgnoreImage, r.IgnoreEnvs)
		}
		return nil
	},
}

func init() {
	policyCmd.Flags().StringVarP(&policyFile, "file", "f", "", "Path to policy YAML file")
	rootCmd.AddCommand(policyCmd)
}
