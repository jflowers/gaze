package adapter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/unbound-force/gaze/v2/internal/adapter"
)

func TestSession_DiscoverPopulatesTestFiles(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio"}, "/tmp/project", []string{"./..."}, &stderr, nil)
	defer func() { _ = session.Close() }()

	providers, err := session.Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	tf := session.DiscoverTestFiles()
	if !tf["tests/test_ops.py"] {
		t.Errorf("DiscoverTestFiles() = %v, want entry for tests/test_ops.py", tf)
	}

	results, err := providers.Complexity.Analyze([]string{"./..."}, "/tmp/project")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Analyze() = %d funcs, want 3 (test-file entry filtered)", len(results))
	}
	for _, r := range results {
		if r.File == "tests/test_ops.py" {
			t.Errorf("test file tests/test_ops.py leaked into CRAP scoring")
		}
	}
}

func TestSession_DiscoverCapabilityDisabled(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio", "--no-discover", "--report-counts"}, "/tmp/project", []string{"./..."}, &stderr, nil)

	providers, err := session.Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	if tf := session.DiscoverTestFiles(); tf != nil {
		t.Errorf("DiscoverTestFiles() = %v, want nil when discover capability absent", tf)
	}

	results, err := providers.Complexity.Analyze([]string{"./..."}, "/tmp/project")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Analyze() = %d funcs, want 4 (no filtering when discover capability is disabled)", len(results))
	}

	// Assert discover was never invoked.
	if err := session.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if counts := session.Client().Stderr(); strings.Contains(counts, `"discover"`) {
		t.Errorf("discover should not be called when capability absent; counts = %q", counts)
	}
}

func TestSession_DiscoverError(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio", "--discover-error"}, "/tmp/project", []string{"./..."}, &stderr, nil)
	defer func() { _ = session.Close() }()

	providers, err := session.Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	if tf := session.DiscoverTestFiles(); tf != nil {
		t.Errorf("DiscoverTestFiles() = %v, want nil when discover fails", tf)
	}
	if !strings.Contains(stderr.String(), "discover failed") {
		t.Errorf("stderr = %q, want substring %q", stderr.String(), "discover failed")
	}
	// Graceful fallback: CRAP analysis still completes with all functions
	// scored (no data loss when discover fails).
	results, err := providers.Complexity.Analyze([]string{"./..."}, "/tmp/project")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Analyze() = %d funcs, want 4 (no filtering when discover fails)", len(results))
	}
}

func TestSession_DiscoverCalledOnce(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio", "--report-counts"}, "/tmp/project", []string{"./..."}, &stderr, nil)

	if _, err := session.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	counts := session.Client().Stderr()
	if !strings.Contains(counts, `"discover":1`) {
		t.Errorf("subprocess stderr = %q, want discover invoked exactly once", counts)
	}
}

func TestSession_DiscoverEmptyTestFiles(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio", "--empty-discover"}, "/tmp/project", []string{"./..."}, &stderr, nil)
	defer func() { _ = session.Close() }()

	providers, err := session.Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	tf := session.DiscoverTestFiles()
	if tf == nil {
		t.Fatalf("DiscoverTestFiles() = nil, want non-nil empty map when discover returns empty test_files")
	}
	if len(tf) != 0 {
		t.Errorf("DiscoverTestFiles() = %v, want empty map", tf)
	}

	// An empty test-file set disables filtering: all 4 functions are scored.
	results, err := providers.Complexity.Analyze([]string{"./..."}, "/tmp/project")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Analyze() = %d funcs, want 4 (no filtering when test_files is empty)", len(results))
	}
}

func TestSession_QualitySentinelEndToEnd(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio"}, "/tmp/project", []string{"./..."}, &stderr, nil)
	defer func() { _ = session.Close() }()

	providers, err := session.Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	mappings, err := adapter.FetchTestMappings(session.Client(), []string{"./..."}, "/tmp/project")
	if err != nil {
		t.Fatalf("FetchTestMappings: %v", err)
	}
	results, err := providers.SideEffects.AllResults()
	if err != nil {
		t.Fatalf("AllResults: %v", err)
	}

	reports, summary := adapter.BuildQualityFromMappings(mappings, results, session.DiscoverTestFiles())

	found := false
	for _, r := range reports {
		if r.TestFunction != "test_add" {
			continue
		}
		found = true
		if !r.ContractCoverage.NoContractExpected {
			t.Errorf("test_add ContractCoverage.NoContractExpected = false, want true")
		}
		if r.ContractCoverage.Reason != "test_function_no_target_effects" {
			t.Errorf("test_add ContractCoverage.Reason = %q, want %q", r.ContractCoverage.Reason, "test_function_no_target_effects")
		}
		if r.ContractCoverage.Percentage != 0 {
			t.Errorf("test_add ContractCoverage.Percentage = %v, want 0", r.ContractCoverage.Percentage)
		}
	}
	if !found {
		names := make([]string, len(reports))
		for i, r := range reports {
			names[i] = r.TestFunction
		}
		t.Fatalf("no test_add report produced; reports = %v", names)
	}
	if summary.TotalTests != 4 {
		t.Errorf("summary.TotalTests = %d, want 4", summary.TotalTests)
	}
}

func TestSession_CognitiveComplexityCapability(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio"}, "/tmp/project", []string{"./..."}, &stderr, nil)
	defer func() { _ = session.Close() }()

	providers, err := session.Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	if providers.CognitiveComplexity == nil {
		t.Fatal("providers.CognitiveComplexity = nil, want non-nil when analyzer advertises cognitive_complexity")
	}

	stats, err := providers.CognitiveComplexity.Analyze([]string{"./..."}, "/tmp/project")
	if err != nil {
		t.Fatalf("cognitive Analyze: %v", err)
	}

	byFunc := make(map[string]int, len(stats))
	for _, st := range stats {
		byFunc[st.Function] = st.CognitiveComplexity
	}
	want := map[string]int{"add": 2, "multiply": 3, "divide": 5}
	for fn, wantCC := range want {
		got, ok := byFunc[fn]
		if !ok {
			t.Errorf("function %q missing from cognitive results, got %v", fn, byFunc)
			continue
		}
		if got != wantCC {
			t.Errorf("%s cognitive_complexity = %d, want %d", fn, got, wantCC)
		}
	}
}

func TestSession_CognitiveComplexityCapabilityDisabled(t *testing.T) {
	var stderr bytes.Buffer
	session := adapter.NewSession(fakeBinaryPath, []string{"--stdio", "--no-cognitive-complexity"}, "/tmp/project", []string{"./..."}, &stderr, nil)
	defer func() { _ = session.Close() }()

	providers, err := session.Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	if providers.CognitiveComplexity != nil {
		t.Error("providers.CognitiveComplexity = non-nil, want nil when analyzer lacks cognitive_complexity capability")
	}
}
