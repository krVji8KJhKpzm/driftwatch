package export

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Format represents a supported export format.
type Format string

const (
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
)

// Options configures the export behaviour.
type Options struct {
	Format    Format
	OutputDir string
	Filename  string
}

// Export writes drift results to the specified format and destination.
func Export(results []drift.Result, opts Options) error {
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("export: create output dir: %w", err)
	}

	name := opts.Filename
	if name == "" {
		name = fmt.Sprintf("drift-report.%s", string(opts.Format))
	}

	path := filepath.Join(opts.OutputDir, name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("export: create file: %w", err)
	}
	defer f.Close()

	switch opts.Format {
	case FormatJSON:
		return writeJSON(f, results)
	case FormatMarkdown:
		return writeMarkdown(f, results)
	case FormatHTML:
		return writeHTML(f, results)
	default:
		return fmt.Errorf("export: unsupported format %q", opts.Format)
	}
}

func writeJSON(w io.Writer, results []drift.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func writeMarkdown(w io.Writer, results []drift.Result) error {
	fmt.Fprintln(w, "# Drift Report")
	fmt.Fprintln(w)
	for _, r := range results {
		status := "✅ Clean"
		if r.Drifted {
			status = "⚠️ Drifted"
		}
		fmt.Fprintf(w, "## %s — %s\n", r.Name, status)
		if r.Drifted {
			for _, d := range r.Diffs {
				fmt.Fprintf(w, "- **%s**: `%s` → `%s`\n", d.Field, d.Expected, d.Actual)
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}

func writeHTML(w io.Writer, results []drift.Result) error {
	fmt.Fprintln(w, "<!DOCTYPE html><html><head><title>Drift Report</title></head><body>")
	fmt.Fprintln(w, "<h1>Drift Report</h1>")
	for _, r := range results {
		class := "clean"
		if r.Drifted {
			class = "drifted"
		}
		fmt.Fprintf(w, "<section class=%q><h2>%s</h2>", class, r.Name)
		if r.Drifted {
			fmt.Fprintln(w, "<ul>")
			for _, d := range r.Diffs {
				fmt.Fprintf(w, "<li><strong>%s</strong>: %s → %s</li>\n",
					d.Field,
					strings.ReplaceAll(d.Expected, "<", "&lt;"),
					strings.ReplaceAll(d.Actual, "<", "&lt;"))
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprintln(w, "</section>")
	}
	fmt.Fprintln(w, "</body></html>")
	return nil
}
