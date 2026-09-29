## ADDED Requirements

### Requirement: Gate-on-Change Flag

The `gaze crap` command SHALL accept a `--gate-on-change` flag that accepts a string value specifying a git ref or `staged`.

#### Scenario: Flag accepts a git ref

- **GIVEN** the user runs `gaze crap --gate-on-change=origin/main ./...`
- **WHEN** the flag is parsed
- **THEN** gaze SHALL compute a diff between the working tree and `origin/main` using `git diff origin/main --unified=0`

#### Scenario: Flag accepts staged

- **GIVEN** the user runs `gaze crap --gate-on-change=staged ./...`
- **WHEN** the flag is parsed
- **THEN** gaze SHALL compute a diff between the index and HEAD using `git diff --staged --unified=0`

#### Scenario: Flag not provided

- **GIVEN** the user runs `gaze crap ./...` without `--gate-on-change`
- **WHEN** the command executes
- **THEN** gaze SHALL behave exactly as before (no diff computation, no changed-function filtering)

---

### Requirement: Diff Parsing

The system SHALL parse `git diff --unified=0` output into structured file changes, each containing a file path and a list of changed line ranges.

#### Scenario: Single file, single hunk

- **GIVEN** a unified diff with one file and one hunk changing lines 10-15
- **WHEN** the diff is parsed
- **THEN** the result SHALL contain one `FileChange` with path matching the diff header and one line range `{start: 10, end: 15}`

#### Scenario: Multiple files, multiple hunks

- **GIVEN** a unified diff with two files, each with multiple hunks
- **WHEN** the diff is parsed
- **THEN** the result SHALL contain two `FileChange` records, each with the correct line ranges from their respective hunks

#### Scenario: Empty diff

- **GIVEN** a unified diff with no changes (empty output)
- **WHEN** the diff is parsed
- **THEN** the result SHALL be an empty list of file changes

#### Scenario: Binary file in diff

- **GIVEN** a unified diff that includes a binary file change (no line hunks)
- **WHEN** the diff is parsed
- **THEN** the binary file SHALL be included in the result with an empty line range list (no functions will match)

---

### Requirement: Changed Function Identification

The system SHALL identify which functions were modified by intersecting changed line ranges with function AST position data.

#### Scenario: Function contains a changed line

- **GIVEN** a function `Foo` spanning lines 10-25 in `pkg/foo.go` and a diff changing line 15 in `pkg/foo.go`
- **WHEN** changed functions are identified
- **THEN** `Foo` SHALL be included in the changed function set

#### Scenario: Function does not contain any changed line

- **GIVEN** a function `Bar` spanning lines 30-50 in `pkg/foo.go` and a diff changing only line 15 in `pkg/foo.go`
- **WHEN** changed functions are identified
- **THEN** `Bar` SHALL NOT be included in the changed function set

#### Scenario: Changed file has no matching functions in CRAP results

- **GIVEN** a diff changing lines in `pkg/generated.go` and no CRAP scores for functions in that file
- **WHEN** changed functions are identified
- **THEN** the result SHALL be an empty changed function set (no error)

#### Scenario: Multiple functions in a changed file

- **GIVEN** functions `A` (lines 1-10), `B` (lines 11-20), `C` (lines 21-30) and a diff changing lines 5 and 25
- **WHEN** changed functions are identified
- **THEN** the changed function set SHALL include `A` and `C` but NOT `B`

---

### Requirement: Gate Evaluation

The system SHALL evaluate the CRAP threshold gate against only the changed functions when `--gate-on-change` is active.

#### Scenario: All changed functions below threshold

- **GIVEN** `--gate-on-change=origin/main` and `--crap-threshold=15`
- **AND** all changed functions have CRAP scores below 15
- **WHEN** the gate is evaluated
- **THEN** the command SHALL exit 0 (success)

#### Scenario: One changed function exceeds threshold

- **GIVEN** `--gate-on-change=origin/main` and `--crap-threshold=15`
- **AND** one changed function has CRAP score 25
- **WHEN** the gate is evaluated
- **THEN** the command SHALL exit 1 (failure)

#### Scenario: No changed functions detected

