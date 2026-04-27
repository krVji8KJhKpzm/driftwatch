package rollback_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/rollback"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "rollback.json")
}

func sampleResults() []drift.Result {
	return []drift.Result{
		{ContainerName: "api", Drifted: true},
		{ContainerName: "worker", Drifted: false},
	}
}

func TestSave_CreatesCheckpoint(t *testing.T) {
	p := tempPath(t)
	if err := rollback.Save(p, "v1", sampleResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp, err := rollback.Get(p, "v1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cp.Name != "v1" {
		t.Errorf("expected name v1, got %s", cp.Name)
	}
	if len(cp.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(cp.Results))
	}
}

func TestSave_DuplicateReturnsError(t *testing.T) {
	p := tempPath(t)
	_ = rollback.Save(p, "v1", sampleResults())
	if err := rollback.Save(p, "v1", sampleResults()); err == nil {
		t.Error("expected error for duplicate checkpoint name")
	}
}

func TestList_SortedByCreatedAtDesc(t *testing.T) {
	p := tempPath(t)
	_ = rollback.Save(p, "v1", sampleResults())
	_ = rollback.Save(p, "v2", sampleResults())
	_ = rollback.Save(p, "v3", sampleResults())

	list, err := rollback.List(p)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 checkpoints, got %d", len(list))
	}
	// most recent first
	if list[0].Name != "v3" {
		t.Errorf("expected v3 first, got %s", list[0].Name)
	}
}

func TestGet_NotFound(t *testing.T) {
	p := tempPath(t)
	_, err := rollback.Get(p, "missing")
	if err == nil {
		t.Error("expected error for missing checkpoint")
	}
}

func TestDelete_RemovesCheckpoint(t *testing.T) {
	p := tempPath(t)
	_ = rollback.Save(p, "v1", sampleResults())
	_ = rollback.Save(p, "v2", sampleResults())

	if err := rollback.Delete(p, "v1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	list, _ := rollback.List(p)
	if len(list) != 1 || list[0].Name != "v2" {
		t.Errorf("expected only v2 remaining, got %+v", list)
	}
}

func TestDelete_NotFound(t *testing.T) {
	p := tempPath(t)
	if err := rollback.Delete(p, "ghost"); err == nil {
		t.Error("expected error deleting non-existent checkpoint")
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	p := tempPath(t)
	_ = os.WriteFile(p, []byte("{bad json"), 0644)
	_, err := rollback.List(p)
	if err == nil {
		t.Error("expected error on corrupt JSON")
	}
}
