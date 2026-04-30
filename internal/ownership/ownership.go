package ownership

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Owner represents a team or individual responsible for a container.
type Owner struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Team  string `json:"team,omitempty"`
}

// Rule maps a container name prefix or exact name to an owner.
type Rule struct {
	Match string `json:"match"`
	Owner Owner  `json:"owner"`
}

// Config holds all ownership rules.
type Config struct {
	Rules []Rule `json:"rules"`
}

// Assignment pairs a drift result with its resolved owner.
type Assignment struct {
	ContainerName string
	Drifted       bool
	Owner         *Owner
}

// LoadConfig reads an ownership config from a JSON file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ownership: read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ownership: parse config: %w", err)
	}
	return &cfg, nil
}

// Assign resolves an owner for each drift result using the provided config.
// Results are sorted by container name for deterministic output.
func Assign(results []drift.Result, cfg *Config) []Assignment {
	assignments := make([]Assignment, 0, len(results))
	for _, r := range results {
		a := Assignment{
			ContainerName: r.ContainerName,
			Drifted:       r.Drifted,
			Owner:         resolve(r.ContainerName, cfg),
		}
		assignments = append(assignments, a)
	}
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].ContainerName < assignments[j].ContainerName
	})
	return assignments
}

// resolve finds the first matching rule for the given container name.
// It prefers exact matches over prefix matches.
func resolve(name string, cfg *Config) *Owner {
	if cfg == nil {
		return nil
	}
	var prefixMatch *Owner
	for i, rule := range cfg.Rules {
		if rule.Match == name {
			owner := cfg.Rules[i].Owner
			return &owner
		}
		if prefixMatch == nil && len(rule.Match) > 0 && len(name) >= len(rule.Match) && name[:len(rule.Match)] == rule.Match {
			owner := cfg.Rules[i].Owner
			prefixMatch = &owner
		}
	}
	return prefixMatch
}
