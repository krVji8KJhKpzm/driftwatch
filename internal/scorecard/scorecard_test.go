package scorecard

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeResult(name string, diffs []drift.Diff) drift.Result {
	return drift.Result{Name: name, Drifted: len(diffs) > 0, Diffs: diffs}
}

func makeDiffs(n int) []drift.Diff {
	d := make([]drift.Diff, n)
	for i := range d {
		d[i] = drift.Diff{Field: "image", Expected: "v1", Actual: "v2"}
	}
	return d
}

func TestBuild_NoDrift(t *testing.T) {
	results := []drift.Result{makeResult("web", nil)}
	entries := Build(results)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Grade != GradeA {
		t.Errorf("expected grade A, got %s", entries[0].Grade)
	}
	if entries[0].DriftCount != 0 {
		t.Errorf("expected 0 drifts, got %d", entries[0].DriftCount)
	}
}

func TestBuild_GradeMapping(t *testing.T) {
	cases := []struct {
		drifts int
		want   Grade
	}{
		{0, GradeA},
		{1, GradeB},
		{2, GradeC},
		{4, GradeD},
		{5, GradeF},
	}
	for _, tc := range cases {
		r := makeResult("svc", makeDiffs(tc.drifts))
		entries := Build([]drift.Result{r})
		if entries[0].Grade != tc.want {
			t.Errorf("drifts=%d: expected %s, got %s", tc.drifts, tc.want, entries[0].Grade)
		}
	}
}

func TestBuild_SortedByDriftCountDesc(t *testing.T) {
	results := []drift.Result{
		makeResult("alpha", makeDiffs(1)),
		makeResult("beta", makeDiffs(5)),
		makeResult("gamma", nil),
	}
	entries := Build(results)
	if entries[0].Name != "beta" {
		t.Errorf("expected beta first, got %s", entries[0].Name)
	}
	if entries[2].Name != "gamma" {
		t.Errorf("expected gamma last, got %s", entries[2].Name)
	}
}

func TestWrite_Output(t *testing.T) {
	results := []drift.Result{
		makeResult("web", makeDiffs(2)),
		makeResult("db", nil),
	}
	entries := Build(results)
	var buf bytes.Buffer
	Write(&buf, entries)
	out := buf.String()
	if !strings.Contains(out, "CONTAINER") {
		t.Error("expected header CONTAINER in output")
	}
	if !strings.Contains(out, "web") {
		t.Error("expected 'web' in output")
	}
	if !strings.Contains(out, string(GradeC)) {
		t.Errorf("expected grade C in output")
	}
	if !strings.Contains(out, string(GradeA)) {
		t.Errorf("expected grade A in output")
	}
}
