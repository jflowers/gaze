package classify_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/jflowers/gaze/internal/analysis"
	"github.com/jflowers/gaze/internal/classify"
	"github.com/jflowers/gaze/internal/config"
	"github.com/jflowers/gaze/internal/taxonomy"
)

// testdataDir returns the absolute path to the testdata/src directory.
func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata", "src")
}

// loadTestPackages loads the test fixture packages for classification
// testing. Returns all packages in the testdata module.
func loadTestPackages(t *testing.T, patterns ...string) []*packages.Package {
	t.Helper()
	dir := testdataDir()

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
		Dir:   dir,
		Tests: false,
	}

	// Load all packages to enable cross-package analysis.
	allPatterns := []string{"./..."}
	if len(patterns) > 0 {
		allPatterns = patterns
	}

	pkgs, err := packages.Load(cfg, allPatterns...)
	if err != nil {
		t.Fatalf("loading test packages: %v", err)
	}

	var valid []*packages.Package
	for _, pkg := range pkgs {
		if len(pkg.Errors) == 0 {
			valid = append(valid, pkg)
		}
	}

	if len(valid) == 0 {
		t.Fatal("no valid test packages loaded")
	}

	return valid
}

// findPackage finds a package by suffix in the loaded packages.
func findPackage(pkgs []*packages.Package, suffix string) *packages.Package {
	for _, pkg := range pkgs {
		if len(pkg.PkgPath) >= len(suffix) &&
			pkg.PkgPath[len(pkg.PkgPath)-len(suffix):] == suffix {
			return pkg
		}
	}
	return nil
}

// TestNamingSignal_ContractualPrefixes tests naming convention
// detection for contractual function names.
func TestNamingSignal_ContractualPrefixes(t *testing.T) {
	tests := []struct {
		name       string
		funcName   string
		effectType taxonomy.SideEffectType
		wantWeight int
	}{
		{"GetData returns", "GetData", taxonomy.ReturnValue, 10},
		{"SaveRecord mutation", "SaveRecord", taxonomy.ErrorReturn, 10},
		{"FetchConfig returns", "FetchConfig", taxonomy.ReturnValue, 10},
		{"DeleteItem error", "DeleteItem", taxonomy.ErrorReturn, 10},
		{"HandleRequest any", "HandleRequest", taxonomy.ReceiverMutation, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := classify.AnalyzeNamingSignal(tt.funcName, tt.effectType)
			if s.Weight != tt.wantWeight {
				t.Errorf("AnalyzeNamingSignal(%q, %s) weight = %d, want %d",
					tt.funcName, tt.effectType, s.Weight, tt.wantWeight)
			}
		})
	}
}

// TestNamingSignal_IncidentalPrefixes tests naming convention
// detection for incidental function names.
func TestNamingSignal_IncidentalPrefixes(t *testing.T) {
	tests := []struct {
		funcName string
	}{
		{"logError"},
		{"LogInfo"},
		{"debugTrace"},
		{"Debug"},
		{"traceRequest"},
		{"printResult"},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			s := classify.AnalyzeNamingSignal(tt.funcName, taxonomy.ReturnValue)
			if s.Weight >= 0 {
				t.Errorf("AnalyzeNamingSignal(%q) weight = %d, want negative",
					tt.funcName, s.Weight)
			}
		})
	}
}

// TestNamingSignal_NoMatch tests that unknown names produce zero
// signal.
func TestNamingSignal_NoMatch(t *testing.T) {
	s := classify.AnalyzeNamingSignal("computeHash", taxonomy.ReturnValue)
	if s.Source != "" {
		t.Errorf("expected zero signal for %q, got source=%q weight=%d",
			"computeHash", s.Source, s.Weight)
	}
}

// TestScoreComputation_BaseConfidence tests that zero signals
// produce a score of 50 (the base confidence).
func TestScoreComputation_BaseConfidence(t *testing.T) {
	c := classify.ComputeScore(nil, nil)
	if c.Confidence != 50 {
		t.Errorf("zero signals: confidence = %d, want 50", c.Confidence)
	}
	if c.Label != taxonomy.Ambiguous {
		t.Errorf("zero signals: label = %q, want %q", c.Label, taxonomy.Ambiguous)
	}
}

