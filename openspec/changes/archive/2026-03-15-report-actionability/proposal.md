## Why

The gaze report tells users *what* is wrong (high CRAP scores, low contract coverage, Q4 quadrant) but not *why* or *how to fix it*. Users waste time on the wrong remediation strategy:

- Functions with complexity >= 15 can never reach CRAP < 15 even at 100% coverage — they need **decomposition**, not more tests. Users discover this only after writing tests that don't move the needle.
- Functions at 0% contract coverage may have 94% line coverage — they need **contract assertions** in existing tests, not new test functions. Users write redundant tests instead of strengthening assertions.
- Low contract coverage may be caused by the classification engine labeling effects as "ambiguous" (confidence just below threshold) — no amount of test work can fix this. Users need to use `--ai` reclassification or adjust thresholds.

These three issues (#42, #44, #48) all stem from the same gap: the CRAP score data model lacks actionable metadata.

## What Changes

Add two new fields to the CRAP score data model:

1. **`FixStrategy`** on `Score` — a deterministic label computed from complexity, coverage, and quadrant that tells the user (or agent) what kind of fix is needed: `decompose`, `add_tests`, `add_assertions`, or `decompose_and_test`.

2. **`FixStrategyCounts`** on `Summary` — aggregated counts per strategy, giving a project-level view of how many functions need each type of fix.

Update the text report to group the "Worst Offenders" section by fix strategy and include strategy counts in the summary.

## Capabilities

### New Capabilities
- `FixStrategy` field on `crap.Score`: Deterministic label based on complexity and coverage thresholds. One of: `decompose` (complexity >= CRAPThreshold, coverage alone can't help), `add_tests` (0% line coverage, complexity < CRAPThreshold), `add_assertions` (has line coverage but contract coverage is low/zero), `decompose_and_test` (complexity >= CRAPThreshold AND 0% coverage).
- `FixStrategyCounts` field on `crap.Summary`: Map of fix strategy to count, giving project-level remediation planning data.

### Modified Capabilities
- `computeScores`: Computes `FixStrategy` for each score based on complexity and coverage thresholds.
- `buildSummary`: Accumulates `FixStrategyCounts` from scores.
- `WriteText`: Adds strategy counts to the summary section and groups worst offenders by strategy.
- `WriteJSON`: Automatically includes new fields (no code change needed — struct serialization).

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/crap/crap.go` | Add `FixStrategy` type, constants, field on `Score`, `FixStrategyCounts` on `Summary` |
| `internal/crap/analyze.go` | Compute `FixStrategy` in `computeScores`, accumulate in `buildSummary` |
| `internal/crap/report.go` | Display strategy counts in summary, group worst offenders by strategy |
| `internal/crap/crap_test.go` | Test strategy assignment and summary counts |
| `AGENTS.md` | Update Recent Changes |

No changes to `aireport` package — the CRAP JSON flows through `ReportPayload.CRAP` as raw bytes, so new fields propagate automatically.

## Constitution Alignment

### I. Autonomous Collaboration
**PASS** — New fields are self-describing JSON with clear semantics. Agents can consume `fix_strategy` to plan work without human interpretation.

### II. Composability First
**PASS** — Additive fields with `omitempty`. Existing consumers that don't know about the new fields are unaffected.

### III. Observable Quality
**PASS** — The fix strategy is deterministic from the CRAP formula. No ambiguity about what the label means. JSON output is machine-parseable.

### IV. Testability
**PASS** — Strategy assignment is a pure function of complexity and coverage values — directly testable with synthetic inputs via the existing `computeScores` test infrastructure.
