package summary

import (
	"fmt"
	"io"
	"sort"

	"github.com/yourorg/driftwatch/internal/drift"
)

// ContainerTrend holds drift counts across multiple snapshots for a single container.
type ContainerTrend struct {
	Name       string
	DriftCount int
	Total      int
	Rate       float64 // percentage of checks that showed drift
}

// TrendReport summarises drift frequency per container across a slice of result sets.
type TrendReport struct {
	Entries []ContainerTrend
}

// BuildTrend aggregates multiple result sets (e.g. from history) into a TrendReport.
// Each inner slice represents one point-in-time scan.
func BuildTrend(snapshots [][]drift.Result) TrendReport {
	type acc struct {
		drifted int
		total   int
	}
	counts := make(map[string]*acc)

	for _, snap := range snapshots {
		for _, r := range snap {
			if _, ok := counts[r.Name]; !ok {
				counts[r.Name] = &acc{}
			}
			counts[r.Name].total++
			if r.Drifted {
				counts[r.Name].drifted++
			}
		}
	}

	var entries []ContainerTrend
	for name, a := range counts {
		rate := 0.0
		if a.total > 0 {
			rate = float64(a.drifted) / float64(a.total) * 100.0
		}
		entries = append(entries, ContainerTrend{
			Name:       name,
			DriftCount: a.drifted,
			Total:      a.total,
			Rate:       rate,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Rate != entries[j].Rate {
			return entries[i].Rate > entries[j].Rate
		}
		return entries[i].Name < entries[j].Name
	})

	return TrendReport{Entries: entries}
}

// WriteTrendReport writes a human-readable trend table to w.
func WriteTrendReport(w io.Writer, r TrendReport) {
	if len(r.Entries) == 0 {
		fmt.Fprintln(w, "No trend data available.")
		return
	}
	fmt.Fprintf(w, "%-30s %8s %8s %8s\n", "CONTAINER", "DRIFTED", "TOTAL", "RATE")
	fmt.Fprintf(w, "%-30s %8s %8s %8s\n", "----------", "-------", "-----", "----")
	for _, e := range r.Entries {
		fmt.Fprintf(w, "%-30s %8d %8d %7.1f%%\n", e.Name, e.DriftCount, e.Total, e.Rate)
	}
}
