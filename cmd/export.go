package cmd

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/baseline"
	"github.com/driftwatch/internal/export"
	"github.com/spf13/cobra"
)

var (
	exportFormat    string
	exportOutputDir string
	exportFilename  string
	exportBaseline  string
)

func init() {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export drift results to a file (json, markdown, html)",
		RunE:  runExport,
	}

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Output format: json, markdown, html")
	exportCmd.Flags().StringVarP(&exportOutputDir, "output-dir", "o", ".", "Directory to write the report into")
	exportCmd.Flags().StringVar(&exportFilename, "filename", "", "Override output filename (default: drift-report.<format>)")
	exportCmd.Flags().StringVar(&exportBaseline, "baseline", "", "Path to baseline file to compare against")

	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	fmt, err := export.ParseFormat(exportFormat)
	if err != nil {
		return err
	}

	if exportBaseline == "" {
		return fmt.Errorf("--baseline is required")
	}

	results, err := baseline.Load(exportBaseline)
	if err != nil {
		return fmt.Errorf("export: load baseline: %w", err)
	}

	opts := export.Options{
		Format:    fmt,
		OutputDir: exportOutputDir,
		Filename:  exportFilename,
	}

	if err := export.Export(results, opts); err != nil {
		return err
	}

	dest := exportOutputDir
	if exportFilename != "" {
		dest = exportOutputDir + "/" + exportFilename
	} else {
		dest = exportOutputDir + "/drift-report." + exportFormat
	}
	fmt.Fprintf(os.Stdout, "Report written to %s\n", dest)
	return nil
}
