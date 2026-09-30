package crap

import (
	"path/filepath"
	"strings"

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

// ChangeGateResult is the output of the change gate evaluation.
type ChangeGateResult struct {
	ChangedFunctions []ChangedFunction       `json:"changed_functions"`
	Summary          ChangedFunctionsSummary `json:"changed_functions_summary"`
	Passed           bool                    `json:"passed"`
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
		p := s.CRAP < crapThreshold
		if p && s.GazeCRAP != nil && gazeCRAPThreshold > 0 {
			p = *s.GazeCRAP < gazeCRAPThreshold
		}
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
		m[fc.Path] = fc.Ranges
	}
	return m
}

func isFunctionChanged(s Score, cm changeMap) bool {
	for diffPath, ranges := range cm {
		if !fileMatches(s.File, diffPath) {
			continue
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
	}
	return false
}

func fileMatches(scoreFile, diffPath string) bool {
	if scoreFile == diffPath {
		return true
	}
	if strings.HasSuffix(scoreFile, string(filepath.Separator)+diffPath) {
		return true
	}
	if strings.HasSuffix(diffPath, string(filepath.Separator)+scoreFile) {
		return true
	}
	scoreRel := strings.TrimPrefix(scoreFile, "./")
	diffRel := strings.TrimPrefix(diffPath, "./")
	return scoreRel == diffRel
}
