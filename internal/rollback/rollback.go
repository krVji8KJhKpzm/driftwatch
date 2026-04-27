package rollback

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Checkpoint represents a named rollback point capturing drift results at a moment in time.
type Checkpoint struct {
	Name      string             `json:"name"`
	CreatedAt time.Time          `json:"created_at"`
	Results   []drift.Result     `json:"results"`
}

// Store holds a collection of checkpoints persisted to disk.
type Store struct {
	Checkpoints []Checkpoint `json:"checkpoints"`
}

// Save writes a named checkpoint to the given file path.
func Save(path, name string, results []drift.Result) error {
	store, err := loadStore(path)
	if err != nil {
		return err
	}
	for _, cp := range store.Checkpoints {
		if cp.Name == name {
			return fmt.Errorf("checkpoint %q already exists", name)
		}
	}
	store.Checkpoints = append(store.Checkpoints, Checkpoint{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Results:   results,
	})
	return writeStore(path, store)
}

// List returns all checkpoints sorted by creation time descending.
func List(path string) ([]Checkpoint, error) {
	store, err := loadStore(path)
	if err != nil {
		return nil, err
	}
	sort.Slice(store.Checkpoints, func(i, j int) bool {
		return store.Checkpoints[i].CreatedAt.After(store.Checkpoints[j].CreatedAt)
	})
	return store.Checkpoints, nil
}

// Get retrieves a checkpoint by name.
func Get(path, name string) (*Checkpoint, error) {
	store, err := loadStore(path)
	if err != nil {
		return nil, err
	}
	for _, cp := range store.Checkpoints {
		if cp.Name == name {
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("checkpoint %q not found", name)
}

// Delete removes a checkpoint by name.
func Delete(path, name string) error {
	store, err := loadStore(path)
	if err != nil {
		return err
	}
	filtered := store.Checkpoints[:0]
	found := false
	for _, cp := range store.Checkpoints {
		if cp.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, cp)
	}
	if !found {
		return fmt.Errorf("checkpoint %q not found", name)
	}
	store.Checkpoints = filtered
	return writeStore(path, store)
}

func loadStore(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Store{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read rollback store: %w", err)
	}
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parse rollback store: %w", err)
	}
	return &store, nil
}

func writeStore(path string, store *Store) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rollback store: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
