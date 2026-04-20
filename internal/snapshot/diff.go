package snapshot

import (
	"fmt"

	"github.com/driftwatch/internal/drift"
)

// DiffEntry describes a change in drift status between two snapshots.
type DiffEntry struct {
	Name     string
	Previous []drift.Difference
	Current  []drift.Difference
}

// Diff compares two snapshots and returns entries whose drift state changed.
func Diff(prev, curr *Snapshot) []DiffEntry {
	prevMap := indexResults(prev.Results)
	var entries []DiffEntry

	for _, cur := range curr.Results {
		p, ok := prevMap[cur.Name]
		if !ok {
			// new container not seen before
			if cur.Drifted {
				entries = append(entries, DiffEntry{
					Name:     cur.Name,
					Previous: nil,
					Current:  cur.Differences,
				})
			}
			continue
		}

		if driftChanged(p, cur) {
			entries = append(entries, DiffEntry{
				Name:     cur.Name,
				Previous: p.Differences,
				Current:  cur.Differences,
			})
		}
	}

	return entries
}

// Summary returns a human-readable summary of a DiffEntry slice.
func Summary(entries []DiffEntry) string {
	if len(entries) == 0 {
		return "no drift changes between snapshots"
	}
	s := fmt.Sprintf("%d container(s) changed drift state:\n", len(entries))
	for _, e := range entries {
		s += fmt.Sprintf("  %s: %d -> %d differences\n", e.Name, len(e.Previous), len(e.Current))
	}
	return s
}

func indexResults(results []drift.Result) map[string]drift.Result {
	m := make(map[string]drift.Result, len(results))
	for _, r := range results {
		m[r.Name] = r
	}
	return m
}

func driftChanged(prev, curr drift.Result) bool {
	return prev.Drifted != curr.Drifted || len(prev.Differences) != len(curr.Differences)
}
