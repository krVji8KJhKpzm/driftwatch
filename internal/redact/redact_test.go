package redact

import (
	"testing"
)

func TestNew_NilConfig(t *testing.T) {
	r, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ShouldRedact("PASSWORD") {
		t.Error("expected no redaction with nil config")
	}
}

func TestNew_InvalidPattern(t *testing.T) {
	cfg := &Config{Patterns: []string{"[invalid"}}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestShouldRedact_ByKey(t *testing.T) {
	cfg := &Config{Keys: []string{"PASSWORD", "SECRET"}}
	r, _ := New(cfg)

	if !r.ShouldRedact("password") {
		t.Error("expected case-insensitive match on 'password'")
	}
	if !r.ShouldRedact("SECRET") {
		t.Error("expected match on 'SECRET'")
	}
	if r.ShouldRedact("HOST") {
		t.Error("expected no match on 'HOST'")
	}
}

func TestShouldRedact_ByPattern(t *testing.T) {
	cfg := &Config{Patterns: []string{"(?i)token", "(?i)api_key"}}
	r, _ := New(cfg)

	if !r.ShouldRedact("AUTH_TOKEN") {
		t.Error("expected pattern match on 'AUTH_TOKEN'")
	}
	if !r.ShouldRedact("api_key") {
		t.Error("expected pattern match on 'api_key'")
	}
	if r.ShouldRedact("PORT") {
		t.Error("expected no match on 'PORT'")
	}
}

func TestRedact_ReplacesValue(t *testing.T) {
	cfg := &Config{Keys: []string{"DB_PASS"}}
	r, _ := New(cfg)

	got := r.Redact("DB_PASS", "supersecret")
	if got != redactedPlaceholder {
		t.Errorf("expected %q, got %q", redactedPlaceholder, got)
	}

	got = r.Redact("DB_HOST", "localhost")
	if got != "localhost" {
		t.Errorf("expected 'localhost', got %q", got)
	}
}

func TestRedactMap(t *testing.T) {
	cfg := &Config{Keys: []string{"SECRET"}}
	r, _ := New(cfg)

	input := map[string]string{
		"SECRET": "abc123",
		"HOST":   "example.com",
	}
	out := r.RedactMap(input)

	if out["SECRET"] != redactedPlaceholder {
		t.Errorf("expected SECRET to be redacted, got %q", out["SECRET"])
	}
	if out["HOST"] != "example.com" {
		t.Errorf("expected HOST unchanged, got %q", out["HOST"])
	}
	if input["SECRET"] != "abc123" {
		t.Error("original map should not be mutated")
	}
}
