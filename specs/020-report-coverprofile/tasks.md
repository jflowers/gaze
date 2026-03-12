# Tasks: Pass Pre-Generated Coverage Profile to gaze report

**Input**: Design documents from `specs/020-report-coverprofile/`
**Branch**: `020-report-coverprofile`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, quickstart.md ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- All paths are relative to the repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new infrastructure is needed. This feature adds fields and a flag to existing code. Phase 1 is a single verification step to confirm the baseline builds and tests pass before any changes.

- [ ] T001 Verify baseline: run `go build ./...` and `go test -race -count=1 -short ./...` — confirm all pass before any changes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Thread `CoverProfile string` through the two internal layers (`RunnerOptions` and `runCRAPStep`) that sit between the CLI flag and `crap.Analyze`. These changes must be complete before any user-story work begins because US2 and US3 tests exercise these layers.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T002 Add `CoverProfile string` field to `aireport.RunnerOptions` in `internal/aireport/runner.go` with GoDoc comment: "CoverProfile is the path to a pre-generated Go coverage profile. When non-empty, the CRAP analysis step uses this file directly instead of spawning go test internally (FR-001, FR-002). Empty string means generate internally (FR-003)."
- [ ] T003 Add `coverProfile string` as third parameter to `runCRAPStep` in `internal/aireport/runner_steps.go` (before `stderr io.Writer`); set `opts.CoverProfile = coverProfile` inside the function body
- [ ] T004 Update the `runCRAPStep` call site in `runProductionPipeline` in `internal/aireport/runner.go` to pass `opts.CoverProfile` as the new third argument
- [ ] T005 Verify: `go build ./...` passes after foundational changes — confirms all call sites compile

**Checkpoint**: Foundational plumbing complete — `CoverProfile` flows from `RunnerOptions` through to `crap.Analyze`. User story implementation can now begin.

---

## Phase 3: User Story 1 — Supply existing coverage profile to skip internal test run (Priority: P1) 🎯 MVP

**Goal**: Add `--coverprofile` to `gaze report`, wire it through `reportParams` and `runReport` to `RunnerOptions.CoverProfile`, and verify it produces a correct report without spawning `go test` internally.

**Independent Test**: Run `gaze report ./... --format=json --coverprofile=coverage.out` with a pre-generated profile. Verify the command exits 0, stdout contains JSON with CRAP data, and no `go test` subprocess is spawned during the CRAP step.

### Implementation for User Story 1

- [ ] T005b [US1] Add pre-flight `--coverprofile` path validation to `runReport` in `cmd/gaze/main.go`: if `p.coverProfile != ""`, call `os.Stat(p.coverProfile)`; if stat fails (not found), return `fmt.Errorf("--coverprofile %q: %w", p.coverProfile, err)`; if `info.IsDir()`, return `fmt.Errorf("--coverprofile %q is a directory, not a file", p.coverProfile)`. This validation runs before `aireport.Run` is called, satisfying FR-006 (hard exit, non-zero) without conflicting with `runProductionPipeline`'s partial-failure architecture. (Decision 3, Option A, 2026-03-12.)

- [ ] T006 [US1] Add `coverProfile string` field to `reportParams` struct in `cmd/gaze/main.go` (after the `minContractCoverage *int` field)
- [ ] T007 [US1] Add `var coverProfile string` local variable to `newReportCmd` in `cmd/gaze/main.go` alongside the other flag variables
- [ ] T008 [US1] Register the flag in `newReportCmd` in `cmd/gaze/main.go`: `cmd.Flags().StringVar(&coverProfile, "coverprofile", "", "path to a pre-generated coverage profile (skips internal go test run)")`
- [ ] T009 [US1] Add `coverProfile: coverProfile` to the `reportParams{}` literal in the `RunE` closure of `newReportCmd` in `cmd/gaze/main.go`
- [ ] T010 [US1] Add `CoverProfile: p.coverProfile` to the `aireport.RunnerOptions{}` literal in `runReport` in `cmd/gaze/main.go`
- [ ] T011 [US1] Add a `--coverprofile` usage example to the `Long` field of `newReportCmd` in `cmd/gaze/main.go`: `gaze report ./... --ai=claude --coverprofile=coverage.out`

### Tests for User Story 1

