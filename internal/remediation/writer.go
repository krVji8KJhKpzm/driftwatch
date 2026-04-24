package remediation

import (
	"encoding/json"
	"fmt"
	"io"
)

// Write outputs suggestions in the given format to w.
func Write(w io.Writer, suggestions []Suggestion, format string) error {
	switch format {
	case "json":
		return writeJSON(w, suggestions)
	default:
		return writeText(w, suggestions)
	}
}

func writeText(w io.Writer, suggestions []Suggestion) error {
	if len(suggestions) == 0 {
		_, err := fmt.Fprintln(w, "No remediation suggestions.")
		return err
	}
	for _, s := range suggestions {
		_, err := fmt.Fprintf(w, "[%s] %s\n  expected: %s\n  actual:   %s\n  hint:     %s\n\n",
			s.Container, s.Field, s.Expected, s.Actual, s.Hint)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(w io.Writer, suggestions []Suggestion) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(suggestions)
}
