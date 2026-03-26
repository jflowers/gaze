# Implementation Plan: Fix Missing Effect Detection

**Branch**: `036-fix-missing-effects` | **Date**: 2026-03-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/036-fix-missing-effects/spec.md`

## Summary

Fix the `no_effects_detected` problem for functions with clear observable side effects by: (1) adding an AST-based mutation detection fallback when SSA construction fails, (2) distinguishing "no test" from "no effects" in the contract coverage pipeline, and (3) surfacing SSA degradation diagnostics. Changes span `internal/analysis/` (AST fallback) and `internal/crap/` (coverage reason).

## Technical Context

**Language/Version**: Go 1.25+ (module minimum per go.mod directive)
**Primary Dependencies**: `go/ast`, `go/types`, `golang.org/x/tools/go/ssa`, `golang.org/x/tools/go/packages` (all existing; no new dependencies)
**Storage**: N/A — no persistence changes
**Testing**: Standard library `testing` package; existing mutation analysis tests, SSA build tests, analysis acceptance tests
**Target Platform**: Cross-platform CLI (darwin/linux x amd64/arm64)
**Project Type**: Single Go module
**Performance Goals**: Static analysis — no runtime performance constraints. AST fallback is O(n) where n is the number of statements in the function body, same as existing AST analyzers.
**Constraints**: AST fallback must not produce effects when SSA succeeds (FR-005). Must not introduce false positives for local variable assignments (FR-003) or value receivers (FR-004). Existing effect detection must not regress (SC-003).
**Scale/Scope**: Primary changes in `internal/analysis/mutation.go` (AST fallback function + insertion). Secondary changes in `internal/crap/contract.go` (new coverage reason). Test files in both packages.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Accuracy — PASS

- The AST fallback directly reduces false negatives: void methods with receiver mutations that were silently missed (due to SSA failure) will now be detected.
- FR-003, FR-004, and FR-009 prevent false positives by restricting detection to pointer receiver field assignments and method calls.
- SC-003 enforces zero regression on existing effect detection.
- The new `no_test_coverage` reason improves accuracy of coverage reporting — "no test" is a fundamentally different situation from "no effects."

### II. Minimal Assumptions — PASS

- The AST fallback detects the most common mutation patterns (field assignment, map/slice access, method calls on receiver fields) without requiring any user annotation.
- No new assumptions about project structure, test framework, or coding style.
- The fallback acknowledges its lower fidelity in the Assumptions section — it does not claim SSA-level precision.

### III. Actionable Output — PASS

- Functions previously invisible (zero effects) will now appear with detected mutations, enabling contract coverage analysis.
- The `no_test_coverage` reason tells the user exactly what to do: write a test for this function.
- SSA degradation diagnostics (FR-007) tell the user which packages are affected and why.

### IV. Testability — PASS

- The AST fallback is independently testable by passing `ssaPkg = nil` to `AnalyzeMutations` — no dependency injection required (the nil check is the injection point).
- Existing test fixtures (`testdata/src/mutation/`) contain all the patterns needed: pointer receiver mutations, value receiver traps, pointer arg mutations, read-only access.
- **Coverage strategy**: Unit tests for the new `analyzeASTMutations` function, integration tests via `AnalyzeMutations(nil)`, non-regression via existing analysis acceptance tests.

## Project Structure

### Documentation (this feature)

```text
specs/036-fix-missing-effects/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
internal/analysis/
├── mutation.go              # Primary: new analyzeASTMutations function + fallback insertion
├── mutation_test.go         # Tests for AST fallback
└── testdata/src/mutation/   # Existing fixtures (sufficient)

