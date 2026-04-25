package diff

import (
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResult(name string, drifted bool, differences []string) drift.Result {
	return drift.Result{
		Name:        name,
		Drifted:     drifted,
		Differences: differences,
	}
}

func TestHighlight_NoDrift(t *testing.T) {
	r := makeResult("web", false, nil)
	diffs := Highlight(r)
	if len(diffs) != 0 {
		t.Fatalf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestHighlight_ParsesFields(t *testing.T) {
	r := makeResult("web", true, []string{
		"image: nginx:1.21: nginx:1.25",
		"env.PORT: 8080: 9090",
	})
	diffs := Highlight(r)
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(diffs))
	}
	if diffs[0].Field != "image" {
		t.Errorf("expected field 'image', got %q", diffs[0].Field)
	}
	if diffs[0].Expected != "nginx:1.21" {
		t.Errorf("unexpected expected value: %q", diffs[0].Expected)
	}
	if diffs[0].Actual != "nginx:1.25" {
		t.Errorf("unexpected actual value: %q", diffs[0].Actual)
	}
}

func TestHighlight_UnparsedField(t *testing.T) {
	r := makeResult("api", true, []string{"unknown change"})
	diffs := Highlight(r)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "unknown change" {
		t.Errorf("unexpected field: %q", diffs[0].Field)
	}
}

func TestFormatAll_NoDrift(t *testing.T) {
	r := makeResult("svc", false, nil)
	out := FormatAll(r)
	if !strings.Contains(out, "no drift detected") {
		t.Errorf("expected no-drift message, got: %s", out)
	}
}

func TestFormatAll_WithDrift(t *testing.T) {
	r := makeResult("svc", true, []string{
		"image: old:1: new:2",
	})
	out := FormatAll(r)
	if !strings.Contains(out, "drift detected") {
		t.Errorf("expected drift header, got: %s", out)
	}
	if !strings.Contains(out, "image") {
		t.Errorf("expected image field in output, got: %s", out)
	}
	if !strings.Contains(out, "expected:") {
		t.Errorf("expected 'expected:' label in output, got: %s", out)
	}
}
