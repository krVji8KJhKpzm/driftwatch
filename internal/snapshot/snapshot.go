package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Snapshot holds a timestamped drift detection result set.
type Snapshot struct {
	Timestamp time.Time          `json:"timestamp"`
	Results   []drift.Result     `json:"results"`
}

// Save writes a snapshot to the given file path as JSON.
func Save(path string, results []drift.Result) error {
	snap := Snapshot{
		Timestamp: time.Now().UTC(),
		Results:   results,
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("snapshot: write failed: %w", err)
	}

	return nil
}

// Load reads a snapshot from the given file path.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: read failed: %w", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("snapshot: unmarshal failed: %w", err)
	}

	return &snap, nil
}
