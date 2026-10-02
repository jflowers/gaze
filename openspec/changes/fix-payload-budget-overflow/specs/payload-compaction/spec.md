## ADDED Requirements

### Requirement: CRAP scores array exclusion

The compact payload MUST NOT include the full `scores` array from the CRAP output. The CRAP compact output MUST retain the `summary` object (including the bounded worst-offender lists) and MUST NOT contain a `"scores"` field.

#### Scenario: Scores array omitted
- **GIVEN** a CRAP report with `scores` containing 2000 entries and `summary.worst_crap` containing 5 entries
- **WHEN** `CompactForAI()` is called
- **THEN** the CRAP output MUST contain a `"summary"` object with `"worst_crap"` and MUST NOT contain a `"scores"` field

### Requirement: Quality reports bounding

The compact payload MUST bound the `quality_reports` array to a fixed maximum independent of the total number of test-target pairs. The compact output MUST include only actionable reports (those with at least one coverage gap, discarded return, ambiguous effect, or unmapped assertion), ordered by ascending contract coverage, capped at 50 entries.

#### Scenario: Reports bounded to actionable entries
- **GIVEN** a Quality output with 1500 test-target pairs, of which 400 are actionable (have gaps or unmapped assertions)
- **WHEN** `CompactForAI()` is called
- **THEN** the `"quality_reports"` array MUST contain at most 50 entries, each being an actionable report, and MUST NOT include reports with no gaps, discarded returns, ambiguous effects, or unmapped assertions

### Requirement: Classify results exclusion

The compact payload MUST NOT include the full `results` array from the Classify output. The `"classify"` field MUST instead contain a counts-only object with `contractual`, `ambiguous`, and `incidental` integer counts, plus the `version` string. These counts MUST be sourced from the top-level `summary` counts, not by re-scanning the full results.

#### Scenario: Classify output is counts-only
- **GIVEN** a Classify result with 5000 side effects across 800 targets, and top-level summary counts `contractual=3100`, `ambiguous=1200`, `incidental=700`
- **WHEN** `CompactForAI()` is called
- **THEN** the `"classify"` field MUST be an object with `"contractual": 3100`, `"ambiguous": 1200`, `"incidental": 700`, and MUST NOT contain a `"results"` field

#### Scenario: Classify step failure preserved
- **GIVEN** a `ReportPayload` where the Classify step failed (Classify is nil)
- **WHEN** `CompactForAI()` is called
- **THEN** the `"classify"` field MUST be `null`

### Requirement: Self-contained quality side-effect projection

When the compact payload projects a quality side effect (a coverage gap, a discarded return, or an ambiguous effect), it MUST emit a minimal side-effect object carrying `id`, `type`, `tier`, `location`, `description`, and `target`, and MUST NOT include the `classification` field. This projection MUST carry enough detail for the reporter to render "function name, effect type, and description" without cross-referencing the Classify output.

#### Scenario: Gap carries description inline
- **GIVEN** a Quality report whose `contract_coverage.gaps` contains a side effect with `id="se-a1b2c3d4"`, `type="ErrorReturn"`, `tier="P0"`, `location="file.go:42"`, `description="returns an error that is never asserted"`, `target="DoThing"`, and a `classification` object
- **WHEN** `CompactForAI()` is called
- **THEN** the projected gap MUST contain `"id": "se-a1b2c3d4"`, `"type": "ErrorReturn"`, `"tier": "P0"`, `"location": "file.go:42"`, `"description": "returns an error that is never asserted"`, `"target": "DoThing"`, and MUST NOT contain a `"classification"` field

## MODIFIED Requirements

### Requirement: Quality data compaction

The compact payload MUST replace full `SideEffect` objects in Quality report fields with arrays of self-contained minimal side-effect objects (see "Self-contained quality side-effect projection"). `contract_coverage.gaps` MUST be projected to `contract_coverage.gaps`, `contract_coverage.discarded_returns` MUST be projected to `contract_coverage.discarded_returns`, and `ambiguous_effects` MUST be projected to `ambiguous_effects`. Each projected object MUST carry `id`, `type`, `tier`, `location`, `description`, and `target`, and MUST omit `classification` and `signals`. The `contract_coverage.gap_hints` and `contract_coverage.discarded_return_hints` fields MUST be preserved as-is, and all scalar fields on `ContractCoverage` and `QualityReport` MUST be preserved.

Previously: The compact payload replaced full `SideEffect` objects with arrays of side effect ID strings (`contract_coverage.gap_ids`, `contract_coverage.discarded_return_ids`, `ambiguous_effect_ids`).

