package metrics_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/metrics"
)

func makeResult(name string, drifted bool, fields ...string) drift.Result {
	r := drift.Result{Name: name, Drifted: drifted}
	for _, f := range fields {
		r.Diffs = append(r.Diffs, drift.Diff{Field: f, Expected: "a", Actual: "b"})
	}
	return r
}

func TestCompute_NoDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("app", false),
		makeResult("db", false),
	}
	s := metrics.Compute(results)
	if s.Total != 2 || s.Drifted != 0 || s.Clean != 2 {
		t.Fatalf("unexpected counts: %+v", s)
	}
	if s.DriftRate != 0 {
		t.Errorf("expected 0%% drift rate, got %.1f", s.DriftRate)
	}
}

func TestCompute_WithDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("app", true, "image", "env.PORT"),
		makeResult("db", false),
		makeResult("cache", true, "image"),
	}
	s := metrics.Compute(results)
	if s.Total != 3 || s.Drifted != 2 || s.Clean != 1 {
		t.Fatalf("unexpected counts: %+v", s)
	}
	if s.ByField["image"] != 2 {
		t.Errorf("expected image count 2, got %d", s.ByField["image"])
	}
	if s.ByField["env.PORT"] != 1 {
		t.Errorf("expected env.PORT count 1, got %d", s.ByField["env.PORT"])
	}
}

func TestCompute_DriftRate(t *testing.T) {
	results := []drift.Result{
		makeResult("a", true, "image"),
		makeResult("b", false),
		makeResult("c", false),
		makeResult("d", false),
	}
	s := metrics.Compute(results)
	if s.DriftRate != 25.0 {
		t.Errorf("expected 25.0%% drift rate, got %.1f", s.DriftRate)
	}
}

func TestCompute_TopDrifted(t *testing.T) {
	results := []drift.Result{
		makeResult("a", true, "image", "image", "env.X"),
		makeResult("b", true, "image"),
	}
	s := metrics.Compute(results)
	if len(s.TopDrifted) == 0 || s.TopDrifted[0] != "image" {
		t.Errorf("expected image as top drifted field, got %v", s.TopDrifted)
	}
}

func TestWrite_Text(t *testing.T) {
	s := metrics.Summary{
		Total: 4, Drifted: 1, Clean: 3, DriftRate: 25.0,
		ByField:    map[string]int{"image": 1},
		TopDrifted: []string{"image"},
	}
	var buf bytes.Buffer
	if err := metrics.Write(&buf, s, "text"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "25.0%") {
		t.Errorf("expected drift rate in output, got: %s", out)
	}
	if !strings.Contains(out, "image") {
		t.Errorf("expected field name in output, got: %s", out)
	}
}

func TestWrite_JSON(t *testing.T) {
	s := metrics.Summary{
		Total: 2, Drifted: 1, Clean: 1, DriftRate: 50.0,
		ByField:    map[string]int{"env.PORT": 1},
		TopDrifted: []string{"env.PORT"},
	}
	var buf bytes.Buffer
	if err := metrics.Write(&buf, s, "json"); err != nil {
		t.Fatal(err)
	}
	var got metrics.Summary
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.DriftRate != 50.0 {
		t.Errorf("expected 50.0, got %v", got.DriftRate)
	}
}
