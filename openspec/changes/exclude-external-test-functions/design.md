## Context

External analyzers report complexity for both source and test files, but only instrument source files for coverage (e.g., Python `--cov=src`). Test functions therefore have 0% coverage and inflate CRAPload/quadrant counts/fix-strategy counts with phantom entries. The analyzer protocol's `discover` method already returns `test_files` but Gaze never consumes it. Go mode is unaffected because `go test -coverprofile` instruments `_test.go` files.

## Goals / Non-Goals

### Goals

- Exclude external-analyzer test files from CRAP scoring when `discover` capability is available.
- Store the `discover` result on the `Session` so the complexity provider and CLI wiring can access it without a second protocol call.
- Gracefully degrade when `discover` is unsupported or returns a protocol error — no hard errors, no data loss. (A `discover` timeout is the exception: the transport kills the shared analyzer subprocess on deadline, so `Initialize` fails fast rather than continuing with a poisoned client.)
- Distinguish "test function with no target effects" from "production function with unasserted contracts" in quality reports, without emitting schema-invalid JSON.
- Ensure Go-mode behavior is unchanged (zero-impact on existing code paths).

### Non-Goals

- Changing the `_test.go` skip in `computeScores` — that path remains Go-specific and is correct for Go.
- Adding a `discover` call to the Go-native provider path (goprovider).
- Modifying the analyzer protocol wire format — `DiscoverResult` is already defined in protocol 1.1.0.
- Filtering test files from `analyze` (side-effect detection) — test files legitimately contain side effects to analyze for contract coverage purposes.
- Handling test-file coverage in the `coverage` protocol method — the analyzer is already responsible for scoping its coverage instrumentation.

## Decisions

### D1: Filter at the complexity-provider level, not in `computeScores`

The `computeScores` function in `internal/crap/analyze.go` already has a Go-specific `_test.go` skip. Adding a general-purpose exclusion list to `crap.Options` and threading it through `computeScores` would require an `ExcludedFiles map[string]bool` field on `Options`, which couples the universal scoring engine to a language-specific discovery mechanism.

Instead, the `ExternalComplexityProvider.Analyze` method filters the returned `[]FunctionComplexity` before returning it to the CRAP pipeline. This keeps the filtering at the adapter layer, where it belongs, and avoids adding a general-purpose field to `Options` that only external-analyzer callers would populate.

The provider receives the excluded-file set via constructor injection: `NewExternalComplexityProvider(client, testFiles)`. `Session.Initialize()` calls `discover` BEFORE constructing the complexity provider and passes `s.testFiles` to the constructor — guaranteeing the filter is populated before `Analyze` is ever invoked (deterministic ordering, no silent no-op).

**Constitution alignment**: Minimal Assumptions — the filtering is an adapter concern; the universal scoring engine is unchanged. Accuracy — the provider derives its filter from the analyzer's own self-describing `discover` output.

### D2: Call `discover` in `Session.Initialize()` after the `initialize` handshake

The `discover` protocol method returns `source_files`/`test_files` — metadata that is stable for the session lifetime and is needed before any scoring operation. Calling it in `Initialize()` ensures the test-file set is available to all providers that need it without introducing a separate lifecycle step or requiring the CLI to call it explicitly.

The call is gated on `caps.Discover`. When the capability is absent, the `testFiles` field remains `nil` and all downstream filtering is a no-op. The call uses `protocol.ShortTimeout` (30s), consistent with the `initialize` handshake and the protocol's documented grouping of `discover` under the short-timeout constant — NOT `AnalysisTimeout` (5m), which would stall a hung analyzer's user-facing invocation.

This follows the same pattern as `DocCoverage()` and `Analyze()` on `Session` — capability-gated protocol calls that return `nil, nil` when the feature is unavailable.

**Constitution alignment**: Minimal Assumptions — analyzers without `discover` are unaffected. Actionable Output — graceful degradation with a stderr warning for protocol errors; a `discover` timeout fails fast (the transport kills the shared subprocess on deadline, so fail-fast is the only safe response).

### D3: Store `testFiles` as a `map[string]bool` of cleaned relative paths

