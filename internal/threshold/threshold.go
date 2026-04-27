package threshold

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driftwatch/internal/drift"
)

// Config defines drift thresholds that trigger violations.
type Config struct {
	MaxDriftCount  int      `json:"max_drift_count"`  // max number of drifted containers
	MaxDriftRate   float64  `json:"max_drift_rate"`   // 0.0–1.0 fraction of total containers
	BlockedFields  []string `json:"blocked_fields"`   // fields that must never drift
}

// Violation represents a single threshold breach.
type Violation struct {
	Rule    string
	Message string
}

// Evaluate checks drift results against the config and returns any violations.
func Evaluate(cfg *Config, results []drift.Result) []Violation {
	if cfg == nil {
		return nil
	}

	var violations []Violation

	total := len(results)
	driftedCount := 0
	for _, r := range results {
		if r.Drifted {
			driftedCount++
		}
	}

	if cfg.MaxDriftCount > 0 && driftedCount > cfg.MaxDriftCount {
		violations = append(violations, Violation{
			Rule:    "max_drift_count",
			Message: fmt.Sprintf("drifted containers (%d) exceeds threshold (%d)", driftedCount, cfg.MaxDriftCount),
		})
	}

	if cfg.MaxDriftRate > 0 && total > 0 {
		rate := float64(driftedCount) / float64(total)
		if rate > cfg.MaxDriftRate {
			violations = append(violations, Violation{
				Rule:    "max_drift_rate",
				Message: fmt.Sprintf("drift rate (%.2f) exceeds threshold (%.2f)", rate, cfg.MaxDriftRate),
			})
		}
	}

	for _, r := range results {
		if !r.Drifted {
			continue
		}
		for _, d := range r.Diffs {
			if isBlocked(d.Field, cfg.BlockedFields) {
				violations = append(violations, Violation{
					Rule:    "blocked_field",
					Message: fmt.Sprintf("container %q has drift on blocked field %q", r.Name, d.Field),
				})
			}
		}
	}

	return violations
}

// LoadConfig reads a threshold config from a JSON file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("threshold: read %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("threshold: parse %s: %w", path, err)
	}
	return &cfg, nil
}

func isBlocked(field string, blocked []string) bool {
	for _, b := range blocked {
		if b == field {
			return true
		}
	}
	return false
}
