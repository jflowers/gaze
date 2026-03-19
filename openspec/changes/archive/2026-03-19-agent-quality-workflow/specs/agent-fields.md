# Agent-Facing JSON Fields Spec

## ADDED Requirements

### Requirement: CRAP summary MUST include sorted `recommended_actions` list

The `crap.Summary` JSON output MUST include a `recommended_actions` array of objects, each containing `function`, `package`, `file`, `line`, `fix_strategy`, `crap`, `gaze_crap` (optional), `complexity`, and `quadrant` (optional). The list MUST be sorted by fix strategy priority (add_tests, add_assertions, decompose_and_test, decompose), then by CRAP score descending within each group. The list MUST be limited to the top 20 entries.

#### Scenario: Agent reads recommended actions
- **GIVEN** a module with 10 CRAPload functions (3 add_tests, 5 add_assertions, 2 decompose)
- **WHEN** an agent runs `gaze crap --format=json ./...`
- **THEN** the JSON `summary.recommended_actions` contains 10 entries, with the 3 `add_tests` entries first (sorted by CRAP desc), then 5 `add_assertions`, then 2 `decompose`

#### Scenario: Empty recommended_actions when no CRAPload
- **GIVEN** a module where all functions are below the CRAP threshold
- **WHEN** `gaze crap --format=json` runs
- **THEN** `summary.recommended_actions` is an empty array or absent

#### Scenario: recommended_actions truncated at 20
- **GIVEN** a module with 50 CRAPload functions
- **WHEN** `gaze crap --format=json` runs
- **THEN** `summary.recommended_actions` contains exactly 20 entries (the top 20 by priority)

### Requirement: Quality report MUST include `assertion_count` per test-target pair

Each `QualityReport` in the quality JSON output MUST include an `assertion_count` integer field representing the total number of detected assertion sites for that test-target pair.

#### Scenario: Test with assertions
- **GIVEN** a test function `TestAdd` with 3 assertions (2 equality checks + 1 error check)
- **WHEN** `gaze quality --format=json` runs
- **THEN** the quality report for `TestAdd` has `assertion_count: 3`

#### Scenario: Test with no assertions
- **GIVEN** a test function `TestEmpty` that calls the target but makes no assertions
- **WHEN** `gaze quality --format=json` runs
- **THEN** the quality report for `TestEmpty` has `assertion_count: 0`

#### Scenario: Distinguishing coverage gap from assertion gap
- **GIVEN** function `Foo` with `line_coverage: 90%` and `contract_coverage: 0%`
- **WHEN** an agent reads the quality report
- **THEN** if `assertion_count > 0`, the agent knows to add CONTRACT assertions (not more tests); if `assertion_count == 0`, the agent knows the test function exists but has no assertions at all

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
