# Data Model: Fix Missing Effect Detection

**Branch**: `036-fix-missing-effects`
**Date**: 2026-03-25

## Overview

This feature adds one new function to `internal/analysis/mutation.go` and one new coverage reason string to `internal/crap/contract.go`. No new data structures, JSON schemas, or output formats are introduced. The data model impact is:

1. A new internal function `analyzeASTMutations` (not exported)
2. A new coverage reason value `"no_test_coverage"`
3. AST-fallback-detected effects annotated with `" (AST fallback)"` in their Description field

## Entities

### Existing: SideEffect (unchanged)

The `taxonomy.SideEffect` struct is the output of mutation detection. No fields are added or modified. AST-fallback effects use the same type/tier values as SSA-detected effects:

| Field | Value for AST Fallback |
|-------|----------------------|
| `Type` | `taxonomy.ReceiverMutation` or `taxonomy.PointerArgMutation` |
| `Tier` | `P0` |
| `Description` | Standard description + `" (AST fallback)"` suffix |
| `Target` | Receiver or parameter name |

### New: Coverage Reason Value

The `ContractCoverageInfo.Reason` field gains a new documented value:

| Reason | When | Action |
|--------|------|--------|
| `"no_effects_detected"` | Zero effects detected for the function | Investigate: does the function have side effects? Is SSA degrading? |
| `"all_effects_ambiguous"` | Effects detected but all classified as ambiguous | Add GoDoc, improve naming conventions |
| `"no_test_coverage"` | **NEW** — Effects detected but no test targets this function | Write a test that calls this function |
| `"no_assertions_mapped"` | Test exists but assertions don't map to effects | Add assertions for the function's side effects |

## Schema Impact

No changes to:
- Analysis JSON output format (effects are the same `SideEffect` type)
- Classification JSON output format
- Quality JSON Schema
- CRAP JSON output format

The `Description` field on AST-fallback effects will contain the `" (AST fallback)"` suffix, which is a string value change within the existing schema. The `Reason` field on `ContractCoverageInfo` gains a new string value, which is within the existing free-form string type.

## State Transitions

N/A — no lifecycle or state machines. Analysis is a pure function: Go packages in, effects out.
