package groupby_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/groupby"
)

func makeResult(name string, drifted bool, diffs []drift.Diff) drift.Result {
	return drift.Result{Name: name, Drifted: drifted, Diffs: diffs}
}

func makeDiff(field, want, got string) drift.Diff {
	return drift.Diff{Field: field, Want: want, Got: got}
}

func TestBy_Status(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", true, nil),
		makeResult("beta", false, nil),
		makeResult("gamma", true, nil),
	}
	res, err := groupby.By(results, groupby.FieldStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(res.Groups))
	}
	// drifted group should be first (count=2)
	if res.Groups[0].Key != "drifted" || res.Groups[0].Count != 2 {
		t.Errorf("expected drifted group first with count 2, got %+v", res.Groups[0])
	}
}

func TestBy_Image(t *testing.T) {
	results := []drift.Result{
		makeResult("a", true, []drift.Diff{makeDiff("image", "nginx:1.24", "nginx:1.25")}),
		makeResult("b", true, []drift.Diff{makeDiff("image", "nginx:1.24", "nginx:1.25")}),
		makeResult("c", true, []drift.Diff{makeDiff("image", "redis:7", "redis:8")}),
	}
	res, err := groupby.By(results, groupby.FieldImage)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(res.Groups))
	}
	if res.Groups[0].Key != "nginx:1.25" || res.Groups[0].Count != 2 {
		t.Errorf("expected nginx:1.25 first with count 2, got %+v", res.Groups[0])
	}
}

func TestBy_EnvKey(t *testing.T) {
	results := []drift.Result{
		makeResult("x", true, []drift.Diff{makeDiff("env.PORT", "8080", "9090")}),
		makeResult("y", true, []drift.Diff{makeDiff("env.PORT", "8080", "9091"), makeDiff("env.HOST", "a", "b")}),
	}
	res, err := groupby.By(results, groupby.FieldEnvKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys := map[string]int{}
	for _, g := range res.Groups {
		keys[g.Key] = g.Count
	}
	if keys["PORT"] != 2 {
		t.Errorf("expected PORT count=2, got %d", keys["PORT"])
	}
	if keys["HOST"] != 1 {
		t.Errorf("expected HOST count=1, got %d", keys["HOST"])
	}
}

func TestBy_UnknownField(t *testing.T) {
	_, err := groupby.By([]drift.Result{makeResult("a", false, nil)}, groupby.Field("bogus"))
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestWrite_Text(t *testing.T) {
	res := groupby.Result{
		Field: groupby.FieldStatus,
		Groups: []groupby.Group{
			{Key: "drifted", Count: 1, Names: []string{"alpha"}},
		},
	}
	var buf bytes.Buffer
	if err := groupby.Write(&buf, res, "text"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "drifted") || !strings.Contains(out, "alpha") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestWrite_JSON(t *testing.T) {
	res := groupby.Result{
		Field: groupby.FieldStatus,
		Groups: []groupby.Group{
			{Key: "clean", Count: 1, Names: []string{"beta"}},
		},
	}
	var buf bytes.Buffer
	if err := groupby.Write(&buf, res, "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"field"`) {
		t.Errorf("expected JSON output, got: %s", buf.String())
	}
}
