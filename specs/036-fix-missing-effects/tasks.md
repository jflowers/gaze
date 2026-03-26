# Tasks: Fix Missing Effect Detection

**Input**: Design documents from `/specs/036-fix-missing-effects/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: Tests are included — the spec requires zero regression on effect detection (SC-003) and the constitution mandates testability (Principle IV).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: No setup needed — all changes are within existing packages. No new files, dependencies, or infrastructure.

(No tasks in this phase.)

---

## Phase 2: Foundational

**Purpose**: No foundational blocking prerequisites — the feature modifies existing code within established packages.

(No tasks in this phase.)

---

## Phase 3: User Story 1 — AST-Based Mutation Fallback (Priority: P1) MVP

**Goal**: When SSA construction fails, detect receiver mutations via AST analysis instead of silently reporting zero effects. Resolves 3 of the 4 functions from issue #79.

**Independent Test**: Run `go test -race -count=1 -run TestAnalyzeASTMutations ./internal/analysis/...` — the AST fallback should detect `ReceiverMutation` effects when `ssaPkg = nil`.

### Implementation for User Story 1

- [x] T001 [US1] Implement `analyzeASTMutations` function in `internal/analysis/mutation.go`. Signature: `func analyzeASTMutations(fset *token.FileSet, fd *ast.FuncDecl, fnObj *types.Func, pkgPath string, funcName string) []taxonomy.SideEffect`. The algorithm: (1) Check if `fd.Recv` is non-nil (must be a method). If nil, return nil (package-level functions have no receiver to mutate). (2) Check if the receiver type is a pointer (`*ast.StarExpr`). If not, return nil (FR-004: value receiver mutations are not observable). (3) Extract the receiver name from `fd.Recv.List[0].Names[0].Name`. If the receiver is unnamed (no names), return nil. (4) Walk `fd.Body` with `ast.Inspect` looking for `*ast.AssignStmt` nodes where any LHS expression has a root identifier matching the receiver name (via a helper `exprRootIdent` that unwraps `SelectorExpr`, `IndexExpr`, `IndexListExpr`, `StarExpr` to find the base `*ast.Ident`). (5) Also detect `*ast.CallExpr` nodes where `Fun` is a `*ast.SelectorExpr` whose `X` is a `*ast.SelectorExpr` with root ident matching the receiver name (method calls on receiver fields, e.g., `si.index.Delete(key)` — FR-008). (6) If any receiver mutation is found, emit a single `taxonomy.SideEffect{Type: taxonomy.ReceiverMutation, Tier: "P0", Description: "mutates receiver " + receiverName + " (AST fallback)", Target: funcName}`. Deduplicate to one effect per receiver. (7) Similarly, walk `fd.Type.Params` for pointer-typed parameters. For each pointer param, check if the body contains assignments to its fields. Emit `taxonomy.PointerArgMutation` for each. Add GoDoc comment per project conventions.

- [x] T002 [US1] Implement `exprRootIdent` helper function in `internal/analysis/mutation.go`. Signature: `func exprRootIdent(expr ast.Expr) *ast.Ident`. Unwraps `*ast.SelectorExpr` (return X), `*ast.IndexExpr` (return X), `*ast.StarExpr` (return X), `*ast.ParenExpr` (return X) recursively until it reaches a `*ast.Ident`, which it returns. Returns nil if the expression chain does not resolve to an identifier. Add GoDoc comment.

- [x] T003 [US1] Modify `AnalyzeMutations` in `internal/analysis/mutation.go` at line 75 (the `ssaPkg == nil` guard). Replace `return nil` with `return analyzeASTMutations(fset, fd, fnObj, pkgPath, funcName)`. This is the single-line insertion that activates the AST fallback when SSA is unavailable. FR-005 is satisfied because this branch only executes when `ssaPkg == nil` — when SSA succeeds, the code continues to the SSA path below.

- [x] T004 [US1] Add `TestAnalyzeASTMutations_PointerReceiverFieldAssignment` test in `internal/analysis/mutation_test.go`. Call `AnalyzeMutations` with `ssaPkg = nil` on the existing `Counter.Increment` fixture (from `testdata/src/mutation/mutation.go` — a pointer receiver method that assigns `c.count++`). Verify the result contains at least one `ReceiverMutation` effect with `" (AST fallback)"` in the description.

- [x] T005 [US1] Add `TestAnalyzeASTMutations_ValueReceiverNoEffect` test in `internal/analysis/mutation_test.go`. Call `AnalyzeMutations` with `ssaPkg = nil` on the existing `ValueReceiverTrap.Set` fixture (from `testdata/src/mutation/mutation.go` — a value receiver method). Verify the result is nil (FR-004: no effects for value receivers).

- [x] T006 [US1] Add `TestAnalyzeASTMutations_PointerArgMutation` test in `internal/analysis/mutation_test.go`. Call `AnalyzeMutations` with `ssaPkg = nil` on the existing `Normalize` fixture (from `testdata/src/mutation/mutation.go` — takes `*Point` and assigns to its fields). Verify the result contains at least one `PointerArgMutation` effect.

- [x] T007 [US1] Add `TestAnalyzeASTMutations_NoMutation` test in `internal/analysis/mutation_test.go`. Call `AnalyzeMutations` with `ssaPkg = nil` on the existing `ReadOnlyPointer.String` fixture (or equivalent read-only function from the mutation testdata). Verify the result is nil (FR-009: no false positives for read-only access).

- [x] T008 [US1] Add `TestAnalyzeASTMutations_SSAPrecedence` test in `internal/analysis/mutation_test.go`. Call `AnalyzeMutations` with a valid `ssaPkg` (from `loadTestPackageWithSSA`) on a pointer receiver method. Verify the result does NOT contain `" (AST fallback)"` in any description — confirming SSA takes precedence (FR-005).

**Checkpoint**: AST mutation fallback works. Run `go test -race -count=1 -run TestAnalyzeASTMutations ./internal/analysis/...` to verify.

---

## Phase 4: User Story 2 — `no_test_coverage` Coverage Reason (Priority: P2)

**Goal**: Functions with detected effects but no test coverage appear in the contract coverage report with reason `no_test_coverage` instead of being invisible or mislabeled as `no_effects_detected`.

**Independent Test**: Run `go test -race -count=1 -run TestNoTestCoverage ./internal/crap/...` — functions with effects but no test should show `no_test_coverage` reason.

### Implementation for User Story 2

- [x] T009 [US2] Add an `effectsSet` map to `BuildContractCoverageFunc` in `internal/crap/contract.go`. After `analyzePackageCoverage` runs at lines 30-50, iterate the classified analysis results (`classifiedResults`) and build a `map[string]bool` keyed by `"pkgPath.FuncName"` for functions that have > 0 side effects. Store this alongside the existing `coverageMap`. In the returned closure (lines 56-111), when the coverage map lookup returns `ok == false`, check if the function key exists in `effectsSet`. If yes, return `ContractCoverageInfo{Reason: "no_test_coverage"}`. If no, return `ContractCoverageInfo{Reason: "no_effects_detected"}`.

- [x] T010 [US2] Update the `ContractCoverageInfo.Reason` documentation comment in `internal/crap/analyze.go` (around line 66) to include `"no_test_coverage"` as a documented reason value: "no_test_coverage — effects were detected but no test targets this function."

- [x] T011 [US2] Add `TestBuildContractCoverageFunc_NoTestCoverage` test in `internal/crap/contract_test.go` (or the appropriate test file). Create a scenario where a function has detected effects (e.g., a `ReturnValue`) but is not in the coverage map (no test targets it). Call the `ContractCoverageFunc` closure for that function key and verify the returned `ContractCoverageInfo` has `Reason: "no_test_coverage"`.

- [x] T012 [US2] Add `TestBuildContractCoverageFunc_NoEffects` test in `internal/crap/contract_test.go`. Create a scenario where a function has zero detected effects and is not in the coverage map. Call the closure and verify the returned `ContractCoverageInfo` has `Reason: "no_effects_detected"` (existing behavior preserved).

**Checkpoint**: Coverage reason distinction works. Run `go test -race -count=1 ./internal/crap/...` to verify.

---

## Phase 5: User Story 3 — SSA Degradation Diagnostics (Priority: P3)

**Goal**: Users can see when SSA failed and the AST fallback was used, via the effect description annotation and existing SSA degradation tracking.

**Independent Test**: Run `go test -race -count=1 -run TestASTFallbackAnnotation ./internal/analysis/...` — AST-fallback effects should contain "(AST fallback)" in their Description field.

### Implementation for User Story 3

- [x] T013 [US3] Verify that the `" (AST fallback)"` annotation from T001 is correctly included in AST-detected effects by running the tests from T004 and T008. The annotation is already implemented in T001 as part of the `Description` field. This task validates it end-to-end: load the mutation fixture, call `AnalyzeMutations(nil)`, check the Description field of each returned effect.

- [x] T014 [US3] Verify that the existing `SSADegradedPackages` tracking in `internal/crap/contract.go` (lines 187-190 of `analyzePackageCoverage`) correctly reports the degraded package path when the AST fallback is used. Run `go test -race -count=1 ./internal/crap/...` to confirm existing SSA degradation tests pass with no changes needed — the mechanism already tracks `pkgPath` when SSA fails.

**Checkpoint**: Diagnostics working. Run `go test -race -count=1 -short ./internal/analysis/... ./internal/crap/...` to verify both packages.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, non-regression, and cleanup.

- [x] T015 Run the full analysis test suite: `go test -race -count=1 ./internal/analysis/...`. All existing tests must pass. Verify SC-003: no function that previously had detected effects loses any effects.
- [x] T016 Run `golangci-lint run ./internal/analysis/ ./internal/crap/` and fix any lint issues introduced by the new code.
- [x] T017 Run the full CI-equivalent test suite: `go test -race -count=1 -short ./...` and verify all tests pass.
- [x] T018 Verify quickstart.md validation — run each command listed in `specs/036-fix-missing-effects/quickstart.md` and confirm they produce the expected results.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Empty — no setup needed
- **Foundational (Phase 2)**: Empty — no prerequisites
- **User Story 1 (Phase 3)**: Can start immediately — changes `internal/analysis/mutation.go`
- **User Story 2 (Phase 4)**: Can start immediately — changes `internal/crap/contract.go` (independent from US1)
- **User Story 3 (Phase 5)**: Depends on US1 (validates the annotation from T001)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent — changes `internal/analysis/` only
- **User Story 2 (P2)**: Independent — changes `internal/crap/` only
- **User Story 3 (P3)**: Depends on US1 — validates AST fallback annotation

### Within Each User Story

- T001 and T002 must complete before T003 (functions must exist before insertion point uses them)
- T003 must complete before T004-T008 (fallback must be wired up before tests can validate it)
- T009 must complete before T010-T012 (logic must exist before tests and docs)

### Parallel Opportunities

```text
Phase 3 (US1):  T001 ─┐
                T002 ─┘ → T003 → T004 ─┐
                                  T005 ─┤ (parallel — independent test cases)
                                  T006 ─┤
                                  T007 ─┤
                                  T008 ─┘

