<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Add Confidence Computation Function

- [x] 1.1 Add `computeMappingClassificationConfidence(mappings []protocol.AssertionMappingData, pkg, fn string) int` to `internal/adapter/contract.go`. The function filters mappings to those matching the target `(pkg, fn)` pair (across all test functions), counts total and recognized (where `AssertionType != ""`), and returns `recognized * 100 / total` (0 when total is 0). Add GoDoc comment documenting the mapping-row proxy semantics (NOT identical to the Go-native `quality.computeDetectionConfidence`, which uses all detected assertion sites as the denominator). File: `internal/adapter/contract.go`.

## 2. Store and Expose Detection Confidence

- [x] 2.1 Add unexported `classificationConfidence map[string]int` field to `ExternalContractCoverageProvider` struct (keyed by `pkg + "/" + function` where function is the **target function** name). File: `internal/adapter/contract.go`.
- [x] 2.2 Add exported `MappingClassificationConfidence(pkg, function string) int` method on `ExternalContractCoverageProvider` that looks up the per-target-function confidence from the map, returning 0 if not found or if the map is nil (before `Build`). The `pkg` and `function` parameters refer to the **target** package and function. File: `internal/adapter/contract.go`.
- [x] 2.3 In `Build`, reset `classificationConfidence` to nil at the start (before any degraded early-return path) to prevent stale values from a prior successful build, then after `buildContractLookup` returns, iterate unique `(TargetPackage, TargetFunction)` pairs in the mappings and call `computeMappingClassificationConfidence(mappings, pkg, fn)` for each. Store results in the `classificationConfidence` map keyed by `pkg + "/" + function`. File: `internal/adapter/contract.go`.

## 3. Populate Confidence in Quality Report Construction

- [x] 3.1 In `buildExternalQualityReports` (`cmd/gaze/main.go`), use the comma-ok type assertion `ecp, ok := providers.ContractCoverage.(*adapter.ExternalContractCoverageProvider)` to safely access `MappingClassificationConfidence(result.Target.Package, result.Target.Function)`. When the assertion succeeds, populate `AssertionDetectionConfidence` on each `QualityReport` entry during construction. When it fails, confidence remains at the default 0 (safe because the type assertion can only fail in test scenarios using mock providers, never in the `--analyzer` production path). Also populate `AssertionDetectionConfidence` on the summary via arithmetic mean of per-report values (matching the Go-native aggregation in `quality.Assess`). Add a `// COUPLING` comment noting the relationship with mapping classification confidence computation in `internal/adapter/contract.go`. File: `cmd/gaze/main.go`.

## 4. Testing

- [x] 4.1 Add `TestComputeMappingClassificationConfidence` table-driven test to `internal/adapter/contract_internal_test.go`. Cover: nil mappings (0), empty slice (0), all recognized (100), none recognized (0), mixed recognized/unknown (correct ratio), integer truncation (1/3 = 33), multiple test functions filtered by name, empty `AssertionType` counts as unrecognized. Coverage target: 100% branch coverage. File: `internal/adapter/contract_internal_test.go`.
- [x] 4.2 Add `TestMappingClassificationConfidence` unit test verifying the method returns stored values, 0 for unknown functions, and 0 (no panic) when called before `Build` (nil map). Also add a regression test verifying `Build` resets the confidence map to nil on degraded paths (e.g., a subsequent `Build` with `TestMapping` unsupported after a seeded map). File: `internal/adapter/contract_internal_test.go`.
- [x] 4.3 Add integration test via fake analyzer verifying `QualityReport` entries have specific expected `AssertionDetectionConfidence` values (e.g., 100 for 1/1 recognized) and the summary mean. Extend fake analyzer's `test_mapping` response with additional mappings including some with empty `assertion_type` to test mixed recognition. The `QualityReport`-level assertions live in `cmd/gaze/external_analyzer_test.go` (`TestQualityWithExternalAnalyzer`); the provider-level integration test lives in `internal/adapter/adapter_test.go` (`TestMappingClassificationConfidence_Integration`).

## 5. Verification

- [x] 5.1 Run `go build ./...` -- MUST compile with zero errors.
- [x] 5.2 Run `go test -race -count=1 -short ./...` -- all tests MUST pass.
- [x] 5.3 Run `golangci-lint run` -- MUST pass with zero findings.

## 6. Documentation

- [x] 6.1 Update `AGENTS.md` Recent Changes section with `fix-assertion-detection-confidence` entry.
- [x] 6.2 No README, CLI docs, or website issue needed -- bug fix with no user-facing behavior change beyond corrected metric display.

## 7. Constitution Alignment Verification

- [x] 7.1 Verify all four principles: Accuracy (fixes inaccurate 0% display), Minimal Assumptions (no protocol changes, works with any analyzer that provides `AssertionType`), Actionable Output (correct confidence enables meaningful quality assessment), Testability (computation function directly tested with table-driven tests).
<!-- spec-review: passed -->
<!-- code-review: passed -->
