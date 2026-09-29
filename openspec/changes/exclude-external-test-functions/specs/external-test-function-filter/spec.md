## ADDED Requirements

### Requirement: External analyzer test-file exclusion from CRAP scoring

When an external analyzer advertises the `discover` capability in its `initialize` response, Gaze MUST call the `discover` protocol method after `initialize` completes and before scoring begins. The returned `test_files` list MUST be used to filter complexity entries out of CRAP scoring: any `FunctionComplexity` whose `File` path (after `filepath.Clean`) matches an entry in `test_files` (also `filepath.Clean`ed) MUST be excluded before scoring (never reaching `computeScores`), and therefore also excluded from quadrant classification, CRAPload/GazeCRAPload totals, fix-strategy counts, `recommended_actions`, and worst-offenders lists.

The `test_files` paths from `discover`, the `complexity` method's `File` values, and the `test_mapping` method's `TestFile` values are all relative to the analyzer's `root_path`. Matching SHALL be relative-to-relative using `filepath.Clean` on both sides (no absolute-path resolution).

#### Scenario: discover-capable analyzer produces clean CRAP report

- **GIVEN** an external analyzer that advertises `discover: true` and returns `test_files: ["tests/test_ops.py"]` from `discover`, whose `complexity` method reports functions in both `src/ops.py` and `tests/test_ops.py`
- **WHEN** `gaze crap --analyzer <name> --language python ./src` is run
- **THEN** functions in `tests/test_ops.py` SHALL NOT appear in CRAP scores, CRAPload, quadrant counts, fix-strategy counts, or `recommended_actions`
- **AND** functions in `src/ops.py` SHALL appear normally
- **AND** the module-wide line-coverage average SHALL reflect only source-file functions

#### Scenario: discover-uncapable analyzer preserves existing behavior

- **GIVEN** an external analyzer that does NOT advertise `discover` capability
- **WHEN** `gaze crap --analyzer <name>` is run
- **THEN** behavior SHALL be identical to current (no test-file filtering)
- **AND** no `discover` protocol call SHALL be made

#### Scenario: discover returns empty test_files

- **GIVEN** an external analyzer whose `discover` method returns `test_files: []`
- **WHEN** `gaze crap --analyzer <name>` is run
- **THEN** no functions SHALL be excluded from CRAP scoring

#### Scenario: discover returns a protocol error and degrades gracefully

- **GIVEN** an external analyzer whose `discover` method returns a protocol error (e.g., method not found despite capability flag)
- **WHEN** `gaze crap --analyzer <name>` is run
- **THEN** Gaze SHALL log a warning to stderr and fall back to no filtering
- **AND** the CRAP analysis SHALL complete with all functions scored (no data loss)

#### Scenario: discover times out and fails fast

- **GIVEN** an external analyzer whose `discover` method hangs and does not respond within `protocol.ShortTimeout`
- **WHEN** `gaze crap --analyzer <name>` is run
- **THEN** `Session.Initialize()` SHALL return an error and the run SHALL abort
- **AND** the error SHALL wrap `context.DeadlineExceeded` so callers can distinguish a discover timeout from a protocol error
- **RATIONALE** a discover timeout kills the shared analyzer subprocess (the transport kills the process on context deadline), so continuing would leave every subsequent provider call operating on a dead client; failing fast is the only safe response

### Requirement: Discover call placement in session initialization

The `discover` protocol call MUST occur during `Session.Initialize()` after the `initialize` handshake succeeds and the capabilities are known, and BEFORE the complexity provider is constructed (so the filter is populated before any `Analyze` call). It SHALL be gated on `caps.Discover` — no call is made when the capability is absent. The call SHALL use `protocol.ShortTimeout` (30s), NOT `AnalysisTimeout`. The resulting `test_files` set SHALL be stored on the `Session` struct for later access by the complexity provider and the CLI wiring.

#### Scenario: discover is called exactly once per session

