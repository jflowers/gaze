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

	if _, ok := byFunc["(*Counter).Increment"]; !ok {
		t.Errorf("expected (*Counter).Increment (pointer receiver), got functions: %v", funcNames)
	}
	if _, ok := byFunc["(Counter).Value"]; !ok {
		t.Errorf("expected (Counter).Value (value receiver), got functions: %v", funcNames)
	}
	if _, ok := byFunc["Normalize"]; !ok {
		t.Errorf("expected Normalize (plain function), got functions: %v", funcNames)
	}
}
