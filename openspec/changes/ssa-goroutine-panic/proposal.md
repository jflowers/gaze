## Why

The `recover()` guards added in spec 021 (issue #29) do not actually catch the SSA builder panic. The panic occurs inside a **goroutine spawned by `ssa.Program.Build()`**, not in the calling goroutine. Go's `recover()` is goroutine-scoped and cannot propagate across goroutine boundaries.

The root cause: `ssa.Program.Build()` in `golang.org/x/tools@v0.43.0` spawns one goroutine per package by default (see `builder.go:3152`). The panic from `go/types.NewSignatureType` with `go-json-experiment/json` generic variadic parameters happens in a child goroutine, bypassing the `safeSSABuild` wrapper entirely.

This means gaze v1.2.11 still crashes in CI on Go 1.25 when analyzing projects that depend on `go-json-experiment/json`. The graceful degradation from issue #30 never triggers because the process terminates before `Assess` can handle the error.

Tracked as GitHub issue #33.

## What Changes

Add `ssa.BuildSerially` to the SSA builder mode flags in both `BuildSSA` and `BuildTestSSA`. This forces `ssa.Program.Build()` to build packages sequentially on the calling goroutine instead of spawning child goroutines. With serial build, the existing `safeSSABuild` `recover()` guard correctly catches the panic.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `BuildSSA` (`internal/analysis/mutation.go`): SSA build mode changes from `ssa.InstantiateGenerics` to `ssa.InstantiateGenerics | ssa.BuildSerially`. Panics from `prog.Build()` are now reliably caught by `safeSSABuild` because all work runs on the calling goroutine.
- `BuildTestSSA` (`internal/quality/pairing.go`): Same change. The `recover()` guard now works as originally intended, and the graceful degradation path (issue #30) activates correctly.

### Removed Capabilities
- None.

## Impact

- **`internal/analysis/mutation.go`**: Change `ssa.InstantiateGenerics` to `ssa.InstantiateGenerics | ssa.BuildSerially` in `ssautil.AllPackages` call.
- **`internal/quality/pairing.go`**: Same change.
- **Performance**: SSA build loses parallelism. For gaze's use case (analyzing one package at a time), the practical impact is minimal -- `ssautil.AllPackages` receives a single input package, and the parallelism was across that package's transitive dependencies. Correctness (not crashing) takes priority over build speed.
- **Spec 021 research**: The assumption that `prog.Build()` is synchronous (research.md R2) was incorrect. The `BuildSerially` fix validates the original `recover()` approach without requiring architectural changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No changes to artifact formats, output schemas, or cross-hero interfaces. The fix is internal to the SSA construction layer.

### II. Composability First

**Assessment**: PASS

Both `BuildSSA` and `BuildTestSSA` remain standalone functions with the same signatures. No new dependencies introduced -- `ssa.BuildSerially` is an existing constant in `golang.org/x/tools/go/ssa`.

### III. Observable Quality

**Assessment**: PASS

The fix ensures the graceful degradation path (issue #30) actually activates, producing machine-parseable `SSADegraded: true` output instead of a process crash. This improves observable quality -- consumers get partial results rather than nothing.

### IV. Testability

**Assessment**: PASS

The existing `safeSSABuild` tests and `BuildSSAFunc` injection point remain valid. The `BuildSerially` flag is a transparent configuration change that doesn't affect the testability seams. A new integration-style test can verify that the panic is caught with the serial build mode.
