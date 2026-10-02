// Package cognitive is a test fixture providing functions with hand-computed
// cognitive complexity values, used by the goprovider integration test to
// verify that GoCognitiveComplexityProvider returns correct values.
package cognitive

// Simple has no control flow and therefore a cognitive complexity of 0.
func Simple() {}

// NestedIf has an if nested inside an if: 1 (outer if) + 1 (nested if at
// nesting depth 1) = 3.
func NestedIf() bool {
	if true {
		if true {
			return true
		}
	}
	return false
}
