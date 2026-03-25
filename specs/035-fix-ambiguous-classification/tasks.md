# Tasks: Fix Ambiguous Classification for Clear Return Types

**Input**: Design documents from `/specs/035-fix-ambiguous-classification/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: Tests are included — the spec requires non-regression verification (SC-002) and the constitution mandates testability (Principle IV).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: No setup needed — all changes are within the existing `internal/classify/` package. No new files, dependencies, or infrastructure.

(No tasks in this phase.)

---

## Phase 2: Foundational

**Purpose**: No foundational blocking prerequisites — the feature modifies existing code within a single package. No database schemas, API routing, or shared infrastructure to set up.

(No tasks in this phase.)

---

## Phase 3: User Story 1 — Constructor Functions Are Classified as Contractual (Priority: P1) MVP

**Goal**: Functions following the `New`/`NewXxx` naming pattern receive a naming convention signal (+10), pushing P0 effects above the contractual threshold (80).

**Independent Test**: Run `go test -race -count=1 -run TestNamingSignal ./internal/classify/...` — the `New` prefix should produce a positive naming signal for `ReturnValue` effects.

### Implementation for User Story 1

- [x] T001 [US1] Add `{"New", []taxonomy.SideEffectType{taxonomy.ReturnValue, taxonomy.ErrorReturn}}` entry to `contractualPrefixes` slice in `internal/classify/naming.go`, after the `"Build"` entry (line ~34). This single line addition follows the exact structure of existing entries. No other changes to naming.go.
- [x] T002 [US1] Add test case `{"New_ReturnValue", "NewClient", taxonomy.ReturnValue, true}` to the `TestNamingSignal_ContractualPrefixes` table-driven test in `internal/classify/classify_test.go`. Verify the test passes — `AnalyzeNamingSignal("NewClient", taxonomy.ReturnValue)` should return a signal with weight +10 and source containing "naming_convention".
- [x] T003 [US1] Add test case `{"New_ExactMatch", "New", taxonomy.ReturnValue, true}` to `TestNamingSignal_ContractualPrefixes` in `internal/classify/classify_test.go`. Verify that the exact function name `"New"` (not just `"NewXxx"` prefix variants) produces the naming signal.
- [x] T004 [US1] Add test case `{"New_ErrorReturn", "NewStore", taxonomy.ErrorReturn, true}` to `TestNamingSignal_ContractualPrefixes` in `internal/classify/classify_test.go`. Verify the `"New"` prefix fires for `ErrorReturn` effects (constructors that return errors).
- [x] T005 [US1] Add test case `{"New_NoMatchMutation", "NewClient", taxonomy.ReceiverMutation, false}` to `TestNamingSignal_ContractualPrefixes` in `internal/classify/classify_test.go`. Verify the `"New"` prefix does NOT fire for `ReceiverMutation` effects (constructors don't imply mutations).

**Checkpoint**: The `"New"` prefix is recognized. Run `go test -race -count=1 -run TestNamingSignal ./internal/classify/...` to verify.

---

## Phase 4: User Story 2 — GoDoc Reduced Signal for Non-Matching Effect Types (Priority: P2)

**Goal**: When GoDoc contains a contractual keyword but the effect type doesn't match the keyword's `impliesFor` list, return a reduced signal (+5) instead of zero. This ensures well-documented functions benefit all their detected effects.

**Independent Test**: Run `go test -race -count=1 -run TestAnalyzeGodocSignal ./internal/classify/...` — a function with GoDoc `"Returns a value"` and a `PointerArgMutation` effect should receive a reduced positive signal (+5).

### Implementation for User Story 2

- [x] T006 [US2] Add `reducedGodocWeight = 5` constant in `internal/classify/godoc.go` alongside the existing `maxGodocWeight = 15` constant (line ~35). Add a GoDoc comment explaining the constant: "reducedGodocWeight is the signal weight when a contractual keyword is found but the effect type does not match the keyword's impliesFor list."
- [x] T007 [US2] Modify the `AnalyzeGodocSignal` function in `internal/classify/godoc.go` to track whether any contractual keyword was found during the keyword loop. Add a `foundContractualKeyword bool` variable before the loop (line ~58). In the existing keyword match block (line ~60), after checking `impliesFor` entries and finding no match for the effect type, set `foundContractualKeyword = true` instead of falling through silently. After the loop ends (line ~73), if `foundContractualKeyword` is true and no full-weight match was returned, return a `taxonomy.Signal` with `Source: "godoc_keyword_indirect"`, `Weight: reducedGodocWeight`, and `Reasoning` describing the indirect match.
- [x] T008 [US2] Add test case to `TestAnalyzeGodocSignal_ContractualKeywords` in `internal/classify/godoc_test.go` for the reduced signal: a function with GoDoc `"Returns a new value"` classified against `PointerArgMutation` effect type should return a signal with weight 5 (reduced) and source `"godoc_keyword_indirect"`.
- [x] T009 [US2] Add test case to `TestAnalyzeGodocSignal_ContractualKeywords` in `internal/classify/godoc_test.go` verifying the full-weight signal is unchanged: a function with GoDoc `"Returns a value"` classified against `ReturnValue` effect type should still return weight 15 (full) and source `"godoc_keyword"`.
- [x] T010 [US2] Add test case to `TestAnalyzeGodocSignal_ContractualKeywords` in `internal/classify/godoc_test.go` verifying that when no contractual keyword is found at all (e.g., GoDoc `"Handles the request"`), the function returns zero signal for `PointerArgMutation` — no spurious reduced signal. Wait — `"handles"` is not in the keyword list. Use GoDoc `"Processes the data efficiently"` which also has no keyword match. Verify zero signal.

**Checkpoint**: Reduced GoDoc signal works. Run `go test -race -count=1 -run TestAnalyzeGodocSignal ./internal/classify/...` to verify.

---

## Phase 5: User Story 3 — No Regression (Priority: P3)

**Goal**: Verify that no existing classification label changes after the scoring adjustments. Both changes (naming prefix, GoDoc reduced signal) only increase scores — they cannot decrease them.

**Independent Test**: Run `go test -race -count=1 ./internal/classify/...` — all existing acceptance tests (TestSC001–TestSC006) must pass unchanged.

### Implementation for User Story 3

- [x] T011 [US3] Run the full classification test suite: `go test -race -count=1 ./internal/classify/...`. All existing tests must pass. If any test fails, identify whether the failure is caused by the naming or GoDoc change, and fix the root cause without reverting the feature.
- [x] T012 [US3] Add `TestSC007_NewPrefixContractual` acceptance test in `internal/classify/classify_test.go` that loads a test fixture containing a function named `NewClient` with a `ReturnValue` P0 effect and verifies it is classified as "contractual" with confidence >= 80. This validates SC-001 directly. If no suitable fixture exists, add a constructor function to the existing `contracts` test fixture (`internal/classify/testdata/src/contracts/`).

**Checkpoint**: Non-regression confirmed. Run `go test -race -count=1 -short ./...` to verify the full suite passes.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup.

- [x] T013 Run `golangci-lint run` and fix any lint issues introduced by the changes in `internal/classify/naming.go` and `internal/classify/godoc.go`.
- [x] T014 Run full CI-equivalent test suite: `go test -race -count=1 -short ./...` and verify all tests pass.
- [x] T015 Verify quickstart.md validation — run each command listed in `specs/035-fix-ambiguous-classification/quickstart.md` and confirm they produce the expected results.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Empty — no setup needed
- **Foundational (Phase 2)**: Empty — no prerequisites
- **User Story 1 (Phase 3)**: Can start immediately — adds naming prefix
- **User Story 2 (Phase 4)**: Can start immediately — independent of US1 (different file: godoc.go)
- **User Story 3 (Phase 5)**: Depends on US1 and US2 being complete (tests verify combined behavior)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent — changes naming.go only
- **User Story 2 (P2)**: Independent — changes godoc.go only
- **User Story 3 (P3)**: Depends on US1 + US2 — validates combined non-regression

### Within Each User Story

- T001 must complete before T002–T005 (function must exist before tests can validate it)
- T006 must complete before T007 (constant must exist before logic uses it)
- T007 must complete before T008–T010 (logic must exist before tests validate it)
- T011 must complete before T012 (full suite validates before new acceptance test is added)

### Parallel Opportunities

```text
Phase 3 (US1):  T001 → T002 ─┐
                       T003 ─┤ (parallel — independent test cases)
                       T004 ─┤
                       T005 ─┘

