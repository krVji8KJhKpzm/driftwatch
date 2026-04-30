package quota

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

// Write outputs quota violations in the requested format ("text" or "json").
func Write(w io.Writer, violations []Violation, format string) error {
	switch format {
	case "json":
		return writeJSON(w, violations)
	default:
		return writeText(w, violations)
	}
}

func writeText(w io.Writer, violations []Violation) error {
	if len(violations) == 0 {
		_, err := fmt.Fprintln(w, "quota: all containers within limits")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "CONTAINER\tDRIFT FIELDS\tLIMIT")
	for _, v := range violations {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\n", v.Container, v.DriftCount, v.Limit)
	}
	return tw.Flush()
}

func writeJSON(w io.Writer, violations []Violation) error {
	type output struct {
		Violations []Violation `json:"violations"`
		Total       int         `json:"total"`
	}
	out := output{Violations: violations, Total: len(violations)}
	if out.Violations == nil {
		out.Violations = []Violation{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
