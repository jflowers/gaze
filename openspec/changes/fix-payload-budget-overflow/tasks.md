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

## 1. Compact CRAP projection

- [x] 1.1 Remove the `Scores []crap.Score` field from `compactCRAPReport` in `internal/aireport/compact.go`.
- [x] 1.2 Restore `WorstCRAP []crap.Score`, `WorstGazeCRAP []crap.Score`, and `RecommendedActions []crap.RecommendedAction` fields (with correct `json` tags `worst_crap`, `worst_gaze_crap`, `recommended_actions`) to `compactCRAPSummary`.
- [x] 1.3 Update `compactCRAPField` to populate the restored worst-offender fields from `full.Summary` and stop copying `full.Scores`.

## 2. Compact Quality projection

- [x] 2.1 Replace `GapIDs []string` / `DiscardedReturnIDs []string` in `compactContractCoverage` with `Gaps []compactSideEffect` / `DiscardedReturns []compactSideEffect` (json tags `gaps`, `discarded_returns`).
- [x] 2.2 Replace `AmbiguousEffectIDs []string` with `AmbiguousEffects []compactSideEffect` (json tag `ambiguous_effects`) in `compactQualityReport`.
- [x] 2.3 Restore `WorstCoverageTests []compactQualityReport` (json tag `worst_coverage_tests`) to `compactPackageSummary`.
- [x] 2.4 Add a projection helper that maps a full `taxonomy.SideEffect` to a `compactSideEffect` (id, type, tier, location, description, target — no classification).
- [x] 2.5 Update `compactQualityField` to project gaps/discarded/ambiguous through the helper, sort `Reports` by ascending contract coverage, keep only actionable entries (gaps, discarded returns, ambiguous effects, or unmapped assertions), and cap at 50.

## 3. Compact Classify projection

- [x] 3.1 Change `compactClassifyField` to stop unmarshalling the full `results` array and instead emit a counts-only object (`version`, `contractual`, `ambiguous`, `incidental`) sourced from `ReportPayload.Summary` counts; keep `null` when `Classify` is nil.

## 4. Tests

- [x] 4.1 Update `TestCompactForAI_QualityGapsReducedToIDs`, `TestCompactForAI_QualityDiscardedReturnsReducedToIDs`, and `TestCompactForAI_QualityAmbiguousEffectsReducedToIDs` to assert self-contained projected objects (`gaps`/`discarded_returns`/`ambiguous_effects`) instead of bare ID arrays.
- [x] 4.2 Invert `TestCompactForAI_CRAPWorstOffendersOmitted` to assert `worst_crap`/`worst_gaze_crap`/`recommended_actions` are preserved and `scores` is absent.
- [x] 4.3 Invert `TestCompactForAI_QualitySummaryDedup` to assert `worst_coverage_tests` is preserved.
- [x] 4.4 Replace `TestCompactForAI_ClassifySignalsStripped` with a counts-only Classify test (asserts `contractual`/`ambiguous`/`incidental` present, `results` absent, and `null` on Classify failure).
- [x] 4.5 Rescale `TestCompactForAI_SizeBudget` to ~2000 functions / ~1500 test-target pairs / ~5000 side effects and keep the `< 300KB` assertion.
- [x] 4.6 Add a test asserting `quality_reports` is bounded to at most 50 actionable entries and excludes non-actionable reports.
- [x] 4.7 Confirm `TestCompactForAI_FullMarshalUnchanged`, `TestRunTextPath_CompactPayloadReceived`, and `TestCompactForAI_MalformedInput` still pass (adjust only if the compact shape changes break them).

## 5. Verification

- [x] 5.1 Run `go build ./cmd/gaze`.
- [x] 5.2 Run `go test -race -count=1 -short ./...`.
- [x] 5.3 Run `golangci-lint run`.
- [x] 5.4 Verify constitution alignment: the change preserves Accuracy (full JSON and thresholds untouched), Minimal Assumptions (bounded budget independent of codebase size), Actionable Output (worst-offender lists and self-contained gap descriptions preserved), and Testability (unit + integration + size-budget tests). Confirm no gate value in AGENTS.md or the constitution was modified.
- [x] 5.5 Manually verify the compact payload size drops below 300KB by running the report step locally (observe `Payload: N bytes (compact)` on stderr) against this repo, then confirm the Go 1.24 `Gaze quality report` CI step passes.

<!-- spec-review: passed -->
<!-- code-review: passed -->
