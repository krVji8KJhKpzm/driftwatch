package diff

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/driftwatch/internal/drift"
)

// Write renders highlighted diffs for all results to w in the given format.
func Write(w io.Writer, results []drift.Result, format string) error {
	switch format {
	case "json":
		return writeJSON(w, results)
	default:
		return writeText(w, results)
	}
}

func writeText(w io.Writer, results []drift.Result) error {
	for _, r := range results {
		line := FormatAll(r)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

type jsonDiff struct {
	Name    string      `json:"name"`
	Drifted bool        `json:"drifted"`
	Fields  []FieldDiff `json:"fields"`
}

func writeJSON(w io.Writer, results []drift.Result) error {
	var out []jsonDiff
	for _, r := range results {
		out = append(out, jsonDiff{
			Name:    r.Name,
			Drifted: r.Drifted,
			Fields:  Highlight(r),
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
