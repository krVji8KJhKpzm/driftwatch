package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Entry represents a single recorded drift check in history.
type Entry struct {
	Timestamp time.Time     `json:"timestamp"`
	Results   []drift.Result `json:"results"`
}

// Record appends a new history entry to the history file at the given path.
func Record(path string, results []drift.Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("history: create directory: %w", err)
	}

	var entries []Entry
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("history: parse existing file: %w", err)
		}
	}

	entries = append(entries, Entry{
		Timestamp: time.Now().UTC(),
		Results:   results,
	})

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("history: write file: %w", err)
	}
	return nil
}

// Load reads all history entries from the given path.
func Load(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("history: file not found: %s", path)
		}
		return nil, fmt.Errorf("history: read file: %w", err)
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("history: parse file: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})
	return entries, nil
}

// Latest returns the most recent history entry, or an error if history is empty.
func Latest(path string) (*Entry, error) {
	entries, err := Load(path)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("history: no entries found")
	}
	return &entries[len(entries)-1], nil
}
