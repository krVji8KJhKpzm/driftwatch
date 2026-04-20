package filter_test

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/filter"
)

func makeResult(name string, drifted bool) drift.Result {
	return drift.Result{Name: name, Drifted: drifted}
}

func TestApply_NoFilter(t *testing.T) {
	input := []drift.Result{makeResult("a", true), makeResult("b", false)}
	out := filter.Apply(input, filter.Options{})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestApply_OnlyDrifted(t *testing.T) {
	input := []drift.Result{makeResult("a", true), makeResult("b", false)}
	out := filter.Apply(input, filter.Options{OnlyDrifted: true})
	if len(out) != 1 || out[0].Name != "a" {
		t.Fatalf("expected only drifted result 'a', got %+v", out)
	}
}

func TestApply_ByName(t *testing.T) {
	input := []drift.Result{makeResult("a", true), makeResult("b", true), makeResult("c", false)}
	out := filter.Apply(input, filter.Options{Names: []string{"b", "c"}})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestApply_OnlyDriftedAndName(t *testing.T) {
	input := []drift.Result{makeResult("a", true), makeResult("b", false), makeResult("c", true)}
	out := filter.Apply(input, filter.Options{OnlyDrifted: true, Names: []string{"b", "c"}})
	if len(out) != 1 || out[0].Name != "c" {
		t.Fatalf("expected only 'c', got %+v", out)
	}
}

func TestApply_EmptyInput(t *testing.T) {
	out := filter.Apply(nil, filter.Options{OnlyDrifted: true})
	if len(out) != 0 {
		t.Fatalf("expected empty result, got %+v", out)
	}
}
