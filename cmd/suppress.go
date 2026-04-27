package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"driftwatch/internal/suppress"
)

func init() {
	var configPath string

	cmd := &cobra.Command{
		Use:   "suppress",
		Short: "List active drift suppression rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSuppress(configPath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "suppress.json", "path to suppression config file")
	rootCmd.AddCommand(cmd)
}

func runSuppress(configPath string) error {
	cfg, err := suppress.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading suppress config: %w", err)
	}

	if len(cfg.Rules) == 0 {
		fmt.Println("No suppression rules defined.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CONTAINER\tFIELD\tREASON\tEXPIRES")
	for _, r := range cfg.Rules {
		expires := "never"
		if !r.Expires.IsZero() {
			expires = r.Expires.Format("2006-01-02")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Container, r.Field, r.Reason, expires)
	}
	return w.Flush()
}
