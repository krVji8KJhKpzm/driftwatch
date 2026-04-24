package history

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func makeExportEntries() []Entry {
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	return []Entry{
		{
			Timestamp: ts,
			Results: []sampleResult(
				[]struct{ name string; drifted bool }{
					{"web", true},
					{"db", false},
				},
			),
		},
	}
}

type sampleResult []struct {
	name    string
	drifted bool
}

func (s sampleResult) toResults() []driftResult {
	out := make([]driftResult, len(s))
	for i, v := range s {
		out[i] = driftResult{Name: v.name, Drifted: v.drifted}
	}
	return out
}

func exportEntries() []Entry {
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	return []Entry{
		{
			Timestamp: ts,
			Results: []driftResult{
				{Name: "web", Drifted: true},
				{Name: "db", Drifted: false},
			},
		},
	}
}

func TestExport_CSV(t *testing.T) {
	var buf bytes.Buffer
	if err := Export(exportEntries(), FormatCSV, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header+data), got %d", len(records))
	}
	if records[0][0] != "timestamp" {
		t.Errorf("expected header 'timestamp', got %q", records[0][0])
	}
	if records[1][1] != "2" {
		t.Errorf("expected total=2, got %q", records[1][1])
	}
	if records[1][2] != "1" {
		t.Errorf("expected drifted=1, got %q", records[1][2])
	}
	if records[1][3] != "1" {
		t.Errorf("expected clean=1, got %q", records[1][3])
	}
}

func TestExport_JSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Export(exportEntries(), FormatJSON, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out []Entry
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("parse json: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if len(out[0].Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(out[0].Results))
	}
}

func TestExport_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Export(exportEntries(), ExportFormat("xml"), &buf)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("unexpected error message: %v", err)
	}
}
