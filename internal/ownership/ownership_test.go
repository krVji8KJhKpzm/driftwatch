package ownership

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func writeTempOwnership(t *testing.T, cfg Config) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	p := filepath.Join(t.TempDir(), "ownership.json")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func makeResult(name string, drifted bool) drift.Result {
	return drift.Result{ContainerName: name, Drifted: drifted}
}

func TestLoadConfig_Valid(t *testing.T) {
	cfg := Config{Rules: []Rule{
		{Match: "web", Owner: Owner{Name: "Alice", Team: "platform"}},
	}}
	p := writeTempOwnership(t, cfg)
	got, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got.Rules))
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/ownership.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestAssign_NilConfig(t *testing.T) {
	results := []drift.Result{makeResult("api", true)}
	assignments := Assign(results, nil)
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
	if assignments[0].Owner != nil {
		t.Errorf("expected nil owner with nil config")
	}
}

func TestAssign_ExactMatch(t *testing.T) {
	cfg := &Config{Rules: []Rule{
		{Match: "api", Owner: Owner{Name: "Bob", Team: "backend"}},
	}}
	assignments := Assign([]drift.Result{makeResult("api", true)}, cfg)
	if assignments[0].Owner == nil {
		t.Fatal("expected owner to be resolved")
	}
	if assignments[0].Owner.Name != "Bob" {
		t.Errorf("expected Bob, got %s", assignments[0].Owner.Name)
	}
}

func TestAssign_PrefixMatch(t *testing.T) {
	cfg := &Config{Rules: []Rule{
		{Match: "worker-", Owner: Owner{Name: "Carol", Team: "infra"}},
	}}
	assignments := Assign([]drift.Result{makeResult("worker-1", false)}, cfg)
	if assignments[0].Owner == nil {
		t.Fatal("expected prefix owner")
	}
	if assignments[0].Owner.Team != "infra" {
		t.Errorf("expected infra, got %s", assignments[0].Owner.Team)
	}
}

func TestAssign_NoMatch(t *testing.T) {
	cfg := &Config{Rules: []Rule{
		{Match: "web", Owner: Owner{Name: "Alice"}},
	}}
	assignments := Assign([]drift.Result{makeResult("db", true)}, cfg)
	if assignments[0].Owner != nil {
		t.Errorf("expected nil owner for unmatched container")
	}
}

func TestAssign_SortedByName(t *testing.T) {
	cfg := &Config{}
	results := []drift.Result{
		makeResult("zebra", false),
		makeResult("alpha", true),
		makeResult("mango", false),
	}
	assignments := Assign(results, cfg)
	names := []string{assignments[0].ContainerName, assignments[1].ContainerName, assignments[2].ContainerName}
	if names[0] != "alpha" || names[1] != "mango" || names[2] != "zebra" {
		t.Errorf("unexpected order: %v", names)
	}
}
