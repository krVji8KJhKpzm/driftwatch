package stale_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/stale"
)

func makeResult(name string, drifted bool) drift.Result {
	return drift.Result{Name: name, Drifted: drifted}
}

func TestEvaluate_NoDrift(t *testing.T) {
	results := []drift.Result{makeResult("web", false)}
	report := stale.Evaluate(results, nil, time.Hour)

	if report.StaleCount() != 0 {
		t.Fatalf("expected 0 stale, got %d", report.StaleCount())
	}
	if report.Entries[0].Stale {
		t.Error("non-drifted container should not be stale")
	}
}

func TestEvaluate_DriftedBelowThreshold(t *testing.T) {
	results := []drift.Result{makeResult("api", true)}
	firstSeen := map[string]time.Time{
		"api": time.Now().Add(-30 * time.Minute),
	}
	report := stale.Evaluate(results, firstSeen, time.Hour)

	if report.StaleCount() != 0 {
		t.Fatalf("expected 0 stale, got %d", report.StaleCount())
	}
}

func TestEvaluate_DriftedAboveThreshold(t *testing.T) {
	results := []drift.Result{makeResult("api", true)}
	firstSeen := map[string]time.Time{
		"api": time.Now().Add(-2 * time.Hour),
	}
	report := stale.Evaluate(results, firstSeen, time.Hour)

	if report.StaleCount() != 1 {
		t.Fatalf("expected 1 stale, got %d", report.StaleCount())
	}
	if !report.Entries[0].Stale {
		t.Error("expected container to be stale")
	}
}

func TestEvaluate_MissingFirstSeen(t *testing.T) {
	results := []drift.Result{makeResult("svc", true)}
	report := stale.Evaluate(results, map[string]time.Time{}, time.Minute)

	if report.StaleCount() != 0 {
		t.Fatalf("expected 0 stale when firstSeen missing, got %d", report.StaleCount())
	}
}

func TestEvaluate_ZeroThresholdNeverStale(t *testing.T) {
	results := []drift.Result{makeResult("db", true)}
	firstSeen := map[string]time.Time{
		"db": time.Now().Add(-72 * time.Hour),
	}
	report := stale.Evaluate(results, firstSeen, 0)

	if report.StaleCount() != 0 {
		t.Fatalf("zero threshold should never mark stale, got %d", report.StaleCount())
	}
}

func TestReport_ThresholdPreserved(t *testing.T) {
	thresh := 45 * time.Minute
	report := stale.Evaluate(nil, nil, thresh)
	if report.Threshold != thresh {
		t.Errorf("expected threshold %v, got %v", thresh, report.Threshold)
	}
}
