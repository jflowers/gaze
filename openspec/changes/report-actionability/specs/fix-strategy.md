## ADDED Requirements

### Requirement: FixStrategy field on Score

`Score` MUST include a `FixStrategy *FixStrategy` field (JSON tag `fix_strategy,omitempty`) that is populated for functions where `CRAP >= CRAPThreshold`. For functions below threshold, the field MUST be nil.

#### Scenario: High complexity function gets "decompose"
- **GIVEN** a function with complexity 20 and line coverage 80%
- **WHEN** `computeScores` is called with CRAPThreshold 15
- **THEN** the `FixStrategy` MUST be `"decompose"` (complexity >= 15, coverage can't help)

#### Scenario: High complexity + zero coverage gets "decompose_and_test"
- **GIVEN** a function with complexity 18 and line coverage 0%
- **WHEN** `computeScores` is called with CRAPThreshold 15
- **THEN** the `FixStrategy` MUST be `"decompose_and_test"`

#### Scenario: Low complexity + zero coverage gets "add_tests"
- **GIVEN** a function with complexity 8 and line coverage 0%
- **WHEN** `computeScores` is called with CRAPThreshold 15
- **THEN** the `FixStrategy` MUST be `"add_tests"`

#### Scenario: Q3 function gets "add_assertions"
- **GIVEN** a function with complexity 6, line coverage 90%, and quadrant Q3 (SimpleButUnderspecified)
- **WHEN** `computeScores` is called
- **THEN** the `FixStrategy` MUST be `"add_assertions"`

#### Scenario: Below-threshold function has nil strategy
- **GIVEN** a function with CRAP score 8.0 (below threshold 15)
- **WHEN** `computeScores` is called
- **THEN** the `FixStrategy` MUST be nil

### Requirement: FixStrategyCounts on Summary

`Summary` MUST include a `FixStrategyCounts map[FixStrategy]int` field (JSON tag `fix_strategy_counts,omitempty`) that counts how many functions need each strategy.

#### Scenario: Summary aggregates strategy counts
- **GIVEN** scores with 3 "decompose", 5 "add_tests", and 2 "add_assertions" functions
- **WHEN** `buildSummary` is called
- **THEN** `FixStrategyCounts["decompose"]` MUST be 3, `FixStrategyCounts["add_tests"]` MUST be 5, `FixStrategyCounts["add_assertions"]` MUST be 2

#### Scenario: No CRAPload functions produces nil counts
- **GIVEN** all scores have CRAP below threshold
- **WHEN** `buildSummary` is called
- **THEN** `FixStrategyCounts` MUST be nil (omitted from JSON)

### Requirement: Text report displays strategy information

`WriteText` MUST display a "Remediation Breakdown" subsection in the summary when `FixStrategyCounts` is non-empty. Each worst-offender entry MUST include a strategy label.

#### Scenario: Remediation breakdown shown
- **GIVEN** a report with non-empty `FixStrategyCounts`
- **WHEN** `WriteText` is called
- **THEN** the output MUST contain "Remediation Breakdown" with counts per strategy

#### Scenario: Worst offenders include strategy labels
- **GIVEN** a report with worst offenders that have `FixStrategy` set
- **WHEN** `WriteText` is called
- **THEN** each worst-offender line MUST include the strategy label (e.g., `[decompose]`)

## MODIFIED Requirements

None — all changes are additive fields.

## REMOVED Requirements

None.
