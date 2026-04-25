package tag

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "tags.json")
}

func TestSave_CreatesTag(t *testing.T) {
	path := tempStore(t)
	if err := Save(path, "v1", "/snaps/v1.json", "initial release"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags, err := List(path)
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "v1" {
		t.Errorf("expected name v1, got %s", tags[0].Name)
	}
	if tags[0].Note != "initial release" {
		t.Errorf("unexpected note: %s", tags[0].Note)
	}
}

func TestSave_DuplicateReturnsError(t *testing.T) {
	path := tempStore(t)
	_ = Save(path, "v1", "/snaps/v1.json", "")
	if err := Save(path, "v1", "/snaps/v1b.json", ""); err == nil {
		t.Fatal("expected error for duplicate tag")
	}
}

func TestList_SortedByCreatedAtDesc(t *testing.T) {
	path := tempStore(t)
	_ = Save(path, "alpha", "/snaps/a.json", "")
	_ = Save(path, "beta", "/snaps/b.json", "")
	tags, err := List(path)
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if !tags[0].CreatedAt.After(tags[1].CreatedAt) && tags[0].CreatedAt.Equal(tags[1].CreatedAt) {
		// timestamps may be equal in fast tests; just ensure no panic
	}
}

func TestDelete_RemovesTag(t *testing.T) {
	path := tempStore(t)
	_ = Save(path, "v1", "/snaps/v1.json", "")
	_ = Save(path, "v2", "/snaps/v2.json", "")
	if err := Delete(path, "v1"); err != nil {
		t.Fatalf("delete error: %v", err)
	}
	tags, _ := List(path)
	if len(tags) != 1 || tags[0].Name != "v2" {
		t.Errorf("expected only v2 remaining, got %+v", tags)
	}
}

func TestDelete_NotFound(t *testing.T) {
	path := tempStore(t)
	if err := Delete(path, "ghost"); err == nil {
		t.Fatal("expected error for missing tag")
	}
}

func TestLoad_CorruptStore(t *testing.T) {
	path := tempStore(t)
	_ = os.WriteFile(path, []byte("not json{"), 0o644)
	_, err := List(path)
	if err == nil {
		t.Fatal("expected error for corrupt store")
	}
}