#### Scenario: Quality gaps projected to self-contained objects
- **GIVEN** a Quality report with a test-target pair whose `contract_coverage.gaps` contains a full `SideEffect` with `id="se-a1b2c3d4"`, `type="MapMutation"`, `description="mutates a shared map"`
- **WHEN** `CompactForAI()` is called
- **THEN** the corresponding quality entry's `contract_coverage` MUST contain `"gaps"` with an entry carrying `"id":"se-a1b2c3d4"`, `"type":"MapMutation"`, `"description":"mutates a shared map"`, and each gap MUST NOT contain a `"classification"` field

#### Scenario: Cross-reference resilience when Classify fails
- **GIVEN** a `ReportPayload` where the Classify step failed (Classify is nil) but Quality has gaps with side effects
- **WHEN** `CompactForAI()` is called
- **THEN** the quality gaps MUST still be emitted with their `type` and `description` inline

### Requirement: Signal stripping in Classify output

The compact payload MUST omit the Classify `results` array entirely and emit only a counts-only `"classify"` object with `contractual`, `ambiguous`, and `incidental` integer counts plus the `version` string. The counts MUST be sourced from the top-level summary.

Previously: The compact payload omitted `Classification.Signals` arrays from all side effects while preserving `label`, `confidence`, and `reasoning`.

#### Scenario: Classify results dropped, counts emitted
- **GIVEN** a Classify result with many side effects and summary counts `contractual=2000`, `ambiguous=900`, `incidental=500`
- **WHEN** `CompactForAI()` is called
- **THEN** the `"classify"` field MUST contain `"contractual": 2000`, `"ambiguous": 900`, `"incidental": 500`, and MUST NOT contain `"results"` or any side-effect or `"signals"` data

### Requirement: CRAP summary deduplication

The compact payload MUST preserve `summary.worst_crap`, `summary.worst_gaze_crap`, and `summary.recommended_actions` (bounded lists of at most 5, 5, and 20 entries respectively). The compact payload MUST omit the full `scores` array. All other summary fields (`total_functions`, `avg_complexity`, `avg_line_coverage`, `avg_crap`, `crapload`, `crap_threshold`, `gaze_crapload`, `gaze_crap_threshold`, `avg_gaze_crap`, `avg_contract_coverage`, `quadrant_counts`, `fix_strategy_counts`, `ssa_degraded_packages`) MUST be preserved.

Previously: The compact payload omitted `summary.worst_crap`, `summary.worst_gaze_crap`, and `summary.recommended_actions` while preserving the full `scores` array.

#### Scenario: Worst offender lists preserved, scores omitted
- **GIVEN** a CRAP report with `summary.worst_crap` containing 5 entries and `scores` containing 2000 entries
- **WHEN** `CompactForAI()` is called
- **THEN** the CRAP output MUST contain `"summary"` with `"worst_crap"` (5 entries), `"worst_gaze_crap"`, and `"recommended_actions"`, and MUST NOT contain a `"scores"` field

### Requirement: Quality summary deduplication

The compact payload MUST preserve `quality_summary.worst_coverage_tests` (bounded to the bottom 5 tests by contract coverage). The compact payload MUST bound the `quality_reports` array to at most 50 actionable entries. All other summary fields (`total_tests`, `average_contract_coverage`, `total_over_specifications`, `assertion_detection_confidence`, `ssa_degraded`, `ssa_degraded_packages`, `skipped_tests`, `skipped_test_names`) MUST be preserved.

Previously: The compact payload omitted `quality_summary.worst_coverage_tests` from the Quality output.

#### Scenario: Worst coverage tests preserved
- **GIVEN** a Quality output with `quality_summary.worst_coverage_tests` containing 5 entries
- **WHEN** `CompactForAI()` is called
- **THEN** the `"quality_summary"` MUST contain `"worst_coverage_tests"` (5 entries), `"total_tests"`, `"average_contract_coverage"`, and `"ssa_degraded"`

### Requirement: Compact payload size budget

The compact payload MUST produce output under 300KB at realistic scale, because the unbounded arrays (`crap.scores`, `quality.quality_reports`, `classify.results`) MUST be excluded or bounded: `crap.scores` and `classify.results` MUST be dropped, and `quality.quality_reports` MUST be capped at 50 actionable entries. A unit test MUST verify that a representative synthetic payload representing 2000 functions, 1500 test-target pairs, and 5000 side effects compacts to under 300KB.

Previously: The compact payload had to stay under 300KB for projects with up to 500 analyzed functions, verified with a 200-function / 100-pair / 30-doc synthetic fixture.

#### Scenario: Size budget met at realistic scale
- **GIVEN** a synthetic `ReportPayload` representing 2000 functions, 1500 test-target pairs, 5000 side effects, and 30 documentation files with 15KB content each
- **WHEN** `CompactForAI()` is called
- **THEN** the output size MUST be less than 300KB

## REMOVED Requirements

None.
