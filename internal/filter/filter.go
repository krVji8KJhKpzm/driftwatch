package filter

import "github.com/yourorg/driftwatch/internal/drift"

// Options holds filtering criteria for drift results.
type Options struct {
	OnlyDrifted bool
	Names       []string
}

// Apply returns a filtered slice of drift results based on the given options.
func Apply(results []drift.Result, opts Options) []drift.Result {
	var out []drift.Result
	for _, r := range results {
		if opts.OnlyDrifted && !r.Drifted {
			continue
		}
		if len(opts.Names) > 0 && !containsName(opts.Names, r.Name) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func containsName(names []string, target string) bool {
	for _, n := range names {
		if n == target {
			return true
		}
	}
	return false
}
