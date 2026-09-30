package crap

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/unbound-force/gaze/internal/diff"
)

func TestFilterChangedFunctions_FunctionContainsChangedLine(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 15, End: 15}}},
	}
	result := FilterChangedFunctions(scores, changes)
	if len(result) != 1 {
		t.Fatalf("got %d, want 1", len(result))
	}
	if result[0].Function != "Foo" {
		t.Errorf("function = %q, want %q", result[0].Function, "Foo")
	}
}

func TestFilterChangedFunctions_FunctionNotChanged(t *testing.T) {
	scores := []Score{
		{Function: "Bar", File: "pkg/foo.go", Line: 30, EndLine: 50, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 15, End: 15}}},
	}
	result := FilterChangedFunctions(scores, changes)
	if len(result) != 0 {
		t.Fatalf("got %d, want 0", len(result))
	}
}

func TestFilterChangedFunctions_NoMatchingFunctions(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "pkg/generated.go", Ranges: []diff.LineRange{{Start: 1, End: 10}}},
	}
	result := FilterChangedFunctions(scores, changes)
	if len(result) != 0 {
		t.Fatalf("got %d, want 0", len(result))
	}
}

func TestFilterChangedFunctions_MultipleFunctionsInChangedFile(t *testing.T) {
	scores := []Score{
		{Function: "A", File: "pkg/foo.go", Line: 1, EndLine: 10, CRAP: 5},
		{Function: "B", File: "pkg/foo.go", Line: 11, EndLine: 20, CRAP: 5},
		{Function: "C", File: "pkg/foo.go", Line: 21, EndLine: 30, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 5, End: 5}, {Start: 25, End: 25}}},
	}
	result := FilterChangedFunctions(scores, changes)
	if len(result) != 2 {
		t.Fatalf("got %d, want 2", len(result))
	}
	if result[0].Function != "A" || result[1].Function != "C" {
		t.Errorf("got %q and %q, want A and C", result[0].Function, result[1].Function)
	}
}

func TestEvaluateChangeGate_AllPassing(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 5},
		{Function: "Bar", File: "pkg/foo.go", Line: 30, EndLine: 50, CRAP: 8},
	}
	changes := []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 15, End: 15}, {Start: 35, End: 35}}},
	}
	result := EvaluateChangeGate(scores, changes, 15, 15)
	if !result.Passed {
		t.Error("expected gate to pass")
	}
	if result.Summary.Total != 2 {
		t.Errorf("total = %d, want 2", result.Summary.Total)
	}
	if result.Summary.Failed != 0 {
		t.Errorf("failed = %d, want 0", result.Summary.Failed)
	}
}

func TestEvaluateChangeGate_OneFailing(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 25},
		{Function: "Bar", File: "pkg/foo.go", Line: 30, EndLine: 50, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 15, End: 15}, {Start: 35, End: 35}}},
	}
	result := EvaluateChangeGate(scores, changes, 15, 15)
	if result.Passed {
		t.Error("expected gate to fail")
	}
	if result.Summary.Failed != 1 {
		t.Errorf("failed = %d, want 1", result.Summary.Failed)
	}
}

func TestEvaluateChangeGate_GazeCRAPThreshold(t *testing.T) {
	gazeCRAP := 20.0
	scores := []Score{
		{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 5, GazeCRAP: &gazeCRAP},
	}
	changes := []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 15, End: 15}}},
	}
	result := EvaluateChangeGate(scores, changes, 15, 15)
	if result.Passed {
		t.Error("expected gate to fail due to GazeCRAP")
	}
}

func TestEvaluateChangeGate_NoChangedFunctions(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 25},
	}
	changes := []diff.FileChange{
		{Path: "pkg/other.go", Ranges: []diff.LineRange{{Start: 1, End: 5}}},
	}
	result := EvaluateChangeGate(scores, changes, 15, 15)
	if !result.Passed {
		t.Error("expected gate to pass when no functions changed")
	}
	if result.Summary.Total != 0 {
		t.Errorf("total = %d, want 0", result.Summary.Total)
	}
}

