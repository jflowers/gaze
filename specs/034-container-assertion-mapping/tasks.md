# Tasks: Container Type Assertion Mapping

**Input**: Design documents from `/specs/034-container-assertion-mapping/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: Tests are included as they are integral to this feature (the spec requires ratchet non-regression and acceptance tests for the container pattern).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create test fixture infrastructure that all user stories depend on.

- [x] T001 Create test fixture target file `internal/quality/testdata/src/containerunwrap/containerunwrap.go` with a `Container` struct type wrapping a JSON body field and three exported functions: (1) `WrapJSON(key, value string) *Container` that returns a Container with a JSON-encoded body, (2) `WrapMultiField(fields map[string]string) *Container` that returns a Container with multiple JSON fields, (3) `WrapNestedJSON(key, innerKey, value string) *Container` that returns a Container with nested JSON. Include a `TextContent` struct with a `Text string` field and a `Result` struct with a `Content []TextContent` field to mirror the MCP pattern. Add a `WrapMCPStyle(key, value string) *Result` function.
- [x] T002 Create test fixture test file `internal/quality/testdata/src/containerunwrap/containerunwrap_test.go` with tests that exercise the container-unwrap-assert pattern: (1) `TestWrapJSON_BasicUnmarshal` — assigns return value, accesses `.Body` field, calls `json.Unmarshal` into `map[string]any`, asserts on `data["key"]`, (2) `TestWrapJSON_StructUnmarshal` — same pattern but unmarshals into a typed struct and asserts on struct fields (FR-006), (3) `TestWrapMultiField_MultipleAssertions` — 4 assertions on different map keys from a single unmarshal, (4) `TestWrapMCPStyle_DeepChain` — mirrors the MCP pattern: `result.Content[0].Text` → type conversion → `json.Unmarshal` → assert on map keys (4+ intermediate steps per FR-004). Each test must use stdlib `testing` with `t.Errorf`/`t.Fatalf` assertions (no testify).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Internal helper functions that the main mapping pass depends on.

**CRITICAL**: These must be complete before User Story 1 implementation.

- [x] T003 Implement `isTransformationCall` function in `internal/quality/mapping.go` that accepts an `*ast.CallExpr` and `*types.Info`, returns `(byteArgIdx int, ptrDestIdx int, ok bool)`. The function extracts the call's type signature via `info.Types[call.Fun].Type` cast to `*types.Signature`, iterates `sig.Params()` checking each parameter type: `[]byte` (via `*types.Slice` with `Elem().(*types.Basic).Kind() == types.Byte`), `string` (via `*types.Basic` with `Kind() == types.String`), `io.Reader` (via `types.NewMethodSet` checking for `Read` method), or pointer (`*types.Pointer`). Returns the positional indices of the byte-like and pointer parameters. Returns `ok=false` if both are not found. Add GoDoc comment per project conventions.
- [x] T004 Implement `containsObject` function in `internal/quality/mapping.go` that accepts an `ast.Expr`, a `types.Object`, and `*types.Info`, returns `bool`. Uses `ast.Inspect` to walk the expression tree looking for any `*ast.Ident` whose `info.Uses[ident]` (or `info.Defs[ident]`) equals the target object by pointer identity. Short-circuits on first match. Add GoDoc comment.
- [x] T005 Implement `extractPointerDest` function in `internal/quality/mapping.go` that accepts an `*ast.CallExpr`, the pointer parameter index (from `isTransformationCall`), and `*types.Info`, returns `types.Object`. Extracts the call argument at the pointer index, unwraps `*ast.UnaryExpr` with `Op == token.AND` to get the underlying `*ast.Ident`, then returns `info.Uses[ident]` (or `info.Defs[ident]`). Returns nil if the argument is not an addressable identifier. Add GoDoc comment.

---

## Phase 3: User Story 1 — Assertions on Unwrapped Container Fields Count Toward Contract Coverage (Priority: P1) MVP

**Goal**: When a test assigns a function's return value, accesses a field, passes it through a transformation (like JSON unmarshal), and asserts on the result, those assertions map to the ReturnValue effect at confidence 55.

**Independent Test**: Run `go test -race -count=1 -run TestSC003_MappingAccuracy ./internal/quality/...` — the containerunwrap fixture assertions should be mapped, and the overall ratchet should hold at >= 85.0%.

### Implementation for User Story 1

- [x] T006 [US1] Implement `matchContainerUnwrap` function in `internal/quality/mapping.go`. Signature: `func matchContainerUnwrap(site AssertionSite, objToEffectID map[types.Object]string, effectMap map[string]*taxonomy.SideEffect, testPkg *packages.Package, returnEffectID string) *taxonomy.AssertionMapping`. The algorithm: (1) Collect all `types.Object` keys from `objToEffectID` that map to a ReturnValue effect as the initial tracked variable set. (2) For up to 6 iterations (chain depth limit per R5), walk test package AST via `ast.Inspect` looking for `*ast.AssignStmt` nodes after the tracked variables' positions. For each assignment where any RHS expression references a tracked variable (using `containsObject`): (a) if the RHS is or contains a `*ast.CallExpr` matching `isTransformationCall`, extract the pointer destination via `extractPointerDest` and add it to the tracked set for the next iteration; (b) otherwise, add each LHS identifier's `types.Object` (via `TypesInfo.Defs`/`Uses`) to the tracked set. (3) After building the tracked set, check if the assertion site's expression contains any tracked variable using `ast.Inspect` + `info.Uses` identity check, or if `resolveExprRoot` of the assertion expression resolves to a tracked variable. (4) If matched, return an `AssertionMapping` with `Confidence: 55`, `SideEffectID: returnEffectID`, and the effect from `effectMap`. (5) If no match after all iterations, return nil.
- [x] T007 [US1] Insert `matchContainerUnwrap` call into the mapping pipeline in `mapAssertionsToEffectsImpl` in `internal/quality/mapping.go`. Add the call after the `matchInlineCall` check (after line ~114) and before the `tryAIMapping` check (line ~116). Pass `site`, `objToEffectID`, `effectMap`, `testPkg`, and `returnEffectID`. The insertion follows the existing short-circuit pattern: `if mapping == nil && returnEffectID != "" { mapping = matchContainerUnwrap(...) }`.
- [x] T008 [US1] Add `containerunwrap` fixture to the `TestSC003_MappingAccuracy` ratchet test in `internal/quality/quality_test.go`. Add `"containerunwrap"` to the fixture list (around line 988). Verify the test passes with the new fixture — accuracy should remain >= 85.0% (SC-002) and the containerunwrap assertions should be mapped (contributing to the total).

**Checkpoint**: At this point, the core container mapping works. Run `go test -race -count=1 -run TestSC003 ./internal/quality/...` to verify.

---

## Phase 4: User Story 2 — Multi-Step Unwrap Chains Are Traced (Priority: P2)

**Goal**: Chains with 4+ intermediate steps (field access, type assertion, index, unmarshal) are traced correctly, matching the real-world MCP test pattern.

**Independent Test**: The `TestWrapMCPStyle_DeepChain` fixture test exercises a 4-step chain. Run `go test -race -count=1 -run TestSC003_MappingAccuracy ./internal/quality/...` — the deep-chain assertions should be mapped.

### Implementation for User Story 2

- [x] T009 [US2] Extend the `containerunwrap_test.go` fixture with `TestWrapMCPStyle_ErrorExclusion` — a test where the unmarshal call's error return is checked (`if err := json.Unmarshal(..., &data); err != nil { t.Fatal(err) }`) and then map fields are asserted. This validates FR-009: the error assertion should NOT be mapped to ReturnValue, but the data field assertions SHOULD be.
- [x] T010 [US2] Verify that `matchContainerUnwrap` correctly handles the multi-step chain in `TestWrapMCPStyle_DeepChain`: `result := WrapMCPStyle(...)` → `result.Content[0].Text` → `[]byte(...)` → `json.Unmarshal(..., &data)` → `data["key"]`. If the initial implementation from T006 does not trace through this full chain (e.g., because intermediate expressions span field access + index + type assertion in a single statement), adjust the forward tracing to handle compound expressions by using `resolveExprRoot` to check whether an assignment RHS's root ident is in the tracked set (not just direct `containsObject` matches).
- [x] T011 [US2] Add a unit test `TestMatchContainerUnwrap_ChainDepth` in `internal/quality/mapping_test.go` that constructs a synthetic scenario with 4+ intermediate steps and verifies the mapping succeeds. Also test that a chain exceeding the depth limit (> 6 steps) returns nil (no mapping).

**Checkpoint**: Deep chains work. Run `go test -race -count=1 ./internal/quality/...` to verify all quality tests pass.

---

## Phase 5: User Story 3 — Container Mapping Works Alongside Existing Mapping Passes (Priority: P3)

**Goal**: The new pass does not alter any existing mappings — zero regression on confidence levels or effect IDs for the 8 existing test fixtures.

**Independent Test**: Run `go test -race -count=1 -run TestSC003_MappingAccuracy ./internal/quality/...` and verify that the accuracy ratchet holds at >= 85.0% AND no previously-mapped assertion changes its confidence or effect ID.

### Implementation for User Story 3

- [x] T012 [US3] Add `TestSC004_NoRegressionOnExistingFixtures` acceptance test in `internal/quality/quality_test.go` that runs the mapping pipeline on all 8 existing fixtures (welltested, undertested, overspecd, tabledriven, helpers, multilib, indirectmatch, mainpkg) and captures the full list of mapped assertion `{SideEffectID, Confidence, AssertionLocation}` tuples. Compare against a golden snapshot to verify zero changes. If any mapped assertion's confidence or effect ID differs from the baseline, the test fails with a descriptive error message identifying the regression.
- [x] T013 [US3] Add `TestMatchContainerUnwrap_ReturnsNil_WhenDirectMatchExists` unit test in `internal/quality/mapping_test.go` verifying that for an assertion already matchable by direct identity (confidence 75), the pipeline returns the direct match and `matchContainerUnwrap` is never reached (the short-circuit ensures this). Use a fixture where the assertion directly references the return variable (no container unwrap needed).

**Checkpoint**: Non-regression confirmed. Run full test suite: `go test -race -count=1 -short ./...`

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation, and cleanup.

- [x] T014 [P] Export `matchContainerUnwrap` for external testing via `internal/quality/export_test.go` if needed by any test in `mapping_test.go` that calls it directly. If all tests go through the pipeline entry point (`MapAssertionsToEffects`), skip this task.
- [x] T015 Run `golangci-lint run` and fix any lint issues introduced by the new code in `internal/quality/mapping.go`.
- [x] T016 Run full CI-equivalent test suite: `go test -race -count=1 -short ./...` and verify all tests pass.
- [x] T017 Update the ratchet comment block in `internal/quality/quality_test.go` (around lines 1050-1070) to document the new accuracy milestone: add a row for "After container unwrap mapping" with the new mapped/total count.
- [x] T018 Verify quickstart.md validation — run each command listed in `specs/034-container-assertion-mapping/quickstart.md` and confirm they produce the expected results.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (fixture files must exist for the helpers to be testable against)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (helper functions must exist)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (core implementation must exist to extend)
- **User Story 3 (Phase 5)**: Depends on Phase 3 (pipeline insertion must exist to verify non-regression)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational only — core implementation
- **User Story 2 (P2)**: Depends on US1 — extends the algorithm for deep chains
- **User Story 3 (P3)**: Depends on US1 — verifies non-regression of existing behavior

### Within Each Phase

- T001 and T002 can run in parallel (different files)
- T003, T004, T005 can run in parallel (independent helper functions in the same file, but no dependencies between them)
- T006 must complete before T007 (function must exist before pipeline insertion)
- T007 must complete before T008 (pipeline must include the pass before ratchet test can validate it)
- T012 and T013 can run in parallel (different test files/functions)
- T014, T015, T017 can run in parallel (different files/concerns)

### Parallel Opportunities

```text
Phase 1:  T001 ─┐
          T002 ─┘ (parallel — different files)

