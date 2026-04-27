package suppress

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Rule defines a suppression rule that silences drift for a specific container and field.
type Rule struct {
	Container string    `json:"container"`
	Field     string    `json:"field"`
	Reason    string    `json:"reason"`
	Expires   time.Time `json:"expires,omitempty"`
}

// Config holds all suppression rules.
type Config struct {
	Rules []Rule `json:"rules"`
}

// LoadConfig reads a suppression config from the given JSON file path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("suppress config not found: %s", path)
		}
		return nil, fmt.Errorf("reading suppress config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing suppress config: %w", err)
	}
	return &cfg, nil
}

// IsSuppressed returns true if the given container+field combination is
// covered by an active (non-expired) suppression rule.
func IsSuppressed(cfg *Config, container, field string) bool {
	if cfg == nil {
		return false
	}
	now := time.Now()
	for _, r := range cfg.Rules {
		if r.Container != container && r.Container != "*" {
			continue
		}
		if r.Field != field && r.Field != "*" {
			continue
		}
		if !r.Expires.IsZero() && r.Expires.Before(now) {
			continue
		}
		return true
	}
	return false
}
