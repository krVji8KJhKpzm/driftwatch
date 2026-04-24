package remediation

import (
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResult(name string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Name:    name,
		Drifted: len(diffs) > 0,
		Diffs:   diffs,
	}
}

func TestGenerate_NoDrift(t *testing.T) {
	results := []drift.Result{makeResult("web", nil)}
	got := Generate(results)
	if len(got) != 0 {
		t.Fatalf("expected 0 suggestions, got %d", len(got))
	}
}

func TestGenerate_ImageDrift(t *testing.T) {
	diffs := []drift.Diff{{Field: "image", Expected: "nginx:1.25", Actual: "nginx:1.24"}}
	results := []drift.Result{makeResult("web", diffs)}
	got := Generate(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(got))
	}
	if got[0].Field != "image" {
		t.Errorf("expected field 'image', got %q", got[0].Field)
	}
	if got[0].Container != "web" {
		t.Errorf("expected container 'web', got %q", got[0].Container)
	}
}

func TestGenerate_EnvDrift(t *testing.T) {
	diffs := []drift.Diff{{Field: "env:LOG_LEVEL", Expected: "info", Actual: "debug"}}
	results := []drift.Result{makeResult("api", diffs)}
	got := Generate(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(got))
	}
	if got[0].Hint == "" {
		t.Error("expected non-empty hint")
	}
}

func TestGenerate_SkipsNonDrifted(t *testing.T) {
	results := []drift.Result{
		makeResult("ok", nil),
		makeResult("bad", []drift.Diff{{Field: "image", Expected: "a", Actual: "b"}}),
	}
	got := Generate(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(got))
	}
	if got[0].Container != "bad" {
		t.Errorf("expected container 'bad', got %q", got[0].Container)
	}
}

func TestBuildHint_UnknownField(t *testing.T) {
	hint := buildHint("svc", "replicas", "3")
	if hint == "" {
		t.Error("expected non-empty hint for unknown field")
	}
}
