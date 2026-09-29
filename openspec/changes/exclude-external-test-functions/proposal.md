## Why

When Gaze analyzes a non-Go project through an external analyzer, `gaze crap` and `gaze quality` score the project's **test** functions as if they were production code. The analyzer protocol's `discover` method already returns a source/test split (`source_files` / `test_files`), and the analyzer's `complexity` method returns entries for both sets — but the analyzer's `coverage` method only instruments its `src/` package (e.g. Python `--cov=src`). Test functions therefore have no coverage entry, default to 0% line coverage, and are CRAP-scored as uncovered production code.

On the `zero-dot-force/snake-eyes` repo this produces:

- 993 functions analyzed, **724 of which are test functions**.
- Module-wide CRAPload = 188; **172 of 188 Q4 (dangerous) functions are test functions**.
- Module-wide avg line coverage reads 24.9% while `src/`-only coverage is 92.1% (src-only CRAPload = 16).
- Remediation breakdown reports **174 phantom `add_tests` flags**, essentially all test functions.

This noise defeats the tool: CRAP/quality reports are dominated by test functions, CI gates driven by `gaze crap --max-crapload` on Python/Rust/TS projects trip on phantom violations, and users must manually filter "test functions with 0% coverage" out of every report.

Go mode is unaffected because Go's `-coverprofile` instruments `_test.go` files, so Go test functions carry real coverage data. The gap is specific to external analyzers whose coverage scope excludes tests by convention.

## What Changes

- **`crap` scoring excludes test files**: when an external analyzer advertises the `discover` capability, Gaze fetches `test_files` and filters those files out of the complexity/CRAP scoring pipeline, so test functions no longer contribute 0%-coverage CRAP entries, CRAPload, quadrant counts, or `add_tests` remediation flags.
- **Graceful degradation**: when `discover` is unsupported or unavailable, behavior falls back to today's semantics (no filtering) — no hard failure, no regression for analyzers without the capability.
- **`quality` distinguishes "test function, no contract expected"**: when a test function is confirmed (via `discover` `test_files`) to live in a test file and its target has no detected effects, contract coverage is reported as `no_contract_expected: true` with `reason: "test_function_no_target_effects"` and a `percentage` of 0 — distinct from a production function whose contract is unasserted (a real 0% gap).

## Capabilities

### New Capabilities

- `external-test-function-filter`: external-analyzer test files reported via `discover` are excluded from CRAP scoring, and test functions without an expected contract are distinguished from production functions with unasserted contracts.

## Impact

- `internal/adapter/` — the external complexity/coverage/quality providers gain a `discover`-aware filtering step.
- `internal/protocol/` — no wire changes; `DiscoverResult` (`source_files`/`test_files`) is finally consumed.
- `internal/crap/analyze.go` — unchanged; filtering happens at the adapter layer (`ExternalComplexityProvider`), so `computeScores` is never reached for test-file entries.
- `internal/quality/` — contract-coverage status for test functions distinguishes "no contract expected".
- `cmd/gaze/main.go` — the external `crap`/`quality` paths thread the discovered test-file set into the providers.
- Reports: CRAPload, quadrant counts, fix-strategy counts, and `recommended_actions` stop being dominated by test functions.

## Constitution Alignment

Assessed against the project constitution (`.specify/memory/constitution.md`): Accuracy, Minimal Assumptions, Actionable Output, Testability.

### I. Accuracy

**Assessment**: PASS

The change removes phantom test-function CRAP entries, restoring accurate CRAPload, quadrant, and fix-strategy counts. The quality "no contract expected" sentinel is gated on `discover`-confirmed `test_files` membership, so it cannot misclassify a production function with analyzer-undetected effects. Accuracy claims are backed by regression tests (§6 of `tasks.md`).

### II. Minimal Assumptions

**Assessment**: PASS

The test-file filter derives entirely from the analyzer's own self-describing `discover` output — no file-name heuristics, no user annotation, no source restructuring. When `discover` is unavailable, Gaze degrades gracefully to today's behavior (no filtering), making no assumptions about the analyzer.

### III. Actionable Output

**Assessment**: PASS

Reports again identify real production-code risk: `recommended_actions` and fix-strategy counts no longer recommend `add_tests` for test functions. The new `no_contract_expected`/`reason` fields give machine-readable JSON consumers a precise, non-ambiguous signal, and `percentage` remains within its documented 0–100 range.

### IV. Testability

**Assessment**: PASS

The filtering is a pure function of (complexity entries, test-file set) and the summary exclusion is a pure function of the report set — both unit-testable in isolation with synthetic data. Integration coverage flows through the existing fake analyzer, and Go-mode behavior is protected by the purely additive nature of the change plus the unchanged Go-mode test suite. Coverage strategy is specified in `design.md`/`tasks.md` (§6).
