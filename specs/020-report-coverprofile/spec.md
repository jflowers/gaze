# Feature Specification: Pass Pre-Generated Coverage Profile to gaze report

**Feature Branch**: `020-report-coverprofile`
**Created**: 2026-03-12
**Status**: Draft
**Input**: User description: "Option A — add --coverprofile to gaze report. Add a --coverprofile flag to the gaze report subcommand so users can pass a pre-generated coverage profile instead of having gaze run go test internally. This eliminates the double test run in CI."

## Overview

Go developers running `gaze report` in CI currently experience two full test suite executions per job: one from their existing test step (with race detection and verbose output), and a second triggered internally by `gaze report` to collect coverage data. This doubles CI latency, weakens the coverage signal (the internal run uses `-short` and no race detector), and is invisible to the user. This feature adds a `--coverprofile` flag to `gaze report` so users can supply the profile already produced by their test step.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Supply existing coverage profile to skip internal test run (Priority: P1)

A developer has a CI workflow that already runs tests with race detection and generates a coverage profile. They want `gaze report` to use that profile instead of running tests again. They pass `--coverprofile=coverage.out` to `gaze report` and the report is generated without a second test execution.

**Why this priority**: This is the entire motivation for the feature. Eliminating the double test run is the primary user value.

**Independent Test**: Run `gaze report ./... --coverprofile=coverage.out` with a pre-generated profile. Verify no internal test subprocess is spawned, the report is produced, and it reflects the coverage data in the supplied file.

**Acceptance Scenarios**:

1. **Given** a valid coverage profile at `coverage.out`, **When** the user runs `gaze report ./... --ai=claude --coverprofile=coverage.out`, **Then** gaze produces a complete report using the supplied profile without spawning any test subprocess internally.
2. **Given** a valid coverage profile, **When** the user runs `gaze report` with `--coverprofile` and `--format=json`, **Then** the JSON output contains CRAP scores computed from the supplied profile.
3. **Given** no `--coverprofile` flag, **When** the user runs `gaze report ./...`, **Then** existing behavior is preserved: gaze generates coverage internally via its own test run.

---

### User Story 2 - Receive a clear error when the supplied profile path is invalid (Priority: P2)

A developer accidentally passes a nonexistent or unreadable path to `--coverprofile`. They need a clear, actionable error message rather than a silent failure or cryptic internal error.

**Why this priority**: Usability gate. If bad input is silently mishandled, users lose trust in the flag.

**Independent Test**: Run `gaze report ./... --coverprofile=/nonexistent/coverage.out`. Verify the error references the bad path and distinguishes the problem from an analysis failure.

**Acceptance Scenarios**:

1. **Given** `--coverprofile` points to a file that does not exist, **When** the user runs `gaze report`, **Then** gaze exits non-zero with an error message that identifies the path and states the file was not found.
2. **Given** `--coverprofile` points to a directory, **When** the user runs `gaze report`, **Then** gaze exits non-zero with a descriptive error that identifies the path is not a regular file.
3. **Given** `--coverprofile` points to a file with invalid coverage profile content, **When** the user runs `gaze report`, **Then** gaze exits non-zero with an error message referencing the parse failure.

---

### User Story 3 - Flag is self-documenting so no external docs are needed to use it (Priority: P3)

A developer integrating `gaze report` into an existing CI workflow can discover and use `--coverprofile` from `--help` alone, without reading a separate guide.

**Why this priority**: Adoption depends on discoverability. If the flag is not clearly described in the help output, users will not find or trust it.

**Independent Test**: Run `gaze report --help` and read only the flag descriptions. Verify a developer can infer the correct invocation without consulting additional documentation.

**Acceptance Scenarios**:

1. **Given** a user runs `gaze report --help`, **When** they look for coverage-related flags, **Then** `--coverprofile` is listed with a description that explains it accepts a pre-generated profile and bypasses the internal test run.
2. **Given** a user reads the README, **When** they look for a CI integration example, **Then** they find an example showing `go test -coverprofile=coverage.out` followed by `gaze report --coverprofile=coverage.out`.

---

### Edge Cases

