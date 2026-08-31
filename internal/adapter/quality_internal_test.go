package adapter

import (
	"testing"

	"github.com/unbound-force/gaze/internal/protocol"
	"github.com/unbound-force/gaze/internal/taxonomy"
)

// ---------------------------------------------------------------------------
// BuildQualityFromMappings tests (table-driven)
// ---------------------------------------------------------------------------

func TestBuildQualityFromMappings(t *testing.T) {
	// Helper to create classified side effects.
	contractual := func(id, typ string) taxonomy.SideEffect {
		return taxonomy.SideEffect{
			ID:   id,
			Type: taxonomy.SideEffectType(typ),
			Classification: &taxonomy.Classification{
				Label:      taxonomy.Contractual,
				Confidence: 90,
			},
		}
	}
	incidental := func(id, typ string) taxonomy.SideEffect {
		return taxonomy.SideEffect{
			ID:   id,
			Type: taxonomy.SideEffectType(typ),
			Classification: &taxonomy.Classification{
				Label:      taxonomy.Incidental,
				Confidence: 90,
			},
		}
	}
	ambiguous := func(id, typ string) taxonomy.SideEffect {
		return taxonomy.SideEffect{
			ID:   id,
			Type: taxonomy.SideEffectType(typ),
			Classification: &taxonomy.Classification{
				Label:      taxonomy.Ambiguous,
				Confidence: 55,
			},
		}
	}
	unclassified := func(id, typ string) taxonomy.SideEffect {
		return taxonomy.SideEffect{
			ID:   id,
			Type: taxonomy.SideEffectType(typ),
		}
	}

	t.Run("full classification with contract coverage and over-specification", func(t *testing.T) {
		results := []taxonomy.AnalysisResult{
			{
				Target: taxonomy.FunctionTarget{
					Package:  "math_utils",
					Function: "multiply",
					Location: "math_utils.py:10",
				},
				SideEffects: []taxonomy.SideEffect{
					contractual("se-001", "ReturnValue"),
					contractual("se-002", "ErrorReturn"),
					incidental("se-003", "LogOutput"),
				},
			},
		}
		mappings := []protocol.AssertionMappingData{
			{
				TestFunction:      "test_multiply",
				TestFile:          "tests/test_math.py",
				AssertionLocation: "tests/test_math.py:15",
				AssertionType:     "equality",
				TargetFunction:    "multiply",
				TargetPackage:     "math_utils",
				SideEffectType:    "ReturnValue",
				Confidence:        80,
			},
			{
				TestFunction:      "test_multiply",
				TestFile:          "tests/test_math.py",
				AssertionLocation: "tests/test_math.py:16",
				AssertionType:     "equality",
				TargetFunction:    "multiply",
				TargetPackage:     "math_utils",
				SideEffectType:    "LogOutput",
				Confidence:        70,
			},
		}

		reports, summary := BuildQualityFromMappings(mappings, results)

		if len(reports) != 1 {
			t.Fatalf("got %d reports, want 1", len(reports))
		}
		r := reports[0]
		if r.TestFunction != "test_multiply" {
			t.Errorf("TestFunction = %q, want %q", r.TestFunction, "test_multiply")
		}
		// Contract coverage: 1 of 2 contractual effects covered (ReturnValue asserted, ErrorReturn not)
		if r.ContractCoverage.Percentage != 50 {
			t.Errorf("ContractCoverage.Percentage = %v, want 50", r.ContractCoverage.Percentage)
		}
		if r.ContractCoverage.CoveredCount != 1 {
			t.Errorf("CoveredCount = %d, want 1", r.ContractCoverage.CoveredCount)
		}
		if r.ContractCoverage.TotalContractual != 2 {
			t.Errorf("TotalContractual = %d, want 2", r.ContractCoverage.TotalContractual)
		}
		// Over-specification: 1 assertion on incidental effect (LogOutput)
		if r.OverSpecification.Count != 1 {
			t.Errorf("OverSpecification.Count = %d, want 1", r.OverSpecification.Count)
		}
		if r.AssertionCount != 2 {
			t.Errorf("AssertionCount = %d, want 2", r.AssertionCount)
		}
		if r.AssertionDetectionConfidence != 0 {
			t.Errorf("AssertionDetectionConfidence = %d, want 0", r.AssertionDetectionConfidence)
		}

		// Summary
		if summary.TotalTests != 1 {
			t.Errorf("summary.TotalTests = %d, want 1", summary.TotalTests)
		}
		if summary.AverageContractCoverage != 50 {
			t.Errorf("summary.AverageContractCoverage = %v, want 50", summary.AverageContractCoverage)
		}
		if summary.TotalOverSpecifications != 1 {
			t.Errorf("summary.TotalOverSpecifications = %d, want 1", summary.TotalOverSpecifications)
		}
	})

	t.Run("no mappings produces empty report", func(t *testing.T) {
		results := []taxonomy.AnalysisResult{
			{
				Target: taxonomy.FunctionTarget{Package: "pkg", Function: "fn"},
				SideEffects: []taxonomy.SideEffect{
					contractual("se-001", "ReturnValue"),
				},
			},
		}

		reports, summary := BuildQualityFromMappings(nil, results)

		if len(reports) != 0 {
			t.Fatalf("got %d reports, want 0", len(reports))
		}
		if summary.TotalTests != 0 {
			t.Errorf("summary.TotalTests = %d, want 0", summary.TotalTests)
		}
	})

	t.Run("empty mappings slice produces empty report", func(t *testing.T) {
		results := []taxonomy.AnalysisResult{
			{
				Target: taxonomy.FunctionTarget{Package: "pkg", Function: "fn"},
				SideEffects: []taxonomy.SideEffect{
					contractual("se-001", "ReturnValue"),
				},
			},
		}

		reports, summary := BuildQualityFromMappings([]protocol.AssertionMappingData{}, results)

		if len(reports) != 0 {
			t.Fatalf("got %d reports, want 0", len(reports))
		}
		if summary.TotalTests != 0 {
			t.Errorf("summary.TotalTests = %d, want 0", summary.TotalTests)
		}
	})

	t.Run("unclassified effects treated as contractual", func(t *testing.T) {
		results := []taxonomy.AnalysisResult{
			{
				Target: taxonomy.FunctionTarget{Package: "pkg", Function: "fn"},
				SideEffects: []taxonomy.SideEffect{
					unclassified("se-001", "ReturnValue"),
					unclassified("se-002", "ErrorReturn"),
				},
			},
		}
		mappings := []protocol.AssertionMappingData{
			{
				TestFunction:   "test_fn",
				TestFile:       "test.py",
				TargetFunction: "fn",
				TargetPackage:  "pkg",
				SideEffectType: "ReturnValue",
				Confidence:     80,
			},
		}

		reports, _ := BuildQualityFromMappings(mappings, results)

		if len(reports) != 1 {
			t.Fatalf("got %d reports, want 1", len(reports))
		}
		// 1 of 2 unclassified (= contractual) effects covered
		if reports[0].ContractCoverage.Percentage != 50 {
			t.Errorf("ContractCoverage.Percentage = %v, want 50", reports[0].ContractCoverage.Percentage)
		}
		if reports[0].ContractCoverage.TotalContractual != 2 {
			t.Errorf("TotalContractual = %d, want 2", reports[0].ContractCoverage.TotalContractual)
		}
	})

	t.Run("multiple test functions produce separate reports", func(t *testing.T) {
		results := []taxonomy.AnalysisResult{
			{
				Target: taxonomy.FunctionTarget{Package: "pkg", Function: "fn"},
				SideEffects: []taxonomy.SideEffect{
					contractual("se-001", "ReturnValue"),
				},
			},
		}
		mappings := []protocol.AssertionMappingData{
			{
				TestFunction:   "test_fn_basic",
				TestFile:       "test.py",
				TargetFunction: "fn",
				TargetPackage:  "pkg",
				SideEffectType: "ReturnValue",
				Confidence:     90,
			},
			{
				TestFunction:   "test_fn_edge",
				TestFile:       "test.py",
				TargetFunction: "fn",
				TargetPackage:  "pkg",
				SideEffectType: "ReturnValue",
				Confidence:     85,
			},
		}

		reports, summary := BuildQualityFromMappings(mappings, results)

		if len(reports) != 2 {
			t.Fatalf("got %d reports, want 2", len(reports))
		}
		// Reports sorted by test function name
		if reports[0].TestFunction != "test_fn_basic" {
			t.Errorf("reports[0].TestFunction = %q, want %q", reports[0].TestFunction, "test_fn_basic")
		}
		if reports[1].TestFunction != "test_fn_edge" {
			t.Errorf("reports[1].TestFunction = %q, want %q", reports[1].TestFunction, "test_fn_edge")
		}
		if summary.TotalTests != 2 {
			t.Errorf("summary.TotalTests = %d, want 2", summary.TotalTests)
		}
		// Both tests cover the same function's effect → 100% each
		if summary.AverageContractCoverage != 100 {
			t.Errorf("summary.AverageContractCoverage = %v, want 100", summary.AverageContractCoverage)
		}
	})

	t.Run("ambiguous effects excluded from contract coverage", func(t *testing.T) {
		results := []taxonomy.AnalysisResult{
			{
				Target: taxonomy.FunctionTarget{Package: "pkg", Function: "fn"},
				SideEffects: []taxonomy.SideEffect{
					contractual("se-001", "ReturnValue"),
					ambiguous("se-002", "MapMutation"),
				},
			},
		}
		mappings := []protocol.AssertionMappingData{
			{
				TestFunction:   "test_fn",
				TestFile:       "test.py",
				TargetFunction: "fn",
				TargetPackage:  "pkg",
				SideEffectType: "ReturnValue",
				Confidence:     80,
			},
		}

		reports, _ := BuildQualityFromMappings(mappings, results)

		if len(reports) != 1 {
			t.Fatalf("got %d reports, want 1", len(reports))
		}
		// Ambiguous effect excluded → 1 contractual, 1 covered = 100%
		if reports[0].ContractCoverage.Percentage != 100 {
			t.Errorf("ContractCoverage.Percentage = %v, want 100", reports[0].ContractCoverage.Percentage)
		}
		if len(reports[0].AmbiguousEffects) != 1 {
			t.Errorf("AmbiguousEffects = %d, want 1", len(reports[0].AmbiguousEffects))
		}
	})

	t.Run("worst coverage tests in summary", func(t *testing.T) {
		results := []taxonomy.AnalysisResult{
			{
				Target: taxonomy.FunctionTarget{Package: "pkg", Function: "fn"},
				SideEffects: []taxonomy.SideEffect{
					contractual("se-001", "ReturnValue"),
					contractual("se-002", "ErrorReturn"),
				},
			},
		}
		// Two tests: one covers both effects (100%), one covers none (0%)
		mappings := []protocol.AssertionMappingData{
			{
				TestFunction:   "test_fn_full",
				TestFile:       "test.py",
				TargetFunction: "fn",
				TargetPackage:  "pkg",
				SideEffectType: "ReturnValue",
				Confidence:     90,
			},
			{
				TestFunction:   "test_fn_full",
				TestFile:       "test.py",
				TargetFunction: "fn",
				TargetPackage:  "pkg",
				SideEffectType: "ErrorReturn",
				Confidence:     90,
			},
			{
				TestFunction:      "test_fn_partial",
				TestFile:          "test.py",
				AssertionLocation: "test.py:30",
				AssertionType:     "equality",
				TargetFunction:    "fn",
				TargetPackage:     "pkg",
				SideEffectType:    "UnknownType", // won't match any effect
				Confidence:        50,
			},
		}

		_, summary := BuildQualityFromMappings(mappings, results)

		if len(summary.WorstCoverageTests) != 2 {
			t.Fatalf("WorstCoverageTests = %d, want 2", len(summary.WorstCoverageTests))
		}
		// Worst first
		if summary.WorstCoverageTests[0].ContractCoverage.Percentage != 0 {
			t.Errorf("worst[0].Percentage = %v, want 0", summary.WorstCoverageTests[0].ContractCoverage.Percentage)
		}
		if summary.WorstCoverageTests[1].ContractCoverage.Percentage != 100 {
			t.Errorf("worst[1].Percentage = %v, want 100", summary.WorstCoverageTests[1].ContractCoverage.Percentage)
		}
	})

	t.Run("nil results with mappings still produces reports", func(t *testing.T) {
		mappings := []protocol.AssertionMappingData{
			{
				TestFunction:   "test_fn",
				TestFile:       "test.py",
				TargetFunction: "fn",
				TargetPackage:  "pkg",
				SideEffectType: "ReturnValue",
				Confidence:     80,
			},
		}

		reports, summary := BuildQualityFromMappings(mappings, nil)

		if len(reports) != 1 {
			t.Fatalf("got %d reports, want 1", len(reports))
		}
		// No effects → zero coverage
		if reports[0].ContractCoverage.Percentage != 0 {
			t.Errorf("ContractCoverage.Percentage = %v, want 0", reports[0].ContractCoverage.Percentage)
		}
		if summary.TotalTests != 1 {
			t.Errorf("summary.TotalTests = %d, want 1", summary.TotalTests)
		}
	})
}

