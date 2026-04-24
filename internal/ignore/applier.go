package ignore

import (
	"strings"

	"driftwatch/internal/drift"
)

// Apply filters out drift entries that match ignore rules in cfg.
// If cfg is nil, results are returned unchanged.
func Apply(cfg *Config, results []drift.Result) []drift.Result {
	if cfg == nil || len(cfg.Rules) == 0 {
		return results
	}
	out := make([]drift.Result, 0, len(results))
	for _, r := range results {
		out = append(out, applyToResult(cfg, r))
	}
	return out
}

func applyToResult(cfg *Config, r drift.Result) drift.Result {
	if !r.Drifted {
		return r
	}
	filtered := r.Diffs[:0]
	for _, d := range r.Diffs {
		field := classifyField(d.Field)
		if !ShouldIgnoreField(cfg, r.Name, field) {
			filtered = append(filtered, d)
		}
	}
	r.Diffs = filtered
	r.Drifted = len(filtered) > 0
	return r
}

// classifyField normalises a diff field name into the form used in ignore rules
// (e.g. "env.PORT" stays as-is, "image" stays as-is).
func classifyField(field string) string {
	return strings.ToLower(field)
}
