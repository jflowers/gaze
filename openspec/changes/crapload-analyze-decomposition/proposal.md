## Why

`crap.Analyze` has a GazeCRAP score of 43.1 — the highest in the project. It has cyclomatic complexity 15 with approximately 50% line coverage and 50% contract coverage. This puts it squarely in Q4 (Dangerous) of the GazeCRAP quadrant: complex AND under-tested.

The function is already partially decomposed — Steps 1-4 and Step 6 call extracted helpers (`generateCoverProfile`, `resolvePatterns`, `ParseCoverProfile`, `buildCoverMap`, `buildSummary`). However, Step 5 (lines 104-157, 53 lines) is a monolithic loop that joins complexity stats with coverage data, filters generated files, computes CRAP scores, and optionally computes GazeCRAP. This loop contains the majority of the function's branching complexity and is difficult to test in isolation because it requires both `gocyclo.Stat` inputs and a coverage map.

Extracting Step 5 into a `computeScores` function will:
1. Reduce `Analyze`'s complexity from ~15 to ~5, moving it out of Q4
2. Make the score computation loop directly testable with synthetic inputs
3. Improve both line and contract coverage for the CRAP scoring logic
4. Follow the established decomposition pattern from spec 009 (`crapload-reduction`)

## What Changes

Extract the score computation loop (Step 5) from `crap.Analyze` into a separate `computeScores` function. `Analyze` becomes a clean orchestrator that calls extracted functions at each step.

## Capabilities

### New Capabilities
- `computeScores`: Unexported function that joins `gocyclo.Stat` complexity data with a coverage map, applies generated-file filtering, computes CRAP scores via `Formula`, and optionally computes GazeCRAP via `ContractCoverageFunc`. Directly testable with synthetic inputs.

### Modified Capabilities
- `crap.Analyze`: Reduced complexity. The function body becomes a 6-step orchestrator calling extracted helpers. No API surface change — same signature, same return type, same behavior.

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/crap/analyze.go` | Extract `computeScores()` from the Step 5 loop |
| `internal/crap/crap_test.go` | Add direct unit tests for `computeScores()` |
| `internal/crap/export_test.go` | New file to export `computeScores` for external-package tests (or use internal-package tests) |
| `AGENTS.md` | Update Recent Changes |

No changes to callers (`cmd/gaze/main.go`, `internal/aireport/runner_steps.go`). No API surface changes. No output format changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No changes to artifact formats or cross-hero interfaces. The refactoring is internal to the `crap` package. All JSON output and CLI behavior remain identical.

### II. Composability First

**Assessment**: PASS

`crap.Analyze` retains the same function signature. No new dependencies introduced. The extracted `computeScores` is an internal implementation detail — callers are unaffected.

### III. Observable Quality

**Assessment**: PASS

No changes to output formats, JSON schemas, or report structure. The refactoring improves internal code quality (lower complexity, higher testability) without altering any observable outputs.

### IV. Testability

**Assessment**: PASS

This is the primary motivation. The extracted `computeScores` function is directly testable with synthetic `gocyclo.Stat` inputs and in-memory coverage maps — no `go test` subprocess or filesystem access required. This directly addresses the GazeCRAP score by improving both test coverage and reducing complexity.
