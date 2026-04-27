package redact

import (
	"regexp"
	"strings"
)

// Config holds patterns and key names whose values should be redacted.
type Config struct {
	Keys     []string `json:"keys"`
	Patterns []string `json:"patterns"`
}

// Redactor applies redaction rules to drift field values.
type Redactor struct {
	cfg      *Config
	compiled []*regexp.Regexp
}

const redactedPlaceholder = "[REDACTED]"

// New creates a Redactor from the given config. Returns an error if any
// pattern fails to compile.
func New(cfg *Config) (*Redactor, error) {
	if cfg == nil {
		return &Redactor{}, nil
	}
	var compiled []*regexp.Regexp
	for _, p := range cfg.Patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return &Redactor{cfg: cfg, compiled: compiled}, nil
}

// ShouldRedact returns true if the given field key matches a redaction rule.
func (r *Redactor) ShouldRedact(key string) bool {
	if r.cfg == nil {
		return false
	}
	lower := strings.ToLower(key)
	for _, k := range r.cfg.Keys {
		if strings.ToLower(k) == lower {
			return true
		}
	}
	for _, re := range r.compiled {
		if re.MatchString(key) {
			return true
		}
	}
	return false
}

// Redact replaces the value with the redacted placeholder if the key matches
// any rule; otherwise it returns the original value unchanged.
func (r *Redactor) Redact(key, value string) string {
	if r.ShouldRedact(key) {
		return redactedPlaceholder
	}
	return value
}

// RedactMap returns a new map with sensitive values replaced.
func (r *Redactor) RedactMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = r.Redact(k, v)
	}
	return out
}
