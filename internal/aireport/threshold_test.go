package aireport

import (
	"testing"
)

func intPtr(v int) *int { return &v }

// TestEvaluateThresholds_NilConfig verifies that nil thresholds (not provided)
// result in no results and allPassed=true.
func TestEvaluateThresholds_NilConfig(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{CRAPload: 10, GazeCRAPload: 5, AvgContractCoverage: 30},
	}
	results, passed := EvaluateThresholds(ThresholdConfig{}, payload)
	if !passed {
		t.Error("expected passed=true when no thresholds set")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestEvaluateThresholds_ZeroThresholdWithZeroActual verifies that *0 threshold
// with actual=0 passes (zero is a valid live threshold, and 0 <= 0).
func TestEvaluateThresholds_ZeroThresholdWithZeroActual(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{CRAPload: 0},
	}
	cfg := ThresholdConfig{MaxCrapload: intPtr(0)}
	results, passed := EvaluateThresholds(cfg, payload)
	if !passed {
		t.Error("expected passed=true when actual=0 and limit=0")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Error("expected result.Passed=true")
	}
}

// TestEvaluateThresholds_ZeroThresholdWithPositiveActual verifies that *0 threshold
// with actual>0 fails (zero is a live threshold).
func TestEvaluateThresholds_ZeroThresholdWithPositiveActual(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{CRAPload: 3},
	}
	cfg := ThresholdConfig{MaxCrapload: intPtr(0)}
	results, passed := EvaluateThresholds(cfg, payload)
	if passed {
		t.Error("expected passed=false when actual=3 and limit=0")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("expected result.Passed=false")
	}
	if results[0].Actual != 3 {
		t.Errorf("expected Actual=3, got %d", results[0].Actual)
	}
}

// TestEvaluateThresholds_BelowLimit verifies that actual < limit passes.
func TestEvaluateThresholds_BelowLimit(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{CRAPload: 3},
	}
	cfg := ThresholdConfig{MaxCrapload: intPtr(5)}
	results, passed := EvaluateThresholds(cfg, payload)
	if !passed {
		t.Error("expected passed=true when actual < limit")
	}
	if !results[0].Passed {
		t.Error("expected result.Passed=true")
	}
}

// TestEvaluateThresholds_AboveLimit verifies that actual > limit fails.
func TestEvaluateThresholds_AboveLimit(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{CRAPload: 8},
	}
	cfg := ThresholdConfig{MaxCrapload: intPtr(5)}
	results, passed := EvaluateThresholds(cfg, payload)
	if passed {
		t.Error("expected passed=false when actual > limit")
	}
	if results[0].Passed {
		t.Error("expected result.Passed=false")
	}
}

// TestEvaluateThresholds_AllThreeFields verifies that all three threshold
// fields are evaluated independently.
func TestEvaluateThresholds_AllThreeFields(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{
			CRAPload:            4,
			GazeCRAPload:        2,
			AvgContractCoverage: 60,
		},
	}
	cfg := ThresholdConfig{
		MaxCrapload:         intPtr(5),  // pass (4 <= 5)
		MaxGazeCrapload:     intPtr(1),  // fail (2 > 1)
		MinContractCoverage: intPtr(50), // pass (60 >= 50)
	}
	results, passed := EvaluateThresholds(cfg, payload)
	if passed {
		t.Error("expected passed=false (GazeCRAPload exceeds limit)")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	byName := make(map[string]ThresholdResult)
	for _, r := range results {
		byName[r.Name] = r
	}

	if !byName["CRAPload"].Passed {
		t.Error("expected CRAPload to pass")
	}
	if byName["GazeCRAPload"].Passed {
		t.Error("expected GazeCRAPload to fail")
	}
	if !byName["AvgContractCoverage"].Passed {
		t.Error("expected AvgContractCoverage to pass")
	}
}

// TestEvaluateThresholds_BothCRAPloadsFail verifies simultaneous CRAPload
// and GazeCRAPload threshold breaches.
func TestEvaluateThresholds_BothCRAPloadsFail(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{
			CRAPload:     10,
			GazeCRAPload: 7,
		},
	}
	cfg := ThresholdConfig{
		MaxCrapload:     intPtr(5),
		MaxGazeCrapload: intPtr(3),
	}
	results, passed := EvaluateThresholds(cfg, payload)
	if passed {
		t.Error("expected passed=false")
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Passed {
			t.Errorf("expected %s to fail", r.Name)
		}
	}
}

// TestEvaluateThresholds_GazeCRAPloadZeroLiveThreshold verifies the US2
// scenario 7: --max-gaze-crapload=0 with positive actual fails.
func TestEvaluateThresholds_GazeCRAPloadZeroLiveThreshold(t *testing.T) {
	payload := &ReportPayload{
		Summary: ReportSummary{GazeCRAPload: 1},
	}
	cfg := ThresholdConfig{MaxGazeCrapload: intPtr(0)}
	results, passed := EvaluateThresholds(cfg, payload)
	if passed {
		t.Error("expected passed=false when GazeCRAPload=1 and limit=0")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("expected result.Passed=false")
	}
	if results[0].Name != "GazeCRAPload" {
		t.Errorf("expected Name=GazeCRAPload, got %s", results[0].Name)
	}
}

// TestEvaluateThresholds_NilPayload verifies graceful handling of nil payload.
func TestEvaluateThresholds_NilPayload(t *testing.T) {
	cfg := ThresholdConfig{MaxCrapload: intPtr(5)}
	results, passed := EvaluateThresholds(cfg, nil)
	// nil payload → zero-value summary → CRAPload=0 ≤ 5 → pass
	if !passed {
		t.Error("expected passed=true with nil payload and limit=5")
	}
	if len(results) != 1 || !results[0].Passed {
		t.Errorf("unexpected results: %+v", results)
	}
}

// BenchmarkEvaluateThresholds measures the overhead of threshold evaluation.
// EvaluateThresholds is a pure in-memory function with no I/O; its overhead
// must be negligible (well under 1 ms per invocation).
func BenchmarkEvaluateThresholds(b *testing.B) {
	payload := &ReportPayload{
		Summary: ReportSummary{
			CRAPload:            8,
			GazeCRAPload:        3,
			AvgContractCoverage: 72,
		},
	}
	cfg := ThresholdConfig{
		MaxCrapload:         intPtr(10),
		MaxGazeCrapload:     intPtr(5),
		MinContractCoverage: intPtr(60),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateThresholds(cfg, payload)
	}
}