internal/crap/
├── contract.go              # Secondary: new no_test_coverage reason
├── analyze.go               # Document new reason in ContractCoverageInfo.Reason
└── (test files for coverage reason)
```

**Structure Decision**: All changes are within existing packages. No new files for production code. The AST fallback is a new function in `mutation.go`. The coverage reason is a ~5-line change in `contract.go`.

## Design

### Change 1: AST-Based Mutation Fallback

A new function `analyzeASTMutations` in `mutation.go` that detects receiver and pointer argument mutations using AST walking when SSA is unavailable.

**Insertion point**: `AnalyzeMutations` at line 75. When `ssaPkg == nil`, instead of `return nil`, call `analyzeASTMutations(fset, fd, fnObj, pkgPath, funcName)`.

**Algorithm for `analyzeASTMutations`**:

1. **Extract receiver info**: From `fd.Recv`, check if the receiver is a pointer type (`*ast.StarExpr`). If value receiver, return nil immediately (FR-004).

2. **Get receiver name**: The first name in `fd.Recv.List[0].Names` (e.g., `si` in `func (si *SearchIndex) ReindexPage(...)`).

3. **Walk the function body**: Use `ast.Inspect` on `fd.Body` to find:
   - **Direct field assignments** (`*ast.AssignStmt`): Check if the LHS is a selector expression (`*ast.SelectorExpr`) where the root identifier matches the receiver name. This covers `si.field = value` and `si.field[key] = value` (via nested index expressions).
   - **Method calls on receiver fields** (`*ast.CallExpr`): Check if the function being called is a selector on a selector of the receiver (e.g., `si.map.Delete(key)` → `CallExpr.Fun` is `SelectorExpr{X: SelectorExpr{X: Ident("si"), Sel: "map"}, Sel: "Delete"}`).

4. **Deduplicate**: If multiple assignments to the same receiver are found, emit a single `ReceiverMutation` effect (the SSA-based detector also deduplicates per receiver).

5. **Pointer argument mutations**: Walk `fd.Type.Params` for pointer-typed parameters. For each, check if the function body assigns to fields of that parameter. This covers `PointerArgMutation` effects.

6. **Return** the detected effects as `[]taxonomy.SideEffect` with the same type/tier as SSA-detected mutations.

**Receiver name resolution**: The receiver name comes from `fd.Recv.List[0].Names[0].Name`. If the receiver is unnamed (rare but legal: `func (*T) Method()`), the fallback cannot trace assignments and should return nil — this is an acceptable limitation.

**False positive prevention (FR-003, FR-009)**:
- Only assignments where the LHS root identifier matches the receiver or pointer parameter name are detected.
- `_ = si.field` is an assignment but with `_` as the blank identifier on the LHS — not a field assignment.
- `fmt.Println(si.field)` is a call expression, not an assignment — not detected as a mutation.

### Change 2: `no_test_coverage` Reason

In `contract.go`, the `BuildContractCoverageFunc` closure currently has no path that produces `no_test_coverage`. The `no_effects_detected` reason fires when a function has zero effects. The `no_test_coverage` reason would fire when a function has detected effects but no test targets it.

However, looking at the code flow more carefully: when `loadTestPackage` fails at line 174-176, the entire `analyzePackageCoverage` function returns `nil, ""` — the function is never added to the coverage map at all. The CRAP pipeline's `ContractCoverageFunc` lookup returns `ok=false`, and the function gets no coverage data (not even `no_effects_detected`).

The `no_test_coverage` reason addresses a different path: when the quality pipeline runs but target inference (`InferTargets`) does not find a test for a specific function. In this case, the function does not get a `QualityReport`, and it's absent from the coverage map.

**Implementation**: In `BuildContractCoverageFunc` at lines 56-111, when a function key is not found in the coverage map (`ok == false`), check if the function had detected effects in the analysis results. If it did, return `ContractCoverageInfo{Reason: "no_test_coverage"}` instead of the default (which currently means "no data available").

This requires threading the analysis results through to the closure, or keeping a separate set of "functions with effects" alongside the coverage map.

**Simpler approach**: Add the analysis results as a second data source in the closure. After `analyzePackageCoverage` runs, the classified results contain all functions and their effects. When the closure is called for a function not in the coverage map, check if that function appears in the classified results with > 0 effects. If so: `Reason: "no_test_coverage"`. If not: `Reason: "no_effects_detected"`.

### Change 3: SSA Degradation Diagnostics

The existing `SSADegradedPackages` mechanism (added in spec 021) already tracks degraded packages. The AST fallback's diagnostics should integrate with this:

- In `AnalyzeMutations`, when `ssaPkg == nil`, log at warning level (existing behavior) and additionally set a flag or metadata that the caller can read.
- In `analyzeFunction` (`analyzer.go`), after calling `AnalyzeMutations`, if the result came from the AST fallback, annotate the effects with `Source: "ast_fallback"` (or a similar marker).

The simplest approach: add an `ASTFallback bool` field to the `AnalyzeMutations` return, or use a convention where AST-fallback-detected effects have a specific `Description` suffix (e.g., " (AST fallback)").

**Chosen approach**: Return the same `[]taxonomy.SideEffect` type — no new fields. Add `" (AST fallback)"` to the `Description` field of AST-detected mutations. This requires zero schema changes and is visible in both JSON and text output.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
