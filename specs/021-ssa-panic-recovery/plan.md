# Implementation Plan: SSA Panic Recovery

**Branch**: `021-ssa-panic-recovery` | **Date**: 2026-03-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/021-ssa-panic-recovery/spec.md`

## Summary

Add `recover()` guards around the two `prog.Build()` call sites in the gaze
codebase (`analysis.BuildSSA` and `quality.BuildTestSSA`) to convert SSA builder
panics into graceful nil/error returns. This prevents `gaze report`, `gaze
quality`, and `gaze analyze` from crashing under Go 1.25 when the target module's
dependency graph includes packages that trigger upstream SSA type substitution
bugs. Warning-level messages identify skipped packages; debug-level messages
capture the raw panic value for developer troubleshooting.

## Technical Context

**Language/Version**: Go 1.25.0 (module minimum; `go.mod` directive)
**Primary Dependencies**: `golang.org/x/tools@v0.43.0` (SSA builder), `charmbracelet/log` (available but not currently used for runtime logging — project uses `fmt.Fprintf(stderr, ...)`)
**Storage**: N/A (no persistence changes)
**Testing**: Standard library `testing` package; `go test -race -count=1`
**Target Platform**: Cross-platform (darwin/linux x amd64/arm64)
**Project Type**: CLI tool (`cmd/gaze`)
**Performance Goals**: Zero measurable overhead in the non-panic path
**Constraints**: No new external dependencies; must pass existing test suite and benchmarks identically
**Scale/Scope**: 2 functions modified (`BuildSSA`, `BuildTestSSA`), ~10 lines of new code each, plus tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Accuracy — PASS

The recovery guards do not alter analysis logic. When `prog.Build()` succeeds,
behavior is identical. When it panics, the alternative is a process crash with
no output at all — returning nil/error and skipping the affected package is
strictly more accurate than producing zero results. The warning makes the skip
visible so users know their report is partial.

### II. Minimal Assumptions — PASS

No new assumptions about user code, test framework, or coding style. The
recovery is internal to gaze's SSA construction. Users do not need to annotate,
restructure, or change anything. The only assumption is that a nil SSA package
means "SSA unavailable" — this is already the existing contract of `BuildSSA`.

### III. Actionable Output — PASS

The warning message identifies the specific package that was skipped, giving
users actionable information (they can investigate that dependency, upgrade
toolchains, or file upstream bugs). The report still contains results for all
other packages.

### IV. Testability — PASS

Both `BuildSSA` and `BuildTestSSA` are independently testable functions with
clear input/output contracts. The recovery behavior can be tested by injecting
a panic-triggering condition and verifying the return value. Coverage strategy:

- **Unit tests**: Test `BuildSSA` and `BuildTestSSA` panic recovery directly
  (new tests in existing `*_test.go` files)
- **Integration tests**: Verify existing callers handle nil/error returns
  (already covered by existing tests — FR-006)
- **Regression tests**: Full test suite must pass identically (SC-003)
- **Target**: 100% branch coverage of the new `recover()` paths

No constitution violations. Gate passes.

## Project Structure

### Documentation (this feature)

```text
specs/021-ssa-panic-recovery/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── analysis/
│   ├── mutation.go          # BuildSSA — add recover() guard
│   └── mutation_test.go     # Add panic recovery tests
└── quality/
    ├── pairing.go           # BuildTestSSA — add recover() guard
    └── pairing_test.go      # Add panic recovery tests
```

**Structure Decision**: Modifications are scoped to 2 existing files in 2
existing packages (`internal/analysis`, `internal/quality`). No new packages,
directories, or files are required for production code. Test files are co-located
per existing convention.
