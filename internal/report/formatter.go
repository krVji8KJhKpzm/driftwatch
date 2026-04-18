package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/driftwatch/internal/drift"
)

// Format controls the output format of the report.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Write writes a drift report to the provided writer in the specified format.
func Write(w io.Writer, results []drift.Result, format Format) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, results)
	default:
		return writeText(w, results)
	}
}

func writeText(w io.Writer, results []drift.Result) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(w, "No drift detected.")
		return err
	}
	for _, r := range results {
		if !r.Drifted {
			continue
		}
		fmt.Fprintf(w, "[DRIFT] %s\n", r.Name)
		for _, d := range r.Differences {
			fmt.Fprintf(w, "  - %s: expected %q, got %q\n", d.Field, d.Expected, d.Actual)
		}
	}
	return nil
}

func writeJSON(w io.Writer, results []drift.Result) error {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, r := range results {
		if !r.Drifted {
			continue
		}
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("  {\"name\": %q, \"differences\": [", r.Name))
		for j, d := range r.Differences {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("{\"field\": %q, \"expected\": %q, \"actual\": %q}", d.Field, d.Expected, d.Actual))
		}
		sb.WriteString("]}")
	}
	sb.WriteString("\n]")
	_, err := fmt.Fprintln(w, sb.String())
	return err
}
