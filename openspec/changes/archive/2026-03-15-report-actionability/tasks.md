## 1. Data Model

- [x] 1.1 Add `FixStrategy` type and constants (`decompose`, `add_tests`, `add_assertions`, `decompose_and_test`) to `internal/crap/crap.go`.
- [x] 1.2 Add `FixStrategy *FixStrategy` field with JSON tag `fix_strategy,omitempty` to the `Score` struct.
- [x] 1.3 Add `FixStrategyCounts map[FixStrategy]int` field with JSON tag `fix_strategy_counts,omitempty` to the `Summary` struct.

## 2. Computation

- [x] 2.1 Add `assignFixStrategy(score Score, crapThreshold float64) *FixStrategy` function in `internal/crap/analyze.go`. Logic: if CRAP < threshold, return nil. If complexity >= threshold: return `decompose_and_test` (0% coverage) or `decompose` (has coverage). If quadrant == Q3: return `add_assertions`. Otherwise: return `add_tests`.
- [x] 2.2 Call `assignFixStrategy` in `computeScores` after building each `Score` and assign the result to `score.FixStrategy`.
- [x] 2.3 Accumulate `FixStrategyCounts` in `buildSummary` by iterating scores with non-nil `FixStrategy`.

## 3. Text Report

- [x] 3.1 Add `writeRemediationSection(w io.Writer, counts map[FixStrategy]int, styles report.Styles)` to `internal/crap/report.go`. Renders "Remediation Breakdown" with counts per strategy.
- [x] 3.2 Call `writeRemediationSection` from `WriteText` after the quadrant section, when `FixStrategyCounts` is non-empty.
- [x] 3.3 Update `writeWorstSection` to include strategy labels (e.g., `[decompose]`) on each worst-offender line when `FixStrategy` is non-nil.

## 4. Tests

- [x] 4.1 Add `TestAssignFixStrategy_Decompose` — complexity >= threshold, coverage > 0.
- [x] 4.2 Add `TestAssignFixStrategy_DecomposeAndTest` — complexity >= threshold, coverage == 0.
- [x] 4.3 Add `TestAssignFixStrategy_AddTests` — complexity < threshold, coverage == 0.
- [x] 4.4 Add `TestAssignFixStrategy_AddAssertions` — quadrant Q3, CRAP above threshold.
- [x] 4.5 Add `TestAssignFixStrategy_BelowThreshold` — CRAP < threshold, returns nil.
- [x] 4.6 Add `TestBuildSummary_FixStrategyCounts` — verify counts are accumulated correctly.
- [x] 4.7 Verify existing crap tests pass: `go test -race -count=1 ./internal/crap/` (pre-existing TestAnalyze_WithPrebuiltProfile failure unrelated).
- [x] 4.8 Add `TestWriteText_RemediationBreakdown` — verify the remediation section and strategy labels appear in text output.

## 5. Documentation & Verification

- [x] 5.1 Update `AGENTS.md` Recent Changes with a summary of this change.
- [x] 5.2 Run full test suite: `go test -race -count=1 -short ./...`
- [x] 5.3 Run `go build ./...` and `go vet ./...`
