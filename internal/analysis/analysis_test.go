package analysis_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jflowers/gaze/internal/analysis"
	"github.com/jflowers/gaze/internal/taxonomy"
	"golang.org/x/tools/go/packages"
)

// testdataPath returns the absolute path to a testdata fixture package.
func testdataPath(pkgName string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", "src", pkgName)
}

// loadTestdataPackage loads a testdata fixture package using the
// given directory. This is the shared implementation for both test
// and benchmark helpers.
func loadTestdataPackage(pkgName string) (*packages.Package, error) {
	testdataDir := testdataPath(pkgName)

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
		Dir:   testdataDir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages loaded for %q", pkgName)
	}
	if len(pkgs[0].Errors) > 0 {
		return nil, fmt.Errorf("package %q has errors: %v", pkgName, pkgs[0].Errors)
	}
	return pkgs[0], nil
}

// loadTestPackage loads one of the test fixture packages from testdata.
func loadTestPackage(t *testing.T, pkgName string) *packages.Package {
	t.Helper()
	pkg, err := loadTestdataPackage(pkgName)
	if err != nil {
		t.Fatalf("failed to load test package %q: %v", pkgName, err)
	}
	return pkg
}

// loadTestPackageBench loads one of the test fixture packages for benchmarks.
func loadTestPackageBench(b *testing.B, pkgName string) *packages.Package {
	b.Helper()
	pkg, err := loadTestdataPackage(pkgName)
	if err != nil {
		b.Fatalf("failed to load test package %q: %v", pkgName, err)
	}
	return pkg
}

// hasEffect checks if a side effect of the given type exists in the list.
func hasEffect(effects []taxonomy.SideEffect, typ taxonomy.SideEffectType) bool {
	for _, e := range effects {
		if e.Type == typ {
			return true
		}
	}
	return false
}

// countEffects counts effects of a given type.
func countEffects(effects []taxonomy.SideEffect, typ taxonomy.SideEffectType) int {
	count := 0
	for _, e := range effects {
		if e.Type == typ {
			count++
		}
	}
	return count
}

// effectWithTarget finds an effect by type and target string.
func effectWithTarget(effects []taxonomy.SideEffect, typ taxonomy.SideEffectType, target string) *taxonomy.SideEffect {
	for i, e := range effects {
		if e.Type == typ && e.Target == target {
			return &effects[i]
		}
	}
	return nil
}

// --- Return Analyzer Tests ---

