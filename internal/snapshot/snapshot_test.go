package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/snapshot"
)

func sampleResults() []drift.Result {
	return []drift.Result{
		{
			Name:    "web",
			Drifted: true,
			Differences: []drift.Difference{
				{Field: "image", Expected: "nginx:1.24", Actual: "nginx:1.25"},
			},
		},
		{
			Name:    "db",
			Drifted: false,
		},
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")

	results := sampleResults()
	if err := snapshot.Save(path, results); err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}

	snap, err := snapshot.Load(path)
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}

	if len(snap.Results) != len(results) {
		t.Errorf("expected %d results, got %d", len(results), len(snap.Results))
	}
	if snap.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if snap.Results[0].Name != "web" {
		t.Errorf("expected first result name 'web', got %q", snap.Results[0].Name)
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path/snap.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSave_InvalidPath(t *testing.T) {
	err := snapshot.Save("/nonexistent/dir/snap.json", sampleResults())
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("not json{"), 0644)

	_, err := snapshot.Load(path)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}

func TestSnapshotTimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ts.json")
	before := time.Now().UTC().Add(-time.Second)

	_ = snapshot.Save(path, sampleResults())
	snap, _ := snapshot.Load(path)

	if snap.Timestamp.Before(before) {
		t.Errorf("timestamp %v is before expected lower bound %v", snap.Timestamp, before)
	}
}
