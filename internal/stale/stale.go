package stale

import (
	"time"

	"github.com/driftwatch/internal/drift"
)

// ContainerAge holds staleness metadata for a single container.
type ContainerAge struct {
	Name      string
	Drifted   bool
	DriftAge  time.Duration // how long the container has been in drift
	Stale     bool          // true if drift age exceeds threshold
}

// Report is the result of a staleness evaluation.
type Report struct {
	Entries   []ContainerAge
	Threshold time.Duration
}

// Evaluate checks each drifted result against a staleness threshold.
// firstSeen maps container name to the time drift was first observed.
func Evaluate(results []drift.Result, firstSeen map[string]time.Time, threshold time.Duration) Report {
	now := time.Now()
	report := Report{Threshold: threshold}

	for _, r := range results {
		entry := ContainerAge{
			Name:    r.Name,
			Drifted: r.Drifted,
		}

		if r.Drifted {
			if t, ok := firstSeen[r.Name]; ok {
				entry.DriftAge = now.Sub(t)
				entry.Stale = threshold > 0 && entry.DriftAge >= threshold
			}
		}

		report.Entries = append(report.Entries, entry)
	}

	return report
}

// StaleCount returns the number of containers marked stale.
func (r Report) StaleCount() int {
	count := 0
	for _, e := range r.Entries {
		if e.Stale {
			count++
		}
	}
	return count
}
