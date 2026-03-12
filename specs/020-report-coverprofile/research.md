# Research: 020-report-coverprofile

**Feature**: Pass Pre-Generated Coverage Profile to `gaze report`
**Date**: 2026-03-12
**Status**: Complete — no NEEDS CLARIFICATION markers remain

---

## Decision 1: Where does the flag plumb through the call stack?

**Decision**: Add `coverProfile string` to `reportParams` and `CoverProfile string` to `aireport.RunnerOptions`. Thread it from `newReportCmd` → `reportParams` → `runReport` → `RunnerOptions` → `runProductionPipeline` → `runCRAPStep`.

**Rationale**: The existing `crap.Options.CoverProfile` field is the terminal sink — `crap.Analyze` already branches on it (lines 67–85 of `internal/crap/analyze.go`). Every intermediate layer (`RunnerOptions`, `runCRAPStep`) is the minimal additional surface needed to connect the CLI flag to that existing field. No new validation or business logic is required in the intermediate layers; all validation already lives in `crap.Analyze`.

**Alternatives considered**:
- Validate the path in `runReport` before passing it to `aireport.Run`. Rejected: validation is already complete in `crap.Analyze` (existence check, is-directory check). Duplicating it would create two places to keep in sync.
- Add a `CoverProfile` field directly on `aireport.RunnerOptions` and skip `reportParams`. Rejected: `reportParams` is the params struct for the testable CLI layer; omitting it would make `runReport` untestable with the flag.

---

## Decision 2: Should `runCRAPStep` accept `coverProfile` as a direct parameter or via a modified options struct?

**Decision**: Pass it as a plain `string` parameter to `runCRAPStep(patterns, moduleDir, coverProfile, stderr)`.

**Rationale**: `runCRAPStep` already takes three plain parameters (`patterns`, `moduleDir`, `stderr`). Adding `coverProfile string` maintains the same style — no need to introduce a new options struct for a single additional field. The function is internal to `aireport`, so adding a parameter is a non-breaking change.

**Alternatives considered**:
- Introduce a `crapStepOptions` struct. Rejected: over-engineering for a single field. The existing pattern uses plain parameters for this function.

---

## Decision 3: What validation does `runReport` perform on the supplied path?

**Decision**: Pre-flight validation in `runReport` before calling `aireport.Run`. When `coverProfile != ""`, `runReport` validates: (1) path exists (`os.Stat`), (2) path is a regular file (`!info.IsDir()`). If either check fails, `runReport` returns an error immediately — no analysis steps run. This satisfies FR-006 (hard exit, non-zero, clear error message) without conflicting with `runProductionPipeline`'s partial-failure architecture.

**Rationale**: The partial-failure architecture (`runProductionPipeline` swallows CRAP errors into `payload.Errors.CRAP` and returns `nil`) means that validation errors from `crap.Analyze` are not propagated as Go errors from `runReport`. Pre-flight validation in `runReport` is the only way to satisfy FR-006 without changing `runProductionPipeline`. The validation logic (`os.Stat` + `info.IsDir()`) is ~4 lines and is not duplication of `crap.Analyze` — it is an early gate that prevents silent partial failure on an obviously invalid input.

**Alternatives considered**:
- Delegate validation entirely to `crap.Analyze` (original Decision 3). Rejected: `runProductionPipeline` swallows CRAP errors into `payload.Errors.CRAP`; the error never surfaces as a Go error from `runReport`, so FR-006 cannot be satisfied.
- Option C (distinguish user-supplied vs. internal CRAP failures in `runProductionPipeline`). Rejected: adds branching logic to `runProductionPipeline` that violates the Zero-Waste Mandate; Option A is simpler.

**TOCTOU note**: There is a small window between the `os.Stat` pre-flight check and `ParseCoverProfile`'s `os.Open`. In shared CI environments with parallel jobs writing to the same coverage file, the file could be deleted or replaced in this window. `ParseCoverProfile` will return a parse or I/O error in that case — the error is descriptive and actionable. This race is accepted as a known limitation consistent with standard CLI tool behavior.

---

## Decision 4: How does the flag interact with `--format=json`?

**Decision**: `--coverprofile` works identically for both `--format=text` and `--format=json`. It is passed through to `runCRAPStep` regardless of format. No special handling required.

**Rationale**: The CRAP analysis step runs in both format paths (see `runProductionPipeline`). The flag is orthogonal to output format.

---

## Decision 5: What happens when `--coverprofile` is an empty string (the zero value)?

**Decision**: An empty string means "not provided" — `runCRAPStep` forwards `""` to `crap.Analyze`, which interprets `""` as "generate internally" (existing behavior). No additional nil/zero check needed.

