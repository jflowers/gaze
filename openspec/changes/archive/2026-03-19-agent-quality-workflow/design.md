# Design: Agent Quality Workflow Improvements

## Context

Agents consuming gaze JSON need two things the output doesn't provide:

1. A sorted action list — instead of filtering/sorting `scores[]` by `fix_strategy` and CRAP, agents want a ready-to-use `recommended_actions` list ordered by expected impact.
2. An assertion count — to distinguish "no tests" (assertion_count absent, no quality data) from "tests without assertions" (assertion_count 0 or very low, but line coverage exists).

## Goals / Non-Goals

### Goals
- Add `recommended_actions` to `crap.Summary` — sorted by fix strategy priority, then CRAP score descending
- Add `assertion_count` to `taxonomy.QualityReport` — total detected assertion sites per test-target pair
- Both fields appear in JSON output of `gaze crap --format=json` and `gaze report --format=json`

### Non-Goals
- Adding `--with-quality` to `gaze crap` (deferred — `gaze report --format=json` already covers the use case)
- Adding automated decomposition hints (out of scope per issue #48)
- Changing the AI-formatted text report (the agent prompt can consume the new JSON fields as-is)

## Decisions

### D1: `RecommendedAction` type

```go
type RecommendedAction struct {
    Function    string       `json:"function"`
    Package     string       `json:"package"`
    File        string       `json:"file"`
    Line        int          `json:"line"`
    FixStrategy FixStrategy  `json:"fix_strategy"`
    CRAP        float64      `json:"crap"`
    GazeCRAP    *float64     `json:"gaze_crap,omitempty"`
    Complexity  int          `json:"complexity"`
    Quadrant    *Quadrant    `json:"quadrant,omitempty"`
}
```

This is a subset of `Score` — only the fields an agent needs to plan and execute a fix. It excludes coverage percentages (the agent can look those up in the full scores if needed) and includes the `FixStrategy` as a top-level field (not optional — every recommended action has a strategy).

### D2: Sorting order for `recommended_actions`

Priority order (matching the agent's discovered ordering from issue #48):
1. `add_tests` first (easiest wins — zero coverage, any test moves the needle)
2. `add_assertions` second (existing tests, just needs contract assertions)
3. `decompose_and_test` third (needs refactoring + tests)
4. `decompose` last (no tests can help, must reduce complexity first)

Within each strategy group, sort by CRAP score descending (worst first).

### D3: `AssertionCount` field

Add `AssertionCount int` to `taxonomy.QualityReport` with JSON tag `"assertion_count"`. Populated from `len(assertionSites)` in `quality.Assess` — the value is the total number of detected assertion sites (all kinds: stdlib comparison, error check, testify, gocmp) for the test-target pair.

This count is available right after `DetectAssertions` returns and before the mapper consumes the sites. Currently the count is discarded. The change is to store it in the report.

### D4: Populate `recommended_actions` in `buildSummary`

The `buildSummary` function in `internal/crap/analyze.go` already iterates all scores to compute `WorstCRAP`, `WorstGazeCRAP`, and `FixStrategyCounts`. Adding `RecommendedActions` is a simple extension: filter to scores with non-nil `FixStrategy`, create `RecommendedAction` from each, sort by the priority order, and truncate to a reasonable limit (top 20, matching the `WorstCRAP` limit).

## Risks / Trade-offs

- **JSON size increase**: `recommended_actions` adds up to 20 entries (~2KB). Acceptable for machine-readable output.
- **Redundancy with `worst_crap`**: `recommended_actions` overlaps with `worst_crap` but provides different ordering (by strategy priority, not CRAP score) and includes only actionable entries. Both lists serve different purposes.
