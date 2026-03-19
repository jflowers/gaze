# Tasks: Agent Quality Workflow Improvements

## 1. Add `RecommendedAction` type and field

- [x] 1.1 Add `RecommendedAction` struct to `internal/crap/crap.go` with fields: `Function string`, `Package string`, `File string`, `Line int`, `FixStrategy FixStrategy`, `CRAP float64`, `GazeCRAP *float64`, `Complexity int`, `Quadrant *Quadrant`. Add JSON tags matching field names in snake_case.
- [x] 1.2 Add `RecommendedActions []RecommendedAction` field to `crap.Summary` with JSON tag `"recommended_actions,omitempty"`.

## 2. Populate `recommended_actions` in `buildSummary`

- [x] 2.1 In `internal/crap/analyze.go`, in the `buildSummary` function, after computing `FixStrategyCounts`, iterate all scores with non-nil `FixStrategy`. Create a `RecommendedAction` from each. Sort by: (1) fix strategy priority order (`add_tests` < `add_assertions` < `decompose_and_test` < `decompose`), then (2) CRAP score descending within each group. Truncate to 20 entries. Assign to `summary.RecommendedActions`.
- [x] 2.2 Add a `fixStrategyPriority` helper function that maps each `FixStrategy` to an integer for sorting: `add_tests`=0, `add_assertions`=1, `decompose_and_test`=2, `decompose`=3.

## 3. Add `AssertionCount` to quality report

- [x] 3.1 Add `AssertionCount int` field to `taxonomy.QualityReport` in `internal/taxonomy/types.go` with JSON tag `"assertion_count"`.
- [x] 3.2 In `internal/quality/quality.go`, in the `Assess` function, after calling `DetectAssertions`, store `len(assertionSites)` in the `QualityReport.AssertionCount` field for each test-target pair.

## 4. Update JSON Schema

- [x] 4.1 Update the CRAP JSON Schema in `internal/report/schema.go` to include `recommended_actions` array with the `RecommendedAction` field definitions.
- [x] 4.2 Update the Quality JSON Schema in `internal/report/schema.go` to include `assertion_count` integer field on quality reports.

## 5. Tests

- [x] 5.1 Add `TestBuildSummary_RecommendedActions` in `internal/crap/analyze_internal_test.go` that verifies: (a) recommended_actions contains only CRAPload functions, (b) sorted by strategy priority then CRAP desc, (c) truncated to 20.
- [x] 5.2 Add `TestAssertionCount_PopulatedInQualityReport` in `internal/quality/quality_test.go` that runs `Assess` on the `welltested` fixture and verifies `AssertionCount > 0` on the resulting reports.
- [x] 5.3 Add `TestAssertionCount_ZeroForTestWithoutAssertions` — if a fixture exists with a test that calls the target but makes no assertions, verify `AssertionCount == 0`. If no such fixture exists, create a minimal one.

## 6. Verification

- [x] 6.1 Run `go build ./cmd/gaze && go vet ./...` to verify compilation.
- [x] 6.2 Run `go test -race -count=1 -short ./internal/crap/... ./internal/quality/... ./internal/report/...` to verify tests pass.
- [x] 6.3 Run `gaze crap --format=json ./internal/config` and verify `recommended_actions` appears in the JSON summary with correctly sorted entries.
- [x] 6.4 Run `gaze quality --format=json ./internal/config` and verify `assertion_count` appears on each quality report entry.
