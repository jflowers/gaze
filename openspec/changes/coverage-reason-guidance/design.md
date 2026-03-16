## Context

`ContractCoverageFunc` in `crap.Options` returns `(float64, bool)` — just a percentage and availability flag. The rich data from `quality.ComputeContractCoverage` (which includes `TotalContractual`, `Gaps`, and per-effect `Classification`) is collapsed to a single number by `buildContractCoverageFunc` in `cmd/gaze/main.go`. The reason for 0% coverage is lost at this boundary.

## Goals / Non-Goals

### Goals
- Surface *why* contract coverage is 0% (or low) at the per-function CRAP score level
- Provide actionable recommendation based on the reason
- Include confidence range for the `all_effects_ambiguous` case so users can see how close they are to the threshold

### Non-Goals
- Changing the classification engine or thresholds
- Adding new CLI flags for threshold adjustment
- Modifying the quality pipeline's classification logic

## Decisions

### D1: ContractCoverageInfo replaces bare float64

```go
type ContractCoverageInfo struct {
    Percentage    float64
    Reason        string // "", "all_effects_ambiguous", "no_effects_detected", "no_assertions_mapped"
    MinConfidence int    // lowest classification confidence (0 if no effects)
    MaxConfidence int    // highest classification confidence (0 if no effects)
}
```

`ContractCoverageFunc` changes from `func(pkg, fn string) (float64, bool)` to `func(pkg, fn string) (ContractCoverageInfo, bool)`.

**Rationale**: A struct is extensible. The signature change breaks existing callers but there are only 2: `buildContractCoverageFunc` in `cmd/gaze/main.go` and test mocks. Both are internal.

### D2: Reason computation in buildContractCoverageFunc

The reason is computed where the full quality data is available — in `buildContractCoverageFunc` / `analyzePackageCoverage`. The logic:

```
if TotalContractual == 0:
    if all effects have Classification.Label == "ambiguous":
        reason = "all_effects_ambiguous"
        compute min/max confidence from effect Classifications
    elif len(SideEffects) == 0:
        reason = "no_effects_detected"
    else:
        reason = "no_assertions_mapped"  // effects exist but none are contractual
else:
    reason = ""  // normal coverage (may be partial)
```

### D3: Display in text report

The worst-offenders section already shows `[fix_strategy]` labels. Add the reason after the strategy label when non-empty:

```
1. 72.0  LoadConversations [add_assertions] (all effects ambiguous, confidence 78-79)
```

And in the summary, when many functions share the same reason, add a note:

```
Note: 24 functions have 0% contract coverage because all effects are
classified as ambiguous (confidence 78-79, threshold 80).
Use --ai for reclassification or --contractual-threshold=75.
```

## Risks / Trade-offs

### R1: ContractCoverageFunc signature change breaks callers
Only 2 internal callers exist. Both are updated in this change. The signature change is backward-incompatible but the function type is unexported-equivalent (only used within the `crap` package's `Options` struct).
