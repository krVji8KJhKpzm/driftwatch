package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Entry represents a scheduled drift-check job.
type Entry struct {
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval_seconds"`
	Manifest string        `json:"manifest"`
	OutputFmt string       `json:"output_format,omitempty"`
	Enabled  bool          `json:"enabled"`
	LastRun  time.Time     `json:"last_run,omitempty"`
}

// Config holds the full schedule configuration.
type Config struct {
	Entries []Entry `json:"schedules"`
}

// LoadConfig reads a schedule config from path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("schedule config not found: %s", path)
		}
		return nil, fmt.Errorf("reading schedule config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing schedule config: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(cfg *Config) error {
	seen := make(map[string]bool)
	for i, e := range cfg.Entries {
		if e.Name == "" {
			return fmt.Errorf("schedule entry %d missing name", i)
		}
		if seen[e.Name] {
			return fmt.Errorf("duplicate schedule name: %s", e.Name)
		}
		seen[e.Name] = true
		if e.Interval <= 0 {
			return fmt.Errorf("schedule %q has invalid interval", e.Name)
		}
		if e.Manifest == "" {
			return fmt.Errorf("schedule %q missing manifest path", e.Name)
		}
	}
	return nil
}

// Due returns all enabled entries whose next run time has passed.
func Due(cfg *Config, now time.Time) []Entry {
	var due []Entry
	for _, e := range cfg.Entries {
		if !e.Enabled {
			continue
		}
		next := e.LastRun.Add(e.Interval)
		if e.LastRun.IsZero() || now.Equal(next) || now.After(next) {
			due = append(due, e)
		}
	}
	return due
}
