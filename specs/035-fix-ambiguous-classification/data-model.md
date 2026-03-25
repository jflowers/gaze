# Data Model: Fix Ambiguous Classification for Clear Return Types

**Branch**: `035-fix-ambiguous-classification`
**Date**: 2026-03-25

## Overview

This feature modifies no data structures, JSON schemas, or output formats. It changes the internal scoring arithmetic within the existing `classify` package. The data model impact is limited to:

1. One new entry in the `contractualPrefixes` slice (naming.go)
2. A new constant for the reduced GoDoc weight (godoc.go)
3. A new signal source label for the reduced GoDoc signal

## Entities

### Existing: contractualPrefixes (modified)

The `contractualPrefixes` slice in `naming.go` gains one new entry:

| Prefix | ImpliesFor | Status |
|--------|-----------|--------|
| `"New"` | `ReturnValue`, `ErrorReturn` | **NEW** |
| `"Get"` | `ReturnValue` | Existing |
| `"Build"` | `ReturnValue`, `ErrorReturn` | Existing |
| (13 others) | (various) | Existing |

### New Constant: reducedGodocWeight

A new weight constant in `godoc.go` for the non-matching type signal:

| Constant | Value | Description |
|----------|-------|-------------|
| `reducedGodocWeight` | 5 | Weight for GoDoc keyword match where the effect type does not appear in the keyword's `impliesFor` list |

### New Signal Source Label

When the reduced GoDoc signal fires, it uses a distinct source label to differentiate from the full-weight signal in the reasoning output:

| Source | Weight | When |
|--------|--------|------|
| `"godoc_keyword"` | +15 | Keyword matches AND effect type matches `impliesFor` (existing) |
| `"godoc_keyword_indirect"` | +5 | Keyword matches BUT effect type does NOT match `impliesFor` (new) |

## Schema Impact

No changes to:
- Classification JSON output format
- Quality JSON Schema
- CRAP JSON output format
- Report JSON payload

The existing `confidence` field in classifications will contain higher values for affected functions (this is the desired outcome, not a format change). The `signals` array in verbose output will contain the new `"godoc_keyword_indirect"` source label for the reduced signal.

## State Transitions

N/A — no lifecycle or state machines. Classification is a pure function: analysis results in, classified results out.
