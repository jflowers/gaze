## 1. Data Model

- [ ] 1.1 Add `ContractCoverageInfo` struct to `internal/crap/crap.go` with `Percentage`, `Reason`, `MinConfidence`, `MaxConfidence`.
- [ ] 1.2 Change `ContractCoverageFunc` type from `func(string, string) (float64, bool)` to `func(string, string) (ContractCoverageInfo, bool)`.
- [ ] 1.3 Add `ContractCoverageReason *string` with JSON tag `contract_coverage_reason,omitempty` to `Score`.
- [ ] 1.4 Add `EffectConfidenceRange *[2]int` with JSON tag `effect_confidence_range,omitempty` to `Score`.

## 2. Computation

- [ ] 2.1 Update `computeScores` to use `ContractCoverageInfo` from the callback and populate `ContractCoverageReason` and `EffectConfidenceRange` on each `Score`.
- [ ] 2.2 Update `buildContractCoverageFunc` in `cmd/gaze/main.go` to compute the reason by examining classification labels and confidence values from quality reports. Store `ContractCoverageInfo` in the coverage map instead of bare `float64`.
- [ ] 2.3 Update all callers of `ContractCoverageFunc` that use the old signature (test mocks in `crap_test.go` and `analyze_internal_test.go`).

## 3. Text Report

- [ ] 3.1 Update `writeWorstSection` in `internal/crap/report.go` to display the reason and confidence range after the `[fix_strategy]` label.

## 4. Tests

- [ ] 4.1 Add `TestComputeScores_CoverageReason_AllAmbiguous` — verify reason is set when callback returns `all_effects_ambiguous`.
- [ ] 4.2 Add `TestComputeScores_CoverageReason_Normal` — verify reason is nil for normal coverage.
- [ ] 4.3 Update existing `TestComputeScores_GazeCRAP` and related tests to use the new callback signature.
- [ ] 4.4 Run full test suite.

## 5. Documentation & Verification

- [ ] 5.1 Update `AGENTS.md` Recent Changes.
- [ ] 5.2 Run `go build ./...` and `go vet ./...`.
