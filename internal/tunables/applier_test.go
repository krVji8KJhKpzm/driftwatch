package tunables

import (
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeResult(name string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Name:    name,
		Drifted: len(diffs) > 0,
		Diffs:   diffs,
	}
}

func envDiff(key, want, got string) drift.Diff {
	return drift.Diff{Field: "env", Key: key, Expected: want, Actual: got}
}

func imageDiff(want, got string) drift.Diff {
	return drift.Diff{Field: "image", Expected: want, Actual: got}
}

func TestApply_NilConfig(t *testing.T) {
	results := []drift.Result{makeResult("web", []drift.Diff{imageDiff("a", "b")})}
	out := Apply(results, nil)
	if len(out[0].Diffs) != 1 {
		t.Errorf("nil config should not filter diffs")
	}
}

func TestApply_IgnoreImageTag(t *testing.T) {
	cfg := &Config{IgnoreImageTag: true}
	results := []drift.Result{makeResult("web", []drift.Diff{imageDiff("a", "b"), envDiff("PORT", "80", "90")})}
	out := Apply(results, cfg)
	if len(out[0].Diffs) != 1 {
		t.Errorf("want 1 diff after ignoring image, got %d", len(out[0].Diffs))
	}
	if out[0].Diffs[0].Field != "env" {
		t.Errorf("remaining diff should be env")
	}
}

func TestApply_EnvKeyPrefixFilter(t *testing.T) {
	cfg := &Config{EnvKeyPrefixes: []string{"DEBUG_", "TEST_"}}
	diffs := []drift.Diff{
		envDiff("DEBUG_LEVEL", "1", "2"),
		envDiff("APP_PORT", "80", "90"),
	}
	out := Apply([]drift.Result{makeResult("svc", diffs)}, cfg)
	if len(out[0].Diffs) != 1 {
		t.Errorf("want 1 diff, got %d", len(out[0].Diffs))
	}
	if out[0].Diffs[0].Key != "APP_PORT" {
		t.Errorf("unexpected key: %s", out[0].Diffs[0].Key)
	}
}

func TestApply_MaxEnvDiffs(t *testing.T) {
	cfg := &Config{MaxEnvDiffs: 2}
	diffs := []drift.Diff{
		envDiff("A", "1", "2"),
		envDiff("B", "1", "2"),
		envDiff("C", "1", "2"),
	}
	out := Apply([]drift.Result{makeResult("svc", diffs)}, cfg)
	if len(out[0].Diffs) != 2 {
		t.Errorf("want 2 env diffs capped, got %d", len(out[0].Diffs))
	}
}

func TestApply_DriftedFlagUpdated(t *testing.T) {
	cfg := &Config{IgnoreImageTag: true}
	results := []drift.Result{makeResult("web", []drift.Diff{imageDiff("a", "b")})}
	out := Apply(results, cfg)
	if out[0].Drifted {
		t.Error("drifted flag should be false after all diffs filtered")
	}
}