The protocol's `discover` `test_files`, the complexity method's `FunctionComplexityData.File`, and `test_mapping`'s `AssertionMappingData.TestFile` all use paths relative to the analyzer's `root_path` (the same value sent as `RootPath` in each method's params). The adapters copy these paths through verbatim, so no absolute-path resolution is required — matching is relative-to-relative.

The session normalizes each `test_files` entry with `filepath.Clean` and stores the set as a `map[string]bool` for O(1) lookup. Consumers (the complexity filter and the quality sentinel) also apply `filepath.Clean` to the path being tested before the lookup, so `./tests/test_ops.py` and `tests/test_ops.py` match.

### D4: `no_contract_expected` boolean + `reason` for "no contract expected" in quality reports

A test function whose target has zero detected side effects cannot have a meaningful contract coverage score. Rather than encoding this as an out-of-range percentage (a `-1` sentinel would violate the JSON schema's `percentage` `minimum: 0, maximum: 100` constraint and emit schema-invalid JSON), Gaze adds two fields to `taxonomy.ContractCoverage` (which currently has neither):

- `NoContractExpected bool` (JSON tag `no_contract_expected,omitempty`) — the machine-readable signal.
- `Reason string` (JSON tag `reason,omitempty`) — carries `"test_function_no_target_effects"`.

When the sentinel fires, `Percentage` remains `0` (a legal value). The sentinel fires only when BOTH conditions hold:

1. The test function's unioned target effects are empty (no detected side effects), AND
2. The test function's file is confirmed as a test file via `discover` (`testFiles` membership).

Condition 2 is the accuracy guard: it prevents a production function with analyzer-undetected effects from being silently marked "no contract expected". When `discover` is unavailable (`testFiles == nil`), the sentinel does NOT fire and behavior falls back to today's 0% "needs assertions" reporting. This introduces a bounded residual risk: an analyzer whose side-effect detection is incomplete could under-report a test target's effects, and the sentinel's correctness is bounded by that detection accuracy (documented as a known limitation).

The package summary's `AverageContractCoverage` and `WorstCoverageTests` exclude `NoContractExpected` reports. When ALL reports are no-contract-expected, `AverageContractCoverage` is `0` (no division) and `WorstCoverageTests` is empty — avoiding `0/0 → NaN` which would make `json.Marshal` fail. `TotalTests` and `AssertionDetectionConfidence` still count every report (they are still test functions).

Note: this `Reason` field shares the JSON key `reason` with the pre-existing `taxonomy.PackageSummary.Reason` (domains `test_mapping_unavailable` / `test_mapping_error`), but at a different nesting level (`quality_reports[].contract_coverage.reason` vs `quality_summary.reason`); both domains are documented in `docs/reference/json-schemas.md`.

### D5: No wire-protocol changes

`DiscoverResult` is already defined in protocol 1.1.0 with `source_files`, `test_files`, and `framework` fields. The `Capabilities.Discover` boolean is already defined. The `DiscoverParams` struct already exists. No protocol version bump is needed — this is a consumer-side change that finally uses the existing protocol surface.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Analyzer's `test_files` list is incomplete and misses some test files | Low risk — the list is the analyzer's own declaration of what it considers test files. If it's wrong, that's an analyzer bug, not a Gaze bug. Users can still see the full report via Go mode (which is unaffected). |
| `discover` call adds latency to session initialization | The call is made once per session under `ShortTimeout` (30s), and most external analyzers return `discover` results in O(ms). The benefit (correct CRAP reports) outweighs the cost. |
| Path mismatch between `test_files` and complexity/`test_mapping` paths (e.g., an analyzer reports absolute paths or a different root) | Use `filepath.Clean` on both sides. All three path sources originate from the same analyzer under the same `root_path`, so relative-to-relative matching is correct; if an analyzer reports absolute paths, `filepath.Clean` comparison still works. |
| Quality sentinel misclassifies a production function with analyzer-undetected effects | Gated on `discover`-confirmed `test_files` membership (condition 2 of D4), so a production function's file is never in `test_files`. Residual risk is bounded by analyzer side-effect-detection accuracy and documented as a known limitation. |
| Existing JSON consumers ignore the new `no_contract_expected` boolean | The new fields are additive and optional; `percentage` stays within 0–100, so no existing schema or consumer breaks. A schema update (`no_contract_expected`, `reason`) documents the new fields. |
