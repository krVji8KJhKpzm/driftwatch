package suppress

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempSuppress(t *testing.T, cfg Config) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	p := filepath.Join(t.TempDir(), "suppress.json")
	if err := os.WriteFile(p, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestLoadConfig_Valid(t *testing.T) {
	cfg := Config{Rules: []Rule{{Container: "web", Field: "image", Reason: "known"}}}
	p := writeTempSuppress(t, cfg)
	got, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(got.Rules))
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/suppress.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIsSuppressed_NilConfig(t *testing.T) {
	if IsSuppressed(nil, "web", "image") {
		t.Error("nil config should never suppress")
	}
}

func TestIsSuppressed_Match(t *testing.T) {
	cfg := &Config{Rules: []Rule{{Container: "web", Field: "image", Reason: "ok"}}}
	if !IsSuppressed(cfg, "web", "image") {
		t.Error("expected suppression for web/image")
	}
}

func TestIsSuppressed_NoMatch(t *testing.T) {
	cfg := &Config{Rules: []Rule{{Container: "web", Field: "image", Reason: "ok"}}}
	if IsSuppressed(cfg, "api", "image") {
		t.Error("should not suppress different container")
	}
}

func TestIsSuppressed_Expired(t *testing.T) {
	cfg := &Config{Rules: []Rule{
		{Container: "web", Field: "image", Expires: time.Now().Add(-time.Hour)},
	}}
	if IsSuppressed(cfg, "web", "image") {
		t.Error("expired rule should not suppress")
	}
}

func TestIsSuppressed_Wildcard(t *testing.T) {
	cfg := &Config{Rules: []Rule{{Container: "*", Field: "env.DEBUG", Reason: "global"}}}
	if !IsSuppressed(cfg, "any-container", "env.DEBUG") {
		t.Error("wildcard container should match any container")
	}
}
