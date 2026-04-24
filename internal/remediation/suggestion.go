package remediation

import (
	"fmt"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Suggestion represents a remediation hint for a drifted field.
type Suggestion struct {
	Container string
	Field     string
	Expected  string
	Actual    string
	Hint      string
}

// Generate produces remediation suggestions for all drifted results.
func Generate(results []drift.Result) []Suggestion {
	var suggestions []Suggestion
	for _, r := range results {
		if !r.Drifted {
			continue
		}
		for _, d := range r.Diffs {
			suggestions = append(suggestions, Suggestion{
				Container: r.Name,
				Field:     d.Field,
				Expected:  d.Expected,
				Actual:    d.Actual,
				Hint:      buildHint(r.Name, d.Field, d.Expected),
			})
		}
	}
	return suggestions
}

func buildHint(container, field, expected string) string {
	switch {
	case field == "image":
		return fmt.Sprintf("Update container '%s' to use image: %s", container, expected)
	case strings.HasPrefix(field, "env:"):
		key := strings.TrimPrefix(field, "env:")
		return fmt.Sprintf("Set env var %s=%s on container '%s'", key, expected, container)
	default:
		return fmt.Sprintf("Reconcile field '%s' on container '%s' to: %s", field, container, expected)
	}
}
