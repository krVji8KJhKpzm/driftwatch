package alert

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Rule defines when an alert should be triggered.
type Rule struct {
	OnImageDrift bool   `json:"on_image_drift" yaml:"on_image_drift"`
	OnEnvDrift   bool   `json:"on_env_drift"   yaml:"on_env_drift"`
	MinDrifted   int    `json:"min_drifted"    yaml:"min_drifted"`
	Level        Level  `json:"level"          yaml:"level"`
	Message      string `json:"message"        yaml:"message"`
}

// Alert is a triggered alert for a container.
type Alert struct {
	Container string
	Level     Level
	Reason    string
}

// Evaluate checks drift results against rules and returns triggered alerts.
func Evaluate(results []drift.Result, rules []Rule) []Alert {
	var alerts []Alert
	for _, r := range results {
		if !r.Drifted {
			continue
		}
		for _, rule := range rules {
			if triggered, reason := matches(r, rule); triggered {
				alerts = append(alerts, Alert{
					Container: r.Name,
					Level:     rule.Level,
					Reason:    reason,
				})
				break
			}
		}
	}
	return alerts
}

// Write prints alerts to w; returns non-nil error if any alert is LevelError.
func Write(w io.Writer, alerts []Alert) error {
	if len(alerts) == 0 {
		fmt.Fprintln(w, "no alerts triggered")
		return nil
	}
	var hasError bool
	for _, a := range alerts {
		fmt.Fprintf(w, "[%s] %s: %s\n", strings.ToUpper(string(a.Level)), a.Container, a.Reason)
		if a.Level == LevelError {
			hasError = true
		}
	}
	if hasError {
		return fmt.Errorf("one or more error-level alerts triggered")
	}
	return nil
}

// WriteToStdout is a convenience wrapper around Write using os.Stdout.
func WriteToStdout(alerts []Alert) error {
	return Write(os.Stdout, alerts)
}

func matches(r drift.Result, rule Rule) (bool, string) {
	var reasons []string
	if rule.OnImageDrift && r.ImageDrift {
		reasons = append(reasons, "image drift detected")
	}
	if rule.OnEnvDrift && len(r.EnvDiff) > 0 {
		reasons = append(reasons, fmt.Sprintf("%d env var(s) drifted", len(r.EnvDiff)))
	}
	if len(reasons) == 0 {
		return false, ""
	}
	return true, strings.Join(reasons, "; ")
}