// TestScoreComputation_Contractual tests that strong positive
// signals produce a contractual classification.
func TestScoreComputation_Contractual(t *testing.T) {
	signals := []taxonomy.Signal{
		{Source: "interface", Weight: 30},
		{Source: "visibility", Weight: 10},
	}

	c := classify.ComputeScore(signals, nil)
	// 50 + 30 + 10 = 90 >= 80 = contractual.
	if c.Label != taxonomy.Contractual {
		t.Errorf("label = %q, want %q", c.Label, taxonomy.Contractual)
	}
	if c.Confidence != 90 {
		t.Errorf("confidence = %d, want 90", c.Confidence)
	}
}

// TestScoreComputation_Incidental tests that negative signals
// produce an incidental classification.
func TestScoreComputation_Incidental(t *testing.T) {
	signals := []taxonomy.Signal{
		{Source: "naming", Weight: -10},
	}

	c := classify.ComputeScore(signals, nil)
	// 50 - 10 = 40 < 50 = incidental.
	if c.Label != taxonomy.Incidental {
		t.Errorf("label = %q, want %q", c.Label, taxonomy.Incidental)
	}
	if c.Confidence != 40 {
		t.Errorf("confidence = %d, want 40", c.Confidence)
	}
}

// TestScoreComputation_Contradiction tests that contradicting
// signals apply a penalty.
func TestScoreComputation_Contradiction(t *testing.T) {
	signals := []taxonomy.Signal{
		{Source: "interface", Weight: 30},
		{Source: "naming", Weight: -10},
	}

	c := classify.ComputeScore(signals, nil)
	// 50 + 30 - 10 - 20 (contradiction) = 50.
	if c.Confidence != 50 {
		t.Errorf("contradiction: confidence = %d, want 50", c.Confidence)
	}
	if c.Label != taxonomy.Ambiguous {
		t.Errorf("contradiction: label = %q, want %q", c.Label, taxonomy.Ambiguous)
	}
}

// TestScoreComputation_ClampToZero tests that very negative scores
// clamp to 0.
func TestScoreComputation_ClampToZero(t *testing.T) {
	signals := []taxonomy.Signal{
		{Source: "naming", Weight: -10},
		{Source: "godoc", Weight: -15},
		{Source: "another", Weight: -30},
	}

	c := classify.ComputeScore(signals, nil)
	if c.Confidence != 0 {
		t.Errorf("clamp: confidence = %d, want 0", c.Confidence)
	}
}

// TestScoreComputation_ClampTo100 tests that very positive scores
// clamp to 100.
func TestScoreComputation_ClampTo100(t *testing.T) {
	signals := []taxonomy.Signal{
		{Source: "interface", Weight: 30},
		{Source: "visibility", Weight: 20},
		{Source: "caller", Weight: 15},
		{Source: "naming", Weight: 10},
		{Source: "godoc", Weight: 15},
	}

	c := classify.ComputeScore(signals, nil)
	// 50 + 30 + 20 + 15 + 10 + 15 = 140, clamped to 100.
	if c.Confidence != 100 {
		t.Errorf("clamp: confidence = %d, want 100", c.Confidence)
	}
}

// TestScoreComputation_CustomThresholds tests that custom
// thresholds from config are respected.
func TestScoreComputation_CustomThresholds(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Classification.Thresholds.Contractual = 90
	cfg.Classification.Thresholds.Incidental = 40

	signals := []taxonomy.Signal{
		{Source: "interface", Weight: 30},
		{Source: "visibility", Weight: 10},
	}

	c := classify.ComputeScore(signals, cfg)
	// 50 + 30 + 10 = 90 >= 90 = contractual with custom threshold.
	if c.Label != taxonomy.Contractual {
		t.Errorf("custom threshold: label = %q, want %q",
			c.Label, taxonomy.Contractual)
	}
}

// TestScoreComputation_Determinism verifies that identical inputs
// produce identical outputs (FR-011).
func TestScoreComputation_Determinism(t *testing.T) {
	signals := []taxonomy.Signal{
		{Source: "interface", Weight: 30},
		{Source: "naming", Weight: 10},
	}

	c1 := classify.ComputeScore(signals, nil)
	c2 := classify.ComputeScore(signals, nil)

	if c1.Label != c2.Label {
		t.Errorf("determinism: labels differ: %q vs %q", c1.Label, c2.Label)
	}
	if c1.Confidence != c2.Confidence {
		t.Errorf("determinism: confidence differs: %d vs %d",
			c1.Confidence, c2.Confidence)
	}
}

