package threshold_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/threshold"
)

func makeResults(drifted, clean int, fields ...string) []drift.Result {
	var results []drift.Result
	for i := 0; i < drifted; i++ {
		diffs := make([]drift.Diff, len(fields))
		for j, f := range fields {
			diffs[j] = drift.Diff{Field: f, Expected: "a", Actual: "b"}
		}
		results = append(results, drift.Result{Name: "c", Drifted: true, Diffs: diffs})
	}
	for i := 0; i < clean; i++ {
		results = append(results, drift.Result{Name: "ok", Drifted: false})
	}
	return results
}

func TestEvaluate_NilConfig(t *testing.T) {
	v := threshold.Evaluate(nil, makeResults(5, 0))
	if len(v) != 0 {
		t.Fatalf("expected no violations, got %d", len(v))
	}
}

func TestEvaluate_NoDrift(t *testing.T) {
	cfg := &threshold.Config{MaxDriftCount: 1, MaxDriftRate: 0.5}
	v := threshold.Evaluate(cfg, makeResults(0, 4))
	if len(v) != 0 {
		t.Fatalf("expected no violations, got %d", len(v))
	}
}

func TestEvaluate_ExceedsCount(t *testing.T) {
	cfg := &threshold.Config{MaxDriftCount: 2}
	v := threshold.Evaluate(cfg, makeResults(3, 0))
	if len(v) != 1 || v[0].Rule != "max_drift_count" {
		t.Fatalf("expected max_drift_count violation, got %+v", v)
	}
}

func TestEvaluate_ExceedsRate(t *testing.T) {
	cfg := &threshold.Config{MaxDriftRate: 0.4}
	v := threshold.Evaluate(cfg, makeResults(3, 2))
	if len(v) != 1 || v[0].Rule != "max_drift_rate" {
		t.Fatalf("expected max_drift_rate violation, got %+v", v)
	}
}

func TestEvaluate_BlockedField(t *testing.T) {
	cfg := &threshold.Config{BlockedFields: []string{"image"}}
	v := threshold.Evaluate(cfg, makeResults(1, 0, "image"))
	if len(v) != 1 || v[0].Rule != "blocked_field" {
		t.Fatalf("expected blocked_field violation, got %+v", v)
	}
}

func TestEvaluate_MultipleViolations(t *testing.T) {
	cfg := &threshold.Config{MaxDriftCount: 1, MaxDriftRate: 0.1, BlockedFields: []string{"env"}}
	v := threshold.Evaluate(cfg, makeResults(3, 1, "env"))
	if len(v) < 3 {
		t.Fatalf("expected at least 3 violations, got %d", len(v))
	}
}

func TestLoadConfig_Valid(t *testing.T) {
	cfg := threshold.Config{MaxDriftCount: 5, MaxDriftRate: 0.25, BlockedFields: []string{"image"}}
	data, _ := json.Marshal(cfg)
	p := filepath.Join(t.TempDir(), "thresh.json")
	_ = os.WriteFile(p, data, 0644)

	loaded, err := threshold.LoadConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.MaxDriftCount != 5 {
		t.Errorf("expected MaxDriftCount=5, got %d", loaded.MaxDriftCount)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := threshold.LoadConfig("/no/such/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
