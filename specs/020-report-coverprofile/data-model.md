# Data Model: 020-report-coverprofile

**Feature**: Pass Pre-Generated Coverage Profile to `gaze report`
**Date**: 2026-03-12

This feature introduces no new persistent data structures or storage. It threads an existing string value (a filesystem path) through the existing call stack. The data model changes are purely additive fields on existing structs.

---

## Modified Entities

### `reportParams` (cmd/gaze/main.go)

The params struct for the testable CLI layer of `gaze report`.

| Field | Type | Added? | Description |
|-------|------|--------|-------------|
| `patterns` | `[]string` | — | Package patterns to analyze |
| `format` | `string` | — | Output format: "text" or "json" |
| `adapterName` | `string` | — | AI adapter name |
| `modelName` | `string` | — | Model name |
| `aiTimeout` | `time.Duration` | — | AI adapter timeout |
| `maxCrapload` | `*int` | — | CRAPload threshold |
| `maxGazeCrapload` | `*int` | — | GazeCRAPload threshold |
| `minContractCoverage` | `*int` | — | Min contract coverage threshold |
| `stdout` | `io.Writer` | — | Output destination |
| `stderr` | `io.Writer` | — | Error/progress destination |
| `runnerFunc` | `func(RunnerOptions) error` | — | Test override |
| **`coverProfile`** | **`string`** | **NEW** | **Path to pre-generated coverage profile. Empty string = generate internally.** |

**Validation rules**:
- Empty string: valid, means "use internal generation" (existing behavior)
- Non-empty: forwarded as-is to `crap.Analyze`; validated there (existence, is-regular-file)

---

### `aireport.RunnerOptions` (internal/aireport/runner.go)

The options struct passed to `aireport.Run`.

| Field | Type | Added? | Description |
|-------|------|--------|-------------|
| `Patterns` | `[]string` | — | Package patterns |
| `ModuleDir` | `string` | — | Go module root directory |
| `Adapter` | `AIAdapter` | — | AI adapter |
| `AdapterCfg` | `AdapterConfig` | — | Adapter configuration |
| `SystemPrompt` | `string` | — | AI formatting instructions |
| `Format` | `string` | — | "text" or "json" |
| `Stdout` | `io.Writer` | — | Output destination |
| `Stderr` | `io.Writer` | — | Error/progress destination |
| `Thresholds` | `ThresholdConfig` | — | CI gate configuration |
| `StepSummaryPath` | `string` | — | GITHUB_STEP_SUMMARY path |
| `AnalyzeFunc` | `func([]string, string) (*ReportPayload, error)` | — | Test override |
| **`CoverProfile`** | **`string`** | **NEW** | **Path to pre-generated coverage profile. Forwarded to `runCRAPStep`. Empty = generate internally.** |

---

### `runCRAPStep` signature (internal/aireport/runner_steps.go)

This internal function gains one parameter.

| Parameter | Type | Added? | Description |
|-----------|------|--------|-------------|
| `patterns` | `[]string` | — | Package patterns |
| `moduleDir` | `string` | — | Module root directory |
| **`coverProfile`** | **`string`** | **NEW** | **Forwarded to `crap.Options.CoverProfile`. Empty = generate internally.** |
| `stderr` | `io.Writer` | — | Warning destination |

Return type unchanged: `(*crapStepResult, error)`.

---

## Unchanged Entities

### `crap.Options` (internal/crap/analyze.go)

Already has `CoverProfile string`. No changes required.

### `crap.Analyze` (internal/crap/analyze.go)

Already branches on `opts.CoverProfile == ""`. No changes required.

### `ReportPayload`, `ThresholdConfig`, `AdapterConfig`

Unaffected. The coverage profile path is consumed entirely within the CRAP step and never serialized to the report payload.

---

## Data Flow

```
CLI flag --coverprofile=<path>
    ↓ (string, "" if omitted)
newReportCmd (cobra RunE)
    ↓
reportParams.coverProfile
    ↓
runReport(p reportParams)
    ↓
aireport.RunnerOptions.CoverProfile
    ↓
aireport.Run → runProductionPipeline
    ↓
runCRAPStep(patterns, moduleDir, coverProfile, stderr)
    ↓
crap.Options{CoverProfile: coverProfile}
    ↓
crap.Analyze (existing validation + profile use)
```

The path value is never written to disk, never logged, and never included in the report output. It is a read-only input consumed transiently during the CRAP analysis step.
