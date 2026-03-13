package quality_test

import (
	"errors"
	"testing"

	"github.com/unbound-force/gaze/internal/quality"
)

// ---------------------------------------------------------------------------
// safeSSABuild tests (quality package copy)
//
// safeSSABuild is duplicated in internal/quality because Go's package
// system does not allow importing unexported symbols across internal
// packages. The duplication is intentional and bounded to this one
// 6-line function — see specs/021-ssa-panic-recovery/research.md R3.
// ---------------------------------------------------------------------------

// TestSafeSSABuild_NoPanic verifies the quality package's
// safeSSABuild returns nil for non-panicking functions.
func TestSafeSSABuild_NoPanic(t *testing.T) {
	result := quality.SafeSSABuild(func() {})
	if result != nil {
		t.Errorf("safeSSABuild returned %v, want nil", result)
	}
}

// TestSafeSSABuild_PanicString verifies the quality package's
// safeSSABuild recovers a string panic.
func TestSafeSSABuild_PanicString(t *testing.T) {
	result := quality.SafeSSABuild(func() {
		panic("test panic")
	})
	s, ok := result.(string)
	if !ok {
		t.Fatalf("safeSSABuild returned %T, want string", result)
	}
	if s != "test panic" {
		t.Errorf("safeSSABuild returned %q, want %q", s, "test panic")
	}
}

// TestSafeSSABuild_PanicError verifies the quality package's
// safeSSABuild recovers an error-typed panic.
func TestSafeSSABuild_PanicError(t *testing.T) {
	errPanic := errors.New("SSA builder error")
	result := quality.SafeSSABuild(func() {
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
// SC-001: BuildTestSSA panic recovery contract
//
// Note: BuildTestSSA's panic recovery cannot be tested end-to-end
// because prog.Build() is a concrete method on *ssa.Program that
// cannot be mocked. The recovery pattern is verified through the
// safeSSABuild helper tests above. BuildTestSSA's logging behavior
// is verified by code inspection.
// ---------------------------------------------------------------------------
