package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/driftwatch/internal/drift"
)

func driftedResult(name string, diffs []drift.Difference) drift.Result {
	return drift.Result{Name: name, Drifted: true, Differences: diffs}
}

func TestWriteText_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, []drift.Result{}, FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No drift detected") {
		t.Errorf("expected no-drift message, got: %s", buf.String())
	}
}

func TestWriteText_WithDrift(t *testing.T) {
	results := []drift.Result{
		driftedResult("web", []drift.Difference{
			{Field: "image", Expected: "nginx:1.25", Actual: "nginx:1.24"},
		}),
	}
	var buf bytes.Buffer
	err := Write(&buf, results, FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[DRIFT] web") {
		t.Errorf("expected drift header, got: %s", out)
	}
	if !strings.Contains(out, "nginx:1.25") {
		t.Errorf("expected expected value in output, got: %s", out)
	}
}

func TestWriteJSON_WithDrift(t *testing.T) {
	results := []drift.Result{
		driftedResult("api", []drift.Difference{
			{Field: "env.PORT", Expected: "8080", Actual: "9090"},
		}),
	}
	var buf bytes.Buffer
	err := Write(&buf, results, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"name"`) {
		t.Errorf("expected JSON name field, got: %s", out)
	}
	if !strings.Contains(out, "env.PORT") {
		t.Errorf("expected field name in JSON, got: %s", out)
	}
}

func TestWriteText_SkipsNonDrifted(t *testing.T) {
	results := []drift.Result{
		{Name: "clean", Drifted: false},
	}
	var buf bytes.Buffer
	_ = Write(&buf, results, FormatText)
	if strings.Contains(buf.String(), "clean") {
		t.Errorf("non-drifted container should not appear in output")
	}
}
