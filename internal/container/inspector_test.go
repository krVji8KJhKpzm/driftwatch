package container

import (
	"context"
	"errors"
	"testing"
)

type mockRunner struct {
	output []byte
	err    error
}

func (m mockRunner) Run(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return m.output, m.err
}

const sampleInspect = `[{
  "Id": "abc123",
  "Name": "/myapp",
  "Config": {
    "Image": "myapp:latest",
    "Env": ["PORT=8080", "DEBUG=true"],
    "Labels": {"version": "1.2.3"}
  }
}]`

func TestInspect_Valid(t *testing.T) {
	insp := &Inspector{Runner: mockRunner{output: []byte(sampleInspect)}}
	info, err := insp.Inspect(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Image != "myapp:latest" {
		t.Errorf("expected image myapp:latest, got %q", info.Image)
	}
	if info.Env["PORT"] != "8080" {
		t.Errorf("expected PORT=8080, got %q", info.Env["PORT"])
	}
	if info.Labels["version"] != "1.2.3" {
		t.Errorf("expected label version=1.2.3")
	}
}

func TestInspect_RunnerError(t *testing.T) {
	insp := &Inspector{Runner: mockRunner{err: errors.New("exec failed")}}
	_, err := insp.Inspect(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInspect_EmptyResult(t *testing.T) {
	insp := &Inspector{Runner: mockRunner{output: []byte("[]")}}
	_, err := insp.Inspect(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error for empty result")
	}
}

func TestParseEnv(t *testing.T) {
	raw := []string{"KEY=value", "EMPTY=", "NOEQUALS"}
	m := parseEnv(raw)
	if m["KEY"] != "value" {
		t.Errorf("KEY mismatch")
	}
	if m["EMPTY"] != "" {
		t.Errorf("EMPTY mismatch")
	}
	if _, ok := m["NOEQUALS"]; ok {
		t.Errorf("NOEQUALS should not be parsed")
	}
}
