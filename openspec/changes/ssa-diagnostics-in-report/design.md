## Context

The SSADegraded infrastructure (issue #30) provides graceful degradation when SSA construction fails, but the failure details are only visible in stderr log messages. The data model carries a single boolean — the per-package identity and total package count are lost during aggregation.

## Goals / Non-Goals

### Goals
- Track which packages had SSA failures (package paths)
- Surface degradation at the ReportPayload level (not buried in quality JSON)
- Show diagnostics in the quality text report
- Update the JSON schema for the new fields

### Non-Goals
- Go version compatibility detection (separate concern, future enhancement)
- Fixing SSA failures (addressed by specs 021, 033)
- Changing the SSA recovery mechanism

## Decisions

### D1: SSADegradedPackages on PackageSummary

```go
SSADegradedPackages []string `json:"ssa_degraded_packages,omitempty"`
```

Set by `Assess` when SSA fails — contains the single package path. During aggregation in `runQualityStep`, degraded package paths from all per-package calls are collected into the aggregate summary.

### D2: SSADegraded fields on ReportSummary

```go
type ReportSummary struct {
    CRAPload            int
    GazeCRAPload        int
    AvgContractCoverage int
    SSADegraded         bool     // new
    SSADegradedPackages []string // new
}
```

`ReportSummary` is `json:"-"` on `ReportPayload` (internal only, not serialized). It's used for threshold evaluation and the step summary. Adding SSA fields here means `evaluateAndPrintThresholds` and `WriteStepSummary` can access degradation info.

### D3: runQualityForPackage returns package path on degradation

Change the second return from `bool` to `string` — empty string means not degraded, non-empty means degraded (the string is the package path). This avoids adding a third return value.

### D4: Diagnostics in quality text report

When `SSADegraded` is true, `WriteText` (quality) adds a warning section:

```
⚠ SSA construction failed for 2 packages:
  - github.com/example/pkg/chrome
  - github.com/example/pkg/exporter
  Quality metrics for these packages are partial (AST-only).
```

## Risks / Trade-offs

### R1: Minimal risk — additive fields
All new fields use `omitempty`. Existing consumers that don't know about them are unaffected.
