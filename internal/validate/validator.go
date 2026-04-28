// Package validate provides manifest validation against live container state,
// checking that required fields are present and conform to expected formats
// before drift detection is performed.
package validate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Severity indicates how serious a validation finding is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Finding represents a single validation issue found in a drift result.
type Finding struct {
	Container string   `json:"container"`
	Field     string   `json:"field"`
	Message   string   `json:"message"`
	Severity  Severity `json:"severity"`
}

// Report holds all findings produced by a validation run.
type Report struct {
	Findings []Finding `json:"findings"`
	Valid    bool      `json:"valid"`
}

var imageRefPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-/:@]*$`)

// Validate inspects each drift result for structural and semantic issues.
// It returns a Report containing any warnings or errors found, and a non-nil
// error only if validation itself could not be performed.
func Validate(results []drift.Result) (Report, error) {
	if results == nil {
		return Report{}, errors.New("validate: nil results slice")
	}

	var findings []Finding

	for _, r := range results {
		if strings.TrimSpace(r.Container) == "" {
			findings = append(findings, Finding{
				Container: "(unknown)",
				Field:     "container",
				Message:   "container name is empty",
				Severity:  SeverityError,
			})
			continue
		}

		if strings.TrimSpace(r.ExpectedImage) == "" {
			findings = append(findings, Finding{
				Container: r.Container,
				Field:     "expected_image",
				Message:   "expected image is empty; manifest may be missing an image field",
				Severity:  SeverityError,
			})
		} else if !imageRefPattern.MatchString(r.ExpectedImage) {
			findings = append(findings, Finding{
				Container: r.Container,
				Field:     "expected_image",
				Message:   fmt.Sprintf("expected image %q does not look like a valid image reference", r.ExpectedImage),
				Severity:  SeverityWarning,
			})
		}

		if strings.TrimSpace(r.ActualImage) == "" {
			findings = append(findings, Finding{
				Container: r.Container,
				Field:     "actual_image",
				Message:   "actual image is empty; container may not be running",
				Severity:  SeverityWarning,
			})
		}

		for _, d := range r.Diffs {
			if strings.TrimSpace(d.Field) == "" {
				findings = append(findings, Finding{
					Container: r.Container,
					Field:     "diff.field",
					Message:   "a diff entry has an empty field name",
					Severity:  SeverityError,
				})
			}
		}
	}

	return Report{
		Findings: findings,
		Valid:    !hasErrors(findings),
	}, nil
}

// hasErrors returns true if any finding has error severity.
func hasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}
