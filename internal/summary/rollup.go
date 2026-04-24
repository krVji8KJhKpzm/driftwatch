package summary

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/yourorg/driftwatch/internal/drift"
)

// ContainerSummary holds aggregated drift statistics for a single container.
type ContainerSummary struct {
	Name        string
	Drifted     bool
	ImageDrift  bool
	EnvDrifts   int
	LabelDrifts int
	Total       int
}

// Rollup aggregates a slice of DetectResult into per-container summaries.
func Rollup(results []drift.DetectResult) []ContainerSummary {
	summaries := make([]ContainerSummary, 0, len(results))

	for _, r := range results {
		s := ContainerSummary{
			Name:    r.Name,
			Drifted: r.Drifted,
		}

		for _, d := range r.Diffs {
			switch d.Field {
			case "image":
				s.ImageDrift = true
				s.Total++
			case "env":
				s.EnvDrifts++
				s.Total++
			case "label":
				s.LabelDrifts++
				s.Total++
			default:
				s.Total++
			}
		}

		summaries = append(summaries, s)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return summaries
}

// WriteRollup writes a human-readable rollup table to w.
func WriteRollup(w io.Writer, summaries []ContainerSummary) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CONTAINER\tDRIFTED\tIMAGE\tENV\tLABEL\tTOTAL")

	for _, s := range summaries {
		drifted := "no"
		if s.Drifted {
			drifted = "yes"
		}
		imageDrift := "no"
		if s.ImageDrift {
			imageDrift = "yes"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%d\n",
			s.Name, drifted, imageDrift, s.EnvDrifts, s.LabelDrifts, s.Total)
	}

	tw.Flush()
}