**Rationale**: This is exactly how the existing `gaze crap --coverprofile` flag works (line 694 of `main.go`). Consistent behavior, no special-casing.

---

## Decision 6: Does `RunnerOptions.AnalyzeFunc` (the testing override) need to be updated?

**Decision**: No. `AnalyzeFunc` replaces the entire `runProductionPipeline` call (including `runCRAPStep`). Tests that use `AnalyzeFunc` to inject a `FakeAdapter` pipeline bypass `runCRAPStep` entirely and are unaffected by the new field. Only tests that exercise the production pipeline path would exercise `--coverprofile`.

**Rationale**: The existing test architecture already isolates the production pipeline behind `AnalyzeFunc`. The new field is plumbed through `RunnerOptions` but only used when `AnalyzeFunc == nil`.

**Known limitation**: `AnalyzeFunc` has signature `func(patterns []string, moduleDir string) (*ReportPayload, error)` — it does **not** receive `RunnerOptions`. Tests that use `AnalyzeFunc` therefore cannot directly observe `opts.CoverProfile`. To verify `CoverProfile` wiring at the `cmd/gaze` layer, use the `runnerFunc` spy (the `reportParams.runnerFunc` override, which does receive `RunnerOptions`). To verify the `RunnerOptions→runCRAPStep→crap.Options` wiring, use a real `runCRAPStep` call with a static fixture (T013).

---

## Decision 7: What tests are required?

**Decision**: Three test categories in `cmd/gaze/main_test.go` (contract tests for the CLI layer) and one in `internal/aireport/runner_steps_test.go` (unit test for `runCRAPStep`):

1. **`TestRunReport_CoverProfile_ValidPath`** (unit — spy pattern): use a spy `runnerFunc` (the `reportParams.runnerFunc` override) that captures the `RunnerOptions` passed to it. Assert `capturedOpts.CoverProfile == suppliedPath` and `callCount == 1`. **Not** guarded by `testing.Short()` — the spy intercepts before `aireport.Run`, so no subprocess or package loading occurs. *Note: this supersedes the earlier "integration / real profile / mtime check" approach; the spy pattern is more reliable and avoids a slow test.* The `runnerFunc` spy intercepts at the `cmd/gaze` layer only — the `RunnerOptions→runCRAPStep` wiring leg is covered by test 4 below.

2. **`TestRunReport_CoverProfile_NonexistentPath`** (unit): call `runReport` with `coverProfile = "/nonexistent/path.out"`. Verify error contains the path and a "not found" indicator. No subprocess guard needed.

3. **`TestRunReport_CoverProfile_DirectoryPath`** (unit): call `runReport` with `coverProfile = t.TempDir()`. Verify error indicates "directory" or "not a file".

4. **`TestRunCRAPStep_WithCoverProfile`** (unit): call `runCRAPStep` directly with a pre-generated profile. Verify the result is non-nil and `CRAPload` is populated. Guarded by `testing.Short()`.

**Coverage strategy**: Unit + integration. The CRAP analysis path already has deep coverage; new tests target only the new plumbing surface. No e2e tests required — the existing `TestRunSelfCheck` exercises the full binary path and will pick up the flag once implemented.

---

## Decision 8: Does the `gaze crap --coverprofile` flag need any changes?

**Decision**: No. `gaze crap --coverprofile` is a separate flag on a separate subcommand and is not affected by this feature.

---

## Summary of findings

| # | Question | Answer |
|---|----------|--------|
| 1 | Call stack plumbing path | `newReportCmd` → `reportParams` → `runReport` → `RunnerOptions` → `runCRAPStep` → `crap.Analyze.CoverProfile` |
| 2 | `runCRAPStep` API change | Add `coverProfile string` as 3rd parameter (before `stderr`) |
| 3 | Validation location | Pre-flight in `runReport` (`os.Stat` + `info.IsDir()` before `aireport.Run`); `crap.Analyze` handles parse errors via partial-failure mode |
| 4 | `--format=json` interaction | Transparent — flag works for both formats |
| 5 | Empty string semantics | `""` = not provided = generate internally (existing behavior unchanged) |
| 6 | `AnalyzeFunc` impact | None — override bypasses `runCRAPStep` entirely |
| 7 | Test strategy | 4 CLI contract tests (T012 via `runnerFunc` spy — no `testing.Short()`; T014/T015/T016 — TBD per FR-006 architectural decision) + 1 unit test for `runCRAPStep` (T013, `testing.Short()`) |
| 8 | `gaze crap` changes | None |
