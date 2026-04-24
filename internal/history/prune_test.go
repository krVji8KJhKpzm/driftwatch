package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeEntries(t *testing.T, entries []Entry) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func makeEntries(n int, base time.Time) []Entry {
	out := make([]Entry, n)
	for i := range out {
		out[i] = Entry{Timestamp: base.Add(time.Duration(i) * time.Hour)}
	}
	return out
}

func TestPrune_NotFound(t *testing.T) {
	removed, err := Prune("/nonexistent/history.json", PruneOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestPrune_KeepLast(t *testing.T) {
	base := time.Now().Add(-5 * time.Hour)
	entries := makeEntries(5, base)
	path := writeEntries(t, entries)

	removed, err := Prune(path, PruneOptions{KeepLast: 3})
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 3 {
		t.Errorf("expected 3 entries, got %d", len(loaded))
	}
}

func TestPrune_OlderThan(t *testing.T) {
	base := time.Now().Add(-10 * time.Hour)
	entries := makeEntries(6, base) // hours -10,-9,-8,-7,-6,-5 relative to now
	path := writeEntries(t, entries)

	cutoff := base.Add(3 * time.Hour) // removes first 3
	removed, err := Prune(path, PruneOptions{OlderThan: cutoff})
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if removed != 3 {
		t.Errorf("expected 3 removed, got %d", removed)
	}
}

func TestPrune_NothingToRemove(t *testing.T) {
	base := time.Now()
	entries := makeEntries(3, base)
	path := writeEntries(t, entries)

	removed, err := Prune(path, PruneOptions{KeepLast: 10})
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}
