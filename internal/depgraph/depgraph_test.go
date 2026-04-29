package depgraph

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResult(name, image string, drifted bool, fields ...string) drift.Result {
	r := drift.Result{Name: name, ActualImage: image, Drifted: drifted}
	for _, f := range fields {
		r.Diffs = append(r.Diffs, drift.Diff{Field: f, Expected: "a", Actual: "b"})
	}
	return r
}

func TestBuild_Empty(t *testing.T) {
	g := Build(nil)
	if len(g.Nodes) != 0 || len(g.Edges) != 0 {
		t.Fatalf("expected empty graph, got %+v", g)
	}
}

func TestBuild_NoDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", "nginx:1", false),
		makeResult("beta", "redis:7", false),
	}
	g := Build(results)
	if len(g.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 0 {
		t.Fatalf("expected no edges, got %d", len(g.Edges))
	}
}

func TestBuild_SharedEnvCreatesEdge(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", "nginx:1", true, "DATABASE_URL", "image"),
		makeResult("beta", "redis:7", true, "DATABASE_URL"),
	}
	g := Build(results)
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	e := g.Edges[0]
	if e.Shared != "DATABASE_URL" {
		t.Errorf("expected shared key DATABASE_URL, got %s", e.Shared)
	}
	if e.From != "alpha" || e.To != "beta" {
		t.Errorf("unexpected edge endpoints: %s <-> %s", e.From, e.To)
	}
}

func TestBuild_ImageFieldNotShared(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", "nginx:1", true, "image"),
		makeResult("beta", "redis:7", true, "image"),
	}
	g := Build(results)
	if len(g.Edges) != 0 {
		t.Fatalf("image field should not create edges, got %d edges", len(g.Edges))
	}
}

func TestBuild_DriftCountAccurate(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", "nginx:1", true, "A", "B", "image"),
	}
	g := Build(results)
	if g.Nodes[0].DriftCount != 3 {
		t.Errorf("expected drift count 3, got %d", g.Nodes[0].DriftCount)
	}
}

func TestWrite_TextFormat(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", "nginx:1", true, "PORT"),
		makeResult("beta", "redis:7", true, "PORT"),
	}
	g := Build(results)
	var buf bytes.Buffer
	if err := Write(g, "text", &buf); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("output missing container names: %s", out)
	}
	if !strings.Contains(out, "PORT") {
		t.Errorf("output missing shared key PORT: %s", out)
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", "nginx:1", true, "PORT"),
	}
	g := Build(results)
	var buf bytes.Buffer
	if err := Write(g, "json", &buf); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	var out Graph
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out.Nodes) != 1 || out.Nodes[0].Name != "alpha" {
		t.Errorf("unexpected JSON output: %+v", out)
	}
}
