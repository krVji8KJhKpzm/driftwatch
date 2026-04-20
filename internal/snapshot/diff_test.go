package snapshot_test

import (
	"strings"
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/snapshot"
)

func makeSnap(results []drift.Result) *snapshot.Snapshot {
	return &snapshot.Snapshot{
		Timestamp: time.Now().UTC(),
		Results:   results,
	}
}

func TestDiff_NoDrift(t *testing.T) {
	results := []drift.Result{{Name: "web", Drifted: false}}
	prev := makeSnap(results)
	curr := makeSnap(results)

	entries := snapshot.Diff(prev, curr)
	if len(entries) != 0 {
		t.Errorf("expected 0 diff entries, got %d", len(entries))
	}
}

func TestDiff_NewDrift(t *testing.T) {
	prev := makeSnap([]drift.Result{{Name: "web", Drifted: false}})
	curr := makeSnap([]drift.Result{
		{Name: "web", Drifted: true, Differences: []drift.Difference{{Field: "image"}}},
	})

	entries := snapshot.Diff(prev, curr)
	if len(entries) != 1 {
		t.Fatalf("expected 1 diff entry, got %d", len(entries))
	}
	if entries[0].Name != "web" {
		t.Errorf("expected entry name 'web', got %q", entries[0].Name)
	}
}

func TestDiff_DriftResolved(t *testing.T) {
	prev := makeSnap([]drift.Result{
		{Name: "api", Drifted: true, Differences: []drift.Difference{{Field: "env"}}},
	})
	curr := makeSnap([]drift.Result{{Name: "api", Drifted: false}})

	entries := snapshot.Diff(prev, curr)
	if len(entries) != 1 {
		t.Fatalf("expected 1 diff entry, got %d", len(entries))
	}
	if len(entries[0].Current) != 0 {
		t.Errorf("expected no current differences, got %d", len(entries[0].Current))
	}
}

func TestDiff_NewContainer(t *testing.T) {
	prev := makeSnap([]drift.Result{})
	curr := makeSnap([]drift.Result{
		{Name: "cache", Drifted: true, Differences: []drift.Difference{{Field: "image"}}},
	})

	entries := snapshot.Diff(prev, curr)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry for new drifted container, got %d", len(entries))
	}
}

func TestSummary_Empty(t *testing.T) {
	out := snapshot.Summary(nil)
	if !strings.Contains(out, "no drift changes") {
		t.Errorf("unexpected summary: %q", out)
	}
}

func TestSummary_WithEntries(t *testing.T) {
	entries := []snapshot.DiffEntry{
		{Name: "web", Previous: nil, Current: []drift.Difference{{Field: "image"}}},
	}
	out := snapshot.Summary(entries)
	if !strings.Contains(out, "web") {
		t.Errorf("expected 'web' in summary, got: %q", out)
	}
}
