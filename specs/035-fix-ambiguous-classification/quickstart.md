# Quickstart: Fix Ambiguous Classification for Clear Return Types

**Branch**: `035-fix-ambiguous-classification`
**Date**: 2026-03-25

## What This Changes

Gaze's classification scoring engine is updated so that Go constructor functions (`New`/`NewXxx`) and functions with GoDoc contractual keywords receive higher confidence scores. Functions that were previously classified as "ambiguous" despite clear return types and documentation will now be correctly classified as "contractual."

## Before vs After

**Before**: `func New(baseURL string) *Client` with GoDoc "Returns a *Client..." → confidence 59, classification "ambiguous", contract coverage 0%.

**After**: The same function → confidence 100, classification "contractual", contract coverage reflects actual test assertions.

## Files Modified

| File | Change |
|------|--------|
| `internal/classify/naming.go` | Add `"New"` to `contractualPrefixes` (1 line) |
| `internal/classify/godoc.go` | Add reduced signal (+5) for non-matching effect types (~5 lines) |
| `internal/classify/classify_test.go` | Add test cases for `New` prefix and non-regression |
| `internal/classify/godoc_test.go` | Add test cases for reduced GoDoc signal |

## How to Verify

```bash
# Run the classification test suite
go test -race -count=1 ./internal/classify/...

# Run the full unit+integration suite
go test -race -count=1 -short ./...
```

## No User-Facing Changes

- No new CLI flags or commands
- No new configuration options
- No changes to JSON output schema
- No changes to text report format
- Classification confidence values will increase for affected functions (this is the desired outcome, not a format change)
