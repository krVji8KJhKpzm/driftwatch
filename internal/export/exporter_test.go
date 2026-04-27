package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResults() []drift.Result {
	return []drift.Result{
		{
			Name:    "api",
			Drifted: true,
			Diffs: []drift.Diff{
				{Field: "image", Expected: "api:1.0", Actual: "api:2.0"},
			},
		},
		{
			Name:    "worker",
			Drifted: false,
		},
	}
}

func TestExport_JSON(t *testing.T) {
	dir := t.TempDir()
	err := Export(makeResults(), Options{Format: FormatJSON, OutputDir: dir, Filename: "out.json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "out.json"))
	var got []drift.Result
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 results, got %d", len(got))
	}
}

func TestExport_Markdown(t *testing.T) {
	dir := t.TempDir()
	err := Export(makeResults(), Options{Format: FormatMarkdown, OutputDir: dir, Filename: "out.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "out.md"))
	content := string(data)
	if !strings.Contains(content, "# Drift Report") {
		t.Error("missing markdown header")
	}
	if !strings.Contains(content, "⚠️ Drifted") {
		t.Error("expected drifted marker")
	}
	if !strings.Contains(content, "api:1.0") {
		t.Error("expected expected value in diff")
	}
}

func TestExport_HTML(t *testing.T) {
	dir := t.TempDir()
	err := Export(makeResults(), Options{Format: FormatHTML, OutputDir: dir, Filename: "out.html"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "out.html"))
	content := string(data)
	if !strings.Contains(content, "<h1>Drift Report</h1>") {
		t.Error("missing HTML heading")
	}
	if !strings.Contains(content, "class=\"drifted\"") {
		t.Error("expected drifted class")
	}
}

func TestExport_DefaultFilename(t *testing.T) {
	dir := t.TempDir()
	err := Export(makeResults(), Options{Format: FormatJSON, OutputDir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "drift-report.json")); err != nil {
		t.Error("default filename not created")
	}
}

func TestExport_UnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	err := Export(makeResults(), Options{Format: "csv", OutputDir: dir})
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}
