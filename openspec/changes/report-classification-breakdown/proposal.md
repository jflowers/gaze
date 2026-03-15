## Why

The gaze report shows contract coverage percentages but doesn't explain *why* contract coverage is low. When the classification engine labels most side effects as "ambiguous" (confidence just below the contractual threshold), contract coverage is structurally zero regardless of how many test assertions exist. Users discover this only after wasting effort adding assertions that can't improve the metric.

Issue #42 reports a real case: 55 assertion fixes moved CI contract coverage from 10% to 11% because 73% of side effects were classified "ambiguous" at confidence 78-79 (threshold 80). The bottleneck was classification, not assertions.

## What Changes

1. Add a `CountLabels` helper to the `classify` package that counts contractual/ambiguous/incidental from classified `[]AnalysisResult`.
2. Enrich `runClassifyStep` to return typed classification counts (like `runCRAPStep` returns `CRAPload`).
3. Propagate counts to `ReportSummary` so the AI adapter and threshold evaluation can see them.
4. Add classification counts to the CRAP text report summary when available.

## Capabilities

### New Capabilities
- `classify.CountLabels`: Helper function that counts side effects by classification label from `[]AnalysisResult`.
- `ClassificationCounts` on `ReportSummary`: Contractual, Ambiguous, Incidental counts at the report payload level.
- Classification breakdown in CRAP text report summary when classification data is available.

### Modified Capabilities
- `runClassifyStep`: Returns typed `classifyStepResult` with JSON + counts (not bare `json.RawMessage`).
- `runProductionPipeline`: Propagates classification counts to `ReportSummary`.

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/classify/classify.go` | Add `CountLabels` helper |
| `internal/aireport/runner_steps.go` | Add `classifyStepResult`, update `runClassifyStep` |
| `internal/aireport/payload.go` | Add classification fields to `ReportSummary` |
| `internal/aireport/runner.go` | Propagate counts, update `pipelineStepFuncs` type |
| `internal/crap/report.go` | Display classification counts in text summary |
| `internal/crap/crap.go` | Add `ClassificationCounts` to `Summary` |
| Tests | New tests for CountLabels, updated pipeline tests |
| `AGENTS.md` | Update Recent Changes |

## Constitution Alignment

### I. Accuracy — PASS
The counts are derived directly from the classification labels already computed. No new analysis or assumptions.

### II. Minimal Assumptions — PASS
No new user-facing requirements. Additive fields.

### III. Actionable Output — PASS
This is the primary motivation — users can now see "contract coverage is 0% because all effects are ambiguous" and know to use `--ai` or adjust thresholds instead of adding more assertions.

### IV. Testability — PASS
`CountLabels` is a pure function testable with synthetic inputs.
