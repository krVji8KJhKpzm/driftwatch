package diff

import (
	"fmt"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// FieldDiff represents a single field-level difference.
type FieldDiff struct {
	Field    string
	Expected string
	Actual   string
}

// Highlight extracts per-field diffs from a drift result for human-readable display.
func Highlight(result drift.Result) []FieldDiff {
	var diffs []FieldDiff

	if !result.Drifted {
		return diffs
	}

	for _, d := range result.Differences {
		parts := strings.SplitN(d, ":", 3)
		if len(parts) == 3 {
			diffs = append(diffs, FieldDiff{
				Field:    strings.TrimSpace(parts[0]),
				Expected: strings.TrimSpace(parts[1]),
				Actual:   strings.TrimSpace(parts[2]),
			})
		} else {
			diffs = append(diffs, FieldDiff{
				Field:    d,
				Expected: "",
				Actual:   "",
			})
		}
	}

	return diffs
}

// FormatDiff renders a FieldDiff as a colored diff-style string.
func FormatDiff(fd FieldDiff) string {
	if fd.Expected == "" && fd.Actual == "" {
		return fmt.Sprintf("  ~ %s (changed)", fd.Field)
	}
	return fmt.Sprintf("  ~ %s\n    - expected: %s\n    + actual:   %s", fd.Field, fd.Expected, fd.Actual)
}

// FormatAll renders all FieldDiffs for a result into a single string block.
func FormatAll(result drift.Result) string {
	diffs := Highlight(result)
	if len(diffs) == 0 {
		return fmt.Sprintf("[%s] no drift detected", result.Name)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] drift detected:\n", result.Name))
	for _, d := range diffs {
		sb.WriteString(FormatDiff(d))
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}
