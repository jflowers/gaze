# Research: SSA Panic Recovery

**Feature**: 021-ssa-panic-recovery
**Date**: 2026-03-13

## R1: Logging Pattern for Recovery Warnings

### Decision
Use `fmt.Fprintf(stderr, "warning: ...")` for user-facing warnings and
`log.Debug(...)` (from `charmbracelet/log`) for the raw panic value at debug
level.

### Rationale
The gaze codebase has an established pattern of `fmt.Fprintf(stderr, "warning: ...")`
for user-facing warnings (see `internal/crap/coverage.go:73`,
`internal/aireport/output.go:26`). There is no structured logger in use at
runtime. The `charmbracelet/log` dependency is declared in `go.mod` but only
used by the TUI layer (Bubble Tea), not for analysis pipeline logging.

For the debug-level panic value, `charmbracelet/log` provides level-filtered
output and is already a dependency. However, introducing it in `internal/analysis`
and `internal/quality` would create a new coupling for a single log line.

**Alternative approach**: Use `log.Printf` from the standard library `log`
package (already imported transitively) with a conditional debug check. However,
the standard `log` package has no level filtering.

**Selected approach**: For both `BuildSSA` and `BuildTestSSA`, the functions do
not currently accept a stderr writer or logger parameter. Rather than changing
the function signatures (which would break all callers), the warning will be
emitted via Go's standard `log` package (`log.Printf`) for the warning level.
The raw panic value will be included in the same log line but formatted in a way
that is informational rather than alarming (e.g., `"SSA build skipped for
<pkg>: recovered from internal error"`). For debug-level detail, the raw panic
value will be included as a secondary `log.Printf` call that callers can
suppress by configuring the standard logger.

**Revised approach after further analysis**: The simplest correct approach is:
- `BuildSSA` already returns `nil` on failure — callers already handle this.
  Add `log.Printf("warning: SSA build skipped for %s: internal panic recovered",
  pkg.PkgPath)` in the recover block.
- `BuildTestSSA` already returns `error` — the error message carries the
  package path. Add the same `log.Printf` in the recover block.
- For the raw panic value at debug level, use a second `log.Printf` prefixed
  with `"debug: "`. This is the lightest-weight approach that requires no new
  imports, no signature changes, and no new dependencies.

### Alternatives Considered
1. **Add `io.Writer` parameter to `BuildSSA`/`BuildTestSSA`**: Would require
   updating all callers (4+ call sites each). Rejected — disproportionate
   churn for a single warning line.
2. **Use `charmbracelet/log`**: Would create new import in `internal/analysis`
   and `internal/quality`. Rejected — these packages are dependency-light by
   design.
3. **Use `slog` (Go 1.21+)**: Available since gaze targets Go 1.25. This is
   the most idiomatic choice for leveled logging. It provides `slog.Warn` and
   `slog.Debug` with structured fields. However, gaze doesn't use `slog`
   anywhere yet, and introducing it for 2 log lines may be premature.

## R2: recover() Placement Strategy

### Decision
Wrap `prog.Build()` in a helper function with deferred `recover()`. The
recovery converts the panic into a nil/error return.

### Rationale
Go's `recover()` only works inside a deferred function in the same goroutine.

> **Correction (issue #33):** The original assumption that `prog.Build()` is
> synchronous was incorrect. `ssa.Program.Build()` spawns one goroutine per
> package by default (`builder.go:3152`), and panics in child goroutines
> bypass `recover()` in the calling goroutine. The fix is to add
> `ssa.BuildSerially` to the builder mode flags, which forces `Build()` to
> run all SSA construction on the calling goroutine. With `BuildSerially`,
> the `recover()` strategy described here works correctly.

With `ssa.BuildSerially` set, `prog.Build()` runs in the calling goroutine, so a simple
`defer func() { if r := recover(); r != nil { ... } }()` at the top of the
function body works correctly.

The recovery must be placed **before** `prog.Build()` is called, and the
function's return values must be modifiable by the deferred function. Both
`BuildSSA` and `BuildTestSSA` use named or implicit return values that can be
set in the recovery path:

- `BuildSSA`: Returns `*ssa.Package`. Recovery sets it to `nil` (already the
  zero value). Use a named return to allow the deferred function to modify it.
- `BuildTestSSA`: Returns `(*ssa.Program, *ssa.Package, error)`. Recovery sets
  `error` to a descriptive message. Use named returns.

### Alternatives Considered
1. **Separate goroutine with channel**: Run `prog.Build()` in a goroutine and
   catch panics via `recover()` in that goroutine. Originally rejected as
   unnecessary complexity. In retrospect this would have worked, but
   `ssa.BuildSerially` is a simpler fix (issue #33).
2. **Wrapping `ssautil.AllPackages` too**: The panic is in `prog.Build()`, not
   in `ssautil.AllPackages`. Wrapping both would be defensive but unnecessary
   based on the current stack trace. Can be extended later if needed.

## R3: Testing Panic Recovery

### Decision
Test panic recovery by introducing a test-only mechanism that triggers a panic
during SSA building.

### Rationale
The actual panic is triggered by a specific combination of Go 1.25 + generic
variadic types in `go-json-experiment/json`. Reproducing this in a test would
require either:

1. **Loading the actual `go-json-experiment/json` package**: This creates a
   large, fragile test fixture that depends on external module resolution and
   would be slow. Not suitable for unit tests.
2. **Mocking `prog.Build()`**: `prog.Build()` is a method on `*ssa.Program`,
   which is a concrete type from `x/tools`. It cannot be mocked without an
   interface.
3. **Testing the recovery pattern in isolation**: Extract the panic-recovery
   wrapper into a testable helper function that accepts a `func()` to call.
   Test the helper with a deliberately panicking function.

**Selected approach**: Option 3 — extract a `safeSSABuild` helper (unexported)
that accepts `func()` and returns a recovered panic value (or nil). Test this
helper directly with:
- A function that completes normally → returns nil
- A function that panics with a string → returns the string
- A function that panics with an error → returns the error

Then `BuildSSA` and `BuildTestSSA` call `safeSSABuild(prog.Build)` and check
the result.

### Alternatives Considered
1. **testdata package that triggers the panic**: Would require crafting a Go
   package with the exact generic variadic pattern. Fragile — depends on the
   specific `x/tools` bug persisting. If `x/tools` fixes the bug upstream,
   the test would stop testing the recovery path.
2. **No unit test for recovery; rely on integration test**: Violates
   Constitution Principle IV (Testability) — every function must be testable
   in isolation.
