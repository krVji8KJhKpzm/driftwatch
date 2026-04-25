package tag

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// Tag represents a named label attached to a drift snapshot.
type Tag struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	SnapshotPath string  `json:"snapshot_path"`
	Note      string    `json:"note,omitempty"`
}

// TagStore holds a collection of tags persisted to disk.
type TagStore struct {
	Tags []Tag `json:"tags"`
}

// Save persists a new tag to the given store file.
func Save(storePath, name, snapshotPath, note string) error {
	store, err := loadStore(storePath)
	if err != nil {
		return fmt.Errorf("load tag store: %w", err)
	}
	for _, t := range store.Tags {
		if t.Name == name {
			return fmt.Errorf("tag %q already exists", name)
		}
	}
	store.Tags = append(store.Tags, Tag{
		Name:         name,
		CreatedAt:    time.Now().UTC(),
		SnapshotPath: snapshotPath,
		Note:         note,
	})
	return writeStore(storePath, store)
}

// List returns all tags sorted by creation time descending.
func List(storePath string) ([]Tag, error) {
	store, err := loadStore(storePath)
	if err != nil {
		return nil, fmt.Errorf("load tag store: %w", err)
	}
	sort.Slice(store.Tags, func(i, j int) bool {
		return store.Tags[i].CreatedAt.After(store.Tags[j].CreatedAt)
	})
	return store.Tags, nil
}

// Delete removes a tag by name from the store.
func Delete(storePath, name string) error {
	store, err := loadStore(storePath)
	if err != nil {
		return fmt.Errorf("load tag store: %w", err)
	}
	next := store.Tags[:0]
	found := false
	for _, t := range store.Tags {
		if t.Name == name {
			found = true
			continue
		}
		next = append(next, t)
	}
	if !found {
		return fmt.Errorf("tag %q not found", name)
	}
	store.Tags = next
	return writeStore(storePath, store)
}

func loadStore(path string) (*TagStore, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &TagStore{}, nil
	}
	if err != nil {
		return nil, err
	}
	var store TagStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("corrupt tag store: %w", err)
	}
	return &store, nil
}

func writeStore(path string, store *TagStore) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
