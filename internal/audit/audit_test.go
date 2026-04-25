package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.log")
}

func TestRecord_CreatesAndAppends(t *testing.T) {
	path := tempPath(t)
	l := NewLogger(path)

	evt := Event{
		Timestamp:     time.Now().UTC(),
		Type:          EventDriftDetected,
		ContainerName: "web",
		Details:       "1 field(s) drifted",
	}
	if err := l.Record(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ContainerName != "web" {
		t.Errorf("expected container 'web', got %q", events[0].ContainerName)
	}
}

func TestFromResults_DriftedAndClean(t *testing.T) {
	path := tempPath(t)
	l := NewLogger(path)

	results := []drift.Result{
		{ContainerName: "api", Drifted: true, Fields: []drift.FieldDiff{{Field: "image"}}, },
		{ContainerName: "db", Drifted: false},
	}
	if err := l.FromResults(results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != EventDriftDetected {
		t.Errorf("expected drift_detected, got %q", events[0].Type)
	}
	if events[1].Type != EventScanCompleted {
		t.Errorf("expected scan_completed, got %q", events[1].Type)
	}
}

func TestLoadEvents_NotFound(t *testing.T) {
	events, err := LoadEvents("/nonexistent/audit.log")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if events != nil {
		t.Errorf("expected nil events, got %v", events)
	}
}

func TestLoadEvents_CorruptLine(t *testing.T) {
	path := tempPath(t)
	if err := os.WriteFile(path, []byte("not-json\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadEvents(path)
	if err == nil {
		t.Error("expected error for corrupt JSON line")
	}
}

func TestRecord_InvalidPath(t *testing.T) {
	l := NewLogger("/no/such/dir/audit.log")
	err := l.Record(Event{Type: EventScanCompleted})
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
