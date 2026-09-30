<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Cognitive Complexity Core

- [x] 1.1 Create `internal/cognitive/` package with `FuncCognitiveComplexity` struct and `AnalyzeFunc(fset, funcDecl) int` signature
- [x] 1.2 Implement Sonar algorithm AST walker: if/else if/else, switch, type switch, for, range, select increments with nesting penalties
- [x] 1.3 Implement logical operator sequence detection (`&&`, `||`) with mixed-sequence counting
- [x] 1.4 Implement recursion detection (CallExpr matching enclosing function name)
- [x] 1.5 Implement goto statement increment
- [x] 1.6 Add `AnalyzeFile(fset, file) []FuncCognitiveComplexity` batch function
- [x] 1.7 Write unit tests for all increment rules using synthetic AST (table-driven, target 100% branch coverage)
- [x] 1.8 Write unit tests for nesting penalty combinations (deeply nested structures)
- [x] 1.9 Write unit tests for logical operator sequences (same-operator chain, mixed operators)
- [x] 1.10 Write unit tests for recursion detection

## 2. Provider Interface and Go Implementation

- [x] 2.1 Add `FunctionCognitiveComplexity` struct and `CognitiveComplexityProvider` interface to `internal/crap/provider.go`
- [x] 2.2 Create `GoCognitiveComplexityProvider` in `internal/provider/goprovider/` wrapping `cognitive.AnalyzeFile`
- [x] 2.3 Write integration tests with `testdata/src/` fixtures containing functions of known cognitive complexity

## 3. CRAP Pipeline Integration

- [x] 3.1 Add `CognitiveComplexity *int` and `GazeCRAPCC *float64` fields to `crap.Score` struct (omitempty)
- [x] 3.2 Add `CognitiveFormula(complexity int, coveragePct float64) float64` to `internal/crap/crap.go`
- [x] 3.3 Wire `CognitiveComplexityProvider` into `crap.Options` and `computeScores` in `internal/crap/analyze.go`
- [x] 3.4 Add `CognitiveComplexityTotal` and `CognitiveComplexityExceeded` fields to `crap.Summary`
- [x] 3.5 Write unit tests for `CognitiveFormula` (mirroring existing `TestFormula_*` patterns)
- [x] 3.6 Write unit tests for Score struct JSON serialization (omitempty behavior)

## 4. Report Output

- [x] 4.1 Update `crap.WriteText` to include cognitive complexity column in verbose mode
- [x] 4.2 Update JSON reporter to include `cognitive_complexity` and `gaze_crap_cc` per function
- [x] 4.3 Update `internal/report/schema.go` JSON Schema with new fields
- [x] 4.4 Write tests for JSON schema validation with new fields

## 5. CLI Flag and CI Gate

- [x] 5.1 Add `--max-cognitive-complexity` flag to `gaze analyze` command (`cmd/gaze/main.go`)
- [x] 5.2 Add `--max-cognitive-complexity` flag to `gaze crap` command (`cmd/gaze/main.go`)
- [x] 5.3 Implement gate evaluation: exit code 1 when any function exceeds threshold
- [x] 5.4 Write CLI tests for flag pass/fail/not-provided scenarios

## 6. External Analyzer Protocol

- [x] 6.1 Add `cognitive_complexity` method constant and `FunctionCognitiveComplexityData` type to `internal/protocol/types.go`
- [x] 6.2 Implement `FetchCognitiveComplexity` in `internal/adapter/` for external analyzer integration
- [x] 6.3 Wire graceful degradation: nil fields when external analyzer does not implement method
- [x] 6.4 Update fake analyzer in `internal/protocol/testdata/` to support `cognitive_complexity`
- [x] 6.5 Write integration tests for external analyzer cognitive complexity flow

## 7. Documentation and Constitution Verification

- [x] 7.1 Update `docs/protocol.md` with `cognitive_complexity` method documentation
- [x] 7.2 Update README with `--max-cognitive-complexity` flag and GazeCRAP-CC description
- [x] 7.3 Verify constitution alignment: run all tests, confirm no regressions in existing CRAP/GazeCRAP scores
- [x] 7.4 Run CI checks locally (`go test -race -count=1 -short ./...`, `golangci-lint run`)

<!-- spec-review: passed -->

<!-- code-review: passed -->
