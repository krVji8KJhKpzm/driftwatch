package ranking

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeResult(name string, drifted bool, diffs []drift.Diff) drift.Result {
	return drift.Result{Name: name, Drifted: drifted, Diffs: diffs}
}

func TestRank_NoDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("app", false, nil),
	}
	entries := Rank(results)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestRank_SortedByScore(t *testing.T) {
	results := []drift.Result{
		makeResult("low", true, []drift.Diff{{Field: "env", Expected: "a", Actual: "b"}}),
		makeResult("high", true, []drift.Diff{
			{Field: "image", Expected: "x", Actual: "y"},
			{Field: "env", Expected: "a", Actual: "b"},
		}),
	}

	entries := Rank(results)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "high" {
		t.Errorf("expected 'high' first, got %q", entries[0].Name)
	}
	if entries[0].DriftScore != 4 {
		t.Errorf("expected score 4 (image=3 + env=1), got %d", entries[0].DriftScore)
	}
	if entries[1].DriftScore != 1 {
		t.Errorf("expected score 1, got %d", entries[1].DriftScore)
	}
}

func TestRank_TieBreakByName(t *testing.T) {
	results := []drift.Result{
		makeResult("zebra", true, []drift.Diff{{Field: "env", Expected: "a", Actual: "b"}}),
		makeResult("alpha", true, []drift.Diff{{Field: "env", Expected: "a", Actual: "b"}}),
	}
	entries := Rank(results)
	if entries[0].Name != "alpha" {
		t.Errorf("expected 'alpha' first on tie, got %q", entries[0].Name)
	}
}

func TestWriteText_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, []Entry{}, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "no drifted containers") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestWriteText_WithEntries(t *testing.T) {
	entries := []Entry{
		{Name: "svc", DriftScore: 4, Fields: []string{"image", "env"}},
	}
	var buf bytes.Buffer
	if err := Write(&buf, entries, "text"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "svc") {
		t.Errorf("expected container name in output: %q", out)
	}
	if !strings.Contains(out, "4") {
		t.Errorf("expected score in output: %q", out)
	}
}
