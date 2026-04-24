package history

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// TrendEntry holds aggregated drift stats for a single history entry.
type TrendEntry struct {
	Timestamp   time.Time
	Total       int
	Drifted     int
	Resolved    int
	NewDrifts   int
}

// Trend computes a slice of TrendEntry values from a list of history entries,
// showing how drift counts changed over time.
func Trend(entries []Entry) []TrendEntry {
	trend := make([]TrendEntry, 0, len(entries))

	for i, e := range entries {
		te := TrendEntry{
			Timestamp: e.Timestamp,
			Total:     len(e.Results),
		}
		for _, r := range e.Results {
			if r.Drifted {
				te.Drifted++
			}
		}

		if i > 0 {
			prev := trend[i-1]
			if te.Drifted < prev.Drifted {
				te.Resolved = prev.Drifted - te.Drifted
			} else if te.Drifted > prev.Drifted {
				te.NewDrifts = te.Drifted - prev.Drifted
			}
		}

		trend = append(trend, te)
	}

	return trend
}

// WriteTrend writes a human-readable trend table to w.
func WriteTrend(w io.Writer, entries []Entry) error {
	trend := Trend(entries)
	if len(trend) == 0 {
		fmt.Fprintln(w, "no history entries found")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TIMESTAMP\tTOTAL\tDRIFTED\tNEW\tRESOLVED")
	for _, t := range trend {
		fmt.Fprintf(tw, "%s\t%d\t%d\t+%d\t-%d\n",
			t.Timestamp.Format(time.RFC3339),
			t.Total,
			t.Drifted,
			t.NewDrifts,
			t.Resolved,
		)
	}
	return tw.Flush()
}