// TestFilterChangedFunctions_BasenameCollisionNotMatched is a regression
// test for the basename-matching false-positive: two files in different
// directories sharing a basename must not match.
func TestFilterChangedFunctions_BasenameCollisionNotMatched(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "internal/a/util.go", Line: 10, EndLine: 25, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "cmd/b/util.go", Ranges: []diff.LineRange{{Start: 15, End: 15}}},
	}
	result := FilterChangedFunctions(scores, changes)
	if len(result) != 0 {
		t.Fatalf("got %d, want 0 (basename collision must not match)", len(result))
	}
}

// TestFilterChangedFunctions_DotSlashNormalized verifies that a diff path
// with a leading "./" still matches the same canonical score file.
func TestFilterChangedFunctions_DotSlashNormalized(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "foo.go", Line: 10, EndLine: 25, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "./foo.go", Ranges: []diff.LineRange{{Start: 15, End: 15}}},
	}
	result := FilterChangedFunctions(scores, changes)
	if len(result) != 1 {
		t.Fatalf("got %d, want 1 (./ prefix must normalize)", len(result))
	}
}

// TestFilterChangedFunctions_EndLineZeroFallback verifies the EndLine==0
// fallback: a score without an end line uses s.Line as the single-line
// range for change matching.
func TestFilterChangedFunctions_EndLineZeroFallback(t *testing.T) {
	scores := []Score{
		{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 0, CRAP: 5},
	}
	changes := []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 10, End: 10}}},
	}
	result := FilterChangedFunctions(scores, changes)
	if len(result) != 1 {
		t.Fatalf("got %d, want 1 (EndLine 0 must fall back to Line)", len(result))
	}

	changes = []diff.FileChange{
		{Path: "pkg/foo.go", Ranges: []diff.LineRange{{Start: 11, End: 11}}},
	}
	result = FilterChangedFunctions(scores, changes)
	if len(result) != 0 {
		t.Fatalf("got %d, want 0 (line 11 outside single-line range)", len(result))
	}
}

// TestWriteJSONWithChangeGate_NilChangedFunctions verifies the
// nil→[]ChangedFunction{} branch: the emitted changed_functions field must
// be an empty array, not null.
func TestWriteJSONWithChangeGate_NilChangedFunctions(t *testing.T) {
	rpt := &Report{Scores: []Score{}, Summary: Summary{}}
	cgr := &ChangeGateResult{Summary: ChangedFunctionsSummary{}}
	var buf bytes.Buffer
	if err := WriteJSONWithChangeGate(&buf, rpt, cgr); err != nil {
		t.Fatalf("WriteJSONWithChangeGate: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	cf, ok := out["changed_functions"].([]interface{})
	if !ok {
		t.Fatalf("changed_functions is not an array: %T", out["changed_functions"])
	}
	if len(cf) != 0 {
		t.Fatalf("changed_functions len = %d, want 0", len(cf))
	}
}

// TestWriteJSONWithChangeGate_Populated verifies the populated path emits
// changed_functions and changed_functions_summary sections.
func TestWriteJSONWithChangeGate_Populated(t *testing.T) {
	rpt := &Report{
		Scores:  []Score{{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 5}},
		Summary: Summary{},
	}
	cgr := &ChangeGateResult{
		ChangedFunctions: []ChangedFunction{
			{Score: Score{Function: "Foo", File: "pkg/foo.go", Line: 10, EndLine: 25, CRAP: 5}, Passed: true},
		},
		Summary: ChangedFunctionsSummary{Total: 1, Passed: 1, Failed: 0},
	}
	var buf bytes.Buffer
	if err := WriteJSONWithChangeGate(&buf, rpt, cgr); err != nil {
		t.Fatalf("WriteJSONWithChangeGate: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	cf, ok := out["changed_functions"].([]interface{})
	if !ok || len(cf) != 1 {
		t.Fatalf("changed_functions = %v, want 1 element", out["changed_functions"])
	}
	cs, ok := out["changed_functions_summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("changed_functions_summary missing or wrong type: %T", out["changed_functions_summary"])
	}
	if cs["total"] != float64(1) {
		t.Errorf("changed_functions_summary.total = %v, want 1", cs["total"])
	}
}
