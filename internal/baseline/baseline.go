package baseline

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Entry represents a saved baseline for a container.
type Entry struct {
	ContainerName string            `json:"container_name"`
	Image         string            `json:"image"`
	Env           map[string]string `json:"env"`
	RecordedAt    time.Time         `json:"recorded_at"`
}

// Baseline holds baseline entries keyed by container name.
type Baseline struct {
	Entries map[string]Entry `json:"entries"`
}

// Save writes a baseline derived from drift results to the given path.
func Save(path string, results []drift.Result) error {
	b := Baseline{Entries: make(map[string]Entry)}
	for _, r := range results {
		b.Entries[r.ContainerName] = Entry{
			ContainerName: r.ContainerName,
			Image:         r.ActualImage,
			Env:           r.ActualEnv,
			RecordedAt:    time.Now().UTC(),
		}
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("baseline: write %s: %w", path, err)
	}
	return nil
}

// Load reads a baseline from the given path.
func Load(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("baseline: file not found: %s", path)
		}
		return nil, fmt.Errorf("baseline: read %s: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("baseline: parse %s: %w", path, err)
	}
	return &b, nil
}

// Compare checks drift results against a loaded baseline and returns
// only those results whose actual state differs from the baseline.
func Compare(b *Baseline, results []drift.Result) []drift.Result {
	var drifted []drift.Result
	for _, r := range results {
		entry, ok := b.Entries[r.ContainerName]
		if !ok {
			// Container not in baseline — treat as new drift.
			drifted = append(drifted, r)
			continue
		}
		if r.ActualImage != entry.Image {
			drifted = append(drifted, r)
			continue
		}
		if envChanged(entry.Env, r.ActualEnv) {
			drifted = append(drifted, r)
		}
	}
	return drifted
}

func envChanged(a, b map[string]string) bool {
	if len(a) != len(b) {
		return true
	}
	for k, v := range a {
		if b[k] != v {
			return true
		}
	}
	return false
}
