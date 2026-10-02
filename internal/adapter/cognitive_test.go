package adapter

import (
	"testing"

	"github.com/unbound-force/gaze/v2/internal/protocol"
)

func TestConvertCognitiveComplexity(t *testing.T) {
	in := []protocol.FunctionCognitiveComplexityData{
		{Package: "pkg", Name: "Foo", File: "a.go", Line: 1, CognitiveComplexity: 3},
		{Package: "pkg", Name: "Bar", File: "a.go", Line: 10, CognitiveComplexity: 7},
	}
	out := convertCognitiveComplexity(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if out[0].Package != "pkg" || out[0].Function != "Foo" || out[0].File != "a.go" || out[0].Line != 1 || out[0].CognitiveComplexity != 3 {
		t.Errorf("unexpected mapping for Foo: %+v", out[0])
	}
	if out[1].Function != "Bar" || out[1].CognitiveComplexity != 7 {
		t.Errorf("unexpected mapping for Bar: %+v", out[1])
	}
}

func TestConvertCognitiveComplexity_Empty(t *testing.T) {
	out := convertCognitiveComplexity(nil)
	if out == nil || len(out) != 0 {
		t.Errorf("expected empty non-nil slice, got %v", out)
	}
}
