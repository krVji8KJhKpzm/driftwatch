package compare_test

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/compare"
	"github.com/yourorg/driftwatch/internal/drift"
)

// makeContainerResult builds a drift.Result for testing.
func makeContainerResult(name string, drifted bool, diffs []drift.Diff) drift.Result {
	return drift.Result{
		ContainerName: name,
		Drifted:       drifted,
		Diffs:         diffs,
	}
}

func makeDiff(field, expected, actual string) drift.Diff {
	return drift.Diff{
		Field:    field,
		Expected: expected,
		Actual:   actual,
	}
}

func TestBuild_NoDrift(t *testing.T) {
	results := []drift.Result{
		makeContainerResult("web", false, nil),
		makeContainerResult("db", false, nil),
	}

	report := compare.Build(results, results)

	if len(report.Added) != 0 {
		t.Errorf("expected 0 added, got %d", len(report.Added))
	}
	if len(report.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(report.Removed))
	}
	if len(report.Changed) != 0 {
		t.Errorf("expected 0 changed, got %d", len(report.Changed))
	}
}

func TestBuild_AddedContainer(t *testing.T) {
	baseline := []drift.Result{
		makeContainerResult("web", false, nil),
	}
	current := []drift.Result{
		makeContainerResult("web", false, nil),
		makeContainerResult("worker", true, []drift.Diff{makeDiff("image", "app:1.0", "app:2.0")}),
	}

	report := compare.Build(baseline, current)

	if len(report.Added) != 1 {
		t.Fatalf("expected 1 added, got %d", len(report.Added))
	}
	if report.Added[0].ContainerName != "worker" {
		t.Errorf("expected added container 'worker', got %q", report.Added[0].ContainerName)
	}
}

func TestBuild_RemovedContainer(t *testing.T) {
	baseline := []drift.Result{
		makeContainerResult("web", false, nil),
		makeContainerResult("cache", false, nil),
	}
	current := []drift.Result{
		makeContainerResult("web", false, nil),
	}

	report := compare.Build(baseline, current)

	if len(report.Removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(report.Removed))
	}
	if report.Removed[0].ContainerName != "cache" {
		t.Errorf("expected removed container 'cache', got %q", report.Removed[0].ContainerName)
	}
}

func TestBuild_ChangedDrift(t *testing.T) {
	baseline := []drift.Result{
		makeContainerResult("web", true, []drift.Diff{
			makeDiff("image", "app:1.0", "app:1.1"),
		}),
	}
	current := []drift.Result{
		makeContainerResult("web", true, []drift.Diff{
			makeDiff("image", "app:1.0", "app:2.0"),
			makeDiff("env.PORT", "8080", "9090"),
		}),
	}

	report := compare.Build(baseline, current)

	if len(report.Changed) != 1 {
		t.Fatalf("expected 1 changed, got %d", len(report.Changed))
	}
	entry := report.Changed[0]
	if entry.ContainerName != "web" {
		t.Errorf("expected changed container 'web', got %q", entry.ContainerName)
	}
	if len(entry.NewDiffs) != 2 {
		t.Errorf("expected 2 new diffs, got %d", len(entry.NewDiffs))
	}
}

func TestBuild_ParseFieldDiff_EnvKey(t *testing.T) {
	baseline := []drift.Result{
		makeContainerResult("api", true, []drift.Diff{
			makeDiff("env.DEBUG", "false", "true"),
		}),
	}
	current := []drift.Result{
		makeContainerResult("api", true, []drift.Diff{
			makeDiff("env.DEBUG", "false", "true"),
			makeDiff("env.LOG_LEVEL", "info", "debug"),
		}),
	}

	report := compare.Build(baseline, current)

	if len(report.Changed) != 1 {
		t.Fatalf("expected 1 changed entry, got %d", len(report.Changed))
	}
	found := false
	for _, d := range report.Changed[0].NewDiffs {
		if d.Field == "env.LOG_LEVEL" {
			found = true
		}
	}
	if !found {
		t.Error("expected new diff for env.LOG_LEVEL")
	}
}

func TestBuild_DriftResolved(t *testing.T) {
	baseline := []drift.Result{
		makeContainerResult("web", true, []drift.Diff{
			makeDiff("image", "app:1.0", "app:1.1"),
		}),
	}
	current := []drift.Result{
		makeContainerResult("web", false, nil),
	}

	report := compare.Build(baseline, current)

	// A container that was drifted and is now clean should appear in Changed
	// with no new diffs, indicating resolution.
	if len(report.Changed) != 1 {
		t.Fatalf("expected 1 changed (resolved) entry, got %d", len(report.Changed))
	}
	if report.Changed[0].Resolved != true {
		t.Error("expected Resolved to be true for previously drifted container")
	}
}
