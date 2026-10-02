package goprovider_test

import (
	"testing"

	"github.com/unbound-force/gaze/v2/internal/crap"
	"github.com/unbound-force/gaze/v2/internal/provider/goprovider"
)

// Compile-time interface satisfaction check for the cognitive provider.
var _ crap.CognitiveComplexityProvider = (*goprovider.GoCognitiveComplexityProvider)(nil)

// TestGoCognitiveComplexityProvider_Analyze verifies the provider enumerates
// functions via go/packages and emits receiver-qualified method names matching
// the gocyclo convention used by the cyclomatic-complexity provider, so that
// cognitive data joins correctly on the (file, function) key.
func TestGoCognitiveComplexityProvider_Analyze(t *testing.T) {
	provider := goprovider.NewCognitiveComplexityProvider()
	stats, err := provider.Analyze(
		[]string{"github.com/unbound-force/gaze/v2/internal/analysis/testdata/src/mutation"},
		moduleRoot(t),
	)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	byFunc := make(map[string]int, len(stats))
	funcNames := make([]string, 0, len(stats))
	for _, st := range stats {
		byFunc[st.Function] = st.CognitiveComplexity
		funcNames = append(funcNames, st.Function)
	}

	// All mutation-fixture functions have no control flow, so their cognitive
	// complexity is 0. Asserting the exact value catches a mis-mapped field
	// (e.g. copying Line into CognitiveComplexity) that a name-only check would
	// miss.
	for _, fn := range []string{"(*Counter).Increment", "(Counter).Value", "Normalize"} {
		v, ok := byFunc[fn]
		if !ok {
			t.Errorf("expected %s, got functions: %v", fn, funcNames)
			continue
		}
		if v != 0 {
			t.Errorf("%s cognitive complexity = %d, want 0", fn, v)
		}
	}
}

// TestGoCognitiveComplexityProvider_Values verifies the provider returns
// hand-computed cognitive complexity values from a dedicated fixture with
// non-trivial control flow.
func TestGoCognitiveComplexityProvider_Values(t *testing.T) {
	provider := goprovider.NewCognitiveComplexityProvider()
	stats, err := provider.Analyze(
		[]string{"github.com/unbound-force/gaze/v2/internal/provider/goprovider/testdata/src/cognitive"},
		moduleRoot(t),
	)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	byFunc := make(map[string]int, len(stats))
	for _, st := range stats {
		byFunc[st.Function] = st.CognitiveComplexity
	}

	want := map[string]int{
		"Simple":   0,
		"NestedIf": 3,
	}
	for fn, wantCC := range want {
		got, ok := byFunc[fn]
		if !ok {
			t.Errorf("expected function %q, got %v", fn, byFunc)
			continue
		}
		if got != wantCC {
			t.Errorf("%s cognitive complexity = %d, want %d", fn, got, wantCC)
		}
	}
}
