package summary

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeSnap(names []string, drifted []bool) []drift.Result {
	var out []drift.Result
	for i, name := range names {
		out = append(out, drift.Result{Name: name, Drifted: drifted[i]})
	}
	return out
}

func TestBuildTrend_Empty(t *testing.T) {
	report := BuildTrend(nil)
	if len(report.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(report.Entries))
	}
}

func TestBuildTrend_NoDrift(t *testing.T) {
	snaps := [][]drift.Result{
		makeSnap([]string{"web", "db"}, []bool{false, false}),
		makeSnap([]string{"web", "db"}, []bool{false, false}),
	}
	report := BuildTrend(snaps)
	for _, e := range report.Entries {
		if e.DriftCount != 0 || e.Rate != 0 {
			t.Errorf("expected no drift for %s, got count=%d rate=%.1f", e.Name, e.DriftCount, e.Rate)
		}
	}
}

func TestBuildTrend_PartialDrift(t *testing.T) {
	snaps := [][]drift.Result{
		makeSnap([]string{"web"}, []bool{true}),
		makeSnap([]string{"web"}, []bool{false}),
		makeSnap([]string{"web"}, []bool{true}),
		makeSnap([]string{"web"}, []bool{true}),
	}
	report := BuildTrend(snaps)
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Entries))
	}
	e := report.Entries[0]
	if e.DriftCount != 3 {
		t.Errorf("expected DriftCount=3, got %d", e.DriftCount)
	}
	if e.Total != 4 {
		t.Errorf("expected Total=4, got %d", e.Total)
	}
	if e.Rate != 75.0 {
		t.Errorf("expected Rate=75.0, got %.1f", e.Rate)
	}
}

func TestBuildTrend_SortedByRateDesc(t *testing.T) {
	snaps := [][]drift.Result{
		makeSnap([]string{"a", "b"}, []bool{false, true}),
		makeSnap([]string{"a", "b"}, []bool{true, true}),
	}
	report := BuildTrend(snaps)
	if report.Entries[0].Name != "b" {
		t.Errorf("expected 'b' first (100%% rate), got %s", report.Entries[0].Name)
	}
}

func TestWriteTrendReport_Output(t *testing.T) {
	snaps := [][]drift.Result{
		makeSnap([]string{"svc"}, []bool{true}),
	}
	report := BuildTrend(snaps)
	var buf bytes.Buffer
	WriteTrendReport(&buf, report)
	out := buf.String()
	if !strings.Contains(out, "svc") {
		t.Error("expected container name in output")
	}
	if !strings.Contains(out, "100.0%") {
		t.Error("expected 100.0% rate in output")
	}
}

func TestWriteTrendReport_Empty(t *testing.T) {
	var buf bytes.Buffer
	WriteTrendReport(&buf, TrendReport{})
	if !strings.Contains(buf.String(), "No trend data") {
		t.Error("expected empty message")
	}
}
