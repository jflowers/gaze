# Quickstart: Container Type Assertion Mapping

**Branch**: `034-container-assertion-mapping`
**Date**: 2026-03-24

## What This Changes

Gaze's assertion mapping pipeline gains a new pass that traces test assertions through container unwrap patterns. When a test assigns a function's return value, accesses a field, passes it through a transformation (like JSON unmarshal), and asserts on the result, those assertions now count toward contract coverage.

## Before vs After

**Before**: A function returning `*mcp.CallToolResult` with 4 test assertions on JSON fields → 0% contract coverage (Q3: Simple But Underspecified).

**After**: The same function → contract coverage > 0% because assertions on unwrapped JSON fields are traced back to the ReturnValue effect.

## Files Modified

| File | Change |
|------|--------|
| `internal/quality/mapping.go` | New `matchContainerUnwrap` function + pipeline insertion after inline call pass |
| `internal/quality/mapping_test.go` | Unit tests for the new pass |
| `internal/quality/quality_test.go` | Add `containerunwrap` fixture to ratchet test |
| `internal/quality/export_test.go` | Export `matchContainerUnwrap` for testing (if needed) |
| `internal/quality/testdata/src/containerunwrap/containerunwrap.go` | Test fixture: functions returning container types |
| `internal/quality/testdata/src/containerunwrap/containerunwrap_test.go` | Test fixture: unmarshal-and-assert test pattern |

## How to Verify

```bash
# Run the mapping accuracy ratchet test
go test -race -count=1 -run TestSC003_MappingAccuracy ./internal/quality/...

# Run all quality tests
go test -race -count=1 ./internal/quality/...

# Run full test suite
go test -race -count=1 -short ./...
```

## Confidence Level

Container-mapped assertions are assigned confidence **55**, slotting between:
- Inline call pass: 60
- AI mapper pass: 50

This reflects the additional indirection (field access + transformation) compared to direct structural matches, while providing stronger evidence than AI-assisted semantic matching.

## No User-Facing Changes

- No new CLI flags or commands
- No new configuration options
- No changes to JSON output schema
- No changes to text report format
- Contract coverage percentages will increase for affected functions (this is the desired outcome, not a format change)
