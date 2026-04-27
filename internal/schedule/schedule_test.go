package schedule

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempSchedule(t *testing.T, cfg Config) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	p := filepath.Join(t.TempDir(), "schedule.json")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestLoadConfig_Valid(t *testing.T) {
	cfg := Config{Entries: []Entry{
		{Name: "nightly", Interval: time.Hour, Manifest: "manifest.yaml", Enabled: true},
	}}
	p := writeTempSchedule(t, cfg)
	got, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Name != "nightly" {
		t.Errorf("unexpected entries: %+v", got.Entries)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/no/such/file.json")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfig_DuplicateName(t *testing.T) {
	cfg := Config{Entries: []Entry{
		{Name: "job", Interval: time.Minute, Manifest: "m.yaml", Enabled: true},
		{Name: "job", Interval: time.Minute, Manifest: "m.yaml", Enabled: true},
	}}
	p := writeTempSchedule(t, cfg)
	_, err := LoadConfig(p)
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func TestLoadConfig_MissingManifest(t *testing.T) {
	cfg := Config{Entries: []Entry{
		{Name: "job", Interval: time.Minute, Enabled: true},
	}}
	p := writeTempSchedule(t, cfg)
	_, err := LoadConfig(p)
	if err == nil {
		t.Fatal("expected missing manifest error")
	}
}

func TestDue_ReturnsEnabledPastDue(t *testing.T) {
	now := time.Now()
	cfg := &Config{Entries: []Entry{
		{Name: "a", Interval: time.Minute, Manifest: "m.yaml", Enabled: true, LastRun: now.Add(-2 * time.Minute)},
		{Name: "b", Interval: time.Hour, Manifest: "m.yaml", Enabled: true, LastRun: now.Add(-30 * time.Second)},
		{Name: "c", Interval: time.Minute, Manifest: "m.yaml", Enabled: false, LastRun: now.Add(-2 * time.Minute)},
	}}
	due := Due(cfg, now)
	if len(due) != 1 || due[0].Name != "a" {
		t.Errorf("expected only 'a' due, got %+v", due)
	}
}

func TestDue_ZeroLastRunAlwaysDue(t *testing.T) {
	now := time.Now()
	cfg := &Config{Entries: []Entry{
		{Name: "fresh", Interval: time.Hour, Manifest: "m.yaml", Enabled: true},
	}}
	due := Due(cfg, now)
	if len(due) != 1 {
		t.Errorf("expected fresh entry to be due, got %d", len(due))
	}
}
