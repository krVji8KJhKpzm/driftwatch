package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ContainerSpec represents the desired state of a container as defined in a manifest.
type ContainerSpec struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Env         map[string]string `yaml:"env"`
	Ports       []string          `yaml:"ports"`
	Labels      map[string]string `yaml:"labels"`
	RestartPolicy string          `yaml:"restartPolicy"`
}

// Manifest represents a driftwatch manifest file.
type Manifest struct {
	Version    string          `yaml:"version"`
	Containers []ContainerSpec `yaml:"containers"`
}

// LoadFromFile reads and parses a YAML manifest file from the given path.
func LoadFromFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file %q: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest file %q: %w", path, err)
	}

	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest %q: %w", path, err)
	}

	return &m, nil
}

// validate performs basic sanity checks on a parsed manifest.
func (m *Manifest) validate() error {
	if m.Version == "" {
		return fmt.Errorf("manifest must specify a version")
	}
	names := make(map[string]struct{}, len(m.Containers))
	for i, c := range m.Containers {
		if c.Name == "" {
			return fmt.Errorf("container at index %d is missing a name", i)
		}
		if c.Image == "" {
			return fmt.Errorf("container %q is missing an image", c.Name)
		}
		if _, dup := names[c.Name]; dup {
			return fmt.Errorf("duplicate container name %q", c.Name)
		}
		names[c.Name] = struct{}{}
	}
	return nil
}
