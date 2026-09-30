## ADDED Requirements

### Requirement: GazeCRAP-CC Score

The system MUST compute a `GazeCRAP-CC` score per function using the formula `GazeCRAP-CC = CC^2 * (1 - coverage/100)^3 + CC`, where CC is cognitive complexity and coverage is line coverage (0-100).

#### Scenario: fully covered function with low cognitive complexity
- **GIVEN** a function with cognitive complexity 3 and line coverage 100%
- **WHEN** GazeCRAP-CC is computed
- **THEN** the GazeCRAP-CC score MUST be 3.0 (3^2 * 0^3 + 3)

#### Scenario: uncovered function with high cognitive complexity
- **GIVEN** a function with cognitive complexity 10 and line coverage 0%
- **WHEN** GazeCRAP-CC is computed
- **THEN** the GazeCRAP-CC score MUST be 1010.0 (10^2 * 1^3 + 10)

#### Scenario: partially covered function
- **GIVEN** a function with cognitive complexity 5 and line coverage 50%
- **WHEN** GazeCRAP-CC is computed
- **THEN** the GazeCRAP-CC score MUST be 131.25 (5^2 * 0.5^3 + 5)

#### Scenario: cognitive complexity unavailable
- **GIVEN** a function where cognitive complexity was not computed
- **WHEN** GazeCRAP-CC is computed
- **THEN** the GazeCRAP-CC value MUST be nil (not reported)

### Requirement: GazeCRAP-CC in JSON Output

The system MUST include `gaze_crap_cc` per function in JSON output when cognitive complexity and coverage data are available.

#### Scenario: JSON includes GazeCRAP-CC
- **GIVEN** a function with cognitive complexity 5, line coverage 80%, and computed GazeCRAP-CC of 7.4
- **WHEN** JSON output is produced
- **THEN** the function's JSON representation MUST include `"gaze_crap_cc": 7.4`

#### Scenario: JSON omits GazeCRAP-CC when unavailable
- **GIVEN** a function where cognitive complexity was not computed
- **WHEN** JSON output is produced
- **THEN** the function's JSON representation MUST omit the `gaze_crap_cc` field

### Requirement: GazeCRAP-CC in Text Report

The system SHOULD include GazeCRAP-CC in text reports alongside existing CRAP and GazeCRAP scores.

#### Scenario: text report includes GazeCRAP-CC column
- **GIVEN** functions with cognitive complexity data
- **WHEN** text report is produced
- **THEN** the report SHOULD include a GazeCRAP-CC column or section alongside existing CRAP/GazeCRAP scores

## MODIFIED Requirements

### Requirement: Score Struct Extension

The `crap.Score` struct MUST include optional `CognitiveComplexity` and `GazeCRAPCC` fields.

Previously: `Score` contained `Complexity`, `CRAP`, `GazeCRAP`, and related fields.

#### Scenario: new fields are optional
- **GIVEN** a `Score` where cognitive complexity was computed
- **WHEN** the struct is serialized to JSON
- **THEN** `cognitive_complexity` and `gaze_crap_cc` MUST be present

#### Scenario: new fields omitted when nil
- **GIVEN** a `Score` where cognitive complexity was NOT computed
- **WHEN** the struct is serialized to JSON
- **THEN** `cognitive_complexity` and `gaze_crap_cc` MUST be omitted (omitempty)

### Requirement: Summary Aggregate

The `crap.Summary` MUST include aggregate cognitive complexity statistics.

#### Scenario: summary includes cognitive complexity stats
- **GIVEN** a run with multiple functions having cognitive complexity data
- **WHEN** the summary is computed
- **THEN** the summary MUST include total cognitive complexity and count of functions exceeding the cognitive complexity threshold

### Requirement: JSON Schema Update

The JSON Schema MUST be updated to include `cognitive_complexity` and `gaze_crap_cc` fields.

#### Scenario: schema validates new fields
- **GIVEN** JSON output containing `cognitive_complexity` and `gaze_crap_cc`
- **WHEN** validated against the schema
- **THEN** validation MUST pass