Phase 4 (US2):  T006 → T007 → T008 ─┐
                              T009 ─┤ (parallel — independent test cases)
                              T010 ─┘

Phase 3 + 4 can run in PARALLEL (different files: naming.go vs godoc.go)

Phase 5 (US3):  T011 → T012 (sequential — validate first, then add acceptance test)

Phase 6:        T013 ─┐
                T014 ─┘ → T015 (lint+tests parallel, then quickstart)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 3: User Story 1 (T001–T005)
2. **STOP and VALIDATE**: Run `go test -race -count=1 -run TestNamingSignal ./internal/classify/...`
3. At this point, constructor functions named `New`/`NewXxx` receive the naming signal. This alone resolves the primary issue (#78) for most affected functions.

### Incremental Delivery

1. User Story 1 → Naming prefix added → Primary issue resolved (MVP!)
2. User Story 2 → GoDoc reduced signal → Secondary effects also benefit
3. User Story 3 → Non-regression confirmed → Full confidence
4. Polish → Lint, full suite, quickstart validation

---

## Notes

- All code changes are within `internal/classify/` — no cross-package impacts
- US1 and US2 can run in parallel because they modify different files (naming.go vs godoc.go)
- The existing `TestSC001_MechanicalContractualAccuracy` acceptance test is the primary non-regression gate
- Total production code changes: ~6 lines (1 line in naming.go, ~5 lines in godoc.go)
- Total test code changes: ~8 test cases across 2 test files