- What happens when the coverage profile was generated for a different set of packages than the `<patterns>` argument? Missing packages are treated as 0% covered — consistent with current behavior when tests produce no coverage for a package.
- What happens when the profile is empty (zero bytes or no coverage records)? Gaze emits a clear error rather than silently producing a report with all functions at 0%.
- What happens when `--coverprofile` is a symlink? Gaze follows the symlink to the target file (standard file I/O behavior).
- What happens when both `--coverprofile` and `--format=json` are used together? The flag must work for both output formats.
- What happens when the profile was generated by a different version of the Go toolchain? Gaze uses the standard Go coverage profile parser, which is forward-compatible within the standard format.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `gaze report` MUST accept a `--coverprofile <path>` flag designating a pre-generated Go coverage profile file.
- **FR-002**: When `--coverprofile` is supplied, `gaze report` MUST NOT spawn `go test` internally for coverage collection; it MUST use the supplied file directly for CRAP analysis.
- **FR-003**: When `--coverprofile` is omitted, `gaze report` MUST behave exactly as before — generate coverage data internally via its own subprocess and clean up any temporary files after the run.
- **FR-004**: When `--coverprofile` points to a nonexistent file, `gaze report` MUST exit non-zero with an error message that identifies the path and states the file was not found.
- **FR-005**: When `--coverprofile` points to a path that is not a regular file (e.g., a directory), `gaze report` MUST exit non-zero with a descriptive error message identifying the problem.
- **FR-006**: When `--coverprofile` points to a file with invalid or unparseable coverage profile content, `gaze report` MUST exit non-zero with an error message referencing the parse failure.
- **FR-007**: The `--coverprofile` flag MUST appear in `gaze report --help` output with a description that explains it accepts a pre-generated profile and bypasses the internal test run.
- **FR-008**: The flag MUST work for both `--format=text` and `--format=json` output modes.
- **FR-009**: The README MUST include a CI example showing `go test -coverprofile=coverage.out` followed by `gaze report --coverprofile=coverage.out`.

### Key Entities

- **Coverage Profile**: A file in Go's standard `go test -coverprofile` format. Contains per-statement coverage data keyed by source file and line ranges.
- **CRAP Analysis Step**: The internal pipeline step within `gaze report` that parses the coverage profile and combines it with cyclomatic complexity to compute per-function CRAP scores.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A CI workflow using `--coverprofile` requires exactly one test execution per run, not two.
- **SC-002**: `gaze report --help` lists `--coverprofile` with a description that is sufficient for a developer to use the flag correctly without reading external documentation.
- **SC-003**: All three invalid-profile scenarios (nonexistent path, not-a-file, unparseable content) produce distinct, actionable error messages with non-zero exit codes.
- **SC-004**: When a valid pre-generated profile is supplied, CRAP scores in the report are identical to those produced by running `gaze report` without the flag on the same codebase with equivalent test coverage.
- **SC-005**: Omitting `--coverprofile` produces behavior identical to the current release — no regression for existing users.
- **SC-006**: The README CI example is complete and runnable: a developer can copy it into a GitHub Actions workflow step without modification.

## Assumptions

- The coverage profile format is the standard Go `go test -coverprofile` format. No other coverage formats (LCOV, Cobertura, etc.) are in scope.
- The path validation logic (existence check, is-regular-file check) mirrors what `gaze crap --coverprofile` already does; the same implementation is reused or shared.
- The `--coverprofile` flag applies only to the CRAP analysis step within `gaze report`. The quality assessment, side-effect classification, and documentation scan steps are unaffected.
- No caching or persistence of the coverage profile between runs is in scope.
- The flag is optional and additive; no existing flag, behavior, or output format changes.

## Out of Scope

- Support for non-Go coverage formats (LCOV, Cobertura, JaCoCo, etc.).
- Merging multiple coverage profiles; a single profile path is accepted.
- Caching or persisting the profile between runs.
- Any changes to the quality assessment, side-effect classification, or documentation scan steps.
- Validation that the profile was generated for the same package set as the `<patterns>` argument (missing packages are silently treated as 0% covered — existing behavior).

## Clarifications

- **Q: Should `--coverprofile` be validated before any analysis begins (fail-fast), or only when the CRAP step runs?**
  A: Pre-flight validation in `runReport` before calling `aireport.Run`. When `coverProfile` is non-empty, `runReport` validates the path (existence, is-regular-file) before the analysis pipeline starts. This satisfies FR-006 (exit non-zero with clear error) without conflicting with the partial-failure architecture of `runProductionPipeline`. The validation logic mirrors `crap.Analyze`'s checks (`os.Stat`, `info.IsDir()`). If validation fails, `runReport` returns an error immediately — no analysis steps run. If validation passes, the path is forwarded to `RunnerOptions.CoverProfile` as before. (Decision made 2026-03-12; Option A selected over Option B/C.)

- **Q: If the coverage profile references packages not included in `<patterns>`, are the extra records ignored or do they cause an error?**
  A: Extra records are silently ignored; only packages matching `<patterns>` are analyzed. This is the existing behavior of `crap.Analyze`.

- **Q: Does `--coverprofile` interact with `--format=json`?**
  A: Yes — FR-008 explicitly requires the flag to work for both `--format=text` and `--format=json`.
