package threshold

import (
	"encoding/json"
	"fmt"
	"io"
)

// Write renders violations to w in the given format ("text" or "json").
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
		_, err := fmt.Fprintln(w, "threshold: all checks passed")
		return err
	}
	_, err := fmt.Fprintf(w, "threshold: %d violation(s) detected\n", len(violations))
	if err != nil {
		return err
	}
	for _, v := range violations {
		if _, err := fmt.Fprintf(w, "  [%s] %s\n", v.Rule, v.Message); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(w io.Writer, violations []Violation) error {
	type output struct {
		Violations []Violation `json:"violations"`
		Count      int         `json:"count"`
		Passed     bool        `json:"passed"`
	}
	out := output{
		Violations: violations,
		Count:      len(violations),
		Passed:     len(violations) == 0,
	}
	if out.Violations == nil {
		out.Violations = []Violation{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
