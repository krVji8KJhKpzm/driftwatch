package summary

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeResult(name string, drifted bool, diffs []drift.Diff) drift.DetectResult {
	return drift.DetectResult{
		Name:    name,
		Drifted: drifted,
		Diffs:   diffs,
	}
}

func TestRollup_NoDrift(t *testing.T) {
	results := []drift.DetectResult{
		makeResult("web", false, nil),
	}

	summaries := Rollup(results)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Drifted {
		t.Error("expected Drifted=false")
	}
	if summaries[0].Total != 0 {
		t.Errorf("expected Total=0, got %d", summaries[0].Total)
	}
}

func TestRollup_WithDrift(t *testing.T) {
	results := []drift.DetectResult{
		makeResult("api", true, []drift.Diff{
			{Field: "image", Expected: "nginx:1.24", Actual: "nginx:1.25"},
			{Field: "env", Expected: "DEBUG=false", Actual: "DEBUG=true"},
			{Field: "env", Expected: "PORT=8080", Actual: "PORT=9090"},
		}),
	}

	summaries := Rollup(results)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if !s.ImageDrift {
		t.Error("expected ImageDrift=true")
	}
	if s.EnvDrifts != 2 {
		t.Errorf("expected EnvDrifts=2, got %d", s.EnvDrifts)
	}
	if s.Total != 3 {
		t.Errorf("expected Total=3, got %d", s.Total)
	}
}

func TestRollup_Sorted(t *testing.T) {
	results := []drift.DetectResult{
		makeResult("zebra", false, nil),
		makeResult("alpha", true, []drift.Diff{{Field: "image"}}),
	}

	summaries := Rollup(results)
	if summaries[0].Name != "alpha" || summaries[1].Name != "zebra" {
		t.Errorf("expected sorted order, got %s, %s", summaries[0].Name, summaries[1].Name)
	}
}

func TestWriteRollup_Output(t *testing.T) {
	summaries := []ContainerSummary{
		{Name: "web", Drifted: true, ImageDrift: true, EnvDrifts: 1, LabelDrifts: 0, Total: 2},
		{Name: "db", Drifted: false, Total: 0},
	}

	var buf bytes.Buffer
	WriteRollup(&buf, summaries)
	out := buf.String()

	if !strings.Contains(out, "CONTAINER") {
		t.Error("expected header row")
	}
	if !strings.Contains(out, "web") {
		t.Error("expected web row")
	}
	if !strings.Contains(out, "db") {
		t.Error("expected db row")
	}
	if !strings.Contains(out, "yes") {
		t.Error("expected 'yes' for drifted container")
	}
}
