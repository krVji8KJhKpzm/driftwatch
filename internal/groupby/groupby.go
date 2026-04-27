package groupby

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Field represents a grouping dimension.
type Field string

const (
	FieldImage  Field = "image"
	FieldStatus Field = "status"
	FieldEnvKey Field = "env_key"
)

// Group holds a set of container names under a shared key.
type Group struct {
	Key   string   `json:"key"`
	Count int      `json:"count"`
	Names []string `json:"names"`
}

// Result is the output of a grouping operation.
type Result struct {
	Field  Field   `json:"field"`
	Groups []Group `json:"groups"`
}

// By groups drift results by the given field.
func By(results []drift.Result, field Field) (Result, error) {
	index := map[string][]string{}

	for _, r := range results {
		keys, err := extractKeys(r, field)
		if err != nil {
			return Result{}, fmt.Errorf("groupby: %w", err)
		}
		for _, k := range keys {
			index[k] = append(index[k], r.Name)
		}
	}

	groups := make([]Group, 0, len(index))
	for k, names := range index {
		sort.Strings(names)
		groups = append(groups, Group{Key: k, Count: len(names), Names: names})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Count != groups[j].Count {
			return groups[i].Count > groups[j].Count
		}
		return groups[i].Key < groups[j].Key
	})

	return Result{Field: field, Groups: groups}, nil
}

func extractKeys(r drift.Result, field Field) ([]string, error) {
	switch field {
	case FieldImage:
		for _, d := range r.Diffs {
			if d.Field == "image" {
				return []string{d.Got}, nil
			}
		}
		return []string{"<no-image-drift>"}, nil
	case FieldStatus:
		if r.Drifted {
			return []string{"drifted"}, nil
		}
		return []string{"clean"}, nil
	case FieldEnvKey:
		var keys []string
		for _, d := range r.Diffs {
			if strings.HasPrefix(d.Field, "env.") {
				keys = append(keys, strings.TrimPrefix(d.Field, "env."))
			}
		}
		if len(keys) == 0 {
			return []string{"<no-env-drift>"}, nil
		}
		return keys, nil
	default:
		return nil, fmt.Errorf("unknown field %q", field)
	}
}

// Write renders the grouping result to w in the given format ("text" or "json").
func Write(w io.Writer, res Result, format string) error {
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}
	fmt.Fprintf(w, "Grouped by: %s\n", res.Field)
	if len(res.Groups) == 0 {
		fmt.Fprintln(w, "  (no results)")
		return nil
	}
	for _, g := range res.Groups {
		fmt.Fprintf(w, "  [%s] (%d)\n", g.Key, g.Count)
		for _, n := range g.Names {
			fmt.Fprintf(w, "    - %s\n", n)
		}
	}
	return nil
}
