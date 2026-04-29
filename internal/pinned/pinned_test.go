package pinned_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/pinned"
)

func makeResult(name string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Name:    name,
		Drifted: len(diffs) > 0,
		Diffs:   diffs,
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pinned.json")

	store := &pinned.Store{Entries: map[string]pinned.PinnedEntry{
		"web": {Name: "web", Image: "nginx:1.25", PinnedAt: time.Now().UTC()},
	}}

	if err := pinned.Save(path, store); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := pinned.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := loaded.Entries["web"]; !ok {
		t.Error("expected entry for 'web'")
	}
}

func TestLoad_NotFound_ReturnsEmpty(t *testing.T) {
	store, err := pinned.Load("/nonexistent/path/pinned.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.Entries) != 0 {
		t.Errorf("expected empty store, got %d entries", len(store.Entries))
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	f, _ := os.CreateTemp("", "pinned*.json")
	f.WriteString("{not valid json")
	f.Close()
	defer os.Remove(f.Name())

	_, err := pinned.Load(f.Name())
	if err == nil {
		t.Error("expected error for corrupt JSON")
	}
}

func TestPin_AddsEntry(t *testing.T) {
	store := &pinned.Store{Entries: make(map[string]pinned.PinnedEntry)}
	result := makeResult("api", []drift.Diff{
		{Field: "image", Want: "api:1.0", Got: "api:1.1"},
	})
	pinned.Pin(store, result, "approved upgrade")

	entry, ok := store.Entries["api"]
	if !ok {
		t.Fatal("expected entry for 'api'")
	}
	if entry.Image != "api:1.1" {
		t.Errorf("expected image api:1.1, got %s", entry.Image)
	}
	if entry.Comment != "approved upgrade" {
		t.Errorf("unexpected comment: %s", entry.Comment)
	}
}

func TestUnpin_RemovesEntry(t *testing.T) {
	store := &pinned.Store{Entries: map[string]pinned.PinnedEntry{
		"db": {Name: "db"},
	}}
	if !pinned.Unpin(store, "db") {
		t.Error("expected true when removing existing entry")
	}
	if _, ok := store.Entries["db"]; ok {
		t.Error("entry should have been removed")
	}
}

func TestUnpin_MissingReturnsFalse(t *testing.T) {
	store := &pinned.Store{Entries: make(map[string]pinned.PinnedEntry)}
	if pinned.Unpin(store, "ghost") {
		t.Error("expected false for missing entry")
	}
}

func TestIsPinned_MatchesImageDrift(t *testing.T) {
	store := &pinned.Store{Entries: map[string]pinned.PinnedEntry{
		"web": {Name: "web", Image: "nginx:1.25"},
	}}
	result := makeResult("web", []drift.Diff{
		{Field: "image", Want: "nginx:1.24", Got: "nginx:1.25"},
	})
	if !pinned.IsPinned(store, result) {
		t.Error("expected drift to be pinned")
	}
}

func TestIsPinned_NilStore(t *testing.T) {
	result := makeResult("web", []drift.Diff{{Field: "image", Got: "nginx:1.25"}})
	if pinned.IsPinned(nil, result) {
		t.Error("expected false for nil store")
	}
}

func TestIsPinned_NotRegistered(t *testing.T) {
	store := &pinned.Store{Entries: make(map[string]pinned.PinnedEntry)}
	result := makeResult("unknown", []drift.Diff{{Field: "image", Got: "x:1"}})
	if pinned.IsPinned(store, result) {
		t.Error("expected false for unregistered container")
	}
}