Phase 2:  T003 ─┐
          T004 ─┤ (parallel — independent functions)
          T005 ─┘

Phase 3:  T006 → T007 → T008 (sequential — each depends on prior)

Phase 4:  T009 → T010 → T011 (sequential — fixture → fix → test)

Phase 5:  T012 ─┐
          T013 ─┘ (parallel — independent tests)

Phase 6:  T014 ─┐
          T015 ─┤ (parallel — independent concerns)
          T017 ─┘
          T016 → T018 (sequential — full suite then quickstart validation)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (test fixtures)
2. Complete Phase 2: Foundational (helper functions)
3. Complete Phase 3: User Story 1 (core mapping + pipeline insertion + ratchet)
4. **STOP and VALIDATE**: Run `go test -race -count=1 -run TestSC003 ./internal/quality/...`
5. At this point, the core container mapping works for basic patterns

### Incremental Delivery

1. Setup + Foundational → Helpers ready
2. User Story 1 → Core mapping works → Ratchet validates (MVP!)
3. User Story 2 → Deep chains work → MCP pattern fully covered
4. User Story 3 → Non-regression confirmed → Full confidence
5. Polish → Lint, docs, final validation

---

## Notes

- All code changes are within `internal/quality/` — no cross-package impacts
- The existing ratchet test (`TestSC003_MappingAccuracy`) is the primary validation gate
- The 85.0% ratchet floor must hold throughout — check after each phase
- Confidence 55 is a constant; define it as `const containerUnwrapConfidence = 55` alongside the existing confidence constants in `mapping.go`
- The `containerunwrap` fixture follows the pattern of the 8 existing fixtures — same directory structure, same test conventions
