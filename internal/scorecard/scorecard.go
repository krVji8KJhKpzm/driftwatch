package scorecard

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Grade represents a letter grade assigned to a container's drift health.
type Grade string

const (
	GradeA Grade = "A"
	GradeB Grade = "B"
	GradeC Grade = "C"
	GradeD Grade = "D"
	GradeF Grade = "F"
)

// Entry holds the scorecard result for a single container.
type Entry struct {
	Name       string
	DriftCount int
	Grade      Grade
	Summary    string
}

// Build computes a scorecard from a slice of drift results.
func Build(results []drift.Result) []Entry {
	entries := make([]Entry, 0, len(results))
	for _, r := range results {
		count := len(r.Diffs)
		g := grade(count)
		entries = append(entries, Entry{
			Name:       r.Name,
			DriftCount: count,
			Grade:      g,
			Summary:    summary(r.Name, count, g),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].DriftCount != entries[j].DriftCount {
			return entries[i].DriftCount > entries[j].DriftCount
		}
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// Write renders the scorecard to w in a human-readable table.
func Write(w io.Writer, entries []Entry) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CONTAINER\tDRIFTS\tGRADE\tSUMMARY")
	fmt.Fprintln(tw, "---------\t------\t-----\t-------")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", e.Name, e.DriftCount, e.Grade, e.Summary)
	}
	tw.Flush()
}

func grade(driftCount int) Grade {
	switch {
	case driftCount == 0:
		return GradeA
	case driftCount == 1:
		return GradeB
	case driftCount == 2:
		return GradeC
	case driftCount <= 4:
		return GradeD
	default:
		return GradeF
	}
}

func summary(name string, count int, g Grade) string {
	if count == 0 {
		return fmt.Sprintf("%s is fully in sync", name)
	}
	return fmt.Sprintf("%s has %d drift(s), grade %s", name, count, g)
}
