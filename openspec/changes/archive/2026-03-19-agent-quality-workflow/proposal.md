# Agent Quality Workflow Improvements

## Why

Issue #48 identifies four improvements for AI agents consuming gaze JSON output to drive automated quality improvement. Two have been partially addressed:

1. **`fix_strategy` field** — DONE. Added in the `report-actionability` change. Each `crap.Score` carries a `FixStrategy`.
2. **Combined crap+quality mode** — PARTIALLY DONE. `gaze report --format=json` combines all four analysis steps, but it also runs classification and docscan which aren't needed for the "reduce GazeCRAPload" use case. A lightweight `--with-quality` flag on `gaze crap` doesn't exist.

Two remain unaddressed:

3. **`recommended_actions` list** — The JSON has per-function `FixStrategy` and aggregate `FixStrategyCounts`, but no sorted, prioritized action list. An agent has to parse all scores, filter to CRAPload functions, sort by impact, and group by strategy. A `recommended_actions` field on `Summary` would eliminate this work.

4. **Assertion density metric** — The quality data has `AssertionDetectionConfidence` (a ratio, not a count) and `ContractCoverage.CoveredCount` (contractual effects asserted). But there's no total assertion count per test function. The raw `[]AssertionSite` count from `DetectAssertions` is consumed by the mapper and discarded. An `assertion_count` field on `QualityReport` would let agents distinguish "no tests" (count 0) from "tests without assertions" (count > 0 but low coverage).

This change addresses items 3 and 4. Item 2 (combined mode) is deferred — `gaze report --format=json` already provides the combined data, and adding `--with-quality` to `gaze crap` would duplicate the report pipeline's logic.

Partially closes #48 (items 3 and 4).

## What Changes

1. Add `RecommendedActions []RecommendedAction` to `crap.Summary` — a sorted list of functions needing remediation, ordered by expected impact (fix strategy priority, then CRAP score descending).
2. Add `AssertionCount int` to `taxonomy.QualityReport` — the total number of detected assertion sites per test-target pair, populated from `len(DetectAssertions(...))`.

## Capabilities

### New Capabilities
- `recommended_actions` field on CRAP summary JSON — sorted action list with function, package, file, line, fix strategy, CRAP score, and expected impact
- `assertion_count` field on quality report JSON — total assertions per test function

### Modified Capabilities
- `crap.Summary` gains `RecommendedActions` field
- `taxonomy.QualityReport` gains `AssertionCount` field
- `quality.Assess` populates `AssertionCount` from `DetectAssertions` result

### Removed Capabilities
- None

## Impact

- `internal/crap/crap.go` — add `RecommendedAction` type and `RecommendedActions` field to `Summary`
- `internal/crap/analyze.go` — populate `RecommendedActions` in `buildSummary`
- `internal/taxonomy/types.go` — add `AssertionCount int` to `QualityReport`
- `internal/quality/quality.go` — populate `AssertionCount` from `len(assertionSites)` in `Assess`
- `internal/report/schema.go` — update JSON Schema with new fields
- Test files for the above

## Constitution Alignment

### I. Autonomous Collaboration — N/A
### II. Composability First — PASS (additive JSON fields, no breaking changes)
### III. Observable Quality — PASS (makes quality data more actionable for automated consumers)
### IV. Testability — PASS (new fields are pure data, testable with existing patterns)
