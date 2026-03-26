# Quickstart: Fix Missing Effect Detection

**Branch**: `036-fix-missing-effects`
**Date**: 2026-03-25

## What This Changes

Gaze's effect detection gains an AST-based fallback for mutation detection when SSA construction fails. Void methods that mutate receiver fields are no longer silently reported as having zero effects. Additionally, functions with detected effects but no test coverage now show `no_test_coverage` instead of `no_effects_detected`, giving users clearer guidance.

## Before vs After

**Before**: `func (si *SearchIndex) ReindexPage(page *Page)` with SSA failure → `no_effects_detected`, 0% contract coverage.

**After**: The same function → `ReceiverMutation` detected via AST fallback, contract coverage reflects actual test assertions.

**Before**: `func StripBullet(s string) string` with no test calling it → missing from report entirely.

**After**: The same function → `ReturnValue` detected, appears in report with reason `no_test_coverage`.

## Files Modified

| File | Change |
|------|--------|
| `internal/analysis/mutation.go` | New `analyzeASTMutations` function + fallback insertion at line 75 |
| `internal/analysis/mutation_test.go` | Tests for AST fallback with `ssaPkg = nil` |
| `internal/crap/contract.go` | New `no_test_coverage` reason for functions with effects but no test |
| `internal/crap/analyze.go` | Document `no_test_coverage` in Reason field comments |

## How to Verify

```bash
# Run mutation detection tests (includes AST fallback)
go test -race -count=1 ./internal/analysis/...

# Run CRAP/contract coverage tests
go test -race -count=1 ./internal/crap/...

# Run the full unit+integration suite
go test -race -count=1 -short ./...
```

## No User-Facing Changes

- No new CLI flags or commands
- No new configuration options
- No changes to JSON output schema structure
- Effect descriptions for AST-fallback mutations include "(AST fallback)" suffix for transparency
- Coverage reason `no_test_coverage` is a new value in the `reason` field (existing field, new value)