func TestReturns_PureFunction(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "PureFunction")
	if fd == nil {
		t.Fatal("PureFunction not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)
	if len(result.SideEffects) != 0 {
		t.Errorf("PureFunction: expected 0 side effects, got %d: %v",
			len(result.SideEffects), result.SideEffects)
	}
}

func TestReturns_SingleReturn(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "SingleReturn")
	if fd == nil {
		t.Fatal("SingleReturn not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if count := countEffects(result.SideEffects, taxonomy.ReturnValue); count != 1 {
		t.Errorf("expected 1 ReturnValue, got %d", count)
	}
	e := effectWithTarget(result.SideEffects, taxonomy.ReturnValue, "int")
	if e == nil {
		t.Error("expected ReturnValue with target 'int'")
	}
}

func TestReturns_MultipleReturns(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "MultipleReturns")
	if fd == nil {
		t.Fatal("MultipleReturns not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if count := countEffects(result.SideEffects, taxonomy.ReturnValue); count != 2 {
		t.Errorf("expected 2 ReturnValue, got %d", count)
	}
}

func TestReturns_ErrorReturn(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "ErrorReturn")
	if fd == nil {
		t.Fatal("ErrorReturn not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if count := countEffects(result.SideEffects, taxonomy.ReturnValue); count != 1 {
		t.Errorf("expected 1 ReturnValue (int), got %d", count)
	}
	if count := countEffects(result.SideEffects, taxonomy.ErrorReturn); count != 1 {
		t.Errorf("expected 1 ErrorReturn, got %d", count)
	}
}

func TestReturns_ErrorOnly(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "ErrorOnly")
	if fd == nil {
		t.Fatal("ErrorOnly not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if count := countEffects(result.SideEffects, taxonomy.ErrorReturn); count != 1 {
		t.Errorf("expected 1 ErrorReturn, got %d", count)
	}
	if count := countEffects(result.SideEffects, taxonomy.ReturnValue); count != 0 {
		t.Errorf("expected 0 ReturnValue, got %d", count)
	}
}

func TestReturns_TripleReturn(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "TripleReturn")
	if fd == nil {
		t.Fatal("TripleReturn not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if count := countEffects(result.SideEffects, taxonomy.ReturnValue); count != 2 {
		t.Errorf("expected 2 ReturnValue (string, int), got %d", count)
	}
	if count := countEffects(result.SideEffects, taxonomy.ErrorReturn); count != 1 {
		t.Errorf("expected 1 ErrorReturn, got %d", count)
	}
}

func TestReturns_NamedReturns(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "NamedReturns")
	if fd == nil {
		t.Fatal("NamedReturns not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if count := countEffects(result.SideEffects, taxonomy.ReturnValue); count != 1 {
		t.Errorf("expected 1 ReturnValue ([]byte), got %d", count)
	}
	if count := countEffects(result.SideEffects, taxonomy.ErrorReturn); count != 1 {
		t.Errorf("expected 1 ErrorReturn, got %d", count)
	}

	// Verify named return metadata in description.
	for _, e := range result.SideEffects {
		if e.Type == taxonomy.ReturnValue {
			if e.Description == "" {
				t.Error("expected non-empty description for named return")
			}
		}
	}
}

func TestReturns_NamedReturnModifiedInDefer(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "NamedReturnModifiedInDefer")
	if fd == nil {
		t.Fatal("NamedReturnModifiedInDefer not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if !hasEffect(result.SideEffects, taxonomy.DeferredReturnMutation) {
		t.Error("expected DeferredReturnMutation for named return 'err' modified in defer")
	}
	if !hasEffect(result.SideEffects, taxonomy.ErrorReturn) {
		t.Error("expected ErrorReturn")
	}
}

func TestReturns_InterfaceReturn(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "InterfaceReturn")
	if fd == nil {
		t.Fatal("InterfaceReturn not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if count := countEffects(result.SideEffects, taxonomy.ReturnValue); count != 1 {
		t.Errorf("expected 1 ReturnValue (io.Reader), got %d", count)
	}
}

// --- Sentinel Analyzer Tests ---

func TestSentinels_Detection(t *testing.T) {
	pkg := loadTestPackage(t, "sentinel")

	results, err := analysis.Analyze(pkg, analysis.Options{
		IncludeUnexported: true,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Collect all sentinel effects across all results.
	var sentinels []taxonomy.SideEffect
	for _, r := range results {
		for _, e := range r.SideEffects {
			if e.Type == taxonomy.SentinelError {
				sentinels = append(sentinels, e)
			}
		}
	}

	// Should detect: ErrNotFound, ErrPermission, ErrWrapped, errUnexported
	expectedSentinels := map[string]bool{
		"ErrNotFound":   false,
		"ErrPermission": false,
		"ErrWrapped":    false,
		"errUnexported": false,
	}

	for _, s := range sentinels {
		if _, ok := expectedSentinels[s.Target]; ok {
			expectedSentinels[s.Target] = true
		}
	}

	for name, found := range expectedSentinels {
		if !found {
			t.Errorf("expected sentinel %q not detected", name)
		}
	}

	// Should NOT detect NotAnError.
	for _, s := range sentinels {
		if s.Target == "NotAnError" {
			t.Error("NotAnError should not be detected as a sentinel")
		}
	}
}

func TestSentinels_WrappedDetection(t *testing.T) {
	pkg := loadTestPackage(t, "sentinel")

	results, err := analysis.Analyze(pkg, analysis.Options{
		IncludeUnexported: true,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	var wrapped *taxonomy.SideEffect
	for _, r := range results {
		for i, e := range r.SideEffects {
			if e.Type == taxonomy.SentinelError && e.Target == "ErrWrapped" {
				wrapped = &r.SideEffects[i]
			}
		}
	}

	if wrapped == nil {
		t.Fatal("ErrWrapped not detected")
	}
	if wrapped.Description == "" {
		t.Error("expected non-empty description for ErrWrapped")
	}
}

// --- Mutation Analyzer Tests ---

func TestMutation_PointerReceiverIncrement(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Counter", "Increment")
	if fd == nil {
		t.Fatal("(*Counter).Increment not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	e := effectWithTarget(result.SideEffects, taxonomy.ReceiverMutation, "count")
	if e == nil {
		t.Error("expected ReceiverMutation for field 'count'")
	}
}

func TestMutation_PointerReceiverSetName(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Counter", "SetName")
	if fd == nil {
		t.Fatal("(*Counter).SetName not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	e := effectWithTarget(result.SideEffects, taxonomy.ReceiverMutation, "name")
	if e == nil {
		t.Error("expected ReceiverMutation for field 'name'")
	}
}

func TestMutation_PointerReceiverSetBoth(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Counter", "SetBoth")
	if fd == nil {
		t.Fatal("(*Counter).SetBoth not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if countEffects(result.SideEffects, taxonomy.ReceiverMutation) != 2 {
		t.Errorf("expected 2 ReceiverMutation effects, got %d",
			countEffects(result.SideEffects, taxonomy.ReceiverMutation))
	}
	if effectWithTarget(result.SideEffects, taxonomy.ReceiverMutation, "count") == nil {
		t.Error("expected ReceiverMutation for field 'count'")
	}
	if effectWithTarget(result.SideEffects, taxonomy.ReceiverMutation, "name") == nil {
		t.Error("expected ReceiverMutation for field 'name'")
	}
}

func TestMutation_ValueReceiverNoMutation(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "Counter", "Value")
	if fd == nil {
		t.Fatal("(Counter).Value not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if hasEffect(result.SideEffects, taxonomy.ReceiverMutation) {
		t.Error("value receiver should NOT report ReceiverMutation")
	}
	// But it should still report ReturnValue.
	if !hasEffect(result.SideEffects, taxonomy.ReturnValue) {
		t.Error("expected ReturnValue for Value()")
	}
}

func TestMutation_ValueReceiverTrap(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "Counter", "ValueReceiverTrap")
	if fd == nil {
		t.Fatal("(Counter).ValueReceiverTrap not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if hasEffect(result.SideEffects, taxonomy.ReceiverMutation) {
		t.Error("value receiver copy mutation should NOT report ReceiverMutation")
	}
}

func TestMutation_PointerArgument(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindFuncDecl(pkg, "Normalize")
	if fd == nil {
		t.Fatal("Normalize not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	e := effectWithTarget(result.SideEffects, taxonomy.PointerArgMutation, "v")
	if e == nil {
		t.Error("expected PointerArgMutation for argument 'v'")
	}
}

func TestMutation_PointerArgFillSlice(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindFuncDecl(pkg, "FillSlice")
	if fd == nil {
		t.Fatal("FillSlice not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	e := effectWithTarget(result.SideEffects, taxonomy.PointerArgMutation, "dst")
	if e == nil {
		t.Error("expected PointerArgMutation for argument 'dst'")
	}
}

func TestMutation_PointerArgReadOnly(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindFuncDecl(pkg, "ReadOnly")
	if fd == nil {
		t.Fatal("ReadOnly not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if hasEffect(result.SideEffects, taxonomy.PointerArgMutation) {
		t.Error("ReadOnly should NOT report PointerArgMutation (read-only access)")
	}
	// But should report ReturnValue.
	if !hasEffect(result.SideEffects, taxonomy.ReturnValue) {
		t.Error("expected ReturnValue for ReadOnly()")
	}
}

func TestMutation_NestedFieldMutation(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Config", "UpdateConfig")
	if fd == nil {
		t.Fatal("(*Config).UpdateConfig not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	e := effectWithTarget(result.SideEffects, taxonomy.ReceiverMutation, "Timeout")
	if e == nil {
		t.Error("expected ReceiverMutation for field 'Timeout'")
	}
}

func TestMutation_DeepNestedMutation(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Config", "UpdateNested")
	if fd == nil {
		t.Fatal("(*Config).UpdateNested not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	// Should report mutation to the top-level field 'Nested'.
	e := effectWithTarget(result.SideEffects, taxonomy.ReceiverMutation, "Nested")
	if e == nil {
		t.Error("expected ReceiverMutation for field 'Nested' (deep nested mutation)")
	}
}

// --- Analysis Metadata Tests ---

func TestAnalysis_MetadataPopulated(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "SingleReturn")
	if fd == nil {
		t.Fatal("SingleReturn not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if result.Metadata.GazeVersion == "" {
		t.Error("expected non-empty GazeVersion")
	}
	if result.Metadata.GoVersion == "" {
		t.Error("expected non-empty GoVersion")
	}
}

func TestAnalysis_TargetPopulated(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "SingleReturn")
	if fd == nil {
		t.Fatal("SingleReturn not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if result.Target.Function != "SingleReturn" {
		t.Errorf("expected function name 'SingleReturn', got %q",
			result.Target.Function)
	}
	if result.Target.Location == "" {
		t.Error("expected non-empty location")
	}
	if result.Target.Signature == "" {
		t.Error("expected non-empty signature")
	}
}

func TestAnalysis_MethodTargetHasReceiver(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Counter", "Increment")
	if fd == nil {
		t.Fatal("(*Counter).Increment not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	if result.Target.Receiver != "*Counter" {
		t.Errorf("expected receiver '*Counter', got %q",
			result.Target.Receiver)
	}
}

// --- Side Effect ID Tests ---

func TestAnalysis_StableIDs(t *testing.T) {
	pkg := loadTestPackage(t, "returns")
	fd := analysis.FindFuncDecl(pkg, "ErrorReturn")
	if fd == nil {
		t.Fatal("ErrorReturn not found")
	}

	result1 := analysis.AnalyzeFunction(pkg, fd)
	result2 := analysis.AnalyzeFunction(pkg, fd)

	if len(result1.SideEffects) != len(result2.SideEffects) {
		t.Fatalf("different side effect counts: %d vs %d",
			len(result1.SideEffects), len(result2.SideEffects))
	}

	for i := range result1.SideEffects {
		if result1.SideEffects[i].ID != result2.SideEffects[i].ID {
			t.Errorf("unstable ID for effect %d: %q vs %q",
				i, result1.SideEffects[i].ID, result2.SideEffects[i].ID)
		}
	}
}

// --- Analyze() option tests ---

func TestAnalyze_ExportedOnlyByDefault(t *testing.T) {
	pkg := loadTestPackage(t, "returns")

	results, err := analysis.Analyze(pkg, analysis.Options{
		IncludeUnexported: false,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, r := range results {
		if r.Target.Function == "<package>" {
			continue
		}
		// All returned functions should be exported.
		if len(r.Target.Function) > 0 {
			first := r.Target.Function[0]
			if first >= 'a' && first <= 'z' {
				t.Errorf("unexported function %q should not appear with IncludeUnexported=false",
					r.Target.Function)
			}
		}
	}
}

func TestAnalyze_FunctionFilter(t *testing.T) {
	pkg := loadTestPackage(t, "returns")

	results, err := analysis.Analyze(pkg, analysis.Options{
		IncludeUnexported: true,
		FunctionFilter:    "SingleReturn",
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// FunctionFilter also suppresses sentinel analysis.
	if len(results) != 1 {
		t.Fatalf("expected 1 result with FunctionFilter, got %d", len(results))
	}
	if results[0].Target.Function != "SingleReturn" {
		t.Errorf("expected 'SingleReturn', got %q", results[0].Target.Function)
	}
}

// --- All Tiers are P0 ---

func TestAnalysis_AllP0EffectsAreP0(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Counter", "Increment")
	if fd == nil {
		t.Fatal("(*Counter).Increment not found")
	}
	result := analysis.AnalyzeFunction(pkg, fd)

	for _, e := range result.SideEffects {
		if e.Type == taxonomy.ReceiverMutation && e.Tier != taxonomy.TierP0 {
			t.Errorf("ReceiverMutation should be P0, got %s", e.Tier)
		}
	}
}