- **GIVEN** an external analyzer with `discover: true`
- **WHEN** `Session.Initialize()` is called
- **THEN** `discover` SHALL be called exactly once after `initialize`
- **AND** the result SHALL be stored on the Session for the session lifetime

## MODIFIED Requirements

### Requirement: ExternalComplexityProvider gains optional test-file filter

The `ComplexityProvider.Analyze` method signature is unchanged, but the `ExternalComplexityProvider` implementation MUST accept an optional set of excluded file paths (cleaned relative paths) to filter from its returned `[]FunctionComplexity`. The filter SHALL be applied in `Analyze` after `convertComplexity`, comparing `filepath.Clean(File)` against the set, and SHALL be a no-op when the set is empty or nil.

#### Scenario: complexity provider filters test files

- **GIVEN** an `ExternalComplexityProvider` with a test-file set `{"tests/test_ops.py"}`
- **WHEN** `Analyze` is called and the analyzer returns complexity entries for both `src/ops.py` and `tests/test_ops.py`
- **THEN** only the `src/ops.py` entry SHALL appear in the returned `[]FunctionComplexity`

### Requirement: Contract coverage for test functions without targets

When `BuildQualityFromMappings` produces a `QualityReport` for a test function whose unioned target effects are empty AND whose test file is confirmed as a test file via `discover` (`testFiles` membership), the report's contract coverage SHALL be marked "no contract expected": `ContractCoverage.NoContractExpected` SHALL be `true`, `ContractCoverage.Reason` SHALL be `"test_function_no_target_effects"`, and `ContractCoverage.Percentage` SHALL remain `0` (a legal value, not an out-of-range sentinel). When `testFiles` is unavailable (nil) or the test file is not in `testFiles`, the sentinel SHALL NOT fire and behavior SHALL be as before (0% "needs assertions").

Reports marked `NoContractExpected` SHALL be excluded from the package summary's `AverageContractCoverage` and `WorstCoverageTests`. When ALL reports are no-contract-expected, `AverageContractCoverage` SHALL be `0` and `WorstCoverageTests` SHALL be empty. `TotalTests` and `AssertionDetectionConfidence` SHALL still count every report.

#### Scenario: test function with no target effects reports no-contract-expected

- **GIVEN** a `QualityReport` for test function `test_add` in `tests/test_ops.py` (present in `testFiles`) whose target function `add` has zero detected side effects
- **WHEN** `BuildQualityFromMappings` processes the report
- **THEN** `ContractCoverage.NoContractExpected` SHALL be `true`
- **AND** `ContractCoverage.Reason` SHALL be `"test_function_no_target_effects"`
- **AND** `ContractCoverage.Percentage` SHALL be `0`
- **AND** the report SHALL NOT contribute to `AverageContractCoverage` or `WorstCoverageTests`

#### Scenario: test file not confirmed via discover preserves existing behavior

- **GIVEN** a `QualityReport` for test function `test_add` whose target has zero detected side effects, but the test file is NOT in `testFiles` (or `testFiles` is nil because discover is unavailable)
- **WHEN** `BuildQualityFromMappings` processes the report
- **THEN** `ContractCoverage.NoContractExpected` SHALL be `false`
- **AND** `ContractCoverage.Percentage` SHALL be `0` (0% "needs assertions", as before)

#### Scenario: test function with target effects preserves existing behavior

- **GIVEN** a `QualityReport` for test function `test_add` whose target function `add` has detected side effects
- **WHEN** `BuildQualityFromMappings` processes the report
- **THEN** `ContractCoverage.Percentage` SHALL be computed as before (0–100)
- **AND** the report SHALL contribute normally to summary statistics

#### Scenario: all reports no-contract-expected produces valid summary

- **GIVEN** a package whose every test function is no-contract-expected
- **WHEN** `buildQualitySummary` aggregates the reports
- **THEN** `AverageContractCoverage` SHALL be `0` (no division by zero, no NaN)
- **AND** `WorstCoverageTests` SHALL be empty
- **AND** `--format=json` SHALL emit valid JSON
