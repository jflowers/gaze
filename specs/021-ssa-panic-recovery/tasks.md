# Tasks: SSA Panic Recovery

**Input**: Design documents from `/specs/021-ssa-panic-recovery/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Tests**: Included — the spec explicitly requires testability (Constitution Principle IV) and the plan specifies 100% branch coverage of the new recovery paths.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No setup needed — this feature modifies existing files in an existing project. No new packages, directories, or dependencies.

(Phase intentionally empty — proceed directly to Phase 2.)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extract the `safeSSABuild` helper that both `BuildSSA` and `BuildTestSSA` will use. This must be complete before either user story can be implemented.

- [x] T001 Create `safeSSABuild` unexported helper function in `internal/analysis/mutation.go` — accepts `func()`, returns recovered panic value (`any`) or `nil` on success (per research R3)
- [x] T002 Add unit tests for `safeSSABuild` in `internal/analysis/mutation_test.go` — test cases: (a) non-panicking function returns nil, (b) panic with string value returns the string, (c) panic with error value returns the error

**Checkpoint**: `safeSSABuild` helper is tested and ready for use by both `BuildSSA` and `BuildTestSSA`.

---

## Phase 3: User Story 1 - Graceful Degradation on SSA Panic (Priority: P1) MVP

**Goal**: `BuildSSA` and `BuildTestSSA` recover from `prog.Build()` panics instead of crashing the process. They return nil/error respectively, allowing callers to skip the affected package gracefully.

**Independent Test**: Trigger a panic via `safeSSABuild` in each function's recovery path and verify nil/error return instead of crash.

### Tests for User Story 1

- [x] T003 [P] [US1] Add panic recovery test for `BuildSSA` in `internal/analysis/mutation_test.go` — verify that when `prog.Build()` panics, `BuildSSA` returns `nil` (not a crash)
- [x] T004 [P] [US1] Add panic recovery test for `BuildTestSSA` in `internal/quality/pairing_test.go` — verify that when `prog.Build()` panics, `BuildTestSSA` returns `nil, nil, error` (not a crash)

### Implementation for User Story 1

- [x] T005 [P] [US1] Add `recover()` guard to `BuildSSA` in `internal/analysis/mutation.go` — use named return, deferred `recover()` calling `safeSSABuild`, set return to `nil` on panic (FR-001)
- [x] T006 [P] [US1] Add `recover()` guard to `BuildTestSSA` in `internal/quality/pairing.go` — use named returns, deferred `recover()` calling `safeSSABuild`, set error return on panic (FR-002)

**Checkpoint**: Both functions survive `prog.Build()` panics. All callers continue to work because they already handle nil/error. Existing tests pass identically (FR-005, FR-006).

---

## Phase 4: User Story 2 - Transparent Reporting of Skipped Packages (Priority: P2)

**Goal**: When a panic is recovered, emit a warning-level log message with the package path and a debug-level message with the raw panic value.

**Independent Test**: Trigger recovery and verify log output contains the package path (warning) and the raw panic value (debug).

### Tests for User Story 2

- [x] T007 [P] [US2] Add warning message content test in `internal/analysis/mutation_test.go` — verify `BuildSSA` recovery emits log containing the package path
- [x] T008 [P] [US2] Add warning message content test in `internal/quality/pairing_test.go` — verify `BuildTestSSA` recovery emits log containing the package path

### Implementation for User Story 2

- [x] T009 [P] [US2] Add `log.Printf` warning in `BuildSSA` recovery path in `internal/analysis/mutation.go` — format: `"warning: SSA build skipped for %s: internal panic recovered"` with `pkg.PkgPath` (FR-003)
- [x] T010 [P] [US2] Add `log.Printf` debug line in `BuildSSA` recovery path in `internal/analysis/mutation.go` — format: `"debug: SSA panic value: %v"` with recovered value (FR-004)
- [x] T011 [P] [US2] Add `log.Printf` warning in `BuildTestSSA` recovery path in `internal/quality/pairing.go` — same format as T009 (FR-003)
- [x] T012 [P] [US2] Add `log.Printf` debug line in `BuildTestSSA` recovery path in `internal/quality/pairing.go` — same format as T010 (FR-004)

**Checkpoint**: Recovery produces user-visible warning (package path) and developer-visible debug (panic value). Warning does not expose raw panic details to users.

---

## Phase 5: User Story 3 - No Impact on Unaffected Codebases (Priority: P1)

**Goal**: Verify zero behavioral change for codebases that do not trigger SSA panics. The recovery guard is a no-op in the non-panic path.

**Independent Test**: Run the full existing test suite and benchmark suite.

### Verification for User Story 3

- [x] T013 [US3] Run `go build ./...` and confirm success
- [x] T014 [US3] Run `go test -race -count=1 -short ./...` and confirm all existing tests pass identically (SC-003)
- [x] T015 [US3] Verify existing benchmarks in `internal/analysis/bench_test.go` show no measurable performance regression (SC-004)

**Checkpoint**: All existing behavior unchanged. No new output, no performance regression.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T016 Update GoDoc comments on `BuildSSA` in `internal/analysis/mutation.go` — document panic recovery behavior in the function comment
- [x] T017 Update GoDoc comments on `BuildTestSSA` in `internal/quality/pairing.go` — document panic recovery behavior in the function comment
- [x] T018 Update `AGENTS.md` Recent Changes section with 021-ssa-panic-recovery summary
- [x] T019 Assess whether `README.md` needs updates (new Go 1.25 minimum is already declared in `go.mod`; document if user-facing)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Empty — no setup needed
- **Foundational (Phase 2)**: No dependencies — can start immediately. BLOCKS all user stories.
- **User Story 1 (Phase 3)**: Depends on Phase 2 (`safeSSABuild` helper)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (recovery paths must exist before adding logging)
- **User Story 3 (Phase 5)**: Depends on Phases 3 and 4 (all code changes must be in place before regression verification)
- **Polish (Phase 6)**: Depends on Phase 5 (all changes finalized)

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational only — MVP
- **User Story 2 (P2)**: Depends on US1 (logging is added to the recovery paths created in US1)
- **User Story 3 (P1)**: Depends on US1 + US2 (regression verification covers the full change)

### Within Each User Story

- Tests written alongside implementation (test tasks and impl tasks are parallel within same story)
- T005 and T006 can run in parallel (different files)
- T009–T012 can all run in parallel (different files, additive changes)

### Parallel Opportunities

- T003 and T004 can run in parallel (different test files)
- T005 and T006 can run in parallel (different source files)
- T007 and T008 can run in parallel (different test files)
- T009–T012 can all run in parallel (additive log lines in different files)

---

## Parallel Example: User Story 1

```bash
# Tests and implementation can run in parallel (different files):
Task: "T003 [P] [US1] Add panic recovery test for BuildSSA in internal/analysis/mutation_test.go"
Task: "T004 [P] [US1] Add panic recovery test for BuildTestSSA in internal/quality/pairing_test.go"
Task: "T005 [P] [US1] Add recover() guard to BuildSSA in internal/analysis/mutation.go"
Task: "T006 [P] [US1] Add recover() guard to BuildTestSSA in internal/quality/pairing.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (`safeSSABuild` helper + tests)
2. Complete Phase 3: User Story 1 (recovery guards + tests)
3. **STOP and VALIDATE**: `go test -race -count=1 -short ./...` passes
4. `gaze report` no longer crashes on affected codebases

### Incremental Delivery

1. Foundational → `safeSSABuild` tested
2. User Story 1 → Recovery works, callers handle nil/error → Deploy (MVP!)
3. User Story 2 → Warnings visible to users → Deploy
4. User Story 3 → Full regression verified → Deploy
5. Polish → Docs updated → Final
