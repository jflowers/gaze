# Tasks: GazeCRAP Data in Report Pipeline

**Input**: Design documents from `/specs/022-report-gazecrap-pipeline/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Tests**: Test tasks are included because the constitution (Principle IV: Testability) requires coverage strategy for all new code, and research.md R6 defines a specific test plan.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Extract and Relocate)

**Purpose**: Extract `buildContractCoverageFunc` from `cmd/gaze/main.go` to `internal/crap/contract.go` so it can be imported by both `cmd/gaze` and `internal/aireport`. This is the foundational change that all user stories depend on.

- [x] T001 Extract `buildContractCoverageFunc`, `analyzePackageCoverage`, `resolvePackagePaths`, and `extractShortPkgName` from `cmd/gaze/main.go` to `internal/crap/contract.go`. Export only `BuildContractCoverageFunc` — keep `analyzePackageCoverage`, `resolvePackagePaths`, and `extractShortPkgName` as unexported helpers. Preserve all existing behavior including SSA degradation tracking and the `crap.ContractCoverageInfo` return type. Resolve `package main` dependencies: replace `logger.Debug`/`logger.Warn` calls with `fmt.Fprintf(stderr, ...)` (consistent with `internal/aireport` conventions), replace `loadConfig(...)` with `config.Load(...)` with a best-effort fallback (matching `loadGazeConfigBestEffort` in `runner_steps.go`), and omit the `Version` field from `analysis.Options` (it does not affect side effect detection or contract coverage computation).
- [x] T002 Update `cmd/gaze/main.go` to call `crap.BuildContractCoverageFunc` instead of the local `buildContractCoverageFunc`. Remove the local copies of `buildContractCoverageFunc`, `analyzePackageCoverage`, `resolvePackagePaths`, and `extractShortPkgName`. The `coverageFunc` field on `crapParams` must use the new function signature.
- [x] T003 Move existing tests for `buildContractCoverageFunc`, `analyzePackageCoverage`, `resolvePackagePaths`, and `extractShortPkgName` from `cmd/gaze/main_test.go` to `internal/crap/contract_test.go`. Use `package crap` (internal test) for unexported helper tests and `package crap_test` for `BuildContractCoverageFunc` tests. Preserve the critical regression assertion in `TestBuildContractCoverageFunc_WelltestedPackage` (`info.Percentage > 0` for the welltested fixture) — this is the primary regression guard for SC-002. Verify all moved tests pass: `go test -race -count=1 -short ./internal/crap/...`
- [x] T004 Verify `gaze crap` still works identically: run `go build ./cmd/gaze && go test -race -count=1 -short ./cmd/gaze/... ./internal/crap/...` and confirm all existing tests pass with no behavioral changes.

---

## Phase 2: Foundational (Pipeline Plumbing)

**Purpose**: Expand `runCRAPStep` and `pipelineStepFuncs` to accept a `ContractCoverageFunc` parameter. This plumbing change MUST complete before any user story can be implemented.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T005 Add `contractCoverageFunc` parameter (type `func(string, string) (crap.ContractCoverageInfo, bool)`) to `runCRAPStep` in `internal/aireport/runner_steps.go`. When non-nil, set `opts.ContractCoverageFunc = contractCoverageFunc` on the `crap.Options` before calling `crap.Analyze`. When nil, preserve existing behavior (no GazeCRAP data).
- [x] T006 Update the `crapStep` field type in `pipelineStepFuncs` struct in `internal/aireport/runner.go` to match the new `runCRAPStep` signature: `func([]string, string, string, io.Writer, func(string, string) (crap.ContractCoverageInfo, bool)) (*crapStepResult, error)`.
- [x] T007 Update the nil-guard assignments for `steps.crapStep` inside `runProductionPipeline` in `internal/aireport/runner.go` to use the new `runCRAPStep` signature (note: there is no `defaultPipelineSteps()` factory — the guards are inline in `runProductionPipeline`).
- [x] T008 Update the `fakeSteps()` factory function and all `pipelineStepFuncs` mock assignments in `internal/aireport/pipeline_internal_test.go` to include the new `contractCoverageFunc` parameter (accepting and ignoring it with `_`). Update the individual test overrides in `TestRunProductionPipeline_CRAPStepFails`, `TestRunProductionPipeline_MultipleStepsFail`, and `TestRunProductionPipeline_EmptyPatterns`. Verify all pipeline tests pass: `go test -race -count=1 -short ./internal/aireport/...`

**Checkpoint**: Pipeline accepts ContractCoverageFunc but doesn't build one yet — GazeCRAP still absent from reports.

---

## Phase 3: User Story 1 — CI Report Shows GazeCRAP Quadrant Distribution (Priority: P1) MVP

**Goal**: `gaze report --format=json` produces JSON with `gaze_crap`, `quadrant`, `quadrant_counts`, and `gaze_crapload` fields populated.

**Independent Test**: Run `gaze report --format=json --coverprofile=<valid> ./...` and verify the CRAP summary JSON contains non-null `gaze_crapload` and `quadrant_counts`.

### Implementation for User Story 1

- [x] T009 [US1] Update `runProductionPipeline` in `internal/aireport/runner.go` to call `crap.BuildContractCoverageFunc(patterns, moduleDir, stderr)` before the CRAP step. Pass the returned `ccFunc` to `steps.crapStep(...)`. If `BuildContractCoverageFunc` returns nil (all packages failed), pass nil to preserve graceful degradation (FR-005). Pass `degradedPkgs` to `payload.Summary.SSADegradedPackages` (merge with any existing values from the quality step).
- [x] T010 [US1] Add test in `internal/aireport/pipeline_internal_test.go` that verifies: when the mock `crapStep` receives a non-nil `contractCoverageFunc` and returns `crapStepResult{GazeCRAPload: 7}`, `payload.Summary.GazeCRAPload == 7`. This verifies data flows through the pipeline, not just that the callback was passed. Secondary assertion: the callback parameter is non-nil.
- [x] T011 [US1] Verify end-to-end: run `go build ./cmd/gaze` then `go test -race -count=1 -short ./internal/aireport/... ./internal/crap/... ./cmd/gaze/...` to confirm all tests pass. Verify SC-001: `gaze report --format=json --coverprofile=<valid> ./...` produces JSON with non-null `gaze_crapload`.

**Checkpoint**: `gaze report --format=json` now includes GazeCRAP data. SC-001 and SC-002 verifiable.

---

## Phase 4: User Story 2 — GazeCRAPload Threshold Enforcement Works (Priority: P2)

**Goal**: `--max-gaze-crapload` threshold evaluates against actual GazeCRAPload instead of always-zero.

**Independent Test**: Run `gaze report --max-gaze-crapload=0 --coverprofile=<valid> ./...` against a project with Q4 functions and verify non-zero exit code.

### Implementation for User Story 2

- [x] T012 [US2] Add or update test in `cmd/gaze/main_test.go` that verifies SC-003: when GazeCRAPload exceeds `--max-gaze-crapload`, `gaze report` returns a non-zero exit. Use the existing `TestRunReport_GazeCRAPloadThresholds` test pattern. Ensure it covers both PASS (GazeCRAPload=0, threshold=5 → PASS) and FAIL (GazeCRAPload=5, threshold=3 → FAIL) scenarios with synthetic payloads. Note: the synthetic-payload approach is valid for threshold logic testing; SC-002 (exact match with real data) is verified separately in T013.
- [x] T013 [US2] Add a test guarded by `testing.Short()` that runs both `crap.Analyze` (with `BuildContractCoverageFunc` wired) and the report pipeline with the same coverprofile/patterns, and asserts `gaze_crapload` values are identical. This is the automated verification of SC-002 (exact match between `gaze crap` and `gaze report`).

**Checkpoint**: `--max-gaze-crapload` threshold now enforces against actual GazeCRAPload. SC-002 and SC-003 verifiable.

---

## Phase 5: User Story 3 — AI-Formatted Report Includes GazeCRAP Section (Priority: P3)

**Goal**: AI-formatted text report via `--ai=opencode` renders the GazeCRAP Quadrant Distribution table with actual data instead of "N/A".

**Independent Test**: Run `gaze report --ai=opencode --coverprofile=<valid> ./...` and verify the output includes a quadrant distribution table with counts.

### Implementation for User Story 3

- [x] T014 [US3] Verify that the AI agent prompt handles GazeCRAP data correctly by checking: (1) `.opencode/agents/gaze-reporter.md` Quick Reference Example includes a GazeCRAP Quadrant Distribution table, (2) the Scoring Consistency Rules section references `summary.quadrant_counts` (the correct JSON field name — fix if it references `quadrant_distribution`), (3) no prompt changes are needed for rendering. If the prompt references the wrong field name, update it and the scaffold copy.
- [x] T015 [US3] Add test using the existing `FakeAdapter` pattern in `cmd/gaze/main_test.go` that verifies: when the report pipeline produces a payload with non-null `quadrant_counts` in the CRAP JSON, the payload passed to the AI adapter contains those fields. This provides automated coverage for SC-004's data-availability precondition without requiring a real AI model invocation.

**Checkpoint**: AI reports show GazeCRAP quadrant distribution. SC-004 verifiable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, cleanup, and final verification across all stories.

- [x] T016 [P] Update `AGENTS.md` Recent Changes section with a summary of this feature: what files changed, what capability was added.
- [x] T017 [P] Run full CI-equivalent checks: `go build ./cmd/gaze && go test -race -count=1 -short ./... && go vet ./...` to confirm no regressions across the entire codebase.
- [x] T018 Run quickstart.md verification: execute the "After" command from `specs/022-report-gazecrap-pipeline/quickstart.md` and confirm the expected output matches.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (T001-T004) completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 (T005-T008) completion — this is the MVP
- **US2 (Phase 4)**: Depends on Phase 3 (US1 delivers GazeCRAPload data for thresholds to evaluate)
- **US3 (Phase 5)**: Depends on Phase 3 (US1 delivers GazeCRAP data in JSON payload for AI to render)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2). No dependencies on other stories. **This is the MVP.**
- **User Story 2 (P2)**: Depends on US1 (needs GazeCRAPload data to exist before thresholds can evaluate it).
- **User Story 3 (P3)**: Depends on US1 (needs GazeCRAP data in JSON payload before AI can render it). Independent of US2.

### Within Each Phase

- Phase 1: T001 → T002 → T003 → T004 (sequential: extract → update callers → move tests → verify)
- Phase 2: T005 and T006 are parallel [P] (different files). T007 depends on T006. T008 depends on T005+T006+T007.
- Phase 3: T009 → T010 → T011 (sequential: implement → test → verify)
- Phase 4: T012 → T013 (sequential: threshold test → exact-match test)
- Phase 5: T014 → T015 (sequential: verify prompt → add payload assertion test)
- Phase 6: T016 and T017 are parallel [P]. T018 depends on T017.

### Parallel Opportunities

- T005 and T006 can run in parallel (different files: `runner_steps.go` and `runner.go`)
- T016 and T017 can run in parallel (documentation vs testing)
- US2 and US3 can run in parallel after US1 completes (independent of each other)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Extract `BuildContractCoverageFunc` to `internal/crap/contract.go`
2. Complete Phase 2: Expand `runCRAPStep` and `pipelineStepFuncs` signatures
3. Complete Phase 3: Wire `BuildContractCoverageFunc` into `runProductionPipeline`
4. **STOP and VALIDATE**: `gaze report --format=json` produces GazeCRAP data
5. SC-001 and SC-002 verified

### Incremental Delivery

1. MVP → GazeCRAP data in JSON output (US1)
2. Add threshold enforcement → CI gates work (US2)
3. Verify AI rendering → Full report includes quadrant table (US3)
4. Polish → Documentation and cross-codebase verification

---

## Notes

- Constitution Principle IV (Testability) requires tests for all new code — test tasks are included per research.md R6.
- The extracted function must preserve the exact behavior of the original, ensuring SC-002 (exact GazeCRAPload match between `gaze crap` and `gaze report`).
- Quality analysis runs twice (R5 trade-off): once for `BuildContractCoverageFunc` and once in the quality pipeline step. This is accepted overhead — the 30% quality-runtime estimate from R5 is approximate and should be validated during T017 if pipeline runtime is concerning.
- No new CLI flags are added. Existing `--coverprofile` and `--max-gaze-crapload` flags are already wired — they just need real data.
- The `resolvePackagePaths` function in `internal/aireport/runner_steps.go` duplicates the logic being extracted. Consolidation is deferred to a future spec to keep this change focused on the pipeline wiring.
