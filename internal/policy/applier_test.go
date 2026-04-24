package policy_test

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/policy"
)

func makePolicy(rules map[string]policy.Rule) *policy.Policy {
	return &policy.Policy{Rules: rules}
}

func TestApply_NilPolicy(t *testing.T) {
	results := []drift.Result{{Name: "web", Drifted: true, ImageDrift: true}}
	out := policy.Apply(results, nil)
	if !out[0].Drifted {
		t.Error("nil policy should not suppress drift")
	}
}

func TestApply_IgnoreImage(t *testing.T) {
	p := makePolicy(map[string]policy.Rule{
		"web": {Name: "web", IgnoreImage: true},
	})
	results := []drift.Result{{Name: "web", Drifted: true, ImageDrift: true}}
	out := policy.Apply(results, p)
	if out[0].ImageDrift {
		t.Error("expected image drift to be suppressed")
	}
	if out[0].Drifted {
		t.Error("expected drifted to be false after suppression")
	}
}

func TestApply_IgnoreEnv(t *testing.T) {
	p := makePolicy(map[string]policy.Rule{
		"api": {Name: "api", IgnoreEnvs: []string{"SECRET"}},
	})
	results := []drift.Result{{
		Name:    "api",
		Drifted: true,
		EnvDrifts: []drift.EnvDrift{
			{Key: "SECRET", Expected: "x", Actual: "y"},
			{Key: "PORT", Expected: "8080", Actual: "9090"},
		},
	}}
	out := policy.Apply(results, p)
	if len(out[0].EnvDrifts) != 1 {
		t.Fatalf("expected 1 env drift, got %d", len(out[0].EnvDrifts))
	}
	if out[0].EnvDrifts[0].Key != "PORT" {
		t.Errorf("expected PORT drift to remain")
	}
}

func TestApply_NoMatchingRule(t *testing.T) {
	p := makePolicy(map[string]policy.Rule{})
	results := []drift.Result{{Name: "db", Drifted: true, ImageDrift: true}}
	out := policy.Apply(results, p)
	if !out[0].ImageDrift {
		t.Error("expected image drift to remain when no rule matches")
	}
}
