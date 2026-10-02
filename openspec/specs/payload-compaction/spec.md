# Payload Compaction

## Purpose

Define the AI-facing payload compaction for `gaze report` so the payload
delivered to AI adapters stays within model context-window limits while
preserving the fields the reporter needs.

## Requirements

### Requirement: CompactForAI method

`ReportPayload` MUST provide a `CompactForAI() ([]byte, error)` method that produces a reduced JSON representation suitable for AI adapter consumption. The compact output MUST be valid JSON and MUST contain the following top-level fields: `summary`, `crap`, `quality`, `classify`, `docscan`, `errors`.

#### Scenario: Compact output is valid JSON
- **GIVEN** a fully populated `ReportPayload` with all four steps succeeded
- **WHEN** `CompactForAI()` is called
- **THEN** the returned bytes MUST be valid JSON that unmarshals without error

#### Scenario: Compact output includes summary
- **GIVEN** a `ReportPayload` with `Summary.CRAPload=12`, `Summary.GazeCRAPload=3`, `Summary.AvgContractCoverage=45`
- **WHEN** `CompactForAI()` is called
- **THEN** the output MUST contain a top-level `"summary"` object with `"crapload": 12`, `"gaze_crapload": 3`, `"avg_contract_coverage": 45`

#### Scenario: Compact output respects step failures
- **GIVEN** a `ReportPayload` where the Quality step failed (Quality is nil, Errors.Quality is non-nil)
- **WHEN** `CompactForAI()` is called
- **THEN** the `"quality"` field MUST be `null` and `"errors"` MUST contain the quality error message

#### Scenario: All steps failed
- **GIVEN** a `ReportPayload` with all four step fields nil and all four `Errors` fields populated
- **WHEN** `CompactForAI()` is called
- **THEN** the output MUST be valid JSON with all step fields as `null` and the `"errors"` object containing all four error messages

#### Scenario: Step succeeds with empty data
- **GIVEN** a `ReportPayload` where Docscan succeeded but returned an empty array (`[]`)
- **WHEN** `CompactForAI()` is called
- **THEN** the `"docscan"` field MUST be `[]` (empty array), not `null`

### Requirement: Docscan content exclusion

The compact payload MUST strip the `content` field from docscan entries. Each docscan entry MUST retain only `path` and `priority` fields.

#### Scenario: Docscan paths without content
- **GIVEN** a `ReportPayload` with docscan containing 3 documents each with `path`, `content`, and `priority`
- **WHEN** `CompactForAI()` is called
- **THEN** the `"docscan"` array MUST contain 3 entries, each with `"path"` and `"priority"` only, and no `"content"` field

#### Scenario: Docscan step failure preserved
- **GIVEN** a `ReportPayload` where Docscan is nil
- **WHEN** `CompactForAI()` is called
- **THEN** the `"docscan"` field MUST be `null`

### Requirement: Self-contained quality side-effect projection

When the compact payload projects a quality side effect (a coverage gap, a discarded return, or an ambiguous effect), it MUST emit a minimal side-effect object carrying `id`, `type`, `tier`, `location`, `description`, and `target`, and MUST NOT include the `classification` field. This projection MUST carry enough detail for the reporter to render "function name, effect type, and description" without cross-referencing the Classify output.

#### Scenario: Gap carries description inline
- **GIVEN** a Quality report whose `contract_coverage.gaps` contains a side effect with `id="se-a1b2c3d4"`, `type="ErrorReturn"`, `tier="P0"`, `location="file.go:42"`, `description="returns an error that is never asserted"`, `target="DoThing"`, and a `classification` object
- **WHEN** `CompactForAI()` is called
- **THEN** the projected gap MUST contain `"id": "se-a1b2c3d4"`, `"type": "ErrorReturn"`, `"tier": "P0"`, `"location": "file.go:42"`, `"description": "returns an error that is never asserted"`, `"target": "DoThing"`, and MUST NOT contain a `"classification"` field

### Requirement: Quality data compaction

The compact payload MUST replace full `SideEffect` objects in Quality report fields with arrays of self-contained minimal side-effect objects (see "Self-contained quality side-effect projection"). `contract_coverage.gaps` MUST be projected to `contract_coverage.gaps`, `contract_coverage.discarded_returns` MUST be projected to `contract_coverage.discarded_returns`, and `ambiguous_effects` MUST be projected to `ambiguous_effects`. Each projected object MUST carry `id`, `type`, `tier`, `location`, `description`, and `target`, and MUST omit `classification` and `signals`. The `contract_coverage.gap_hints` and `contract_coverage.discarded_return_hints` fields MUST be preserved as-is, and all scalar fields on `ContractCoverage` and `QualityReport` MUST be preserved.

#### Scenario: Quality gaps projected to self-contained objects
- **GIVEN** a Quality report with a test-target pair whose `contract_coverage.gaps` contains a full `SideEffect` with `id="se-a1b2c3d4"`, `type="MapMutation"`, `description="mutates a shared map"`
- **WHEN** `CompactForAI()` is called
- **THEN** the corresponding quality entry's `contract_coverage` MUST contain `"gaps"` with an entry carrying `"id":"se-a1b2c3d4"`, `"type":"MapMutation"`, `"description":"mutates a shared map"`, and each gap MUST NOT contain a `"classification"` field

#### Scenario: Quality hints preserved
- **GIVEN** a Quality report with `contract_coverage.gap_hints: ["assert err != nil", "check return value"]`
- **WHEN** `CompactForAI()` is called
- **THEN** the `contract_coverage.gap_hints` field MUST be `["assert err != nil", "check return value"]`

#### Scenario: Cross-reference resilience when Classify fails
- **GIVEN** a `ReportPayload` where the Classify step failed (Classify is nil) but Quality has gaps with side effects
- **WHEN** `CompactForAI()` is called
- **THEN** the quality gaps MUST still be emitted with their `type` and `description` inline

