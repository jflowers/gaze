## Why

The Boy Scout Rule demands that code you touch must meet quality standards, even if legacy code does not. Currently `gaze crap` computes CRAP scores for the entire codebase, which means CI gates (`--max-crapload`, `--max-gaze-crapload`) can fail due to pre-existing problems in code the PR did not touch. This creates noise and discourages adoption of CRAP gates in projects with legacy debt.

A `--gate-on-change` flag scopes CRAP enforcement to functions modified in the current changeset, enabling incremental quality improvement: every PR must leave the functions it touches at or above the CRAP threshold, without requiring the entire codebase to comply first.

This is Group 2b from [RFC #483](https://github.com/orgs/unbound-force/discussions/483).

## What Changes

- New `--gate-on-change` flag on `gaze crap` that accepts a git ref or diff specification (e.g., `--gate-on-change=origin/main`, `--gate-on-change=HEAD~3`, `--gate-on-change=staged`).
- New `internal/diff/` package that parses `git diff` output to identify changed files and line ranges.
- New `internal/crap/gatechange.go` that maps changed line ranges to functions, filters the CRAP score set to only changed functions, and evaluates the threshold gate.
- JSON output gains a `changed_functions` section with per-function CRAP/GazeCRAP scores and pass/fail status.
- Text output reports which changed functions failed and their scores.
- Exit code 1 when any changed function exceeds the CRAP threshold.

## Capabilities

### New Capabilities
- `gate-on-change`: Git-diff-scoped CRAP gate that identifies functions modified in a changeset and fails CI if any exceed the CRAP threshold.

### Modified Capabilities
- `gaze crap` CLI: New `--gate-on-change` flag added; output extended with `changed_functions` section in JSON and text reports.

## Impact

- **CLI**: New flag on `gaze crap` command. No changes to existing flags or default behavior.
- **Packages**: New `internal/diff/` package (git diff parsing); new `internal/crap/gatechange.go` (changed-function filtering and gate evaluation).
- **Output**: JSON schema extended with optional `changed_functions` array. Text report extended with a "Changed Functions" section when flag is active.
- **CI**: Enables per-PR CRAP gating without requiring full-codebase compliance.
- **Dependencies**: Requires `git` binary on PATH (already an implicit dependency for coverage via `go test`).

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: PASS

The gate-on-change feature reports CRAP scores for functions identified via git diff. It does not alter the CRAP computation itself — it only filters which functions are evaluated. The mapping from diff line ranges to functions must be accurate (no false positives: functions not actually changed must not be included; no false negatives: all functions with changed lines must be included). This will be verified by automated tests with known diff fixtures.

### II. Minimal Assumptions

**Assessment**: PASS

The feature assumes only that `git` is available on PATH (already required for `go test` coverage). It does not require users to annotate code, restructure projects, or configure additional files. The flag is opt-in — existing behavior is unchanged when the flag is omitted.

### III. Actionable Output

**Assessment**: PASS

The output identifies exactly which changed functions failed the CRAP threshold and their scores. JSON output includes machine-parseable `changed_functions` data for CI integration. This directly tells developers which functions to improve — more actionable than a whole-codebase gate failure that may point to unrelated legacy code.

### IV. Testability

**Assessment**: PASS

The diff parser (`internal/diff/`) is a pure function: string input (diff output) → structured output (changed files and line ranges). The changed-function filter is a pure function: (diff ranges, CRAP scores) → filtered scores. Both are testable in isolation with synthetic fixtures. The git invocation is isolated behind an interface for dependency injection.
