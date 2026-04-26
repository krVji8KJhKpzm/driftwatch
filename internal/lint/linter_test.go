package lint

import (
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResult(name string, drifted bool, diffs []drift.Diff) drift.Result {
	return drift.Result{Name: name, Drifted: drifted, Diffs: diffs}
}

func TestRun_NoDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("web", false, nil),
	}
	report := Run(results)
	if report.HasIssues() {
		t.Fatalf("expected no findings, got %d", len(report.Findings))
	}
	if report.ErrorCount != 0 || report.WarningCount != 0 {
		t.Errorf("expected zero counts, got errors=%d warnings=%d", report.ErrorCount, report.WarningCount)
	}
}

func TestRun_ImageDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("api", true, []drift.Diff{
			{Field: "image", Expected: "nginx:1.24", Actual: "nginx:1.25"},
		}),
	}
	report := Run(results)
	if len(report.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(report.Findings))
	}
	f := report.Findings[0]
	if f.Severity != SeverityError {
		t.Errorf("expected error severity, got %s", f.Severity)
	}
	if f.Container != "api" {
		t.Errorf("expected container 'api', got %s", f.Container)
	}
	if report.ErrorCount != 1 {
		t.Errorf("expected ErrorCount=1, got %d", report.ErrorCount)
	}
}

func TestRun_EnvDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("worker", true, []drift.Diff{
			{Field: "env:LOG_LEVEL", Expected: "info", Actual: "debug"},
		}),
	}
	report := Run(results)
	if len(report.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(report.Findings))
	}
	if report.Findings[0].Severity != SeverityWarning {
		t.Errorf("expected warning severity, got %s", report.Findings[0].Severity)
	}
	if report.WarningCount != 1 {
		t.Errorf("expected WarningCount=1, got %d", report.WarningCount)
	}
}

func TestRun_UnknownField(t *testing.T) {
	results := []drift.Result{
		makeResult("db", true, []drift.Diff{
			{Field: "labels:team", Expected: "platform", Actual: "infra"},
		}),
	}
	report := Run(results)
	if len(report.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(report.Findings))
	}
	if report.Findings[0].Severity != SeverityInfo {
		t.Errorf("expected info severity, got %s", report.Findings[0].Severity)
	}
}

func TestRun_SkipsNonDrifted(t *testing.T) {
	results := []drift.Result{
		makeResult("clean", false, []drift.Diff{
			{Field: "image", Expected: "redis:7", Actual: "redis:8"},
		}),
		makeResult("drifted", true, []drift.Diff{
			{Field: "image", Expected: "redis:7", Actual: "redis:8"},
		}),
	}
	report := Run(results)
	if len(report.Findings) != 1 {
		t.Errorf("expected 1 finding (only drifted), got %d", len(report.Findings))
	}
}
