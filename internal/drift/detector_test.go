package drift

import (
	"testing"

	"github.com/user/driftwatch/internal/container"
	"github.com/user/driftwatch/internal/manifest"
)

func baseEntry() manifest.Entry {
	return manifest.Entry{
		Name:   "myapp",
		Image:  "myapp:1.0",
		Env:    map[string]string{"PORT": "8080"},
		Labels: map[string]string{"team": "platform"},
	}
}

func baseInfo() *container.ContainerInfo {
	return &container.ContainerInfo{
		Image:  "myapp:1.0",
		Env:    map[string]string{"PORT": "8080"},
		Labels: map[string]string{"team": "platform"},
	}
}

func TestDetect_NoDrift(t *testing.T) {
	findings := Detect(baseEntry(), baseInfo())
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d: %v", len(findings), findings)
	}
}

func TestDetect_ImageDrift(t *testing.T) {
	info := baseInfo()
	info.Image = "myapp:2.0"
	findings := Detect(baseEntry(), info)
	if len(findings) != 1 || findings[0].Type != DriftImage {
		t.Errorf("expected image drift, got %v", findings)
	}
}

func TestDetect_EnvDrift(t *testing.T) {
	info := baseInfo()
	info.Env["PORT"] = "9090"
	findings := Detect(baseEntry(), info)
	if len(findings) != 1 || findings[0].Type != DriftEnv {
		t.Errorf("expected env drift, got %v", findings)
	}
}

func TestDetect_LabelDrift(t *testing.T) {
	info := baseInfo()
	delete(info.Labels, "team")
	findings := Detect(baseEntry(), info)
	if len(findings) != 1 || findings[0].Type != DriftLabel {
		t.Errorf("expected label drift, got %v", findings)
	}
}

func TestDetect_MultipleDrifts(t *testing.T) {
	info := baseInfo()
	info.Image = "myapp:9.9"
	info.Env["PORT"] = "1111"
	findings := Detect(baseEntry(), info)
	if len(findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(findings))
	}
}
