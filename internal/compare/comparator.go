// Package compare provides field-level comparison utilities for drift results,
// enabling structured diffing between manifest expectations and live container state.
package compare

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourorg/driftwatch/internal/drift"
)

// FieldDiff represents a single field-level difference between expected and actual values.
type FieldDiff struct {
	Field    string
	Expected string
	Actual   string
}

// ContainerComparison holds the full comparison result for a single container.
type ContainerComparison struct {
	Name   string
	Drifted bool
	Diffs  []FieldDiff
}

// Report is the output of comparing a full set of drift results.
type Report struct {
	Containers []ContainerComparison
	TotalDiffs int
}

// Build constructs a Report from a slice of drift results, extracting
// structured field-level diffs for each drifted container.
func Build(results []drift.Result) Report {
	var report Report

	for _, r := range results {
		comp := ContainerComparison{
			Name:    r.Name,
			Drifted: r.Drifted,
		}

		if r.Drifted {
			for _, d := range r.Diffs {
				fd := parseFieldDiff(d)
				comp.Diffs = append(comp.Diffs, fd)
			}
			sort.Slice(comp.Diffs, func(i, j int) bool {
				return comp.Diffs[i].Field < comp.Diffs[j].Field
			})
			report.TotalDiffs += len(comp.Diffs)
		}

		report.Containers = append(report.Containers, comp)
	}

	// Sort containers: drifted first, then alphabetically by name.
	sort.Slice(report.Containers, func(i, j int) bool {
		if report.Containers[i].Drifted != report.Containers[j].Drifted {
			return report.Containers[i].Drifted
		}
		return report.Containers[i].Name < report.Containers[j].Name
	})

	return report
}

// parseFieldDiff extracts a FieldDiff from a raw diff string.
// Expected format: "fieldname: expected=<val> actual=<val>"
// Falls back to a raw representation if the format is unrecognised.
func parseFieldDiff(raw string) FieldDiff {
	// Attempt structured parse.
	colonIdx := strings.Index(raw, ":")
	if colonIdx == -1 {
		return FieldDiff{Field: raw, Expected: "", Actual: raw}
	}

	field := strings.TrimSpace(raw[:colonIdx])
	rest := strings.TrimSpace(raw[colonIdx+1:])

	var expected, actual string
	for _, part := range strings.Fields(rest) {
		if strings.HasPrefix(part, "expected=") {
			expected = strings.TrimPrefix(part, "expected=")
		} else if strings.HasPrefix(part, "actual=") {
			actual = strings.TrimPrefix(part, "actual=")
		}
	}

	if expected == "" && actual == "" {
		return FieldDiff{Field: field, Expected: "", Actual: rest}
	}

	return FieldDiff{Field: field, Expected: expected, Actual: actual}
}

// Summary returns a human-readable one-line summary of the report.
func (r Report) Summary() string {
	drifted := 0
	for _, c := range r.Containers {
		if c.Drifted {
			drifted++
		}
	}
	return fmt.Sprintf("%d container(s) drifted, %d field diff(s) total", drifted, r.TotalDiffs)
}
