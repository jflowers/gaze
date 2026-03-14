## Context

`crap.Analyze` (`internal/crap/analyze.go:61-166`) is a 6-step orchestrator with cyclomatic complexity 15. Steps 1-4 and Step 6 already delegate to extracted helpers. Step 5 (lines 104-157) is a 53-line monolithic loop that:

1. Iterates `gocyclo.Stat` entries from cyclomatic complexity analysis
2. Skips test files (`_test.go` suffix)
3. Skips generated files (with an `isGeneratedFile` cache)
4. Looks up line coverage from the coverage map
5. Computes CRAP score via `Formula(complexity, coverage)`
6. Optionally computes GazeCRAP via `ContractCoverageFunc` callback
7. Classifies into GazeCRAP quadrant
8. Builds a `Score` struct and appends to the result slice

This loop contains 10+ branches (the source of the high complexity) but is only testable through the full `Analyze` function, which requires a coverage profile file and real Go packages.

## Goals / Non-Goals

### Goals
- Extract Step 5 into `computeScores()` — directly testable with synthetic inputs
- Reduce `Analyze`'s cyclomatic complexity from ~15 to ~5
- Add unit tests for `computeScores()` covering all branches (generated files, test files, coverage lookup, GazeCRAP, quadrant classification)
- Improve line and contract coverage for the CRAP scoring logic

### Non-Goals
- Changing the public API of `crap.Analyze` (same signature, same behavior)
- Changing any output formats (JSON, text reports)
- Decomposing other functions (adapter Format methods, runner steps)
- Adding new features or metrics
- Refactoring callers (`cmd/gaze/main.go`, `internal/aireport/runner_steps.go`)

## Decisions

### D1: Extract `computeScores` as unexported function

```go
func computeScores(
    stats []gocyclo.Stat,
    coverMap map[string]float64,
    opts Options,
) []Score
```

The function takes complexity stats, a coverage map, and options (for `IgnoreGenerated`, `ContractCoverageFunc`, `Stderr`). It returns `[]Score`. The generated-file cache is local to this function (not shared state).

**Rationale**: Unexported because this is an internal implementation detail. Callers of `Analyze` don't need to know about the decomposition. The function is testable via internal-package tests (`crap` package test files) or via an `export_test.go` shim.

### D2: Use internal-package tests (not export_test.go)

Test `computeScores` from `crap_test.go` using `package crap` (internal test). This avoids creating an `export_test.go` shim and keeps tests close to the implementation. The existing test file (`crap_test.go`) already uses `package crap_test` (external), so a new internal test file (`analyze_internal_test.go` with `package crap`) is needed.

**Rationale**: `computeScores` is unexported, so external-package tests can't call it directly. An `export_test.go` shim would work but adds boilerplate. A `package crap` internal test file is simpler and follows the project's existing pattern (both internal and external test styles are used per AGENTS.md).

### D3: Preserve the generated-file cache as a local map

The current code uses a `generatedCache` map local to the `Analyze` function to avoid re-reading files for `isGeneratedFile` checks. The extracted `computeScores` will own this cache — it's created at the start of `computeScores` and scoped to the function call.

**Rationale**: No shared state, no global cache. Consistent with the project's "no global state" convention.

### D4: `gocyclo.Stat` is an external type — test with real struct literals

`gocyclo.Stat` is from `github.com/fzipp/gocyclo/v2`. Tests will construct `gocyclo.Stat` struct literals directly (the type has exported fields: `PkgName`, `FuncName`, `Complexity`, `Pos`). No mocking needed.

**Rationale**: `gocyclo.Stat` is a simple value type with exported fields. Constructing test data is straightforward.

## Risks / Trade-offs

### R1: Minimal risk — pure refactoring

This is a pure internal refactoring. No API changes, no behavioral changes, no output changes. The risk is limited to introducing a bug during extraction, which existing integration tests in `crap_test.go` will catch.

### R2: `gocyclo.Stat.Pos` uses `token.Position`

The `Pos` field contains `token.Position{Filename, Offset, Line, Column}`. The coverage map lookup in `lookupCoverage` uses `stat.Pos.Filename` and `stat.FuncName`. Tests need to construct `Pos` with at least `Filename` set. This is straightforward — `token.Position` is a simple struct with exported fields.