- [ ] T012 [US1] Add `TestRunReport_CoverProfile_ValidPath` in `cmd/gaze/main_test.go`: write a minimal valid Go coverage profile to `t.TempDir()`; use a spy `runnerFunc` (the `reportParams.runnerFunc` override) that captures the `RunnerOptions` passed to it and returns nil; call `runReport` with `format:"json"`, `coverProfile` set to that path, and the spy wired into `reportParams`; assert no error and `spy.capturedOpts.CoverProfile == profilePath`. Note: the spy intercepts at the `reportParams.runnerFunc` boundary (before `aireport.Run`), so it proves the `reportParams.coverProfile → RunnerOptions.CoverProfile` wiring — not the downstream `runCRAPStep` wiring (which is covered by T013). No `testing.Short()` guard — spy bypasses all subprocess execution.
- [ ] T013 [US1] Add `TestRunCRAPStep_WithCoverProfile` in `internal/aireport/runner_steps_test.go`: call `runCRAPStep` with a minimal valid coverage profile from `internal/aireport/testdata/` (create a static fixture `testdata/sample.coverprofile` containing `mode: set\n` followed by one coverage record for a real source file in the module, e.g. `github.com/unbound-force/gaze/internal/crap/crap.go:90.57,93.2 1 1`); assert `result != nil`, no error, and `result.JSON != nil` (proves the profile was accepted and CRAP analysis produced output — `crapStepResult` has no `Functions` field; the fields are `JSON json.RawMessage`, `CRAPload int`, `GazeCRAPload int`). Guard with `testing.Short()` since it loads real Go packages via `crap.Analyze`.

- [ ] T013b [US1] Merge SC-001 regression assertion into T012 — add a second assertion to T012: `spy.callCount == 1` (verifying `runnerFunc` was called exactly once). This ensures the SC-001 regression guard is co-located with the wiring assertion and adds no additional test overhead. T013b as a separate task is superseded by this merge; remove T013b from the task list.

**Checkpoint**: `gaze report --coverprofile` works end-to-end for the happy path. US1 is independently testable.

---

## Phase 4: User Story 2 — Clear errors for invalid profile paths (Priority: P2)

**Goal**: Ensure all three invalid-path scenarios (nonexistent, directory, unparseable content) produce distinct, actionable error messages with non-zero exit codes.

**Independent Test**: Run `runReport` with each invalid profile path. Verify error is non-nil and contains the expected diagnostic string. No subprocess guard needed — these tests are pure unit tests.

### Tests for User Story 2

- [ ] T014 [P] [US2] Add `TestRunReport_CoverProfile_NonexistentPath` in `cmd/gaze/main_test.go`: set `coverProfile` to `filepath.Join(t.TempDir(), "nonexistent.out")`, call `runReport` with `format:"json"`, assert `err != nil` (error comes from the pre-flight `os.Stat` check in T005b, not from the pipeline), assert `err.Error()` contains the path string and `"no such file"` (or `"not exist"`). No `testing.Short()` guard needed — no subprocess.
- [ ] T015 [P] [US2] Add `TestRunReport_CoverProfile_DirectoryPath` in `cmd/gaze/main_test.go`: set `coverProfile` to `t.TempDir()` (a directory), call `runReport` with `format:"json"`, assert `err != nil` (error comes from the pre-flight `info.IsDir()` check in T005b, not from the pipeline), assert `err.Error()` contains `"directory"`. No `testing.Short()` guard needed — no subprocess.
- [ ] T016 [P] [US2] Add `TestRunReport_CoverProfile_UnparseableContent` in `cmd/gaze/main_test.go`: write a file with content `"not a coverage profile\n"` to `t.TempDir()`, call `runReport` with `format:"json"` and **no** `runnerFunc` override (let the real pipeline run), capture stdout into a buffer, assert `err == nil` (the partial-failure architecture stores CRAP errors in the JSON payload, not as a Go error), unmarshal the JSON output into `aireport.ReportPayload`, and assert `payload.Errors.CRAP != nil && strings.Contains(*payload.Errors.CRAP, "parsing coverage profile")`. This exercises the real parse-failure path end-to-end. Guard with `testing.Short()` — the real pipeline runs quality, classify, and docscan steps which load Go packages. Note: this test calls `runReport` with a real AI adapter override (`FakeAdapter`) to avoid requiring a real AI CLI in the test environment. The `FakeAdapter` is already available in the test suite.

**Checkpoint**: All three invalid-path error scenarios are covered by automated regression tests. US2 is independently testable alongside US1.

---

## Phase 5: User Story 3 — Flag is self-documenting (Priority: P3)

**Goal**: Ensure `--help` output and README provide sufficient documentation for a developer to use `--coverprofile` without external guidance.

**Independent Test**: Run `gaze report --help` and verify `--coverprofile` appears in the flags list. Read the README and verify a CI example is present.

### Implementation for User Story 3

- [ ] T017 [US3] Update the README `gaze report` section in `README.md`: add a `#### Using a pre-generated coverage profile` subsection with the two-step CI example (`go test -race -count=1 -coverprofile=coverage.out ./...` then `gaze report ./... --ai=claude --coverprofile=coverage.out`) and a GitHub Actions YAML snippet
- [ ] T018 [P] [US3] Add `TestReportCmd_CoverprofileInHelp` in `cmd/gaze/main_test.go`: create the `report` cobra command via `newReportCmd()`, set `cmd.SetArgs([]string{"--help"})` and capture output via `cmd.SetOut(&buf)`, call `cmd.Execute()`, assert `buf.String()` contains `"--coverprofile"` and `"pre-generated"`. Use `cmd.SetArgs`+`Execute` (not `UsageString`) to guarantee the full flag description is rendered.

**Checkpoint**: Flag is discoverable from `--help` and documented in README. US3 is independently testable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: CI parity validation, GoDoc accuracy, and AGENTS.md bookkeeping.

