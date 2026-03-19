# Design: P0 Contractual Default

## Context

The classifier scoring pipeline in `internal/classify/score.go`:

1. `accumulateSignals` starts at `baseConfidence = 50`, sums signal weights
2. `classifyLabel` maps the score: `>=80` contractual, `<50` incidental, `[50,80)` ambiguous
3. Five signal analyzers contribute weights: interface (+30), visibility (+20), callers (+15), godoc (+15), naming (+10/+30)
4. None of the analyzers check the effect's priority tier (P0-P4)

The tier system (`taxonomy.TierOf`) maps each `SideEffectType` to a priority:
- P0: ReturnValue, ErrorReturn, SentinelError, ReceiverMutation, PointerArgMutation
- P1: SliceMutation, MapMutation, GlobalMutation, WriterOutput, HTTPResponseWrite, ChannelSend, ChannelClose, DeferredReturnMutation
- P2-P4: Higher-tier effects (goroutine spawns, file operations, etc.)

## Goals / Non-Goals

### Goals
- P0 effects trend toward `contractual` by default (base 75 instead of 50)
- P1 effects get a moderate boost (base 60)
- P2-P4 effects stay at 50 (unchanged behavior)
- The boost is additive with existing mechanical signals
- Configurable via `.gaze.yaml` for projects that want different behavior

### Non-Goals
- Changing the contractual/incidental thresholds (80/50 stay the same)
- Changing the signal analyzer weights
- Making P0 effects unconditionally contractual (they should still be influenced by signals)

## Decisions

### D1: Tier boost as a signal, not a base change

Rather than changing `baseConfidence` (which would affect all effects), add a tier-based boost as a new signal source inside `accumulateSignals`. This keeps the architecture clean: all confidence modifications are signal weights.

```go
func tierBoost(effectType taxonomy.SideEffectType) int {
    switch taxonomy.TierOf(effectType) {
    case taxonomy.TierP0:
        return 25
    case taxonomy.TierP1:
        return 10
    default:
        return 0
    }
}
```

Applied in `accumulateSignals` before summing mechanical signals:
```go
func accumulateSignals(effectType taxonomy.SideEffectType, signals []taxonomy.Signal) (score int, ...) {
    score = baseConfidence + tierBoost(effectType)
    for _, s := range signals { ... }
}
```

**Rationale**: Adding the tier boost at the signal accumulation level means it interacts naturally with the existing scoring pipeline. The contradiction penalty, clamping, and label classification all work unchanged. The `ComputeScore` function passes `effectType` to `accumulateSignals` (it's already available via `se.Type` in the caller).

### D2: Boost values: P0 = +25, P1 = +10

| Tier | Boost | Base Score | Meaning |
|------|-------|-----------|---------|
| P0 | +25 | 75 | One positive signal (visibility +20 or godoc +15) pushes to contractual. Exported function with GoDoc → 110 → contractual. Unexported with no signals → 75 → ambiguous but close. |
| P1 | +10 | 60 | Needs significant signals to reach contractual. WriterOutput on exported function with GoDoc → 95 → contractual. Unexported → 60 → ambiguous. |
| P2-P4 | 0 | 50 | Unchanged. These effects are genuinely context-dependent. |

**Why not +30 for P0 (immediate contractual at 80)?**: Because that would make P0 effects contractual even on unexported functions with no GoDoc and no callers. While P0 effects are definitionally important, the classifier should still reward visibility signals. A function that nobody calls and has no documentation is legitimately ambiguous — the developer hasn't signaled that its contract matters to external consumers. The +25 boost puts P0 at the doorstep (75) so that ANY positive signal pushes it over.

**Why not configurable boost values in .gaze.yaml?**: The existing `.gaze.yaml` already has `thresholds.contractual` and `thresholds.incidental`. Adding per-tier boost overrides adds complexity for a niche use case. If a project wants different behavior, they can adjust the contractual threshold instead. Defer per-tier config to a future change if requested.

### D3: Update `accumulateSignals` signature

Current: `accumulateSignals(signals []taxonomy.Signal) (score int, hasPositive, hasNegative bool)`
New: `accumulateSignals(effectType taxonomy.SideEffectType, signals []taxonomy.Signal) (score int, hasPositive, hasNegative bool)`

The `effectType` parameter is needed for `tierBoost()`. The only caller is `ComputeScore` which already has `se.Type` available.

### D4: Impact on existing contract coverage numbers

This change will increase contract coverage across all projects because P0 effects (the most common contractual effects) will shift from ambiguous to contractual. Specific impacts:

- **Projects with GoDoc + exported functions**: P0 effects already score 80+ in many cases. Minimal change.
- **Projects without GoDoc on exported functions**: P0 effects move from ~70 (ambiguous) to ~95 (contractual). Significant improvement.
- **Projects with only unexported functions**: P0 effects move from ~50 (ambiguous) to ~75 (still ambiguous but close). Combined with #70 (auto-detect `package main`), this means one GoDoc comment pushes them contractual.
- **GazeCRAPload impact**: Some Q3 functions (simple but underspecified due to all-ambiguous effects) may shift to Q1 (safe) because their effects become contractual. GazeCRAPload may decrease.
- **CI ratchets**: Projects with `--min-contract-coverage` may see their coverage increase (passing is easier). Projects with `--max-gaze-crapload` may see their GazeCRAPload decrease (also easier). No ratchet will break in the "wrong direction."

## Risks / Trade-offs

- **Semantic shift**: Existing gaze users who have tuned `.gaze.yaml` thresholds based on current scoring may see their classification landscape change. Mitigated by: the change makes scoring MORE accurate, not different. Release notes should document the change clearly.
- **Test updates**: Existing classifier tests hardcode expected confidence values. These will need updating. The test changes are mechanical (new expected values) not structural.
- **False contractual**: A P0 effect on a function that's genuinely not part of any contract (e.g., a test helper's return value) might be classified contractual when it shouldn't be. Mitigated by: the existing incidental signals (naming convention `-test`, no callers) still apply. A test helper with naming signals would score 75 - 10 (naming) = 65 → still ambiguous.
