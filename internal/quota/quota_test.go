package quota

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResult(name string, drifted bool, fieldCount int) drift.Result {
	var diffs []drift.Diff
	for i := 0; i < fieldCount; i++ {
		diffs = append(diffs, drift.Diff{Field: fmt.Sprintf("field%d", i)})
	}
	return drift.Result{Name: name, Drifted: drifted, Diffs: diffs}
}

func writeTempConfig(t *testing.T, cfg Config) string {
	t.Helper()
	data, _ := json.Marshal(cfg)
	p := filepath.Join(t.TempDir(), "quota.json")
	_ = os.WriteFile(p, data, 0644)
	return p
}

func TestLoadConfig_Valid(t *testing.T) {
	cfg := Config{
		GlobalMaxFields: 3,
		ContainerRules:  []ContainerRule{{Name: "api", MaxFields: 1}},
	}
	p := writeTempConfig(t, cfg)
	got, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GlobalMaxFields != 3 {
		t.Errorf("expected GlobalMaxFields=3, got %d", got.GlobalMaxFields)
	}
	if len(got.ContainerRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(got.ContainerRules))
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/quota.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestEvaluate_NilConfig(t *testing.T) {
	results := []drift.Result{makeResult("api", true, 5)}
	if v := Evaluate(results, nil); v != nil {
		t.Errorf("expected nil violations for nil config")
	}
}

func TestEvaluate_NoViolations(t *testing.T) {
	cfg := &Config{GlobalMaxFields: 5}
	results := []drift.Result{makeResult("api", true, 3)}
	if v := Evaluate(results, cfg); len(v) != 0 {
		t.Errorf("expected no violations, got %d", len(v))
	}
}

func TestEvaluate_GlobalViolation(t *testing.T) {
	cfg := &Config{GlobalMaxFields: 2}
	results := []drift.Result{makeResult("api", true, 4)}
	v := Evaluate(results, cfg)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Container != "api" || v[0].Limit != 2 || v[0].DriftCount != 4 {
		t.Errorf("unexpected violation: %+v", v[0])
	}
}

func TestEvaluate_PerContainerOverridesGlobal(t *testing.T) {
	cfg := &Config{
		GlobalMaxFields: 10,
		ContainerRules:  []ContainerRule{{Name: "api", MaxFields: 1}},
	}
	results := []drift.Result{makeResult("api", true, 3)}
	v := Evaluate(results, cfg)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Limit != 1 {
		t.Errorf("expected limit=1, got %d", v[0].Limit)
	}
}

func TestEvaluate_SkipsNonDrifted(t *testing.T) {
	cfg := &Config{GlobalMaxFields: 1}
	results := []drift.Result{makeResult("api", false, 5)}
	if v := Evaluate(results, cfg); len(v) != 0 {
		t.Errorf("expected no violations for non-drifted container")
	}
}
