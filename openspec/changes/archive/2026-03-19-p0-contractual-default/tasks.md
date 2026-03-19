# Tasks: P0 Contractual Default

## 1. Add tier boost function

- [x] 1.1 Add `tierBoost(effectType taxonomy.SideEffectType) int` function to `internal/classify/score.go`. Returns +25 for P0, +10 for P1, 0 for P2-P4. Uses `taxonomy.TierOf(effectType)` to determine tier.
- [x] 1.2 Add `TestTierBoost` table-driven test in `internal/classify/score_test.go` covering all five P0 types (ReturnValue, ErrorReturn, SentinelError, ReceiverMutation, PointerArgMutation → 25), representative P1 types (WriterOutput, SliceMutation → 10), and representative P2+ types (→ 0).

## 2. Wire tier boost into scoring pipeline

- [x] 2.1 Change `accumulateSignals` signature in `internal/classify/score.go` from `(signals []taxonomy.Signal)` to `(effectType taxonomy.SideEffectType, signals []taxonomy.Signal)`. Add `score = baseConfidence + tierBoost(effectType)` at the start (replacing `score = baseConfidence`).
- [x] 2.2 Update the `accumulateSignals` caller in `ComputeScore` (`internal/classify/score.go`) to pass `se.Type` as the first argument: `score, hasPositive, hasNegative := accumulateSignals(se.Type, signals)`.
- [x] 2.3 Update `TestAccumulateSignals` tests in `internal/classify/score_test.go` to pass an `effectType` parameter and adjust expected scores (e.g., P0 base is now 75, P1 is 60, P2+ is 50).

## 3. Update existing classifier tests

- [x] 3.1 Update `TestComputeScore` and any other tests in `internal/classify/score_test.go` that hardcode expected confidence values. P0 effects should now score 25 points higher than before. Review each test case and adjust the expected `Confidence` field.
- [x] 3.2 Update integration tests in `internal/classify/classify_test.go` if any assert on specific confidence values for P0 effects. Search for assertions on `Classification.Confidence` and adjust.
- [x] 3.3 Search for classifier confidence assertions in `cmd/gaze/main_test.go` and `internal/quality/` tests. Update any hardcoded confidence expectations.

## 4. Update downstream expectations

- [x] 4.1 Run `go test -race -count=1 -short ./internal/classify/...` and fix any remaining test failures from the confidence shift.
- [x] 4.2 Run `go test -race -count=1 -short ./...` to check for cascading test failures in quality, crap, and report packages. Fix expected values where P0 confidence has changed.
- [x] 4.3 Check if `--max-gaze-crapload=5` and `--max-crapload=35` in `.github/workflows/test.yml` still pass with the new classification. The change should make scoring better (more contractual effects → higher contract coverage → lower GazeCRAPload), so thresholds should still pass. Verify by running `gaze report --format=json --coverprofile=<coverage.out> ./...` locally and checking the new GazeCRAPload value.

## 5. Verification

- [x] 5.1 Run `go build ./cmd/gaze && go vet ./...` to verify compilation.
- [x] 5.2 Run `go test -race -count=1 -short ./internal/classify/... ./internal/quality/... ./internal/crap/... ./cmd/gaze/...` to verify all tests pass.
- [x] 5.3 Run `gaze analyze --classify ./internal/config` and verify P0 effects (ReturnValue, ErrorReturn) now show confidence >= 75 (not 50).
