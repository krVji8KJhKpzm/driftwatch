package baseline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func sampleResults() []drift.Result {
	return []drift.Result{
		{
			ContainerName: "api",
			ActualImage:   "nginx:1.25",
			ActualEnv:     map[string]string{"PORT": "8080"},
			Drifted:       false,
		},
		{
			ContainerName: "worker",
			ActualImage:   "alpine:3.18",
			ActualEnv:     map[string]string{"LOG_LEVEL": "info"},
			Drifted:       true,
		},
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	if err := Save(path, sampleResults()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	b, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(b.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(b.Entries))
	}
	if b.Entries["api"].Image != "nginx:1.25" {
		t.Errorf("unexpected image: %s", b.Entries["api"].Image)
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/baseline.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("not-json{"), 0o644)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCompare_NoDrift(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")
	results := sampleResults()
	_ = Save(path, results)
	b, _ := Load(path)

	out := Compare(b, results)
	if len(out) != 0 {
		t.Errorf("expected no drift vs baseline, got %d", len(out))
	}
}

func TestCompare_ImageDrift(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")
	_ = Save(path, sampleResults())
	b, _ := Load(path)

	changed := sampleResults()
	changed[0].ActualImage = "nginx:1.26"

	out := Compare(b, changed)
	if len(out) != 1 || out[0].ContainerName != "api" {
		t.Errorf("expected drift on api, got %+v", out)
	}
}

func TestCompare_NewContainer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")
	_ = Save(path, sampleResults())
	b, _ := Load(path)

	extra := append(sampleResults(), drift.Result{
		ContainerName: "cache",
		ActualImage:   "redis:7",
		ActualEnv:     map[string]string{},
	})

	out := Compare(b, extra)
	if len(out) != 1 || out[0].ContainerName != "cache" {
		t.Errorf("expected drift only for new container, got %+v", out)
	}
}
