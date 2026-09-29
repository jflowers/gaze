## Context

`gaze crap` computes CRAP scores for all functions in the analyzed packages. CI gates (`--max-crapload`, `--max-gaze-crapload`) evaluate the entire codebase. This means a PR that improves one function can still fail CI because of pre-existing CRAP violations in unrelated legacy code.

The existing `--baseline` flag (spec 009) compares current scores against a saved baseline JSON file, detecting regressions and new-function violations. It operates on the full score set — it does not scope to changed functions.

The `--gate-on-change` flag is complementary: it scopes CRAP evaluation to functions modified in the current changeset, enabling the Boy Scout Rule — code you touch must meet quality standards — without requiring full-codebase compliance.

Constitution alignment is documented in the proposal. All four principles PASS.

## Goals / Non-Goals

### Goals
- Add `--gate-on-change` flag to `gaze crap` that accepts a git ref, diff range, or `staged`
- Parse `git diff` output to identify changed files and line ranges
- Map changed line ranges to function declarations using Go AST position data
- Filter CRAP scores to only changed functions
- Fail CI (exit 1) when any changed function exceeds the CRAP threshold
- Produce JSON output with a `changed_functions` section
- Produce text output listing changed functions and their pass/fail status
- Support both staged changes (`git diff --staged`) and branch comparisons (`git diff <ref>`)

### Non-Goals
- Computing CRAP scores for only the changed lines (CRAP is per-function, not per-line)
- Modifying the existing `--baseline` comparison flow — this is a separate flag with separate semantics
- Supporting non-git VCS (hg, svn) — git is the implicit VCS assumption
- Supporting unified diff format from stdin — the flag always invokes git internally
- Integration with `gaze quality` or `gaze report` — this is a `gaze crap` flag only (Group 2b scope)
- Handling merge commits or rebased histories specially — standard `git diff` semantics apply

## Decisions

### D1: New `internal/diff/` package for git diff parsing

**Decision**: Create a new `internal/diff/` package with a `Parse` function that takes raw `git diff --unified=0` output and returns structured `FileChange` records (file path + list of changed line ranges).

**Rationale**: Separating diff parsing from CRAP logic keeps both testable in isolation. The diff parser is a pure function (string → structs) with no I/O. The git invocation is isolated behind a `Differ` interface for dependency injection in tests.

**Alternative considered**: Using a third-party diff library (e.g., `sergi/go-diff`). Rejected because `git diff --unified=0` output is simple and well-defined — parsing it avoids a new dependency and keeps the tool self-contained.

### D2: Function mapping via AST position intersection

**Decision**: Map changed line ranges to functions by comparing diff line ranges against each function's AST-declared line range (start line to end line). A function is "changed" if any changed line in its file falls within its line range.

**Rationale**: The `crap.Score` struct already has `File` and `Line` fields. The complexity provider (`goprovider.ComplexityProvider`) returns per-function complexity with file and line data. We need to extend this with function end-line information. The `go/ast` package provides `FuncDecl.End()` for this purpose.

**Alternative considered**: Using `git diff` hunk headers to extract function names via regex. Rejected because Go method names can be ambiguous (e.g., `(*Store).Save` vs `Save`), and AST-based mapping is already available through the existing provider infrastructure.

### D3: Flag value semantics

**Decision**: `--gate-on-change` accepts a string value:
- `"staged"` — runs `git diff --staged --unified=0` (compares index to HEAD)
- Any other value — treated as a git ref, runs `git diff <ref> --unified=0` (compares working tree to ref)
- Empty/unset — flag not active, existing behavior unchanged

**Rationale**: The two most common CI use cases are: (1) checking staged changes before commit, and (2) checking all changes in a PR branch against the base branch. A single string flag covers both without requiring separate flags.

### D4: Gate evaluation uses existing CRAP threshold

**Decision**: The gate evaluates changed functions against the existing `--crap-threshold` (default 15). No separate threshold flag is introduced.

**Rationale**: The Boy Scout Rule says "code you touch must meet quality standards" — the same standards as the rest of the codebase. A separate threshold would create confusion about which standard applies. Users who want a different threshold for changed functions can set `--crap-threshold` globally.

### D5: JSON output structure

**Decision**: JSON output gains an optional `changed_functions` array at the top level (alongside `scores` and `summary`). Each entry includes the function's `Score` data plus a `passed` boolean. A `changed_functions_summary` object reports `total`, `passed`, and `failed` counts.

**Rationale**: Machine-parseable output is required for CI integration. The structure mirrors the existing `scores` array format for consistency, with the addition of pass/fail status per function.

### D6: Git invocation behind an interface

**Decision**: Define a `Differ` interface in `internal/diff/`:
```go
type Differ interface {
    Diff(ref string) (string, error)
}
```
Production implementation shells out to `git`. Tests inject a fake that returns canned diff output.

**Rationale**: Follows the existing provider interface pattern (e.g., `ComplexityProvider`, `LineCoverageProvider`). Enables unit testing without requiring git on PATH or a real repository.

### D7: End-line tracking in complexity provider

**Decision**: Extend `FunctionComplexity` (in `internal/crap/provider.go`) with an `EndLine int` field. The Go provider (`goprovider.ComplexityProvider`) populates it from AST data. The mock provider already returns synthetic data.

**Rationale**: Function-to-line-range mapping requires knowing where a function ends, not just where it starts. The `go/ast` `FuncDecl.End()` method provides this. Adding it to the provider interface keeps the abstraction clean.

### D8: Interaction with `--baseline`

**Decision**: `--gate-on-change` and `--baseline` are independent flags. When both are set, both gates are evaluated. Exit 1 if either fails. The JSON output includes both `comparison` (from baseline) and `changed_functions` (from gate-on-change) sections.

**Rationale**: The two flags serve different purposes — baseline detects regressions, gate-on-change enforces the Boy Scout Rule. They are complementary, not mutually exclusive.

## Risks / Trade-offs

### Risk: Function end-line accuracy
The `go/ast` `FuncDecl.End()` returns the end of the function declaration, which includes the full function body. This is correct for our purposes — any line within the function body counts as "changed" if the diff touches it. However, if a function is very long and only one line is changed, the entire function is flagged. This is by design (CRAP is per-function), but may produce surprising results for very long functions.

**Mitigation**: Document this behavior clearly. The text output shows the function's full line range so users understand why it was flagged.

### Risk: Git availability
The feature requires `git` on PATH. This is already an implicit dependency (coverage generation uses `go test`, which requires git for module resolution in most setups). If git is unavailable, the flag produces a clear error message.

**Mitigation**: Check for git binary early and return a descriptive error before attempting diff parsing.

### Trade-off: No per-line CRAP
CRAP scores are inherently per-function. We cannot compute "CRAP for the changed lines only." This means a function with 1000 lines and 1 changed line is evaluated as a whole. This is acceptable because:
1. The Boy Scout Rule applies to the function, not the line
2. Per-line CRAP would require a fundamentally different scoring model
3. The existing `--baseline` flag provides finer-grained regression detection for users who need it

### Trade-off: No support for non-git VCS
The feature is git-only. This matches the project's existing implicit assumptions (Go modules, GitHub CI). Supporting other VCS would require abstracting the diff source, which adds complexity without clear user demand.
