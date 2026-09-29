package crap

import (
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
