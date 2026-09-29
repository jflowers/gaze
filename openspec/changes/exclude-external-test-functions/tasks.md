<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Session discover integration

- [x] 1.1 Add `testFiles map[string]bool` field to `Session` in `internal/adapter/session.go`; add a `DiscoverTestFiles() map[string]bool` accessor that returns `nil` when `discover` was not called
- [x] 1.2 In `Session.Initialize()`, after the `initialize` handshake and capability capture, gate on `caps.Discover` and call the `discover` protocol method once BEFORE constructing the providers; on success normalize each `test_files` entry with `filepath.Clean` and populate `s.testFiles`; on protocol error/timeout log a warning to `s.stderr` (format `"warning: discover failed: %v"`) and leave `testFiles` nil (graceful fallback)
- [x] 1.3 Add `Session.discoverTestFiles()` performing the `discover` call inline (mirroring the `DocCoverage()` capability-gated pattern), using `protocol.ShortTimeout` (30s) for the call context (NOT `AnalysisTimeout`); returns an error on timeout so `Initialize` can fail fast

## 2. Complexity provider filtering

- [x] 2.1 Add a `testFiles map[string]bool` field to `ExternalComplexityProvider` in `internal/adapter/complexity.go`, injected via the constructor `NewExternalComplexityProvider(client, testFiles)`; in `Session.Initialize()` pass `s.testFiles` to the constructor
- [x] 2.2 In `ExternalComplexityProvider.Analyze`, filter the result of `convertComplexity` by dropping any `FunctionComplexity` whose `filepath.Clean(File)` is in `testFiles`; return the filtered slice (no-op when `testFiles` is nil/empty)

## 3. Quality contract-coverage sentinel

