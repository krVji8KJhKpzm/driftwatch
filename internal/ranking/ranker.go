package ranking

import (
	"io"
	"sort"
	"text/tabwriter"
	"fmt"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Entry holds a container name and its drift score.
type Entry struct {
	Name       string
	DriftScore int
	Fields     []string
}

// Rank builds a sorted list of containers ranked by drift severity.
// Each drifted field contributes 1 point; image drift contributes an extra 2.
func Rank(results []drift.Result) []Entry {
	entries := make([]Entry, 0, len(results))

	for _, r := range results {
		if !r.Drifted {
			continue
		}

		score := 0
		fields := make([]string, 0)

		for _, d := range r.Diffs {
			fields = append(fields, d.Field)
			if d.Field == "image" {
				score += 3
			} else {
				score++
			}
		}

		entries = append(entries, Entry{
			Name:       r.Name,
			DriftScore: score,
			Fields:     fields,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].DriftScore != entries[j].DriftScore {
			return entries[i].DriftScore > entries[j].DriftScore
		}
		return entries[i].Name < entries[j].Name
	})

	return entries
}

// Write outputs the ranked entries to w in the given format ("text" or "json").
func Write(w io.Writer, entries []Entry, format string) error {
	if format == "json" {
		return writeJSON(w, entries)
	}
	return writeText(w, entries)
}

func writeText(w io.Writer, entries []Entry) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, "no drifted containers")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RANK\tCONTAINER\tSCORE\tFIELDS")
	for i, e := range entries {
		fmt.Fprintf(tw, "%d\t%s\t%d\t%v\n", i+1, e.Name, e.DriftScore, e.Fields)
	}
	return tw.Flush()
}