- [ ] T019 [P] Update `AGENTS.md` Recent Changes section: add a bullet for `020-report-coverprofile` describing the new `--coverprofile` flag, its effect on CI double-run elimination, and the modified files (`cmd/gaze/main.go`, `internal/aireport/runner.go`, `internal/aireport/runner_steps.go`)
- [ ] T020 [P] Verify GoDoc on `RunnerOptions.CoverProfile` field is accurate and matches the implemented behavior (check `internal/aireport/runner.go`)
- [ ] T021 Run CI parity gate: `go build ./...` and `go test -race -count=1 -short ./...` — all must pass
- [ ] T022 Mark all tasks complete in this `tasks.md` as implemented (update checkboxes to `[x]`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user stories
- **Phase 3 (US1 — valid path)**: Depends on Phase 2 — can start after T005
- **Phase 4 (US2 — error cases)**: Depends on Phase 2 — can run in parallel with Phase 3 (different test file tasks, same production code is already complete after T010)
- **Phase 5 (US3 — docs)**: Independent of Phases 3 and 4 — can start after Phase 2
- **Phase 6 (Polish)**: Depends on all prior phases complete

### Within Phase 2 (strict sequence — same files)

T002 → T003 → T004 → T005

### Within Phase 3 (strict sequence — same file `main.go`)

T005b → T006 → T007 → T008 → T009 → T010 → T011, then T012 and T013 [P] in parallel (T012 is in `main_test.go`; T013 is in `runner_steps_test.go`; T013b was merged into T012)

### Within Phase 4 (all parallel — independent test functions)

T014 [P], T015 [P], T016 [P] can run simultaneously after Phase 2 is complete. Note: T014 and T015 need no `testing.Short()` guard (pre-flight check only). T016 needs `testing.Short()` because it runs the real pipeline.

### Within Phase 5

T017 and T018 [P] can run in parallel (different files)

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 (foundational plumbing). Independent of US2, US3.
- **US2 (P2)**: Depends on Phase 2 only. Error paths are entirely in `crap.Analyze` which is already wired. Independent of US1 and US3.
- **US3 (P3)**: Depends on Phase 2 only (flag must be registered before `--help` test). Independent of US1 and US2.

---

## Parallel Execution Examples

### Phase 4 (all three error tests in parallel)

```text
Task T014: TestRunReport_CoverProfile_NonexistentPath in cmd/gaze/main_test.go
Task T015: TestRunReport_CoverProfile_DirectoryPath in cmd/gaze/main_test.go
Task T016: TestRunReport_CoverProfile_UnparseableContent in cmd/gaze/main_test.go
```

### Phase 3 + Phase 4 + Phase 5 (after Phase 2 complete)

```text
Track A: T005b → T006 → T007 → T008 → T009 → T010 → T011 → T012 → T013b, T013
Track B: T014, T015, T016   (T014/T015 use pre-flight error from T005b — no guard; T016 uses real pipeline + JSON inspection — needs testing.Short())
Track C: T017, T018
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (T001) — baseline verified
2. Complete Phase 2 (T002–T005) — foundational plumbing done
3. Complete Phase 3 (T006–T013) — `--coverprofile` flag works
4. **STOP and VALIDATE**: `go test -race -count=1 -short ./...` passes; `gaze report ./... --format=json --coverprofile=<file>` works
5. This is the full deliverable — US2 and US3 are quality improvements

### Incremental Delivery

1. Phase 1 + Phase 2 → foundational changes compile
2. Phase 3 → flag works, happy path tested (MVP)
3. Phase 4 → error cases locked by regression tests
4. Phase 5 → documentation complete
5. Phase 6 → CI gate passes, bookkeeping done

---

## Notes

- `crap.Analyze` already validates the path (existence, is-directory) — no validation needed in the new code
- `runnerFunc` spy injection (`reportParams.runnerFunc` override) is the preferred test mechanism for `cmd/gaze/main_test.go` tests — avoids spawning real `go test` and removes the need for `testing.Short()` guards on T012. Note: `AnalyzeFunc` (signature `func([]string, string)`) cannot observe `opts.CoverProfile`; only the `runnerFunc` spy (which receives full `RunnerOptions`) can verify the `reportParams→RunnerOptions` wiring
- T013 (`TestRunCRAPStep_WithCoverProfile`) is the only test that needs `testing.Short()` because it calls `crap.Analyze` with a real coverage profile which loads Go packages
- SC-001 regression guard (subprocess non-invocation) is covered by the `spy.callCount == 1` assertion co-located in T012 — T013b was merged into T012 and does not exist as a separate test
- The testdata fixture `internal/aireport/testdata/sample.coverprofile` for T013 is a static text file, not a compiled binary — no build step required
- All production code changes are in 3 files: `cmd/gaze/main.go`, `internal/aireport/runner.go`, `internal/aireport/runner_steps.go`
- [P] tasks = different files or independent functions, no data dependencies on simultaneously-running tasks
