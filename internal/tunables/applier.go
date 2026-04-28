package tunables

import (
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Apply filters or adjusts drift results according to the provided Config.
// If cfg is nil the results are returned unchanged.
func Apply(results []drift.Result, cfg *Config) []drift.Result {
	if cfg == nil {
		return results
	}

	out := make([]drift.Result, 0, len(results))
	for _, r := range results {
		r = applyToResult(r, cfg)
		out = append(out, r)
	}
	return out
}

func applyToResult(r drift.Result, cfg *Config) drift.Result {
	filtered := r.Diffs[:0:0]

	for _, d := range r.Diffs {
		if cfg.IgnoreImageTag && d.Field == "image" {
			continue
		}
		if d.Field == "env" && len(cfg.EnvKeyPrefixes) > 0 && matchesPrefix(d.Key, cfg.EnvKeyPrefixes) {
			continue
		}
		filtered = append(filtered, d)
	}

	if cfg.MaxEnvDiffs > 0 {
		var envCount int
		capped := filtered[:0:0]
		for _, d := range filtered {
			if d.Field == "env" {
				if envCount >= cfg.MaxEnvDiffs {
					continue
				}
				envCount++
			}
			capped = append(capped, d)
		}
		filtered = capped
	}

	r.Diffs = filtered
	r.Drifted = len(filtered) > 0
	return r
}

func matchesPrefix(key string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
