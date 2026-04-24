package policy

import "github.com/yourorg/driftwatch/internal/drift"

// Apply filters drift results according to the loaded policy.
// Env drifts for ignored keys are removed; image drift is cleared when ignore_image is set.
func Apply(results []drift.Result, p *Policy) []drift.Result {
	if p == nil {
		return results
	}
	out := make([]drift.Result, 0, len(results))
	for _, r := range results {
		r = applyToResult(r, p)
		out = append(out, r)
	}
	return out
}

func applyToResult(r drift.Result, p *Policy) drift.Result {
	rule := p.RuleFor(r.Name)

	if rule.IgnoreImage {
		r.ImageDrift = false
	}

	filtered := r.EnvDrifts[:0]
	for _, ed := range r.EnvDrifts {
		if !p.ShouldIgnoreEnv(r.Name, ed.Key) {
			filtered = append(filtered, ed)
		}
	}
	r.EnvDrifts = filtered
	r.Drifted = r.ImageDrift || len(r.EnvDrifts) > 0
	return r
}
