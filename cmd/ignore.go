package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"driftwatch/internal/ignore"
)

var ignoreFile string

func init() {
	ignoreCmd := &cobra.Command{
		Use:   "ignore",
		Short: "Manage and inspect ignore rules for drift detection",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all active ignore rules from the config file",
		RunE:  runIgnoreList,
	}
	listCmd.Flags().StringVar(&ignoreFile, "config", "ignore.json", "path to ignore config file")

	ignoreCmd.AddCommand(listCmd)
	rootCmd.AddCommand(ignoreCmd)
}

func runIgnoreList(cmd *cobra.Command, args []string) error {
	cfg, err := ignore.LoadConfig(ignoreFile)
	if err != nil {
		return fmt.Errorf("loading ignore config: %w", err)
	}

	if len(cfg.Rules) == 0 {
		fmt.Fprintln(os.Stdout, "No ignore rules defined.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CONTAINER\tFIELDS")
	for _, rule := range cfg.Rules {
		fmt.Fprintf(w, "%s\t%v\n", rule.Container, rule.Fields)
	}
	return w.Flush()
}
