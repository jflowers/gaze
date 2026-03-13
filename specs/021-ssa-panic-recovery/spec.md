# Feature Specification: SSA Panic Recovery

**Feature Branch**: `021-ssa-panic-recovery`  
**Created**: 2026-03-13  
**Status**: Implemented  
**Input**: User description: "Add recover() guards around SSA prog.Build() to prevent panics from crashing gaze report under Go 1.25"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Graceful Degradation on SSA Panic (Priority: P1)

A developer runs `gaze report` (or `gaze quality`, `gaze analyze`) under Go 1.25
against a codebase whose dependency graph includes a package that triggers an SSA
builder panic (e.g., `github.com/go-json-experiment/json` with generic variadic
parameters). Instead of the process crashing with a stack trace, the analysis
completes for all other packages and the report is produced. The user sees a
warning identifying the package that was skipped due to the panic.

**Why this priority**: This is the core value of the change. Without it, `gaze
report` is completely unusable under Go 1.25 for affected codebases. Recovering
from the panic converts a total failure into a partial result that is still
actionable.

**Independent Test**: Can be fully tested by triggering a panic inside
`prog.Build()` (via a test double or a known-bad package) and verifying the
caller receives a nil/error return instead of a crash.

**Acceptance Scenarios**:

1. **Given** a Go module that transitively depends on a package causing an SSA
   builder panic, **When** the user runs `gaze report ./...`, **Then** the
   command completes without crashing, produces a report for all analyzable
   packages, and emits a warning identifying the skipped package.

2. **Given** a Go module that transitively depends on a package causing an SSA
   builder panic, **When** the user runs `gaze quality ./...`, **Then** quality
   assessment completes for all other packages and returns results, with a
   warning for the skipped package.

3. **Given** a Go module that transitively depends on a package causing an SSA
   builder panic, **When** the user runs `gaze analyze ./...`, **Then** side
   effect analysis completes for all other packages and returns results, with
   mutation data omitted only for the affected package.

---

### User Story 2 - Transparent Reporting of Skipped Packages (Priority: P2)

When a package is skipped due to an SSA panic, the user needs to understand what
happened and what is missing from their report. The warning message identifies
the package that caused the panic and explains that results for that package may
be incomplete (mutation analysis in `analyze`, quality assessment in `quality`/
`report`).

**Why this priority**: Without clear feedback, users may not realize parts of
their codebase were excluded from analysis, leading to false confidence in the
report's completeness.

**Independent Test**: Can be tested by triggering recovery and verifying the
warning message content includes the package path and a human-readable
explanation.

**Acceptance Scenarios**:

1. **Given** an SSA panic is recovered during analysis, **When** the warning is
   emitted, **Then** the message MUST include the package path that caused the
   panic and indicate that results for that package are incomplete.

2. **Given** an SSA panic is recovered during analysis, **When** the warning is
   emitted, **Then** the warning-level message MUST NOT include the raw panic
   value or stack trace. The raw panic value MUST be logged separately at debug
   level for developer troubleshooting.

---

### User Story 3 - No Impact on Unaffected Codebases (Priority: P1)

Developers running gaze against codebases that do not trigger SSA panics MUST
experience zero behavioral change. The recovery guard is a no-op in the
non-panic path and MUST NOT introduce observable performance overhead,
additional output, or altered results.

**Why this priority**: Existing users must not be affected. This is a
correctness invariant, not a feature.

**Independent Test**: Run the existing test suite and verify all tests pass
identically with the recovery guard in place.

**Acceptance Scenarios**:

1. **Given** a Go module whose packages do not trigger SSA panics, **When** the
   user runs any gaze command, **Then** output and behavior are identical to
   the version without recovery guards.

2. **Given** a Go module whose packages do not trigger SSA panics, **When** the
   user runs `gaze report`, **Then** no warnings about skipped packages appear.

---

### Edge Cases

- What happens when the panic-causing package is the *only* package being
  analyzed? The command completes with an empty result set and a warning. It
  does not crash.
- What happens when multiple packages in the dependency graph each trigger
  separate SSA panics? Each is independently recovered and warned about. Other
  packages are analyzed normally.
- What happens when the panic occurs not in `prog.Build()` but in a later SSA
  operation (e.g., walking SSA blocks)? This change addresses `prog.Build()`
  panics only. Later SSA operations are already bounded to individual functions
  and would need separate treatment if they panic (out of scope).

## Clarifications

### Session 2026-03-13

- Q: Should the raw panic value be logged at a different level for developer debugging, or discarded entirely? → A: Log at debug level (hidden by default, visible with verbose/debug flag).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `BuildSSA` MUST recover from panics during `prog.Build()` and
  return `nil` instead of propagating the panic.
- **FR-002**: `BuildTestSSA` MUST recover from panics during `prog.Build()` and
  return an error with a descriptive message instead of propagating the panic.
- **FR-003**: Both recovery guards MUST emit a warning-level log message
  identifying the package path that caused the panic.
- **FR-004**: The warning message MUST NOT expose the raw panic value or
  internal stack trace to users at the default log level. The raw panic value
  MUST be logged at debug level for developer troubleshooting.
- **FR-005**: The recovery guards MUST be no-ops (zero observable behavior
  change) when `prog.Build()` completes without panicking.
- **FR-006**: All existing callers of `BuildSSA` and `BuildTestSSA` MUST
  continue to handle nil/error returns correctly without regression.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `gaze report` completes without crashing when run against a
  codebase that triggers the Go 1.25 SSA builder panic, producing a report for
  all non-affected packages.
- **SC-002**: A warning is emitted for each package skipped due to SSA panic
  recovery, visible in the command's stderr output.
- **SC-003**: All existing tests pass identically with the recovery guards in
  place — zero regressions.
- **SC-004**: The recovery guard introduces no measurable performance overhead
  on the existing benchmark suite.
