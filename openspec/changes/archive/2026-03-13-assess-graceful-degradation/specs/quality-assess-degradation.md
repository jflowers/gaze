## ADDED Requirements

### Requirement: SSADegraded field on PackageSummary

`PackageSummary` MUST include an `SSADegraded bool` field (JSON tag `ssa_degraded`) that is `true` when SSA construction failed for the package and quality results are partial.

The field MUST default to `false` when SSA construction succeeds.

#### Scenario: SSA build succeeds
- **GIVEN** a test package where `BuildTestSSA` completes without error
- **WHEN** `Assess` is called with the package
- **THEN** the returned `PackageSummary.SSADegraded` MUST be `false`

#### Scenario: SSA build fails
- **GIVEN** a test package where `BuildTestSSA` returns an error
- **WHEN** `Assess` is called with the package
- **THEN** the returned `PackageSummary.SSADegraded` MUST be `true`

### Requirement: QualitySchema includes ssa_degraded

The `QualitySchema` JSON Schema definition for `PackageSummary` MUST include `ssa_degraded` as an optional boolean property. The field MUST NOT be in the `required` array.

#### Scenario: Schema validates degraded output
- **GIVEN** JSON output from `Assess` with `ssa_degraded: true` in the summary
- **WHEN** the output is validated against `QualitySchema`
- **THEN** validation MUST pass

#### Scenario: Schema validates non-degraded output without field
- **GIVEN** JSON output from `Assess` that omits the `ssa_degraded` field
- **WHEN** the output is validated against `QualitySchema`
- **THEN** validation MUST pass (backward compatible)

### Requirement: Degraded warning on stderr

When `BuildTestSSA` fails and `opts.Stderr` is non-nil, `Assess` MUST emit a warning message to `opts.Stderr` that identifies the package and explains SSA analysis was skipped.

#### Scenario: Warning emitted on SSA failure
- **GIVEN** a test package where `BuildTestSSA` returns an error and `opts.Stderr` is a buffer
- **WHEN** `Assess` is called
- **THEN** `opts.Stderr` MUST contain a warning mentioning SSA and the package

#### Scenario: No warning when stderr is nil
- **GIVEN** a test package where `BuildTestSSA` returns an error and `opts.Stderr` is nil
- **WHEN** `Assess` is called
- **THEN** no panic occurs and results are returned normally

## MODIFIED Requirements

### Requirement: Assess error semantics on BuildTestSSA failure

Previously: `Assess` returned `(nil, nil, fmt.Errorf("building test SSA: %w", err))` when `BuildTestSSA` failed.

`Assess` MUST return `(reports, summary, nil)` when `BuildTestSSA` fails, where:
- `reports` contains one `QualityReport` per test function with:
  - `TestFunction` and `TestLocation` populated from AST
  - `AssertionDetectionConfidence` populated from `DetectAssertions`
  - `TargetFunction` set to zero-valued `FunctionTarget{}`
  - `ContractCoverage` set to zero-valued `ContractCoverage{}`
  - `OverSpecification` set to zero-valued `OverSpecificationScore{}`
  - `UnmappedAssertions` and `AmbiguousEffects` set to nil
- `summary` has `SSADegraded: true` and reflects the degraded reports
- The returned error MUST be nil

#### Scenario: Assess returns partial results on SSA failure
- **GIVEN** a test package with 3 test functions where `BuildTestSSA` returns an error
- **WHEN** `Assess` is called
- **THEN** `Assess` returns 3 `QualityReport` entries, a non-nil `PackageSummary` with `SSADegraded: true` and `TotalTests: 3`, and a nil error

#### Scenario: Assess returns full results when SSA succeeds
- **GIVEN** a test package where `BuildTestSSA` completes successfully
- **WHEN** `Assess` is called
- **THEN** `Assess` returns full-fidelity reports with inferred targets, mapped assertions, and `SSADegraded: false` — identical to current behavior

### Requirement: gaze quality CLI exit code on SSA failure

Previously: `runQuality` returned `fmt.Errorf("quality assessment: %w", err)` causing a non-zero exit.

`gaze quality` MUST exit 0 when SSA construction fails, printing available degraded results. The warning on stderr MUST inform the user that results are partial.

#### Scenario: gaze quality exits 0 on SSA failure
- **GIVEN** running `gaze quality <pkg>` where SSA construction fails
- **WHEN** the command completes
- **THEN** exit code is 0, stderr contains an SSA warning, and stdout contains the degraded report

### Requirement: gaze report includes degraded packages

Previously: `runQualityForPackage` returned nil on `Assess` error, silently dropping the package.

`runQualityForPackage` MUST return the degraded reports from `Assess` instead of nil when SSA fails.

#### Scenario: degraded package included in report
- **GIVEN** running `gaze report` where one package's SSA construction fails
- **WHEN** the report quality step processes that package
- **THEN** the package's degraded quality reports are included in the output (not silently dropped)

### Requirement: gaze crap logs warning on SSA failure

Previously: `analyzePackageCoverage` logged at debug level and returned nil.

`analyzePackageCoverage` SHOULD log a visible warning (not just debug) when quality results are degraded, and MUST return the degraded reports instead of nil.

#### Scenario: crap pipeline uses degraded reports
- **GIVEN** running `gaze crap` where one package's SSA construction fails
- **WHEN** contract coverage is computed for that package
- **THEN** degraded quality reports are returned (not nil) and a warning is logged

## REMOVED Requirements

None.
