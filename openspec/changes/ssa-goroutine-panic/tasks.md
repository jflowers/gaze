## 1. Production Fix

- [x] 1.1 In `BuildSSA` (`internal/analysis/mutation.go`), change the `ssautil.AllPackages` call from `ssa.InstantiateGenerics` to `ssa.InstantiateGenerics | ssa.BuildSerially`.
- [x] 1.2 In `BuildTestSSA` (`internal/quality/pairing.go`), change the `ssautil.AllPackages` call from `ssa.InstantiateGenerics` to `ssa.InstantiateGenerics | ssa.BuildSerially`.
- [x] 1.3 Update GoDoc comments on both functions to note that `BuildSerially` is required for `recover()` to work across goroutine boundaries.

## 2. Tests

- [x] 2.1 Verify existing `safeSSABuild` tests still pass (they test the `recover()` pattern in isolation, unaffected by `BuildSerially`).
- [x] 2.2 Verify existing `TestAssess_SSADegraded` and `TestAssess_SSADegraded_NilStderr` still pass (they use `BuildSSAFunc` injection, unaffected by the mode change).
- [x] 2.3 Verify existing `TestAssess_SSASuccess_NotDegraded` still passes (confirms `BuildSerially` does not break normal SSA construction).
- [x] 2.4 Run full test suite: `go test -race -count=1 -short ./...`

## 3. Documentation

- [x] 3.1 Update `specs/021-ssa-panic-recovery/research.md` R2 with a correction note: `prog.Build()` spawns goroutines by default, `ssa.BuildSerially` is required for `recover()` to work.
- [x] 3.2 Update `AGENTS.md` Recent Changes with a summary of this fix.

## 4. Verification

- [x] 4.1 Run `go build ./...` to verify compilation.
- [x] 4.2 Run `go vet ./...` to verify no vet issues.
- [x] 4.3 Verify constitution alignment: (I) Autonomous Collaboration — no artifact changes. (II) Composability — same API surface. (III) Observable Quality — degradation path now actually triggers. (IV) Testability — existing tests unchanged.
