package alert

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempAlert(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "alerts.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempAlert: %v", err)
	}
	return p
}

func TestLoadConfig_Valid(t *testing.T) {
	path := writeTempAlert(t, `
rules:
  - on_image_drift: true
    level: error
  - on_env_drift: true
    level: warn
`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].Level != LevelError {
		t.Errorf("expected error level for rule 0")
	}
}

func TestLoadConfig_DefaultLevel(t *testing.T) {
	path := writeTempAlert(t, `
rules:
  - on_image_drift: true
`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Rules[0].Level != LevelWarn {
		t.Errorf("expected default warn level, got %s", cfg.Rules[0].Level)
	}
}

func TestLoadConfig_InvalidLevel(t *testing.T) {
	path := writeTempAlert(t, `
rules:
  - on_image_drift: true
    level: critical
`)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/alerts.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_EmptyRules(t *testing.T) {
	path := writeTempAlert(t, `rules: []`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules) != 0 {
		t.Errorf("expected 0 rules")
	}
}
