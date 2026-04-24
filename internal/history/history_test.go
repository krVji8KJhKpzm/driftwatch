package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/history"
)

func sampleResults() []drift.Result {
	return []drift.Result{
		{Name: "web", Drifted: true, Diffs: []drift.Diff{{Field: "image", Expected: "nginx:1.24", Actual: "nginx:1.25"}}},
		{Name: "db", Drifted: false},
	}
}

func TestRecord_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	if err := history.Record(path, sampleResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestRecord_AppendsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	if err := history.Record(path, sampleResults()); err != nil {
		t.Fatalf("first record: %v", err)
	}
	if err := history.Record(path, sampleResults()); err != nil {
		t.Fatalf("second record: %v", err)
	}

	entries, err := history.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := history.Load("/nonexistent/path/history.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_SortedByTimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	for i := 0; i < 3; i++ {
		if err := history.Record(path, sampleResults()); err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	entries, err := history.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for i := 1; i < len(entries); i++ {
		if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
			t.Errorf("entries not sorted: index %d before %d", i, i-1)
		}
	}
}

func TestLatest_ReturnsLastEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	if err := history.Record(path, sampleResults()); err != nil {
		t.Fatalf("record: %v", err)
	}

	entry, err := history.Latest(path)
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if len(entry.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(entry.Results))
	}
}

func TestLatest_EmptyHistory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	// Write empty array
	if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := history.Latest(path)
	if err == nil {
		t.Fatal("expected error for empty history")
	}
}
