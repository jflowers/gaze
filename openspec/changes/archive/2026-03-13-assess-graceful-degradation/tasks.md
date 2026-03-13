## 1. Data Model Changes

- [x] 1.1 Add `SSADegraded bool` field with JSON tag `ssa_degraded` to `taxonomy.PackageSummary` in `internal/taxonomy/types.go`. Use `json:"ssa_degraded"` (not `omitempty` — always serialize so consumers can rely on its presence).
- [x] 1.2 Add `ssa_degraded` as an optional `boolean` property to the `PackageSummary` definition in `QualitySchema` (`internal/report/schema.go`). Do NOT add it to the `required` array.

## 2. Core: Refactor quality.Assess

- [x] 2.1 In `quality.Assess` (`internal/quality/quality.go`), replace the hard-failure `BuildTestSSA` error return with a degradation path. When `BuildTestSSA` returns an error: (a) emit a warning to `opts.Stderr` if non-nil, (b) set a local `ssaDegraded` flag, (c) continue instead of returning.
- [x] 2.2 Add a degraded code path after the SSA check. When `ssaDegraded` is true, iterate over `testFuncs`, call `DetectAssertions` for each, compute `AssertionDetectionConfidence`, and build a `QualityReport` with zero-valued `TargetFunction`, `ContractCoverage`, `OverSpecification`, and nil `UnmappedAssertions`/`AmbiguousEffects`. Append to `reports`.
- [x] 2.3 After building reports (both normal and degraded paths), set `summary.SSADegraded = ssaDegraded` on the `PackageSummary` returned by `BuildPackageSummary`. Return `(reports, summary, nil)`.

## 3. Caller Updates

- [x] 3.1 Update `runQuality` in `cmd/gaze/main.go` (~line 877). Since `Assess` no longer returns an error on SSA failure, the existing `if err != nil` block now only fires on genuine errors (nil test package). No change needed to the error handling, but verify the degraded results flow through to `WriteJSON`/`WriteText` correctly. The warning on stderr is already handled by `Assess`.
- [x] 3.2 Update `analyzePackageCoverage` in `cmd/gaze/main.go` (~line 593). Since `Assess` returns non-nil degraded reports instead of an error, the `if err != nil { return nil }` path is only for genuine errors. Degraded reports will flow through naturally. Add a debug log noting when returned reports have degraded summary.
- [x] 3.3 Update `runQualityForPackage` in `internal/aireport/runner_steps.go` (~line 139). Same pattern — degraded reports now flow through instead of being lost to the nil return path.

## 4. Tests

- [x] 4.1 Add `TestAssess_SSADegraded` to `internal/quality/quality_test.go`. Construct a test scenario where `BuildTestSSA` would fail (use a minimal loaded package or mock the failure path), verify `Assess` returns non-nil reports with `SSADegraded: true`, `TotalTests > 0`, and nil error. Verify the warning is written to stderr.
- [x] 4.2 Add `TestAssess_SSASuccess_NotDegraded` to verify `SSADegraded` is `false` when SSA succeeds. This can use an existing test fixture that passes SSA build.
- [x] 4.3 Update `TestQualitySchema_ValidatesSampleOutput` in `internal/report/report_test.go` to include `ssa_degraded` in the sample JSON and verify it validates against the updated schema.
- [x] 4.4 Add a schema validation test with `ssa_degraded: true` to verify the schema accepts degraded output.

## 5. Documentation

- [x] 5.1 Update `AGENTS.md` Recent Changes section with a summary of this change.
- [x] 5.2 Update `AGENTS.md` Active Technologies section if any new technologies are introduced (likely N/A — no new dependencies).
- [x] 5.3 Update GoDoc comments on `Assess`, `PackageSummary`, and any modified caller functions to reflect the new degradation behavior.

## 6. Verification

- [x] 6.1 Run `go test -race -count=1 -short ./...` and verify all tests pass.
- [x] 6.2 Run `golangci-lint run` and verify no lint errors.
- [x] 6.3 Verify constitution alignment: (I) Autonomous Collaboration — outputs remain self-describing with `SSADegraded` field. (II) Composability — `Assess` signature unchanged, additive field. (III) Observable Quality — JSON schema updated, machine-parseable degradation indicator. (IV) Testability — degraded path tested in isolation.
