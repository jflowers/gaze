# Implementation Plan: Container Type Assertion Mapping

**Branch**: `034-container-assertion-mapping` | **Date**: 2026-03-24 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/034-container-assertion-mapping/spec.md`

## Summary

Add a new "container unwrap" mapping pass to Gaze's assertion mapping pipeline that traces data flow through container field access and transformation calls (e.g., JSON unmarshal). When a test assigns a function's return value, accesses a field, passes it through a transformation, and asserts on the output, the new pass connects those assertions to the original ReturnValue side effect at confidence 55. This closes the coverage gap for functions returning generic container/wrapper types.

## Technical Context

**Language/Version**: Go 1.25+ (module minimum per go.mod directive)
**Primary Dependencies**: `go/ast`, `go/types`, `golang.org/x/tools/go/ssa`, `golang.org/x/tools/go/packages` (all existing; no new dependencies)
**Storage**: N/A — no persistence changes
**Testing**: Standard library `testing` package; existing ratchet test `TestSC003_MappingAccuracy` at 85.0% floor (current: 86.4%, 57/66)
**Target Platform**: Cross-platform CLI (darwin/linux x amd64/arm64)
**Project Type**: Single Go module
**Performance Goals**: Static analysis — no runtime performance constraints. Mapping pass should complete in < 1ms per assertion site (consistent with existing passes).
**Constraints**: Must fit within the existing multi-pass pipeline architecture; confidence 55 per FR-003; no new external dependencies.
**Scale/Scope**: Affects `internal/quality/mapping.go` (primary), new test fixture under `internal/quality/testdata/src/containerunwrap/`, ratchet test update.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Accuracy — PASS

- The new pass adds a mapping capability for a pattern that currently produces false negatives (unmapped assertions → understated contract coverage). This directly reduces false negatives.
- FR-005 and SC-004 ensure no existing mappings are altered (no false positives introduced).
- SC-002 enforces the existing 85.0% ratchet floor.
- A new test fixture validates the container pattern specifically (SC-003).

### II. Minimal Assumptions — PASS

- FR-008 ensures the pass is structural, not type-specific — no assumptions about particular container types.
- FR-001 (clarified) uses structural signature pattern recognition for transformation calls, not a hardcoded list.
- No user annotation or restructuring required.

### III. Actionable Output — PASS

- Previously-unmapped assertions will now produce mapped results with confidence 55, directly improving contract coverage percentages in reports.
- The confidence level is visible in JSON output, allowing users to understand the mapping certainty.

### IV. Testability — PASS

- The new `matchContainerUnwrap` function will be independently testable via the existing pipeline test infrastructure.
- A dedicated test fixture (`containerunwrap/`) validates the pass in isolation.
- The ratchet test ensures non-regression.
- **Coverage strategy**: Unit tests for the new function, integration via `TestSC003_MappingAccuracy` ratchet, new fixture-specific acceptance tests for SC-001/SC-003.

## Project Structure

### Documentation (this feature)

```text
specs/034-container-assertion-mapping/
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
internal/quality/
├── mapping.go                      # Primary: new matchContainerUnwrap function + pipeline insertion
├── mapping_test.go                 # Tests for the new mapping pass
├── quality_test.go                 # Ratchet test update (add containerunwrap fixture)
├── export_test.go                  # Export helper for new function if needed
└── testdata/src/
    └── containerunwrap/
        ├── containerunwrap.go      # Target functions returning container types
        └── containerunwrap_test.go # Tests with unwrap-unmarshal-assert pattern
```

**Structure Decision**: All changes are within the existing `internal/quality/` package. No new packages or directories outside of the test fixture. This follows the established pattern of the 8 existing fixtures.

## Design

### Pipeline Integration

The container unwrap pass is inserted into `mapAssertionsToEffectsImpl` in the main matching loop, after the inline call pass (confidence 60) and before the AI mapper pass (confidence 50):

```text
For each assertion site:
  1. matchAssertionToEffect (direct 75, helper bridge 70, indirect 65)
  2. matchInlineCall (60)
  3. matchContainerUnwrap (55)  ← NEW
  4. tryAIMapping (50)
```

This placement ensures:
- Higher-confidence structural matches take priority (FR-005)
- The pass only activates when all existing mechanical passes fail
- The AI mapper remains the last-resort fallback

### Container Unwrap Algorithm

The `matchContainerUnwrap` function traces the data flow from the return value variable through intermediate operations to the assertion variable. The algorithm:

1. **Find the return value variable**: Use the existing `objToEffectID` map to identify which variable holds the target function's return value.

2. **Trace forward through assignments**: Starting from the return value variable, walk subsequent statements in the test function's AST looking for assignment statements where the RHS references the tracked variable (via field access, index, type assertion, or function call argument).

3. **Detect transformation calls**: When a tracked variable flows into a function call as an argument, check if the call matches the transformation signature pattern:
   - The function accepts a byte-like input (`[]byte`, `string`, `io.Reader`)
   - The function accepts a pointer destination argument
   - Example: `json.Unmarshal(body, &data)` — `body` is tracked, `data` becomes the new tracked variable

4. **Bridge across transformation**: When a transformation is detected, the pointer destination's base variable becomes the new tracked variable. Continue tracing from this variable.

5. **Match assertions**: When a tracked variable (or a field/index of it) appears in an assertion expression, produce a mapping at confidence 55 to the ReturnValue effect.

### Transformation Signature Detection

Per the clarification in the spec, transformation calls are recognized structurally:

- A function where at least one parameter has a byte-like type (`[]byte`, `string`) or implements `io.Reader`
- AND at least one parameter is a pointer type (the destination)
- The tracked variable must flow into the byte-like parameter
- The destination pointer's base variable becomes the next link in the chain

This is checked via `go/types` type information on the call expression, not by function name.

### Chain Depth

FR-004 requires tracing through at least 4 intermediate steps. The algorithm uses iterative forward tracing (not recursive) with a configurable maximum depth (default 6, covering the MCP test pattern of result → field → type assert → field → unmarshal → assert with margin).

### Error Exclusion

FR-009 requires excluding transformation error assertions. When a transformation call returns `(T, error)` and the test assigns `err := transform(...)`, the error variable is explicitly excluded from the tracked variable set. Only the destination pointer variable is tracked forward.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
