package alert

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds alert rules loaded from a YAML file.
type Config struct {
	Rules []Rule `yaml:"rules"`
}

// LoadConfig reads alert rules from a YAML file at the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("alert: read config %q: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("alert: parse config %q: %w", path, err)
	}
	for i, r := range cfg.Rules {
		if r.Level == "" {
			cfg.Rules[i].Level = LevelWarn
		}
		if r.Level != LevelWarn && r.Level != LevelError {
			return nil, fmt.Errorf("alert: rule %d has invalid level %q", i, r.Level)
		}
	}
	return &cfg, nil
}
