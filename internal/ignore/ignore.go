package ignore

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Rule defines a single ignore rule for a container or field.
type Rule struct {
	Container string   `json:"container"` // exact name or "*" for all
	Fields    []string `json:"fields"`    // e.g. ["image", "env.PORT"]
}

// Config holds a list of ignore rules loaded from a file.
type Config struct {
	Rules []Rule `json:"rules"`
}

// LoadConfig reads an ignore config from the given JSON file path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ignore config not found: %s", path)
		}
		return nil, fmt.Errorf("reading ignore config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing ignore config: %w", err)
	}
	return &cfg, nil
}

// ShouldIgnoreField returns true if the given field for the named container
// should be suppressed based on the loaded rules.
func ShouldIgnoreField(cfg *Config, container, field string) bool {
	if cfg == nil {
		return false
	}
	for _, rule := range cfg.Rules {
		if rule.Container != "*" && rule.Container != container {
			continue
		}
		for _, f := range rule.Fields {
			if strings.EqualFold(f, field) {
				return true
			}
		}
	}
	return false
}
