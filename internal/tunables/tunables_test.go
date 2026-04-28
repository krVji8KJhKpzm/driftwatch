package tunables

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTempTunables(t *testing.T, cfg map[string]any) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	p := filepath.Join(t.TempDir(), "tunables.json")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxEnvDiffs != 0 || cfg.StrictMode {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}

func TestLoad_Valid(t *testing.T) {
	p := writeTempTunables(t, map[string]any{
		"max_env_diffs":    5,
		"ignore_image_tag": true,
		"strict_mode":      true,
		"env_key_prefixes": []string{"APP_", "SVC_"},
	})
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxEnvDiffs != 5 {
		t.Errorf("max_env_diffs: want 5, got %d", cfg.MaxEnvDiffs)
	}
	if !cfg.IgnoreImageTag {
		t.Error("ignore_image_tag: want true")
	}
	if len(cfg.EnvKeyPrefixes) != 2 {
		t.Errorf("env_key_prefixes: want 2, got %d", len(cfg.EnvKeyPrefixes))
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("/no/such/tunables.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(p, []byte("{not json"), 0o644)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoad_NegativeMaxEnvDiffs(t *testing.T) {
	p := writeTempTunables(t, map[string]any{"max_env_diffs": -1})
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error for negative max_env_diffs")
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxEnvDiffs != 0 {
		t.Errorf("want 0, got %d", cfg.MaxEnvDiffs)
	}
	if cfg.IgnoreImageTag {
		t.Error("want false")
	}
	if cfg.EnvKeyPrefixes == nil {
		t.Error("env_key_prefixes should be non-nil slice")
	}
}
