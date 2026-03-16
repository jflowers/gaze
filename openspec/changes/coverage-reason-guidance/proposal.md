## Why

When a function has 0% contract coverage, gaze doesn't explain why. The three root causes require completely different fixes:

1. **All effects ambiguous** — classification confidence is below the contractual threshold. Fix: use `--ai` reclassification or lower `--contractual-threshold`. Adding assertions is futile.
2. **No effects detected** — the function has no observable side effects. Fix: the function may be pure or the analysis missed effects.
3. **No assertions mapped** — assertions exist but couldn't be traced to effects. Fix: restructure tests or use the AI mapper.

Issue #60 reports a real case: 24 Q3 functions all had effects at confidence 78-79 (threshold 80). The developer spent a full iteration adding 55 assertion fixes with zero improvement because the bottleneck was classification, not assertions.

## What Changes

1. Add `ContractCoverageReason *string` to `crap.Score` — a computed diagnostic explaining why coverage is what it is.
2. When the reason is `all_effects_ambiguous`, include confidence range so users can see how close effects are to the threshold.
3. Display the reason and recommendation in the text report worst-offenders section.

## Capabilities

### New Capabilities
- `ContractCoverageReason` on `crap.Score`: One of `all_effects_ambiguous`, `no_effects_detected`, `no_assertions_mapped`, or nil (normal coverage).
- `EffectConfidenceRange` on `crap.Score`: Min/max classification confidence when reason is `all_effects_ambiguous`.

### Modified Capabilities
- `ContractCoverageFunc` signature: Returns a richer struct (`ContractCoverageInfo`) instead of bare `float64`, carrying the reason and confidence range alongside the percentage.
- `computeScores`: Populates the new fields from the callback result.
- `buildContractCoverageFunc` in `cmd/gaze/main.go`: Computes the reason from the quality report's classification data.
- `writeWorstSection`: Displays the reason and actionable recommendation.

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/crap/crap.go` | Add `ContractCoverageReason`, `EffectConfidenceRange` to `Score` |
| `internal/crap/analyze.go` | Add `ContractCoverageInfo` struct, update `ContractCoverageFunc` type, update `computeScores` |
| `internal/crap/report.go` | Display reason in worst-offenders section |
| `cmd/gaze/main.go` | Update `buildContractCoverageFunc` to compute reason from quality data |
| Tests | Update existing callback tests, add reason-specific tests |
| `AGENTS.md` | Update Recent Changes |

## Constitution Alignment

### I. Accuracy — PASS
The reason is derived directly from existing classification data. No new analysis or assumptions.

### II. Minimal Assumptions — PASS
Additive fields with omitempty. Default behavior unchanged for callers that don't use GazeCRAP.

### III. Actionable Output — PASS
This is the primary motivation. The recommendation tells users exactly what to do instead of letting them waste effort.

### IV. Testability — PASS
The reason computation is a pure function of classification labels and confidence values — directly testable with synthetic inputs.
