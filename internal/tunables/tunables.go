package tunables

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds runtime tunables that adjust drift detection behaviour.
type Config struct {
	MaxEnvDiffs     int      `json:"max_env_diffs"`
	IgnoreImageTag  bool     `json:"ignore_image_tag"`
	EnvKeyPrefixes  []string `json:"env_key_prefixes"`
	StrictMode      bool     `json:"strict_mode"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MaxEnvDiffs:    0,
		IgnoreImageTag: false,
		EnvKeyPrefixes: []string{},
		StrictMode:     false,
	}
}

// Load reads a tunables config from a JSON file at path.
// If path is empty the default config is returned.
func Load(path string) (*Config, error) {
	if path == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tunables: read %q: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("tunables: parse %q: %w", path, err)
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("tunables: invalid config: %w", err)
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.MaxEnvDiffs < 0 {
		return fmt.Errorf("max_env_diffs must be >= 0, got %d", cfg.MaxEnvDiffs)
	}
	return nil
}
