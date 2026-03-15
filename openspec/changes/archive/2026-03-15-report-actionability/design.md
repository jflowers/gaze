## Context

The `crap.Score` struct (`internal/crap/crap.go:16`) carries per-function metrics: complexity, line coverage, CRAP score, and optionally contract coverage, GazeCRAP, and quadrant. The `crap.Summary` struct (line 64) aggregates these into project-level metrics. Neither struct tells the user what action to take.

The fix strategy is deterministic from the CRAP formula: `CRAP(complexity, coverage) = complexity^2 * (1 - coverage/100)^3 + complexity`. At 100% coverage, CRAP equals complexity. Therefore:

- If `complexity >= CRAPThreshold` (default 15): even 100% coverage gives CRAP >= 15. The function **must be decomposed**.
- If `complexity < CRAPThreshold` and `lineCoverage == 0`: coverage alone can bring CRAP below threshold. The function **needs tests**.
- If `complexity < CRAPThreshold` and `lineCoverage > 0` but CRAP is still high: the function needs **more coverage** (edge cases, error paths).
- If a function has GazeCRAP data with high line coverage but low contract coverage: the function needs **contract assertions** (tests exist but don't verify observable behavior).

## Goals / Non-Goals

### Goals
- Add `FixStrategy` field to `crap.Score` — deterministic, computed from existing metrics
- Add `FixStrategyCounts` to `crap.Summary` — aggregated strategy breakdown
- Update text report to display strategy information
- Ensure JSON output automatically includes new fields

### Non-Goals
- Adding classification breakdown to the report (#42 sub-item about confidence histograms — future enhancement)
- Modifying the AI report prompt to consume the new fields (the fields flow through as raw JSON; the AI adapter will see them automatically)
- Adding a `--show-borderline` flag for classification boundary analysis (#43)
- Creating a CRAP JSON Schema (no existing schema; out of scope)

## Decisions

### D1: FixStrategy type and constants

```go
type FixStrategy string

const (
    FixDecompose       FixStrategy = "decompose"
    FixAddTests        FixStrategy = "add_tests"
    FixAddAssertions   FixStrategy = "add_assertions"
    FixDecomposeAndTest FixStrategy = "decompose_and_test"
)
```

Four strategies covering the full decision space. The names are verb-phrases that tell the user exactly what to do.

**Rationale**: String constants follow the project's convention for enumerations (`SideEffectType`, `Tier`, `Quadrant`). The four strategies are mutually exclusive and exhaustive for functions above the CRAP threshold.

### D2: Assignment logic

```
if complexity >= CRAPThreshold:
    if lineCoverage == 0:
        strategy = "decompose_and_test"
    else:
        strategy = "decompose"
elif lineCoverage == 0:
    strategy = "add_tests"
elif contractCoverage != nil && *contractCoverage < 50 && lineCoverage > 50:
    strategy = "add_assertions"
else:
    strategy = "add_tests"  // needs more coverage
```

The `add_assertions` case catches functions where tests exist (line coverage > 50%) but don't verify observable behavior (contract coverage < 50%). This directly addresses the `WriteStepSummary` scenario from issue #42 (94% line coverage, 0% contract coverage).

**Rationale**: The thresholds (50% line, 50% contract) are reasonable defaults. The `add_assertions` strategy only applies when GazeCRAP data is available (contract coverage is non-nil). When GazeCRAP is unavailable, the function is classified as `add_tests` since we can't distinguish test gaps from assertion gaps.

### D3: FixStrategy only on CRAPload functions

`FixStrategy` is only populated on scores where `CRAP >= CRAPThreshold`. Functions below the threshold don't need remediation. This keeps the field `omitempty` — it's nil for healthy functions.

**Rationale**: Avoids noise. A function with CRAP 3.0 doesn't need a fix strategy. The field is only meaningful for functions that appear in the CRAPload count.

### D4: FixStrategyCounts on Summary

```go
FixStrategyCounts map[FixStrategy]int `json:"fix_strategy_counts,omitempty"`
```

Accumulated during `buildSummary` by iterating scores with non-nil `FixStrategy`. Gives project-level planning data: "12 need decomposition, 8 need tests, 3 need assertions."

**Rationale**: Follows the `QuadrantCounts` pattern exactly.

### D5: Text report shows strategy in summary section

Add a "Remediation Breakdown" subsection to the text report summary, after the quadrant breakdown. Lists each strategy with its count. This directly addresses issue #44's request to "group CRAPload functions by remediation strategy."

The worst-offenders section already exists and sorts by CRAP score. Adding a per-entry strategy label (`[decompose]`, `[add_tests]`, etc.) gives context without changing the sort order.

**Rationale**: Minimal text report change. The sorted-by-CRAP ordering is still the most useful view; adding strategy labels lets users scan for their preferred fix type.

## Risks / Trade-offs

### R1: The `add_assertions` heuristic depends on GazeCRAP availability

When `ContractCoverageFunc` is not provided (e.g., `gaze crap` without contract coverage), all high-CRAP functions will be classified as either `decompose`, `add_tests`, or `decompose_and_test`. The `add_assertions` strategy is only available when GazeCRAP data is present. This is acceptable — without contract coverage data, we can't distinguish assertion gaps.

### R2: Fixed thresholds for `add_assertions` detection

The 50% line / 50% contract thresholds for `add_assertions` are hardcoded. A more sophisticated approach would use the GazeCRAP quadrant to determine this. However, the quadrant already captures this information (Q3 = needs assertions), so `add_assertions` could alternatively be assigned to all Q3 functions. Using quadrant directly is simpler and leverages existing classification.

**Resolution**: Use quadrant when available. If `quadrant == Q3`, strategy is `add_assertions`. This is cleaner than independent threshold checks and consistent with the existing quadrant semantics.
