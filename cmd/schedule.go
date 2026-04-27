package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/driftwatch/internal/schedule"
)

func init() {
	var configPath string

	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage and evaluate drift-check schedules",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured schedule entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := schedule.LoadConfig(configPath)
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tINTERVAL\tMANIFEST\tENABLED\tLAST RUN")
			for _, e := range cfg.Entries {
				lastRun := "never"
				if !e.LastRun.IsZero() {
					lastRun = e.LastRun.Format(time.RFC3339)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\n",
					e.Name, e.Interval, e.Manifest, e.Enabled, lastRun)
			}
			return w.Flush()
		},
	}

	dueCmd := &cobra.Command{
		Use:   "due",
		Short: "Show schedule entries that are currently due to run",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := schedule.LoadConfig(configPath)
			if err != nil {
				return err
			}
			due := schedule.Due(cfg, time.Now())
			if len(due) == 0 {
				fmt.Println("No schedules are currently due.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tMANIFEST\tINTERVAL")
			for _, e := range due {
				fmt.Fprintf(w, "%s\t%s\t%s\n", e.Name, e.Manifest, e.Interval)
			}
			return w.Flush()
		},
	}

	for _, sub := range []*cobra.Command{listCmd, dueCmd} {
		sub.Flags().StringVarP(&configPath, "config", "c", "schedule.json", "Path to schedule config file")
		cmd.AddCommand(sub)
	}

	rootCmd.AddCommand(cmd)
}
