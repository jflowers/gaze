package crap

import (
	"path/filepath"

	"github.com/unbound-force/gaze/internal/diff"
)

// ChangedFunction is a Score enriched with a pass/fail indicator
// for the change gate.
type ChangedFunction struct {
	Score
	Passed bool `json:"passed"`
}

// ChangedFunctionsSummary holds aggregate counts for the change gate.
type ChangedFunctionsSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// ChangeGateResult is the output of the change gate evaluation. It is not
// marshaled directly — WriteJSONWithChangeGate and the comparison JSON
// envelope each assemble their own output shape from its fields.
type ChangeGateResult struct {
	ChangedFunctions []ChangedFunction
	Summary          ChangedFunctionsSummary
	Passed           bool
}

// FilterChangedFunctions returns scores for functions whose line
// range intersects any changed line range in fileChanges.
func FilterChangedFunctions(scores []Score, fileChanges []diff.FileChange) []Score {
	changeMap := buildChangeMap(fileChanges)
	var result []Score
	for _, s := range scores {
		if isFunctionChanged(s, changeMap) {
			result = append(result, s)
		}
	}
	return result
}

// EvaluateChangeGate filters scores to changed functions and
// evaluates each against the CRAP and GazeCRAP thresholds.
func EvaluateChangeGate(scores []Score, fileChanges []diff.FileChange, crapThreshold, gazeCRAPThreshold float64) *ChangeGateResult {
	changed := FilterChangedFunctions(scores, fileChanges)

	result := &ChangeGateResult{
		Passed: true,
	}

	passed := 0
	failed := 0
	for _, s := range changed {
		p := !exceedsThreshold(s, crapThreshold, gazeCRAPThreshold)
		if !p {
			result.Passed = false
			failed++
		} else {
			passed++
		}
		result.ChangedFunctions = append(result.ChangedFunctions, ChangedFunction{
			Score:  s,
			Passed: p,
		})
	}

	result.Summary = ChangedFunctionsSummary{
		Total:  len(changed),
		Passed: passed,
		Failed: failed,
	}

	return result
}

type changeMap map[string][]diff.LineRange

func buildChangeMap(fileChanges []diff.FileChange) changeMap {
	m := make(changeMap, len(fileChanges))
	for _, fc := range fileChanges {
		m[cleanPath(fc.Path)] = fc.Ranges
	}
	return m
}

func isFunctionChanged(s Score, cm changeMap) bool {
	ranges, ok := cm[cleanPath(s.File)]
	if !ok {
		return false
	}
	endLine := s.EndLine
	if endLine == 0 {
		endLine = s.Line
	}
	for _, r := range ranges {
		if r.Start <= endLine && r.End >= s.Line {
			return true
		}
	}
	return false
}

// cleanPath normalizes a file path for map-key comparison. Both Score.File
// (relativized to the module root in analyze) and diff.FileChange.Path
// (emitted by git diff relative to the repository root) are module-relative
// in production, so filepath.Clean — which strips "./" prefixes and
// redundant separators — is sufficient for exact key equality. Exact matching
// avoids the false positives of basename or suffix heuristics.
func cleanPath(p string) string {
	return filepath.Clean(p)
}
