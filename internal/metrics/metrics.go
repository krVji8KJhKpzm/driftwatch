package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Summary holds aggregated drift metrics across all inspected containers.
type Summary struct {
	Total        int            `json:"total"`
	Drifted      int            `json:"drifted"`
	Clean        int            `json:"clean"`
	DriftRate    float64        `json:"drift_rate_pct"`
	ByField      map[string]int `json:"by_field"`
	TopDrifted   []string       `json:"top_drifted"`
}

// Compute builds a Summary from a slice of drift results.
func Compute(results []drift.Result) Summary {
	s := Summary{
		ByField: make(map[string]int),
	}

	for _, r := range results {
		s.Total++
		if r.Drifted {
			s.Drifted++
			for _, d := range r.Diffs {
				s.ByField[d.Field]++
			}
		}
	}

	s.Clean = s.Total - s.Drifted
	if s.Total > 0 {
		s.DriftRate = float64(s.Drifted) / float64(s.Total) * 100
	}

	s.TopDrifted = topFields(s.ByField, 5)
	return s
}

// Write serialises the Summary to w in the requested format ("text" or "json").
func Write(w io.Writer, s Summary, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(s)
	default:
		return writeText(w, s)
	}
}

func writeText(w io.Writer, s Summary) error {
	_, err := fmt.Fprintf(w,
		"Drift Metrics\n"+
			"  Total containers : %d\n"+
			"  Drifted          : %d\n"+
			"  Clean            : %d\n"+
			"  Drift rate       : %.1f%%\n",
		s.Total, s.Drifted, s.Clean, s.DriftRate,
	)
	if err != nil {
		return err
	}
	if len(s.TopDrifted) > 0 {
		fmt.Fprintln(w, "  Top drifted fields:")
		for _, f := range s.TopDrifted {
			fmt.Fprintf(w, "    - %s (%d)\n", f, s.ByField[f])
		}
	}
	return nil
}

func topFields(byField map[string]int, n int) []string {
	type kv struct {
		key   string
		count int
	}
	var pairs []kv
	for k, v := range byField {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].key < pairs[j].key
	})
	var out []string
	for i, p := range pairs {
		if i >= n {
			break
		}
		out = append(out, p.key)
	}
	return out
}
