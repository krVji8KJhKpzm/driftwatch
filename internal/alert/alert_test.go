package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResult(name string, imageDrift bool, envDiff []drift.EnvDiff) drift.Result {
	return drift.Result{
		Name:       name,
		Drifted:    imageDrift || len(envDiff) > 0,
		ImageDrift: imageDrift,
		EnvDiff:    envDiff,
	}
}

func TestEvaluate_NoAlerts(t *testing.T) {
	results := []drift.Result{makeResult("web", false, nil)}
	rules := []Rule{{OnImageDrift: true, Level: LevelWarn}}
	alerts := Evaluate(results, rules)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestEvaluate_ImageDrift(t *testing.T) {
	results := []drift.Result{makeResult("web", true, nil)}
	rules := []Rule{{OnImageDrift: true, Level: LevelError}}
	alerts := Evaluate(results, rules)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != LevelError {
		t.Errorf("expected error level, got %s", alerts[0].Level)
	}
	if !strings.Contains(alerts[0].Reason, "image drift") {
		t.Errorf("unexpected reason: %s", alerts[0].Reason)
	}
}

func TestEvaluate_EnvDrift(t *testing.T) {
	envDiff := []drift.EnvDiff{{Key: "PORT", Expected: "8080", Actual: "9090"}}
	results := []drift.Result{makeResult("api", false, envDiff)}
	rules := []Rule{{OnEnvDrift: true, Level: LevelWarn}}
	alerts := Evaluate(results, rules)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Container != "api" {
		t.Errorf("expected container api, got %s", alerts[0].Container)
	}
}

func TestEvaluate_NoDriftSkipped(t *testing.T) {
	results := []drift.Result{makeResult("db", false, nil)}
	rules := []Rule{{OnImageDrift: true, OnEnvDrift: true, Level: LevelError}}
	alerts := Evaluate(results, rules)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts for non-drifted container")
	}
}

func TestWrite_NoAlerts(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no alerts") {
		t.Errorf("expected no-alerts message, got: %s", buf.String())
	}
}

func TestWrite_ErrorLevelReturnsErr(t *testing.T) {
	var buf bytes.Buffer
	alerts := []Alert{{Container: "web", Level: LevelError, Reason: "image drift detected"}}
	err := Write(&buf, alerts)
	if err == nil {
		t.Fatal("expected error for error-level alert")
	}
	if !strings.Contains(buf.String(), "[ERROR]") {
		t.Errorf("expected [ERROR] in output, got: %s", buf.String())
	}
}

func TestWrite_WarnLevelNoErr(t *testing.T) {
	var buf bytes.Buffer
	alerts := []Alert{{Container: "api", Level: LevelWarn, Reason: "env drift"}}
	err := Write(&buf, alerts)
	if err != nil {
		t.Fatalf("unexpected error for warn-level alert: %v", err)
	}
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Errorf("expected [WARN] in output, got: %s", buf.String())
	}
}
