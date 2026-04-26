package lint

import (
	"fmt"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Severity represents the level of a lint finding.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Finding describes a single lint issue found in a drift result.
type Finding struct {
	Container string
	Field     string
	Message   string
	Severity  Severity
}

// Report holds all findings produced by a lint run.
type Report struct {
	Findings []Finding
	ErrorCount   int
	WarningCount int
}

// HasIssues returns true if any findings exist.
func (r *Report) HasIssues() bool {
	return len(r.Findings) > 0
}

// Run inspects drift results and returns a lint report.
func Run(results []drift.Result) Report {
	var report Report
	for _, res := range results {
		if !res.Drifted {
			continue
		}
		for _, d := range res.Diffs {
			f := classify(res.Name, d)
			report.Findings = append(report.Findings, f)
			switch f.Severity {
			case SeverityError:
				report.ErrorCount++
			case SeverityWarning:
				report.WarningCount++
			}
		}
	}
	return report
}

func classify(container string, d drift.Diff) Finding {
	field := strings.ToLower(d.Field)
	switch {
	case field == "image":
		return Finding{
			Container: container,
			Field:     d.Field,
			Message:   fmt.Sprintf("image changed: expected %q, got %q", d.Expected, d.Actual),
			Severity:  SeverityError,
		}
	case strings.HasPrefix(field, "env:"):
		return Finding{
			Container: container,
			Field:     d.Field,
			Message:   fmt.Sprintf("env var %s changed: expected %q, got %q", d.Field, d.Expected, d.Actual),
			Severity:  SeverityWarning,
		}
	default:
		return Finding{
			Container: container,
			Field:     d.Field,
			Message:   fmt.Sprintf("field %s changed: expected %q, got %q", d.Field, d.Expected, d.Actual),
			Severity:  SeverityInfo,
		}
	}
}