- **GIVEN** `--gate-on-change=origin/main`
- **AND** the diff contains no changes to Go source files (or no functions match)
- **WHEN** the gate is evaluated
- **THEN** the command SHALL exit 0 with a message indicating no changed functions were found

#### Scenario: GazeCRAP threshold also evaluated

- **GIVEN** `--gate-on-change=origin/main` and `--gaze-crap-threshold=15`
- **AND** a changed function has GazeCRAP score 20
- **WHEN** the gate is evaluated
- **THEN** the command SHALL exit 1 (failure) — GazeCRAP threshold is also enforced for changed functions

---

### Requirement: JSON Output for Changed Functions

The JSON output SHALL include a `changed_functions` section when `--gate-on-change` is active.

#### Scenario: JSON output with changed functions

- **GIVEN** `--gate-on-change=origin/main --format=json`
- **AND** three functions were changed, two passing and one failing
- **WHEN** JSON output is produced
- **THEN** the output SHALL include a `changed_functions` array with three entries, each containing the function's `Score` data and a `passed` boolean
- **AND** the output SHALL include a `changed_functions_summary` object with `total: 3`, `passed: 2`, `failed: 1`

#### Scenario: JSON output with no changed functions

- **GIVEN** `--gate-on-change=origin/main --format=json`
- **AND** no changed functions were detected
- **WHEN** JSON output is produced
- **THEN** the output SHALL include `changed_functions: []` and `changed_functions_summary: {total: 0, passed: 0, failed: 0}`

---

### Requirement: Text Output for Changed Functions

The text output SHALL include a "Changed Functions" section when `--gate-on-change` is active and changed functions are detected.

#### Scenario: Text output with failing changed functions

- **GIVEN** `--gate-on-change=origin/main` (text format)
- **AND** one changed function `(*Store).Save` has CRAP 25 (threshold 15)
- **WHEN** text output is produced
- **THEN** the output SHALL include a "Changed Functions" section listing `(*Store).Save` with its CRAP score and a failure indicator

#### Scenario: Text output with all passing changed functions

- **GIVEN** `--gate-on-change=origin/main` (text format)
- **AND** all changed functions have CRAP below threshold
- **WHEN** text output is produced
- **THEN** the output SHALL include a "Changed Functions" section indicating all changed functions passed

---

### Requirement: Git Error Handling

The system SHALL handle git invocation errors gracefully.

#### Scenario: Git binary not found

- **GIVEN** `--gate-on-change=origin/main`
- **AND** `git` is not on PATH
- **WHEN** the command attempts to compute the diff
- **THEN** the command SHALL exit 1 with a clear error message: `git binary not found: --gate-on-change requires git`

#### Scenario: Invalid git ref

- **GIVEN** `--gate-on-change=nonexistent-branch`
- **WHEN** the command attempts to compute the diff
- **THEN** the command SHALL exit 1 with an error message including the git stderr output

#### Scenario: Not in a git repository

- **GIVEN** `--gate-on-change=origin/main`
- **AND** the working directory is not inside a git repository
- **WHEN** the command attempts to compute the diff
- **THEN** the command SHALL exit 1 with a clear error message indicating the directory is not a git repository

---

### Requirement: Function End-Line Tracking

The `FunctionComplexity` struct SHALL include an `EndLine` field populated from AST data.

#### Scenario: Go provider populates EndLine

- **GIVEN** a Go function `Foo` declared at line 10 and ending at line 25
- **WHEN** the complexity provider analyzes the package
- **THEN** the returned `FunctionComplexity` for `Foo` SHALL have `StartLine: 10` and `EndLine: 25`

---

## MODIFIED Requirements

### Requirement: CRAP Command Exit Code

The `gaze crap` command SHALL evaluate the `--gate-on-change` gate in addition to existing gates (`--max-crapload`, `--max-gaze-crapload`, `--baseline`) when the flag is provided.

Previously: Exit code was determined by `--max-crapload`, `--max-gaze-crapload`, and `--baseline` gates only.

New text: Exit code is determined by `--max-crapload`, `--max-gaze-crapload`, `--baseline`, and `--gate-on-change` gates. When multiple gates are active, exit 1 if any gate fails.
