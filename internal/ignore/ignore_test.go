package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempIgnore(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "ignore.json")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write temp ignore: %v", err)
	}
	return p
}

func TestLoadConfig_Valid(t *testing.T) {
	p := writeTempIgnore(t, `{"rules":[{"container":"api","fields":["image","env.PORT"]}]}`)
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].Container != "api" {
		t.Errorf("expected container 'api', got %q", cfg.Rules[0].Container)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/ignore.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	p := writeTempIgnore(t, `{bad json`)
	_, err := LoadConfig(p)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestShouldIgnoreField_Match(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{Container: "api", Fields: []string{"image", "env.PORT"}},
		},
	}
	if !ShouldIgnoreField(cfg, "api", "image") {
		t.Error("expected image to be ignored for api")
	}
	if !ShouldIgnoreField(cfg, "api", "env.PORT") {
		t.Error("expected env.PORT to be ignored for api")
	}
}

func TestShouldIgnoreField_NoMatch(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{Container: "api", Fields: []string{"image"}},
		},
	}
	if ShouldIgnoreField(cfg, "worker", "image") {
		t.Error("should not ignore image for unrelated container")
	}
}

func TestShouldIgnoreField_Wildcard(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{Container: "*", Fields: []string{"env.DEBUG"}},
		},
	}
	if !ShouldIgnoreField(cfg, "any-container", "env.DEBUG") {
		t.Error("expected wildcard rule to match any container")
	}
}

func TestShouldIgnoreField_NilConfig(t *testing.T) {
	if ShouldIgnoreField(nil, "api", "image") {
		t.Error("nil config should never ignore")
	}
}
