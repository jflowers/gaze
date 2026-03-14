## 1. Add pipelineStepFuncs and update runProductionPipeline

- [x] 1.1 Add `pipelineStepFuncs` struct to `internal/aireport/runner.go` with four unexported function fields: `crapStep`, `qualityStep`, `classifyStep`, `docscanStep`. Add GoDoc comment.
- [x] 1.2 Update `runProductionPipeline` signature to accept `pipelineStepFuncs` as a parameter. Add nil-check defaulting for each field at the top of the function.
- [x] 1.3 Replace direct calls to `runCRAPStep`, `runQualityStep`, `runClassifyStep`, `runDocscanStep` with calls through the `steps` struct fields.
- [x] 1.4 Update the `Run()` closure (line ~104-109) to pass `pipelineStepFuncs{}` (zero value) to `runProductionPipeline`.
- [x] 1.5 Verify compilation: `go build ./internal/aireport/`

## 2. Tests

- [x] 2.1 Create `internal/aireport/pipeline_internal_test.go` with `package aireport` for internal tests of `runProductionPipeline`.
- [x] 2.2 Add `TestRunProductionPipeline_AllStepsSucceed` — inject fake step functions that return synthetic results, verify all payload sections populated and no errors.
- [x] 2.3 Add `TestRunProductionPipeline_CRAPStepFails` — inject failing CRAP step, verify `payload.Errors.CRAP` non-nil, other sections still populated.
- [x] 2.4 Add `TestRunProductionPipeline_QualityStepFails` — same pattern for quality step.
- [x] 2.5 Add `TestRunProductionPipeline_ClassifyStepFails` — same pattern for classify step.
- [x] 2.6 Add `TestRunProductionPipeline_DocscanStepFails` — same pattern for docscan step.
- [x] 2.7 Add `TestRunProductionPipeline_MultipleStepsFail` — inject failures in CRAP and quality, verify both errors captured and classify/docscan still succeed.
- [x] 2.8 Add `TestRunProductionPipeline_EmptyPatterns` — verify error returned before any step functions are called.
- [x] 2.9 Add `TestRunProductionPipeline_SummaryFields` — verify CRAPload, GazeCRAPload, and AvgContractCoverage are correctly propagated from step results to payload summary.
- [x] 2.10 Run all existing aireport tests: `go test -race -count=1 -run 'Test' ./internal/aireport/`

## 3. Documentation

- [x] 3.1 Update `AGENTS.md` Recent Changes with a summary of this change.

## 4. Verification

- [x] 4.1 Run `go test -race -count=1 -short ./...` and verify all tests pass.
- [x] 4.2 Run `go build ./...` and `go vet ./...` to verify no issues.
- [x] 4.3 Verify constitution alignment: (I) no interface changes, (II) additive-only, (III) no output changes, (IV) pipeline testable in isolation.
