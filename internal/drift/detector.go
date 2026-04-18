package drift

import (
	"fmt"

	"github.com/user/driftwatch/internal/container"
	"github.com/user/driftwatch/internal/manifest"
)

// DriftType categorises what kind of drift was detected.
type DriftType string

const (
	DriftImage  DriftType = "image"
	DriftEnv    DriftType = "env"
	DriftLabel  DriftType = "label"
)

// Finding represents a single drift item.
type Finding struct {
	Type     DriftType
	Key      string
	Expected string
	Actual   string
}

func (f Finding) String() string {
	return fmt.Sprintf("[%s] %s: expected=%q actual=%q", f.Type, f.Key, f.Expected, f.Actual)
}

// Detect compares a manifest entry against live container info and returns findings.
func Detect(entry manifest.Entry, info *container.ContainerInfo) []Finding {
	var findings []Finding

	if entry.Image != "" && entry.Image != info.Image {
		findings = append(findings, Finding{
			Type:     DriftImage,
			Key:      "image",
			Expected: entry.Image,
			Actual:   info.Image,
		})
	}

	for k, expected := range entry.Env {
		actual, ok := info.Env[k]
		if !ok || actual != expected {
			findings = append(findings, Finding{
				Type:     DriftEnv,
				Key:      k,
				Expected: expected,
				Actual:   actual,
			})
		}
	}

	for k, expected := range entry.Labels {
		actual, ok := info.Labels[k]
		if !ok || actual != expected {
			findings = append(findings, Finding{
				Type:     DriftLabel,
				Key:      k,
				Expected: expected,
				Actual:   actual,
			})
		}
	}

	return findings
}
