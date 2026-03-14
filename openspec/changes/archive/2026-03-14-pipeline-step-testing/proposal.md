## Why

`runProductionPipeline` (`internal/aireport/runner.go:196`) has a CRAP score of 42 with 0% line coverage. It is the orchestrator that calls the four analysis step functions (`runCRAPStep`, `runQualityStep`, `runClassifyStep`, `runDocscanStep`) and assembles the `ReportPayload`. It has important partial-failure semantics: if one step fails, the others still run and the error is captured in `payload.Errors`.

This orchestration logic is completely untested because `runProductionPipeline` directly calls the step functions, which in turn call heavy external dependencies (`crap.Analyze`, `analysis.LoadAndAnalyze`, `quality.Assess`, `classify.Classify`, `docscan.Scan`). The only existing test is an integration test guarded by `testing.Short()`.

The project already has an established dependency injection pattern: `RunnerOptions.AnalyzeFunc` allows `Run()` to bypass `runProductionPipeline` entirely. Extending this pattern to inject the four step functions makes the orchestration logic unit-testable with synthetic inputs.

## What Changes

Add four step function fields to a `pipelineStepFuncs` struct that `runProductionPipeline` accepts. When nil, each defaults to the real step function. Tests inject fake step functions that return synthetic results or errors, enabling unit tests for:
- All steps succeed (happy path)
- Individual step failures (error captured, other steps still run)
- Multiple simultaneous failures
- Empty patterns (returns error before calling steps)

## Capabilities

### New Capabilities
- `pipelineStepFuncs`: Internal struct holding injectable step function references for `runProductionPipeline`. Enables unit testing of the orchestration logic without running real analysis.

### Modified Capabilities
- `runProductionPipeline`: Accepts `pipelineStepFuncs` parameter. When step functions are nil, defaults to the real implementations. No behavioral change for production callers.
- `RunnerOptions`: No change — the existing `AnalyzeFunc` nil-check closure that calls `runProductionPipeline` passes the zero-valued `pipelineStepFuncs` (all nil = all default).

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/aireport/runner.go` | Add `pipelineStepFuncs` struct, update `runProductionPipeline` signature to accept it, thread through from `Run()` |
| `internal/aireport/runner_test.go` or new file | Unit tests for `runProductionPipeline` with injected step functions |
| `AGENTS.md` | Update Recent Changes |

No changes to step function signatures. No changes to `RunnerOptions` public API. No output format changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No changes to artifact formats or cross-hero interfaces. The `pipelineStepFuncs` struct is package-private and internal to the orchestration layer.

### II. Composability First

**Assessment**: PASS

`runProductionPipeline` retains its current behavior for all production callers. The injection points are additive — existing code that passes zero-valued `pipelineStepFuncs` gets identical behavior to before.

### III. Observable Quality

**Assessment**: PASS

No changes to output formats, JSON schemas, or report structure. The change improves internal testability without altering any observable outputs.

### IV. Testability

**Assessment**: PASS

This is the primary motivation. The `pipelineStepFuncs` injection follows the established `AnalyzeFunc` pattern and enables testing the orchestration logic (partial failures, error capture, payload assembly) without spawning `go test` subprocesses or loading real Go packages.
