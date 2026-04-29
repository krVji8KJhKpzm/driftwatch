package depgraph

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Node represents a container and its drift relationships.
type Node struct {
	Name       string   `json:"name"`
	Image      string   `json:"image"`
	Drifted    bool     `json:"drifted"`
	DriftCount int      `json:"drift_count"`
	SharedEnvs []string `json:"shared_envs,omitempty"`
}

// Edge represents a shared-environment relationship between two containers.
type Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Shared string `json:"shared_key"`
}

// Graph holds nodes and edges derived from drift results.
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Build constructs a dependency graph from drift results, linking containers
// that share drifted environment variable keys.
func Build(results []drift.Result) Graph {
	nodeMap := make(map[string]*Node)
	envIndex := make(map[string][]string) // envKey -> container names

	for _, r := range results {
		n := &Node{
			Name:    r.Name,
			Image:   r.ActualImage,
			Drifted: r.Drifted,
		}
		for _, d := range r.Diffs {
			n.DriftCount++
			if d.Field != "image" {
				n.SharedEnvs = append(n.SharedEnvs, d.Field)
				envIndex[d.Field] = append(envIndex[d.Field], r.Name)
			}
		}
		nodeMap[r.Name] = n
	}

	seenEdges := make(map[string]bool)
	var edges []Edge

	for key, names := range envIndex {
		if len(names) < 2 {
			continue
		}
		sort.Strings(names)
		for i := 0; i < len(names); i++ {
			for j := i + 1; j < len(names); j++ {
				ek := fmt.Sprintf("%s|%s|%s", names[i], names[j], key)
				if !seenEdges[ek] {
					seenEdges[ek] = true
					edges = append(edges, Edge{From: names[i], To: names[j], Shared: key})
				}
			}
		}
	}

	nodes := make([]Node, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, *n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Name < nodes[j].Name })

	return Graph{Nodes: nodes, Edges: edges}
}

// Write serialises the graph to w in the requested format ("text" or "json").
func Write(g Graph, format string, w io.Writer) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(g)
	default:
		fmt.Fprintf(w, "Dependency Graph\n")
		fmt.Fprintf(w, "Nodes (%d):\n", len(g.Nodes))
		for _, n := range g.Nodes {
			driftMark := " "
			if n.Drifted {
				driftMark = "!"
			}
			fmt.Fprintf(w, "  [%s] %s (%s) drifts=%d\n", driftMark, n.Name, n.Image, n.DriftCount)
		}
		fmt.Fprintf(w, "Edges (%d):\n", len(g.Edges))
		for _, e := range g.Edges {
			fmt.Fprintf(w, "  %s <-> %s  [shared: %s]\n", e.From, e.To, e.Shared)
		}
		return nil
	}
}
