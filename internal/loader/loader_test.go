package loader_test

import (
	"go/token"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jflowers/gaze/internal/loader"
	"golang.org/x/tools/go/packages"
)

// testdataDir returns the absolute path to the analysis testdata
// fixtures, which contain valid Go packages for loading.
func testdataDir(t *testing.T, pkgName string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "analysis", "testdata", "src", pkgName)
}

// loadReturnsPackage loads the "returns" test fixture package.
func loadReturnsPackage(t *testing.T) *loader.Result {
	t.Helper()
	dir := testdataDir(t, "returns")

	// Use packages.Load directly (same as analysis tests) since
	// loader.Load doesn't accept a Dir option.
	cfg := &packages.Config{
		Mode:  loader.LoadMode,
		Dir:   dir,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		t.Fatalf("failed to load returns package: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
	if len(pkgs[0].Errors) > 0 {
		t.Fatalf("package has errors: %v", pkgs[0].Errors)
	}
	return &loader.Result{
		Pkg:  pkgs[0],
		Fset: pkgs[0].Fset,
	}
}

func TestLoad_ValidPackage(t *testing.T) {
	// Load the loader package itself (it's a valid Go package).
	result, err := loader.Load("github.com/jflowers/gaze/internal/loader")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if result.Pkg == nil {
		t.Fatal("expected non-nil Pkg")
	}
	if result.Fset == nil {
		t.Fatal("expected non-nil Fset")
	}
	if result.Pkg.PkgPath != "github.com/jflowers/gaze/internal/loader" {
		t.Errorf("expected pkg path 'github.com/jflowers/gaze/internal/loader', got %q",
			result.Pkg.PkgPath)
	}
}

func TestLoad_InvalidPattern(t *testing.T) {
	_, err := loader.Load("github.com/nonexistent/package/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent package")
	}
}

func TestFindFunction_ExistingFunction(t *testing.T) {
	result := loadReturnsPackage(t)

	fi := result.FindFunction("SingleReturn")
	if fi == nil {
		t.Fatal("FindFunction('SingleReturn') returned nil")
	}
	if fi.Decl == nil {
		t.Error("expected non-nil Decl")
	}
	if fi.Obj == nil {
		t.Error("expected non-nil Obj")
	}
	if fi.Obj.Name() != "SingleReturn" {
		t.Errorf("expected Obj.Name() = 'SingleReturn', got %q", fi.Obj.Name())
	}
	if fi.Pkg == nil {
		t.Error("expected non-nil Pkg")
	}
}

func TestFindFunction_NotFound(t *testing.T) {
	result := loadReturnsPackage(t)

	fi := result.FindFunction("DoesNotExist")
	if fi != nil {
		t.Error("expected nil for nonexistent function")
	}
}

func TestFindFunction_PureFunction(t *testing.T) {
	result := loadReturnsPackage(t)

	fi := result.FindFunction("PureFunction")
	if fi == nil {
		t.Fatal("FindFunction('PureFunction') returned nil")
	}
	if fi.Decl.Name.Name != "PureFunction" {
		t.Errorf("expected Decl.Name.Name = 'PureFunction', got %q",
			fi.Decl.Name.Name)
	}
}

func TestAllFunctions_ExportedOnly(t *testing.T) {
	result := loadReturnsPackage(t)

	funcs := result.AllFunctions(true)
	if len(funcs) == 0 {
		t.Fatal("expected at least one exported function")
	}
	for _, fi := range funcs {
		if !fi.Decl.Name.IsExported() {
			t.Errorf("AllFunctions(true) returned unexported function %q",
				fi.Decl.Name.Name)
		}
	}
}

func TestAllFunctions_IncludeAll(t *testing.T) {
	result := loadReturnsPackage(t)

	all := result.AllFunctions(false)
	exported := result.AllFunctions(true)

	// The returns package has only exported functions, so counts
	// should match. But the important thing is all >= exported.
	if len(all) < len(exported) {
		t.Errorf("AllFunctions(false) returned fewer (%d) than AllFunctions(true) (%d)",
			len(all), len(exported))
	}
}

func TestAllFunctions_ContainsExpectedFunctions(t *testing.T) {
	result := loadReturnsPackage(t)

	funcs := result.AllFunctions(false)
	names := make(map[string]bool)
	for _, fi := range funcs {
		names[fi.Decl.Name.Name] = true
	}

	expected := []string{
		"PureFunction",
		"SingleReturn",
		"MultipleReturns",
		"ErrorReturn",
		"ErrorOnly",
		"TripleReturn",
		"NamedReturns",
		"NamedReturnModifiedInDefer",
		"InterfaceReturn",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("AllFunctions() missing expected function %q", name)
		}
	}
}

func TestFormatPos_ValidPosition(t *testing.T) {
	result := loadReturnsPackage(t)

	fi := result.FindFunction("SingleReturn")
	if fi == nil {
		t.Fatal("SingleReturn not found")
	}

	pos := result.FormatPos(fi.Decl.Pos())
	if pos == "<unknown>" {
		t.Error("expected valid position string, got <unknown>")
	}
	// Should contain file path, line, and column.
	if pos == "" {
		t.Error("expected non-empty position string")
	}
}

func TestFormatPos_InvalidPosition(t *testing.T) {
	result := loadReturnsPackage(t)

	pos := result.FormatPos(token.NoPos)
	if pos != "<unknown>" {
		t.Errorf("expected '<unknown>' for invalid position, got %q", pos)
	}
}
