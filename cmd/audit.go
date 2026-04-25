package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/driftwatch/internal/audit"
	"github.com/spf13/cobra"
)

var (
	auditLogPath string
	auditSince   string
)

func init() {
	auditCmd := &cobra.Command{
		Use:   "audit",
		Short: "View the audit log of drift scan events",
		RunE:  runAudit,
	}
	auditCmd.Flags().StringVar(&auditLogPath, "log", "driftwatch-audit.log", "path to audit log file")
	auditCmd.Flags().StringVar(&auditSince, "since", "", "only show events after this RFC3339 timestamp")
	rootCmd.AddCommand(auditCmd)
}

func runAudit(cmd *cobra.Command, args []string) error {
	events, err := audit.LoadEvents(auditLogPath)
	if err != nil {
		return fmt.Errorf("failed to load audit log: %w", err)
	}
	if len(events) == 0 {
		fmt.Println("No audit events found.")
		return nil
	}

	var since time.Time
	if auditSince != "" {
		since, err = time.Parse(time.RFC3339, auditSince)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tTYPE\tCONTAINER\tDETAILS")

	count := 0
	for _, evt := range events {
		if !since.IsZero() && evt.Timestamp.Before(since) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			evt.Timestamp.Format(time.RFC3339),
			evt.Type,
			evt.ContainerName,
			evt.Details,
		)
		count++
	}
	w.Flush()

	if count == 0 {
		fmt.Println("No events match the given filter.")
	}
	return nil
}