Phase 4 (US2):  T009 → T010 ─┐
                       T011 ─┤ (parallel — T010 is docs, T011/T012 are tests)
                       T012 ─┘

Phase 3 + 4 can run in PARALLEL (different packages: analysis/ vs crap/)

Phase 5 (US3):  T013 → T014 (sequential — validate annotation then diagnostics)

Phase 6:        T015 → T016 → T017 → T018 (sequential — test, lint, full suite, quickstart)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 3: User Story 1 (T001–T008)
2. **STOP and VALIDATE**: Run `go test -race -count=1 -run TestAnalyzeASTMutations ./internal/analysis/...`
3. At this point, void methods with pointer receivers that mutate fields will detect `ReceiverMutation` effects even when SSA fails. This resolves 3 of the 4 functions from issue #79.

### Incremental Delivery

1. User Story 1 → AST mutation fallback → 3 functions fixed (MVP!)
2. User Story 2 → `no_test_coverage` reason → better diagnostics for untested functions
3. User Story 3 → SSA degradation visibility → users can see when fallback was used
4. Polish → Lint, full suite, quickstart validation

---

## Notes

- US1 and US2 can run in parallel (different packages: `internal/analysis/` vs `internal/crap/`)
- The `exprRootIdent` helper (T002) is reusable for future AST analysis needs
- Existing test fixtures in `testdata/src/mutation/` are sufficient — no new fixtures needed
- The `" (AST fallback)"` annotation in effect descriptions requires zero schema changes
- Total production code: ~80-100 lines in `mutation.go`, ~15 lines in `contract.go`
- Total test code: 5 tests in `mutation_test.go`, 2 tests in `contract_test.go` (or equivalent)