### Requirement: Quality reports bounding

The compact payload MUST bound the `quality_reports` array to a fixed maximum independent of the total number of test-target pairs. The compact output MUST include only actionable reports (those with at least one coverage gap, discarded return, ambiguous effect, or unmapped assertion), ordered by ascending contract coverage, capped at 50 entries.

#### Scenario: Reports bounded to actionable entries
- **GIVEN** a Quality output with 1500 test-target pairs, of which 400 are actionable (have gaps or unmapped assertions)
- **WHEN** `CompactForAI()` is called
- **THEN** the `"quality_reports"` array MUST contain at most 50 entries, each being an actionable report, and MUST NOT include reports with no gaps, discarded returns, ambiguous effects, or unmapped assertions

### Requirement: Quality summary deduplication

The compact payload MUST preserve `quality_summary.worst_coverage_tests` (bounded to the bottom 5 tests by contract coverage). The compact payload MUST bound the `quality_reports` array to at most 50 actionable entries. All other summary fields (`total_tests`, `average_contract_coverage`, `total_over_specifications`, `assertion_detection_confidence`, `ssa_degraded`, `ssa_degraded_packages`, `skipped_tests`, `skipped_test_names`) MUST be preserved.

#### Scenario: Worst coverage tests preserved
- **GIVEN** a Quality output with `quality_summary.worst_coverage_tests` containing 5 entries
- **WHEN** `CompactForAI()` is called
- **THEN** the `"quality_summary"` MUST contain `"worst_coverage_tests"` (5 entries), `"total_tests"`, `"average_contract_coverage"`, and `"ssa_degraded"`

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

### Requirement: Signal stripping in Classify output

The compact payload MUST omit the Classify `results` array entirely and emit only a counts-only `"classify"` object with `contractual`, `ambiguous`, and `incidental` integer counts plus the `version` string. The counts MUST be sourced from the top-level summary.

#### Scenario: Classify results dropped, counts emitted
- **GIVEN** a Classify result with many side effects and summary counts `contractual=2000`, `ambiguous=900`, `incidental=500`
- **WHEN** `CompactForAI()` is called
- **THEN** the `"classify"` field MUST contain `"contractual": 2000`, `"ambiguous": 900`, `"incidental": 500`, and MUST NOT contain `"results"` or any side-effect or `"signals"` data

### Requirement: CRAP scores array exclusion

The compact payload MUST NOT include the full `scores` array from the CRAP output. The CRAP compact output MUST retain the `summary` object (including the bounded worst-offender lists) and MUST NOT contain a `"scores"` field.

#### Scenario: Scores array omitted
- **GIVEN** a CRAP report with `scores` containing 2000 entries and `summary.worst_crap` containing 5 entries
- **WHEN** `CompactForAI()` is called
- **THEN** the CRAP output MUST contain a `"summary"` object with `"worst_crap"` and MUST NOT contain a `"scores"` field

### Requirement: CRAP summary deduplication

The compact payload MUST preserve `summary.worst_crap`, `summary.worst_gaze_crap`, and `summary.recommended_actions` (bounded lists of at most 5, 5, and 20 entries respectively). The compact payload MUST omit the full `scores` array. All other summary fields (`total_functions`, `avg_complexity`, `avg_line_coverage`, `avg_crap`, `crapload`, `crap_threshold`, `gaze_crapload`, `gaze_crap_threshold`, `avg_gaze_crap`, `avg_contract_coverage`, `quadrant_counts`, `fix_strategy_counts`, `ssa_degraded_packages`) MUST be preserved.

#### Scenario: Worst offender lists preserved, scores omitted
- **GIVEN** a CRAP report with `summary.worst_crap` containing 5 entries and `scores` containing 2000 entries
- **WHEN** `CompactForAI()` is called
- **THEN** the CRAP output MUST contain `"summary"` with `"worst_crap"` (5 entries), `"worst_gaze_crap"`, and `"recommended_actions"`, and MUST NOT contain a `"scores"` field

### Requirement: Compact payload size budget

The compact payload MUST produce output under 300KB at realistic scale, because the unbounded arrays (`crap.scores`, `quality.quality_reports`, `classify.results`) MUST be excluded or bounded: `crap.scores` and `classify.results` MUST be dropped, and `quality.quality_reports` MUST be capped at 50 actionable entries. A unit test MUST verify that a representative synthetic payload representing 2000 functions, 1500 test-target pairs, and 5000 side effects compacts to under 300KB.

#### Scenario: Size budget met at realistic scale
- **GIVEN** a synthetic `ReportPayload` representing 2000 functions, 1500 test-target pairs, 5000 side effects, and 30 documentation files with 15KB content each
- **WHEN** `CompactForAI()` is called
- **THEN** the output size MUST be less than 300KB

### Requirement: AI adapter payload delivery

`runTextPath()` MUST call `payload.CompactForAI()` instead of `json.Marshal(payload)` when delivering the payload to the AI adapter. The `--format=json` path (`runJSONPath()`) MUST remain unchanged, continuing to use `json.Encoder` with full-fidelity output.

#### Scenario: Text path uses compact payload
- **GIVEN** a report run with `--ai=opencode --format=text`
- **WHEN** the pipeline completes and invokes the AI adapter
- **THEN** the adapter MUST receive the compact payload (with docscan content stripped, signals omitted, etc.) not the full `json.Marshal` output

#### Scenario: JSON path unchanged
- **GIVEN** a report run with `--format=json`
- **WHEN** the pipeline completes
- **THEN** stdout MUST contain the full pretty-printed JSON with all fields including docscan content, signals, worst offender lists, and full side effect objects