- [x] 3.1 Add `NoContractExpected bool` (JSON tag `no_contract_expected,omitempty`) and `Reason string` (JSON tag `reason,omitempty`) fields to `taxonomy.ContractCoverage` in `internal/taxonomy/types.go` (both omitempty so they are absent — not `false`/`""` — for production functions, matching the schema's optionality)
- [x] 3.2 Change `BuildQualityFromMappings` in `internal/adapter/quality.go` to accept a `testFiles map[string]bool` parameter; when a test function's unioned target effects are empty AND `testFiles` is non-nil AND `filepath.Clean(testFile)` is in `testFiles`, set `ContractCoverage.NoContractExpected = true`, `Reason = "test_function_no_target_effects"`, and leave `Percentage` at 0 (a legal value). Update all existing call sites (including ~13 in `quality_internal_test.go`) for the new signature.
- [x] 3.3 In `buildQualitySummary`, exclude `NoContractExpected` reports from `AverageContractCoverage` and `WorstCoverageTests`; when the exclusion leaves zero non-sentinel reports, set `AverageContractCoverage = 0` and `WorstCoverageTests` empty (no division by zero). `TotalTests` and `AssertionDetectionConfidence` still count every report.

## 4. JSON schema update

- [x] 4.1 Add the optional `no_contract_expected` (boolean) and `reason` (string) properties to the contract coverage object in `internal/report/schema.go` (`percentage` stays `minimum: 0, maximum: 100`). Extend an existing schema-validation test (or add one) to validate `no_contract_expected`/`reason` output against the updated schema.

## 5. CLI wiring

- [x] 5.1 Pass the session's discovered `testFiles` (via `DiscoverTestFiles()`) into `BuildQualityFromMappings` in the `gaze quality` external path (`runQualityWithExternalAnalyzer` in `cmd/gaze/main.go`); verify `runCrapWithExternalAnalyzer` and `runQualityWithExternalAnalyzer` compile against the updated types

## 6. Tests

- [x] 6.1 [P] Add unit tests in `internal/adapter/complexity_test.go` for `ExternalComplexityProvider.Analyze` filtering (nil set, empty set, populated set with relative paths, populated set with no match, `./`-prefixed path normalization — the `./` prefix is on the complexity `File` side, matching a pre-cleaned set entry)
- [x] 6.2 Add unit tests in `internal/adapter/session_test.go` (or equivalent) for `Initialize()` discover integration (discover capable → testFiles populated + provider filter set; uncapable → no call; discover error → nil set + warning with asserted stderr substring containing `"discover failed"`). Assert `discover` is called exactly once (fake analyzer counts invocations per method). Depends on 6.4's fake-analyzer flags/counter; runs after 6.4 in the sequential phase.
- [x] 6.3 [P] Add unit tests in `internal/adapter/quality_internal_test.go` for the no-contract-expected sentinel (empty target effects + test_files membership → NoContractExpected + reason + Percentage 0; empty effects but test file NOT in testFiles → normal 0%; discover unavailable (nil) → normal 0%; summary excludes sentinel entries; all-sentinel → AverageContractCoverage 0, no panic; TotalTests and AssertionDetectionConfidence still count sentinel reports; `./`-prefixed `TestFile` `./tests/test_ops.py` still matches a `test_files` entry `tests/test_ops.py`)
- [x] 6.4 Extend the fake analyzer (`internal/protocol/testdata/fake_analyzer/main.go`): add a test-file complexity entry (e.g. `tests/test_ops.py::test_add`) to the `complexity` response, a `--no-discover` capability-disable flag, a discover-error flag, and a per-method invocation counter (e.g. `map[string]int` keyed by method, surfaced via a `--report-counts` flag or stderr marker) so 6.2 can assert discover is called exactly once. Add integration tests for: (a) discover-capable → CRAP report excludes test file, retains source files, and NO `add_tests` fix-strategy entry / quadrant / fix-strategy count is produced for the test file; (b) `--no-discover` → no discover call, no filtering; (c) discover-error → graceful fallback, all functions scored. Repair the existing `TestExternalComplexityProvider` in `internal/adapter/adapter_test.go` (hard-codes `want 3`): the new test-file complexity entry raises the unfiltered count to 4, so either call `SetTestFiles` with the test-file set (filtering back to 3) or update the expected count.
- [x] 6.6 Add a quality-path integration test (extend `cmd/gaze/external_analyzer_test.go`): give the fake analyzer a `test_mapping` entry mapping `test_add` → target `add` with `assertion_type: ""` (unmapped/no-effect assertion) where `add` has empty `side_effects` in the `analyze` response, then assert `gaze quality --analyzer <fake>` emits `no_contract_expected: true` and `reason: "test_function_no_target_effects"` in JSON output (verifies the task 5.1 glue actually threads `DiscoverTestFiles()` through). Repair the existing `TestQualityWithExternalAnalyzer_HappyPath` (hard-codes 3 reports / `TotalTests == 3` / 3-entry confidence map / `AssertionDetectionConfidence != 67`): update expectations for the new 4th sentinel report (`test_add` counted in `TotalTests` and in `AssertionDetectionConfidence`, excluded from `AverageContractCoverage`; recompute the confidence mean deterministically from `test_add`'s `assertion_type: ""`). Not parallel with 6.4 — both edit `fake_analyzer/main.go` (and each edits a distinct test file); runs after 6.4 in document order.
- [x] 6.5 Verify existing Go-mode AND external-analyzer tests are unaffected after the fixture changes (covered by the final full-suite run in 7.4)

## 7. Documentation and verification

- [x] 7.1 Update `docs/protocol.md`: note that `discover` `test_files` is now consumed to filter CRAP scoring; fix the two stale "not currently consumed" statements (the `discover` description and the error-handling table's "no impact" row) to describe the new fallback semantics
- [x] 7.2 Update `docs/reference/json-schemas.md` (and any porting docs pinning contract coverage to "0–100") to document the new `no_contract_expected` and `reason` fields on `ContractCoverage`
- [x] 7.3 File a `unbound-force/website` docs issue tracking the JSON/report output change (user-facing external-analyzer CRAP filtering + additive `no_contract_expected`/`reason` JSON fields — no exemption applies under the Website Documentation Gate)
- [x] 7.4 Run `go build ./cmd/gaze`, `go vet ./...`, `golangci-lint run`, and `go test -race -count=1 -short ./...` to satisfy the CI parity gate (the full-suite run is the final verification covering all new §6 tests)
- [x] 7.5 Verify constitution alignment: confirm the change passes Accuracy (no phantom test-function CRAP entries; sentinel gated on test_files membership), Minimal Assumptions (filter derives from analyzer's own `discover`, no user annotation), Actionable Output (fix-strategy counts no longer polluted), and Testability (new unit tests cover the filtering and sentinel paths)

<!-- spec-review: passed -->
<!-- code-review: passed -->
