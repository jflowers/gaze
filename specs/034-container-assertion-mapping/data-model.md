# Data Model: Container Type Assertion Mapping

**Branch**: `034-container-assertion-mapping`
**Date**: 2026-03-24

## Overview

This feature adds no new persistent data structures or JSON schema fields. It modifies the internal assertion mapping pipeline to produce mappings that already conform to the existing `taxonomy.AssertionMapping` type. The data model impact is limited to:

1. A new confidence tier value (55) in existing mapping output
2. A new test fixture package
3. Internal function parameters and return types

## Entities

### Existing: AssertionMapping (unchanged)

The `taxonomy.AssertionMapping` struct is the output of the mapping pipeline. No fields are added or modified.

| Field | Type | Description |
|-------|------|-------------|
| `SideEffectID` | `string` | ID of the matched side effect |
| `AssertionLocation` | `string` | File:line of the assertion |
| `AssertionType` | `string` | Type of assertion (comparison, error check, etc.) |
| `Confidence` | `int` | Mapping confidence (0-100) |
| `UnmappedReason` | `string` | Why mapping failed (empty if mapped) |

The container unwrap pass produces mappings with `Confidence: 55` and `SideEffectID` pointing to the `ReturnValue` effect. No new `UnmappedReason` values are introduced; unmapped assertions from the container pattern retain existing reason classifications.

### New Internal: Unwrap Chain (ephemeral, not serialized)

An unwrap chain is the sequence of tracked variables between the target call assignment and the assertion. It is an internal concept within the `matchContainerUnwrap` function and is not exposed in any output format.

| Concept | Representation | Description |
|---------|----------------|-------------|
| Tracked variable | `types.Object` | A variable whose value derives from the target function's return value |
| Chain link | Assignment statement | An `*ast.AssignStmt` where the RHS references a tracked variable and the LHS introduces a new tracked variable |
| Transformation | Call expression | An `*ast.CallExpr` within a chain link that matches the transformation signature pattern |
| Destination variable | `types.Object` | The variable pointed to by the pointer argument of a transformation call; becomes the next tracked variable |

### New Internal: Transformation Signature (ephemeral, not serialized)

A function signature recognized as a transformation call.

| Criterion | Description |
|-----------|-------------|
| Byte-like input | At least one parameter is `[]byte`, `string`, or implements `io.Reader` |
| Pointer destination | At least one parameter is a pointer type (`*T`) |
| Match condition | Both criteria met AND the tracked variable flows into the byte-like parameter position |

## State Transitions

N/A — no lifecycle or state machines. The mapping pipeline is a pure function: assertion sites in, mapped/unmapped classifications out.

## Schema Impact

No changes to:
- Quality JSON Schema (`internal/report/schema.go`)
- CRAP JSON output format
- Report JSON payload (`internal/aireport/payload.go`)

The existing `confidence` field in assertion mappings will contain value `55` for container-mapped assertions. This is within the existing [0, 100] range and requires no schema update.

## Test Fixture Data Model

The new `containerunwrap` test fixture requires:

| File | Content |
|------|---------|
| `containerunwrap.go` | Functions returning wrapper/container types with JSON bodies |
| `containerunwrap_test.go` | Tests that assign, extract fields, unmarshal, and assert on fields of the unmarshalled data |

The fixture follows the pattern established by 8 existing fixtures in `internal/quality/testdata/src/`.
