## 1. CountLabels helper

- [x] 1.1 Add `CountLabels(results []taxonomy.AnalysisResult) (contractual, ambiguous, incidental int)` to `internal/classify/classify.go`.
- [x] 1.2 Add `TestCountLabels_Mixed`, `TestCountLabels_NoClassification`, and `TestCountLabels_Empty` tests.

## 2. Pipeline changes

- [x] 2.1 Add `classifyStepResult` struct to `internal/aireport/runner_steps.go` with `JSON`, `Contractual`, `Ambiguous`, `Incidental` fields.
- [x] 2.2 Update `runClassifyStep` to call `classify.CountLabels` and return `*classifyStepResult`.
- [x] 2.3 Update `pipelineStepFuncs.classifyStep` type from `func([]string, string) (json.RawMessage, error)` to `func([]string, string) (*classifyStepResult, error)`.
- [x] 2.4 Add `Contractual`, `Ambiguous`, `Incidental` int fields to `ReportSummary` in `internal/aireport/payload.go`.
- [x] 2.5 Update `runProductionPipeline` to propagate classification counts from `classifyStepResult` to `payload.Summary`.
- [x] 2.6 Update `fakeSteps()` and `ClassifyStepFails` test in `pipeline_internal_test.go` to use `*classifyStepResult`.

## 3. Tests

- [x] 3.1 Update `TestRunProductionPipeline_AllStepsSucceed` to verify classification counts are propagated.
- [x] 3.2 Run all aireport and classify tests — all pass.

## 4. Documentation & Verification

- [x] 4.1 Update `AGENTS.md` Recent Changes.
- [x] 4.2 Run `go build ./...` — clean.
- [x] 4.3 Run tests for affected packages — all pass.
