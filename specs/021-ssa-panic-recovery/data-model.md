# Data Model: SSA Panic Recovery

**Feature**: 021-ssa-panic-recovery
**Date**: 2026-03-13

## Overview

This feature introduces no new persistent data structures, entities, or state.
The change is purely behavioral — wrapping two existing function calls with
panic recovery logic.

## Modified Function Contracts

### BuildSSA

**Package**: `internal/analysis`
**File**: `mutation.go`

| Aspect   | Before                              | After                                           |
|----------|-------------------------------------|-------------------------------------------------|
| Input    | `pkg *packages.Package`             | `pkg *packages.Package` (unchanged)             |
| Output   | `*ssa.Package`                      | `*ssa.Package` (unchanged)                      |
| Panic    | Propagates to caller                | Recovered; returns `nil`                        |
| Logging  | None                                | Warning on recovery (package path); debug (raw panic value) |
| Contract | Returns nil if SSA construction fails | Returns nil if SSA construction fails or panics |

### BuildTestSSA

**Package**: `internal/quality`
**File**: `pairing.go`

| Aspect   | Before                                         | After                                                      |
|----------|-------------------------------------------------|------------------------------------------------------------|
| Input    | `pkg *packages.Package`                         | `pkg *packages.Package` (unchanged)                        |
| Output   | `(*ssa.Program, *ssa.Package, error)`           | `(*ssa.Program, *ssa.Package, error)` (unchanged)          |
| Panic    | Propagates to caller                            | Recovered; returns `nil, nil, error`                       |
| Logging  | None                                            | Warning on recovery (package path); debug (raw panic value) |
| Contract | Returns error if SSA construction produces nil  | Returns error if SSA construction produces nil or panics   |

### safeSSABuild (new, unexported)

**Package**: `internal/analysis` (primary), duplicated or shared in `internal/quality`
**File**: `mutation.go` or a new `ssa_recovery.go` (per research R3)

| Aspect   | Value                                                         |
|----------|---------------------------------------------------------------|
| Input    | `buildFn func()`                                              |
| Output   | `panicVal any` — nil if no panic, recovered value otherwise   |
| Purpose  | Isolate the `recover()` pattern for testability               |

## Caller Impact

All callers of `BuildSSA` and `BuildTestSSA` already handle the nil/error case:

| Caller | File | Handles nil/error via |
|--------|------|-----------------------|
| `Analyze()` | `internal/analysis/analyzer.go:45` | `ssaPkg` passed to `analyzeFunction()`; mutation analysis skipped when nil |
| `AnalyzeFunctionWithSSA()` | `internal/analysis/analyzer.go:127` | Calls `BuildSSA` if nil; skips mutation if still nil |
| `Assess()` | `internal/quality/quality.go:92` | Returns error from `BuildTestSSA`; callers skip quality on error |
| `runQualityForPackage()` | `internal/aireport/runner_steps.go:139` | Returns nil on `Assess` error |
| `runClassifyStep()` | `internal/aireport/runner_steps.go:165` | `LoadAndAnalyze` error → `continue` |

No caller changes required (FR-006 verified at design time).
