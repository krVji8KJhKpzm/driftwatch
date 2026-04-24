package history

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/driftwatch/internal/drift"
)

func makeEntry(ts time.Time, results []drift.Result) Entry {
	return Entry{Timestamp: ts, Results: results}
}

func driftResult(name string, drifted bool) drift.Result {
	return drift.Result{Name: name, Drifted: drifted}
}

func TestTrend_Empty(t *testing.T) {
	trend := Trend(nil)
	if len(trend) != 0 {
		t.Fatalf("expected empty trend, got %d entries", len(trend))
	}
}

func TestTrend_SingleEntry(t *testing.T) {
	entries := []Entry{
		makeEntry(time.Now(), []drift.Result{
			driftResult("web", true),
			driftResult("db", false),
		}),
	}
	trend := Trend(entries)
	if len(trend) != 1 {
		t.Fatalf("expected 1 trend entry, got %d", len(trend))
	}
	if trend[0].Total != 2 {
		t.Errorf("expected Total=2, got %d", trend[0].Total)
	}
	if trend[0].Drifted != 1 {
		t.Errorf("expected Drifted=1, got %d", trend[0].Drifted)
	}
	if trend[0].NewDrifts != 0 || trend[0].Resolved != 0 {
		t.Errorf("first entry should have no new/resolved deltas")
	}
}

func TestTrend_DriftIncreases(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeEntry(now, []drift.Result{driftResult("a", true)}),
		makeEntry(now.Add(time.Minute), []drift.Result{
			driftResult("a", true),
			driftResult("b", true),
		}),
	}
	trend := Trend(entries)
	if trend[1].NewDrifts != 1 {
		t.Errorf("expected NewDrifts=1, got %d", trend[1].NewDrifts)
	}
	if trend[1].Resolved != 0 {
		t.Errorf("expected Resolved=0, got %d", trend[1].Resolved)
	}
}

func TestTrend_DriftDecreases(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeEntry(now, []drift.Result{driftResult("a", true), driftResult("b", true)}),
		makeEntry(now.Add(time.Minute), []drift.Result{driftResult("a", false), driftResult("b", false)}),
	}
	trend := Trend(entries)
	if trend[1].Resolved != 2 {
		t.Errorf("expected Resolved=2, got %d", trend[1].Resolved)
	}
}

func TestWriteTrend_Output(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeEntry(now, []drift.Result{driftResult("web", true)}),
	}
	var buf bytes.Buffer
	if err := WriteTrend(&buf, entries); err != nil {
		t.Fatalf("WriteTrend error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "TIMESTAMP") {
		t.Errorf("expected header in output, got: %s", out)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("expected drifted count in output, got: %s", out)
	}
}

func TestWriteTrend_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTrend(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no history") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
