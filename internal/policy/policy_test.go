package policy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/driftwatch/internal/policy"
)

func writeTempPolicy(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempPolicy: %v", err)
	}
	return p
}

func TestLoadPolicy_Valid(t *testing.T) {
	path := writeTempPolicy(t, `
rules:
  web:
    name: web
    ignore_envs: [SECRET_KEY, DB_PASS]
    ignore_image: false
  worker:
    name: worker
    ignore_image: true
`)
	p, err := policy.LoadPolicy(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(p.Rules))
	}
}

func TestLoadPolicy_NotFound(t *testing.T) {
	_, err := policy.LoadPolicy("/nonexistent/policy.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestShouldIgnoreEnv(t *testing.T) {
	path := writeTempPolicy(t, `
rules:
  web:
    name: web
    ignore_envs: [SECRET_KEY]
`)
	p, _ := policy.LoadPolicy(path)
	if !p.ShouldIgnoreEnv("web", "SECRET_KEY") {
		t.Error("expected SECRET_KEY to be ignored for web")
	}
	if p.ShouldIgnoreEnv("web", "PORT") {
		t.Error("expected PORT not to be ignored")
	}
	if p.ShouldIgnoreEnv("unknown", "SECRET_KEY") {
		t.Error("expected no rule for unknown container")
	}
}

func TestRuleFor_Missing(t *testing.T) {
	p := &policy.Policy{Rules: map[string]policy.Rule{}}
	r := p.RuleFor("missing")
	if r.Name != "" || r.IgnoreImage {
		t.Error("expected empty rule for missing container")
	}
}
