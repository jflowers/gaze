## Context

`runProductionPipeline` (`internal/aireport/runner.go:196-240`) orchestrates the four analysis steps of `gaze report`. It has critical partial-failure semantics: each step runs independently, and failures are captured in `payload.Errors` rather than short-circuiting the pipeline. This orchestration logic has zero test coverage because all four steps call heavy external dependencies.

The project has an established DI pattern: `RunnerOptions.AnalyzeFunc` (line 60) replaces the entire pipeline for `Run()` tests. This change extends the pattern one level deeper — injecting individual step functions into `runProductionPipeline`.

## Goals / Non-Goals

### Goals
- Make `runProductionPipeline` unit-testable with synthetic step function injections
- Test partial-failure semantics (one step fails, others succeed)
- Test error capture in `payload.Errors`
- Test empty-patterns validation
- Follow the established `AnalyzeFunc` nil-check defaulting pattern

### Non-Goals
- Testing individual step functions (`runCRAPStep`, `runQualityStep`, etc.) — those need their own DI and are a separate future change
- Changing the `RunnerOptions` public API
- Changing `Run()` behavior or its existing test infrastructure
- Adding DI for `analyzeModel.Update` (TUI code, architecturally different)

## Decisions

### D1: `pipelineStepFuncs` struct with function fields

```go
type pipelineStepFuncs struct {
    crapStep     func([]string, string, string, io.Writer) (*crapStepResult, error)
    qualityStep  func([]string, string, io.Writer) (*qualityStepResult, error)
    classifyStep func([]string, string) (json.RawMessage, error)
    docscanStep  func(string) (json.RawMessage, error)
}
```

Unexported struct with unexported fields — purely internal. Production callers pass the zero value (all nil), which defaults to the real step functions inside `runProductionPipeline`.

**Rationale**: A struct groups the four injection points into a single parameter rather than adding four parameters to the function signature. Unexported because this is an internal testing seam, not a public API.

### D2: Nil-check defaulting inside `runProductionPipeline`

```go
func runProductionPipeline(..., steps pipelineStepFuncs) (*ReportPayload, error) {
    if steps.crapStep == nil {
        steps.crapStep = runCRAPStep
    }
    // ... same for other three
```

**Rationale**: Follows the exact pattern of `AnalyzeFunc` nil-check defaulting at `runner.go:104-109`. Production callers don't need to know about the step functions — they pass `pipelineStepFuncs{}` and get the real implementations.

### D3: Thread `pipelineStepFuncs` from `Run()` closure

The existing `Run()` closure at line 104-109 creates `runProductionPipeline`:

```go
analyzeFunc = func(patterns []string, moduleDir string) (*ReportPayload, error) {
    return runProductionPipeline(patterns, moduleDir, opts.CoverProfile, opts.Stderr, pipelineStepFuncs{})
}
```

This passes the zero-valued struct (all nil = all defaults). No change to `Run()` behavior.

**Rationale**: The `pipelineStepFuncs` parameter is always zero-valued in production. Only tests construct non-zero values.

### D4: Tests use package-internal access

Tests go in `runner_test.go` (package `aireport_test`) or a new internal test file. Since `pipelineStepFuncs` and `runProductionPipeline` are unexported, tests need package-internal access. Use `package aireport` in a new `pipeline_internal_test.go` file.

**Rationale**: Follows the established pattern from `crap/analyze_internal_test.go` (package `crap`).

## Risks / Trade-offs

### R1: Minimal risk — additive change

Adding a parameter to `runProductionPipeline` is internal. The only caller is `Run()` via a closure. No public API changes.

### R2: Step result types are unexported

`crapStepResult` and `qualityStepResult` are unexported structs. Internal tests can construct them directly since they're in the same package. No export needed.
