package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Rule defines a single policy rule applied during drift detection.
type Rule struct {
	Name        string   `yaml:"name"`
	IgnoreEnvs  []string `yaml:"ignore_envs"`
	IgnoreImage bool     `yaml:"ignore_image"`
}

// Policy holds a collection of named rules keyed by container name.
type Policy struct {
	Rules map[string]Rule `yaml:"rules"`
}

// LoadPolicy reads and parses a policy YAML file from the given path.
func LoadPolicy(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("policy: read file: %w", err)
	}
	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("policy: parse yaml: %w", err)
	}
	if p.Rules == nil {
		p.Rules = make(map[string]Rule)
	}
	return &p, nil
}

// RuleFor returns the Rule for a given container name, or an empty Rule if none exists.
func (p *Policy) RuleFor(containerName string) Rule {
	if r, ok := p.Rules[containerName]; ok {
		return r
	}
	return Rule{}
}

// ShouldIgnoreEnv reports whether the given env key should be ignored for containerName.
func (p *Policy) ShouldIgnoreEnv(containerName, envKey string) bool {
	r := p.RuleFor(containerName)
	for _, k := range r.IgnoreEnvs {
		if k == envKey {
			return true
		}
	}
	return false
}