// TestClassify_ContractsPackage tests end-to-end classification on
// the contracts fixture package.
func TestClassify_ContractsPackage(t *testing.T) {
	allPkgs := loadTestPackages(t)
	contractsPkg := findPackage(allPkgs, "contracts")
	if contractsPkg == nil {
		t.Fatal("contracts package not found")
	}

	// Analyze the contracts package first.
	opts := analysis.Options{
		IncludeUnexported: false,
	}
	results, err := analysis.Analyze(contractsPkg, opts)
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("no analysis results for contracts package")
	}

	// Classify.
	classifyOpts := classify.Options{
		Config:         config.DefaultConfig(),
		ModulePackages: allPkgs,
		TargetPkg:      contractsPkg,
		Verbose:        true,
	}

	classified := classify.Classify(results, classifyOpts)

	// Verify that all side effects have classifications.
	for _, result := range classified {
		for _, se := range result.SideEffects {
			if se.Classification == nil {
				t.Errorf("function %s, effect %s: no classification",
					result.Target.Function, se.Type)
			}
		}
	}

	// Check that interface-implementing methods have high confidence.
	for _, result := range classified {
		if result.Target.Function == "Save" ||
			result.Target.Function == "Delete" ||
			result.Target.Function == "Write" {
			for _, se := range result.SideEffects {
				if se.Classification == nil {
					continue
				}
				if se.Classification.Confidence < 70 {
					t.Errorf("interface method %s.%s: confidence %d, want >= 70",
						result.Target.Function, se.Type,
						se.Classification.Confidence)
				}
			}
		}
	}
}

// TestClassify_IncidentalPackage tests that incidental effects
// are classified with low confidence.
func TestClassify_IncidentalPackage(t *testing.T) {
	allPkgs := loadTestPackages(t)
	incidentalPkg := findPackage(allPkgs, "incidental")
	if incidentalPkg == nil {
		t.Fatal("incidental package not found")
	}

	opts := analysis.Options{
		IncludeUnexported: true,
	}
	results, err := analysis.Analyze(incidentalPkg, opts)
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	classifyOpts := classify.Options{
		Config:         config.DefaultConfig(),
		ModulePackages: allPkgs,
		TargetPkg:      incidentalPkg,
		Verbose:        true,
	}

	classified := classify.Classify(results, classifyOpts)

	// Verify that all side effects have classifications.
	for _, result := range classified {
		for _, se := range result.SideEffects {
			if se.Classification == nil {
				t.Errorf("function %s, effect %s: no classification",
					result.Target.Function, se.Type)
				continue
			}
			// Incidental functions should not be classified as
			// contractual.
			if se.Classification.Label == taxonomy.Contractual {
				t.Errorf("incidental function %s, effect %s: "+
					"classified as contractual (confidence %d)",
					result.Target.Function, se.Type,
					se.Classification.Confidence)
			}
		}
	}
}

// TestClassify_Determinism verifies that classifying the same
// package twice produces identical results (FR-011).
func TestClassify_Determinism(t *testing.T) {
	allPkgs := loadTestPackages(t)
	contractsPkg := findPackage(allPkgs, "contracts")
	if contractsPkg == nil {
		t.Fatal("contracts package not found")
	}

	opts := analysis.Options{IncludeUnexported: false}
	results1, _ := analysis.Analyze(contractsPkg, opts)
	results2, _ := analysis.Analyze(contractsPkg, opts)

	classifyOpts := classify.Options{
		Config:         config.DefaultConfig(),
		ModulePackages: allPkgs,
		TargetPkg:      contractsPkg,
	}

	c1 := classify.Classify(results1, classifyOpts)
	c2 := classify.Classify(results2, classifyOpts)

	if len(c1) != len(c2) {
		t.Fatalf("determinism: result count differs: %d vs %d",
			len(c1), len(c2))
	}

	for i := range c1 {
		for j := range c1[i].SideEffects {
			se1 := c1[i].SideEffects[j]
			se2 := c2[i].SideEffects[j]

			if se1.Classification == nil || se2.Classification == nil {
				continue
			}

			if se1.Classification.Label != se2.Classification.Label {
				t.Errorf("determinism: function %s effect %s: "+
					"labels differ: %q vs %q",
					c1[i].Target.Function, se1.Type,
					se1.Classification.Label,
					se2.Classification.Label)
			}
			if se1.Classification.Confidence != se2.Classification.Confidence {
				t.Errorf("determinism: function %s effect %s: "+
					"confidence differs: %d vs %d",
					c1[i].Target.Function, se1.Type,
					se1.Classification.Confidence,
					se2.Classification.Confidence)
			}
		}
	}
}
