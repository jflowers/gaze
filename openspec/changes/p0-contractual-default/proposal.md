# P0 Contractual Default

## Why

The classifier assigns `ambiguous/50%` as the base confidence for ALL side effects regardless of their priority tier. This means P0 effects — ReturnValue, ErrorReturn, SentinelError, ReceiverMutation, PointerArgMutation — start at the same confidence as a P4 goroutine spawn. With the default contractual threshold at 80, a P0 effect needs +30 from mechanical signals (GoDoc, visibility, callers, naming, interfaces) to reach `contractual`.

In practice, this produces absurd results. A user tested gaze against a 59-function CLI tool and got 153/154 effects classified as `ambiguous/50%`. The only non-ambiguous result was a log call (incidental/40%). Return values and error returns — the definition of a function's contract — were classified as "ambiguous."

The root cause is in `internal/classify/score.go:14`: `baseConfidence = 50`. The `accumulateSignals` function starts from this base and adds signal weights. None of the five signal analyzers check the effect's tier (P0/P1/P2/P3/P4). The tier system exists in `taxonomy.TierOf()` but is completely unused by the classifier.

P0 effects are contractual by definition. A function's return values and error returns ARE its observable contract — there is no context in which they are "incidental." The classifier should reflect this.

Closes #71.

## What Changes

Add a tier-based confidence boost to the classifier scoring. P0 effects get a +25 boost (base 75), P1 effects get a +10 boost (base 60), P2-P4 stay at 50. This means:

- P0 on an exported function with GoDoc: 50 + 25 (tier) + 20 (visibility) + 15 (godoc) = 110 → clamped to 100 → **contractual**
- P0 on an unexported function, no signals: 50 + 25 (tier) = 75 → **ambiguous** but one signal pushes it contractual
- P1 (WriterOutput, SliceMutation): 50 + 10 = 60 → still ambiguous, needs context
- P2-P4 (goroutine, file ops): 50 → unchanged, correctly ambiguous by default

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `internal/classify/score.go`: `accumulateSignals` adds tier-based boost before summing mechanical signals
- Classification output: P0 effects trend toward `contractual` by default; contract coverage becomes non-zero for projects without `.gaze.yaml` or GoDoc

### Removed Capabilities
- None

## Impact

- `internal/classify/score.go` — add tier boost logic
- `internal/classify/score_test.go` — update expected confidence values
- `internal/classify/classify_test.go` — update integration test expectations if P0 scores change
- Downstream: contract coverage numbers will increase across all projects. GazeCRAPload may decrease (fewer Q3/Q4 functions). CI ratchets may need threshold adjustments.

## Constitution Alignment

### I. Autonomous Collaboration — N/A
### II. Composability First — PASS (no new dependencies)

### III. Observable Quality — PASS

This change directly improves accuracy (Constitution Principle I). The current classifier produces 98% ambiguous results for projects without configuration — that's a false signal. Making P0 effects trend toward contractual produces more accurate classification out-of-the-box.

### IV. Testability — PASS

The tier boost is a pure function (`tierBoost(effectType) int`) testable with a simple table-driven test. Integration tests verify the full scoring pipeline.
