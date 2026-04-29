package pinned

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/drift"
)

// PinnedEntry records an expected (approved) drift state for a container.
type PinnedEntry struct {
	Name      string            `json:"name"`
	Image     string            `json:"image,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	PinnedAt  time.Time         `json:"pinned_at"`
	Comment   string            `json:"comment,omitempty"`
}

// Store holds all pinned entries keyed by container name.
type Store struct {
	Entries map[string]PinnedEntry `json:"entries"`
}

// IsPinned returns true when the result's current drift matches a pinned entry.
func IsPinned(store *Store, result drift.Result) bool {
	if store == nil {
		return false
	}
	entry, ok := store.Entries[result.Name]
	if !ok {
		return false
	}
	for _, d := range result.Diffs {
		if d.Field == "image" && entry.Image != "" && d.Got != entry.Image {
			return false
		}
		if d.Field != "image" && entry.Env != nil {
			expected, exists := entry.Env[d.Field]
			if !exists || d.Got != expected {
				return false
			}
		}
	}
	return true
}

// Save persists the store to the given path as JSON.
func Save(path string, store *Store) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("pinned: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads a pinned store from path. Returns an empty store if the file
// does not exist.
func Load(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Store{Entries: make(map[string]PinnedEntry)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pinned: read: %w", err)
	}
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("pinned: parse: %w", err)
	}
	if store.Entries == nil {
		store.Entries = make(map[string]PinnedEntry)
	}
	return &store, nil
}

// Pin adds or updates a pinned entry for the given result.
func Pin(store *Store, result drift.Result, comment string) {
	env := make(map[string]string)
	image := ""
	for _, d := range result.Diffs {
		if d.Field == "image" {
			image = d.Got
		} else {
			env[d.Field] = d.Got
		}
	}
	store.Entries[result.Name] = PinnedEntry{
		Name:     result.Name,
		Image:    image,
		Env:      env,
		PinnedAt: time.Now().UTC(),
		Comment:  comment,
	}
}

// Unpin removes a pinned entry by container name.
func Unpin(store *Store, name string) bool {
	if _, ok := store.Entries[name]; !ok {
		return false
	}
	delete(store.Entries, name)
	return true
}
