package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/driftwatch/internal/lint"
)

func init() {
	var manifestPath string
	var failOnWarning bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint drift results and report issues by severity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(manifestPath, failOnWarning)
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", "manifest.yaml", "path to manifest file")
	cmd.Flags().BoolVar(&failOnWarning, "fail-on-warning", false, "exit non-zero if warnings are found")

	rootCmd.AddCommand(cmd)
}

func runLint(manifestPath string, failOnWarning bool) error {
	results, err := collectResults(manifestPath)
	if err != nil {
		return fmt.Errorf("collecting results: %w", err)
	}

	report := lint.Run(results)

	if !report.HasIssues() {
		fmt.Println("✔ no lint issues found")
		return nil
	}

	printFindings(report)

	fmt.Fprintf(os.Stdout, "\nsummary: %d error(s), %d warning(s)\n",
		report.ErrorCount, report.WarningCount)

	if report.ErrorCount > 0 {
		return fmt.Errorf("lint failed with %d error(s)", report.ErrorCount)
	}
	if failOnWarning && report.WarningCount > 0 {
		return fmt.Errorf("lint failed with %d warning(s)", report.WarningCount)
	}
	return nil
}

// printFindings writes each lint finding to stdout with an icon indicating severity.
func printFindings(report lint.Report) {
	for _, f := range report.Findings {
		icon := severityIcon(f.Severity)
		fmt.Fprintf(os.Stdout, "%s [%s] %s — %s\n", icon, f.Severity, f.Container, f.Message)
	}
}

func severityIcon(s lint.Severity) string {
	switch s {
	case lint.SeverityError:
		return "✖"
	case lint.SeverityWarning:
		return "⚠"
	default:
		return "ℹ"
	}
}
