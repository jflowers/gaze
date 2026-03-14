## 1. Extract computeScores

- [x] 1.1 Create `computeScores` function in `internal/crap/analyze.go` with signature `func computeScores(stats []gocyclo.Stat, coverMap map[string]float64, opts Options) []Score`. Move the Step 5 loop body (lines 104-157) into this function, including the `generatedCache` map and all branching logic.
- [x] 1.2 Replace the Step 5 loop in `Analyze` with a single call to `computeScores(complexityStats, coverMap, opts)`. Assign result to `scores`.
- [x] 1.3 Verify `Analyze`'s cyclomatic complexity is reduced (target: 7 or below). Run `gocyclo -over 7 internal/crap/analyze.go` to confirm.

## 2. Tests

- [x] 2.1 Create `internal/crap/analyze_internal_test.go` with `package crap` for internal tests of `computeScores`.
- [x] 2.2 Add `TestComputeScores_BasicCRAP` — verify CRAP score computation with known complexity and coverage values.
- [x] 2.3 Add `TestComputeScores_SkipsTestFiles` — verify `_test.go` entries are excluded from results.
- [x] 2.4 Add `TestComputeScores_SkipsGeneratedFiles` — verify generated files are excluded when `IgnoreGenerated` is true, and included when false.
- [x] 2.5 Add `TestComputeScores_ZeroCoverage` — verify functions not in the coverage map get `LineCoverage: 0.0`.
- [x] 2.6 Add `TestComputeScores_GazeCRAP` — verify GazeCRAP is computed when `ContractCoverageFunc` is set and returns `(value, true)`.
- [x] 2.7 Add `TestComputeScores_NoGazeCRAP` — verify GazeCRAP fields are nil when `ContractCoverageFunc` is nil.
- [x] 2.8 Add `TestComputeScores_GazeCRAPNotFound` — verify GazeCRAP fields are nil when `ContractCoverageFunc` returns `(0, false)`.
- [x] 2.9 Run existing `crap_test.go` tests to verify no regressions: `go test -race -count=1 ./internal/crap/`

## 3. Documentation

- [x] 3.1 Add GoDoc comment on `computeScores` explaining its role in the pipeline.
- [x] 3.2 Update `AGENTS.md` Recent Changes with a summary of this change.

## 4. Verification

- [x] 4.1 Run `go test -race -count=1 -short ./...` and verify all tests pass.
- [x] 4.2 Run `go build ./...` and `go vet ./...` to verify no issues.
- [x] 4.3 Verify constitution alignment: (I) Autonomous Collaboration — no artifact changes. (II) Composability — same API. (III) Observable Quality — no output changes. (IV) Testability — `computeScores` directly testable in isolation.
