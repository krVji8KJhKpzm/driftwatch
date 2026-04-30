package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Config defines per-container or global drift field quotas.
type Config struct {
	GlobalMaxFields int            `json:"global_max_fields"`
	ContainerRules  []ContainerRule `json:"container_rules"`
}

// ContainerRule sets a max allowed drifted-field count for a named container.
type ContainerRule struct {
	Name         string `json:"name"`
	MaxFields    int    `json:"max_fields"`
}

// Violation records a container that exceeded its quota.
type Violation struct {
	Container  string
	DriftCount int
	Limit      int
}

// LoadConfig reads a quota config from a JSON file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("quota: read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("quota: parse config: %w", err)
	}
	return &cfg, nil
}

// Evaluate checks drift results against quota limits and returns violations.
func Evaluate(results []drift.Result, cfg *Config) []Violation {
	if cfg == nil {
		return nil
	}

	ruleMap := make(map[string]int)
	for _, r := range cfg.ContainerRules {
		ruleMap[r.Name] = r.MaxFields
	}

	var violations []Violation
	for _, res := range results {
		if !res.Drifted {
			continue
		}
		count := len(res.Diffs)
		limit := cfg.GlobalMaxFields
		if perRule, ok := ruleMap[res.Name]; ok {
			limit = perRule
		}
		if limit > 0 && count > limit {
			violations = append(violations, Violation{
				Container:  res.Name,
				DriftCount: count,
				Limit:      limit,
			})
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Container < violations[j].Container
	})
	return violations
}
