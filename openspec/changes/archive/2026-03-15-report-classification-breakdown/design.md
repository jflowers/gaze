## Context

Classification labels (contractual/ambiguous/incidental) are computed per side effect in the classify pipeline but no aggregate counts are produced anywhere. The data flows through as raw JSON and only the resulting contract coverage percentage survives as a typed field. Users can't tell whether low coverage is caused by missing assertions or by classification bottlenecks.

## Goals / Non-Goals

### Goals
- Count classification labels from classified results
- Surface counts at the ReportSummary level
- Display counts in the CRAP text report when available
- Keep the change minimal — reuse existing data, don't modify the classification engine

### Non-Goals
- Adding confidence histograms (e.g., "23 effects at 78-79") — future enhancement
- Detecting whether `--ai` was used — future enhancement
- Modifying classification thresholds or the scoring algorithm
- Adding classification breakdown to per-function CRAP scores

## Decisions

### D1: CountLabels as a classify package helper

```go
func CountLabels(results []taxonomy.AnalysisResult) (contractual, ambiguous, incidental int)
```

Pure function that iterates results and their side effects, counting non-nil classifications by label. Lives in `classify` package because it operates on classification data.

### D2: classifyStepResult struct

```go
type classifyStepResult struct {
    JSON        json.RawMessage
    Contractual int
    Ambiguous   int
    Incidental  int
}
```

Follows the pattern of `crapStepResult` and `qualityStepResult`. The classify step now returns typed counts alongside the raw JSON.

### D3: Update pipelineStepFuncs.classifyStep signature

The classify step function type changes from `func([]string, string) (json.RawMessage, error)` to `func([]string, string) (*classifyStepResult, error)`. This requires updating the `pipelineStepFuncs` struct and the fake steps in tests.

### D4: Classification counts on ReportSummary

```go
type ReportSummary struct {
    ...
    Contractual int
    Ambiguous   int
    Incidental  int
}
```

Populated from `classifyStepResult`. The AI adapter sees these in the top-level summary alongside CRAPload and AvgContractCoverage.

### D5: Display in CRAP text report

When classification counts are available (non-zero total), add a line to the summary section:
```
Classification:      84 contractual, 25 ambiguous, 4 incidental
```

This is added to `writeSummarySection` in `internal/crap/report.go` when `Summary.ClassificationCounts` is non-nil.

Wait — actually the CRAP text report doesn't have access to classification data. The classification counts would need to be on `crap.Summary`, but classification is not computed during `gaze crap` (only during `gaze report`). The CRAP report can't show classification data unless we add a callback similar to `ContractCoverageFunc`.

**Revised approach**: Display classification counts only in the `gaze report` output via the AI adapter (which sees `ReportSummary`). The CRAP text report does not get classification data. This is consistent with the current architecture — `gaze crap` is a lightweight mode that doesn't run classification.

## Risks / Trade-offs

### R1: pipelineStepFuncs type change requires test updates
The classify step signature change affects `fakeSteps()` in `pipeline_internal_test.go`. The test must return `*classifyStepResult` instead of `json.RawMessage`. This is a straightforward mechanical change.