// ---------------------------------------------------------------------------
// computeOverSpecification tests (table-driven)
// ---------------------------------------------------------------------------

func TestComputeOverSpecification(t *testing.T) {
	t.Run("no incidental effects", func(t *testing.T) {
		effects := []taxonomy.SideEffect{
			{
				ID:   "se-001",
				Type: "ReturnValue",
				Classification: &taxonomy.Classification{
					Label:      taxonomy.Contractual,
					Confidence: 90,
				},
			},
		}
		mappings := []taxonomy.AssertionMapping{
			{SideEffectID: "se-001", Confidence: 80},
		}

		os := computeOverSpecification(effects, mappings)

		if os.Count != 0 {
			t.Errorf("Count = %d, want 0", os.Count)
		}
		if os.Ratio != 0 {
			t.Errorf("Ratio = %v, want 0", os.Ratio)
		}
	})

	t.Run("all incidental assertions", func(t *testing.T) {
		effects := []taxonomy.SideEffect{
			{
				ID:   "se-001",
				Type: "LogOutput",
				Classification: &taxonomy.Classification{
					Label:      taxonomy.Incidental,
					Confidence: 90,
				},
			},
		}
		mappings := []taxonomy.AssertionMapping{
			{SideEffectID: "se-001", Confidence: 80},
		}

		os := computeOverSpecification(effects, mappings)

		if os.Count != 1 {
			t.Errorf("Count = %d, want 1", os.Count)
		}
		if os.Ratio != 1.0 {
			t.Errorf("Ratio = %v, want 1.0", os.Ratio)
		}
		if len(os.Suggestions) != 1 {
			t.Errorf("Suggestions = %d, want 1", len(os.Suggestions))
		}
	})

	t.Run("empty mappings", func(t *testing.T) {
		os := computeOverSpecification(nil, nil)

		if os.Count != 0 {
			t.Errorf("Count = %d, want 0", os.Count)
		}
		if os.Ratio != 0 {
			t.Errorf("Ratio = %v, want 0", os.Ratio)
		}
	})
}
