<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Diff Parsing Package

- [x] 1.1 Create `internal/diff/diff.go` with types: `LineRange{Start, End int}`, `FileChange{Path string, Ranges []LineRange}`, and `Differ` interface with `Diff(ref string) (string, error)` method
- [x] 1.2 Implement `Parse(diffOutput string) []FileChange` function in `internal/diff/diff.go` — parses `git diff --unified=0` output into structured `FileChange` records (extract file paths from `+++ b/` headers, extract line ranges from `@@ ... +N,M @@` hunk headers)
- [x] 1.3 Implement `GitDiffer` struct in `internal/diff/git.go` — production `Differ` implementation that shells out to `git diff <ref> --unified=0` or `git diff --staged --unified=0` when ref is `"staged"`. Includes `exec.LookPath("git")` check with descriptive error
- [x] 1.4 Create `internal/diff/diff_test.go` with table-driven tests for `Parse`: single file/single hunk, multiple files/multiple hunks, empty diff, binary file (no hunks), added file (new file mode), deleted file
- [x] 1.5 Create `internal/diff/git_test.go` with tests for `GitDiffer`: git not found error, invalid ref error, staged vs ref argument routing (use `exec.Command` mock or test helper)

## 2. Provider Extension (EndLine)

- [x] 2.1 Add `EndLine int` field to `FunctionComplexity` struct in `internal/crap/provider.go` with GoDoc comment
- [x] 2.2 Update `goprovider.ComplexityProvider` in `internal/provider/goprovider/complexity.go` to populate `EndLine` from `FuncDecl.End()` position
- [x] 2.3 Update mock provider in `internal/provider/mockprovider/` to populate `EndLine` in synthetic test data
- [x] 2.4 Add/update tests for EndLine population in `goprovider` and `mockprovider`

## 3. Changed Function Identification

- [x] 3.1 Create `internal/crap/gatechange.go` with: `ChangedFunction` struct (embeds `Score` + `Passed bool`), `ChangedFunctionsSummary` struct (`Total`, `Passed`, `Failed int`), `ChangeGateResult` struct (`ChangedFunctions []ChangedFunction`, `Summary ChangedFunctionsSummary`, `Passed bool`)
- [x] 3.2 Implement `FilterChangedFunctions(scores []Score, fileChanges []diff.FileChange) []Score` — intersects diff line ranges with function `File`/`Line`/`EndLine` to return only scores for changed functions. Uses `EndLine` from `FunctionComplexity` (threaded through `Score` via a new `EndLine int` field or via a lookup map)
- [x] 3.3 Implement `EvaluateChangeGate(scores []Score, fileChanges []diff.FileChange, crapThreshold, gazeCRAPThreshold float64) *ChangeGateResult` — filters to changed functions, evaluates each against thresholds, computes summary
- [x] 3.4 Create `internal/crap/gatechange_test.go` with table-driven tests: function contains changed line, function does not, no matching functions, multiple functions in changed file, all passing, one failing, GazeCRAP threshold, no changed functions

## 4. Score EndLine Threading

- [x] 4.1 Add `EndLine int` field to `Score` struct in `internal/crap/crap.go` with JSON tag `json:"end_line"`
- [x] 4.2 Update `computeScores` in `internal/crap/analyze.go` to populate `Score.EndLine` from `FunctionComplexity.EndLine`

## 5. CLI Wiring

- [x] 5.1 Add `gateOnChange string` field to `crapParams` struct in `cmd/gaze/main.go`
- [x] 5.2 Add `--gate-on-change` flag registration in `newCrapCmd()` — `StringVar` with empty default, help text: `"fail if changed functions exceed CRAP threshold (git ref or 'staged')"`
- [x] 5.3 Wire `gateOnChange` from flag to `crapParams` in `newCrapCmd` RunE closure
- [x] 5.4 Update `runCrap` in `cmd/gaze/main.go` to: (a) when `gateOnChange` is non-empty, construct a `GitDiffer`, compute diff, parse it, call `EvaluateChangeGate`, (b) integrate gate result into exit code logic (exit 1 if gate fails), (c) thread `ChangeGateResult` to output writers

## 6. JSON Output

- [x] 6.1 Add `ChangedFunctions []ChangedFunction` and `ChangedFunctionsSummary *ChangedFunctionsSummary` fields to the JSON output struct in `internal/crap/` (or the report output layer)
- [x] 6.2 Update JSON marshaling to include `changed_functions` and `changed_functions_summary` when `--gate-on-change` is active
- [x] 6.3 Update JSON Schema in `internal/report/schema.go` with new `changed_functions` array and `changed_functions_summary` object definitions
- [x] 6.4 Add tests for JSON output with changed functions (verify schema compliance)

## 7. Text Output

- [x] 7.1 Add `WriteChangeGateResult` function in `internal/crap/` (or the report output layer) — writes "Changed Functions" section with per-function pass/fail indicators, CRAP scores, and summary counts
- [x] 7.2 Integrate change gate text output into `runCrap` text format path
- [x] 7.3 Add tests for text output formatting (80-column terminal width constraint)

## 8. Error Handling

- [x] 8.1 Verify git binary availability early in `runCrap` when `gateOnChange` is set — return descriptive error before analysis begins
- [x] 8.2 Handle invalid git ref — surface git stderr in error message
- [x] 8.3 Handle non-git-repository directory — detect and return clear error
- [x] 8.4 Add tests for all error paths (git not found, invalid ref, not a repo)

## 9. Baseline Interaction

- [x] 9.1 Verify `--gate-on-change` and `--baseline` work together — both gates evaluated independently, exit 1 if either fails
- [x] 9.2 Add integration test for combined `--gate-on-change` + `--baseline` flags

## 10. Documentation

- [x] 10.1 Update `README.md` with `--gate-on-change` flag documentation in the `gaze crap` command section
- [x] 10.2 Update or create `docs/reference/cli/crap.md` with flag reference
- [x] 10.3 Update JSON Schema documentation for new `changed_functions` fields
- [x] 10.4 Create website issue in `unbound-force/website` for user-facing documentation update (per Website Documentation Gate in AGENTS.md)

## 11. Constitution Verification

- [x] 11.1 Verify Principle I (Accuracy): changed function identification is correct — no false positives (unchanged functions included) or false negatives (changed functions missed). Automated tests with known fixtures
- [x] 11.2 Verify Principle II (Minimal Assumptions): flag is opt-in, no new mandatory dependencies, git already implicit
- [x] 11.3 Verify Principle III (Actionable Output): JSON and text output identify specific changed functions and their scores
- [x] 11.4 Verify Principle IV (Testability): diff parser and change gate filter are pure functions testable in isolation

<!-- spec-review: passed -->

<!-- code-review: passed -->
