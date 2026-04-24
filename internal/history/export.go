package history

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

// ExportFormat defines the supported export formats.
type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatJSON ExportFormat = "json"
)

// Export writes history entries to w in the given format.
func Export(entries []Entry, format ExportFormat, w io.Writer) error {
	switch format {
	case FormatCSV:
		return exportCSV(entries, w)
	case FormatJSON:
		return exportJSON(entries, w)
	default:
		return fmt.Errorf("unsupported export format: %q", format)
	}
}

func exportCSV(entries []Entry, w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write([]string{"timestamp", "total", "drifted", "clean"}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, e := range entries {
		drifted := 0
		for _, r := range e.Results {
			if r.Drifted {
				drifted++
			}
		}
		total := len(e.Results)
		clean := total - drifted

		row := []string{
			e.Timestamp.Format(time.RFC3339),
			strconv.Itoa(total),
			strconv.Itoa(drifted),
			strconv.Itoa(clean),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	return cw.Error()
}

func exportJSON(entries []Entry, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
