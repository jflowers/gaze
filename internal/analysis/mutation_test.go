package analysis_test

import (
	"errors"
	"testing"

	"github.com/unbound-force/gaze/internal/analysis"
)

// ---------------------------------------------------------------------------
// safeSSABuild tests
// ---------------------------------------------------------------------------

// TestSafeSSABuild_NoPanic verifies that safeSSABuild returns nil
// when the build function completes without panicking.
func TestSafeSSABuild_NoPanic(t *testing.T) {
	result := analysis.SafeSSABuild(func() {
		// no panic
	})
	if result != nil {
		t.Errorf("safeSSABuild returned %v, want nil for non-panicking function", result)
	}
}

// TestSafeSSABuild_PanicString verifies that safeSSABuild recovers
// a panic with a string value and returns it.
func TestSafeSSABuild_PanicString(t *testing.T) {
	result := analysis.SafeSSABuild(func() {
		panic("test panic message")
	})
	s, ok := result.(string)
	if !ok {
		t.Fatalf("safeSSABuild returned %T, want string", result)
	}
	if s != "test panic message" {
		t.Errorf("safeSSABuild returned %q, want %q", s, "test panic message")
	}
}

// TestSafeSSABuild_PanicError verifies that safeSSABuild recovers
// a panic with an error value and returns it.
func TestSafeSSABuild_PanicError(t *testing.T) {
	errPanic := errors.New("SSA builder error")
	result := analysis.SafeSSABuild(func() {
		panic(errPanic)
	})
	e, ok := result.(error)
	if !ok {
		t.Fatalf("safeSSABuild returned %T, want error", result)
	}
	if e != errPanic {
		t.Errorf("safeSSABuild returned error %v, want %v", e, errPanic)
	}
}

// ---------------------------------------------------------------------------
// SC-001 / SC-002: panic recovery contract tests
//
// Note: BuildSSA's panic recovery cannot be tested end-to-end because
// prog.Build() is a concrete method on *ssa.Program that cannot be
// mocked or injected. The recovery pattern is verified through the
// safeSSABuild helper tests above (which exercise the identical
// defer/recover logic). BuildSSA's logging behavior is verified by
// code inspection — the log.Printf calls are co-located with the
// safeSSABuild call in the same if-block.
// ---------------------------------------------------------------------------

// TestSC001_BuildSSANoPanicReturnsPackage verifies that BuildSSA
// returns a non-nil *ssa.Package for a valid input package when no
// panic occurs (the normal path). This confirms the recover() guard
// is a no-op in the non-panic case (SC-001, FR-005).
func TestSC001_BuildSSANoPanicReturnsPackage(t *testing.T) {
	pkg := loadTestPackage(t, "mutation")

	ssaPkg := analysis.BuildSSA(pkg)
	if ssaPkg == nil {
		t.Fatal("BuildSSA returned nil for a valid package — recover() guard may have interfered")
	}

	if _, ok := ssaPkg.Members["Normalize"]; !ok {
		t.Error("expected 'Normalize' in SSA members after BuildSSA")
	}
}
