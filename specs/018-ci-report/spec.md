# Feature Specification: AI-Powered CI Quality Report

**Feature Branch**: `018-ci-report`
**Created**: 2026-03-10
**Status**: Draft
**Input**: User description: "018-ci-report"

## Overview

Go developers who use gaze want a single command they can add to a GitHub Actions workflow (or any CI pipeline) to produce the same rich, AI-formatted full quality report that the `/gaze` OpenCode command produces interactively. Today that report is only available inside the OpenCode agent environment. Outside of it — in CI — users have no equivalent: they can run the raw gaze sub-commands, but they receive raw JSON with no interpretation, no health assessment, no prioritized recommendations.

This feature adds a `gaze report` subcommand that orchestrates gaze's four analysis sub-commands, loads a formatting prompt, pipes the gathered data to a user-specified external AI CLI (claude, gemini, or ollama), and writes the resulting formatted markdown report to stdout. When running inside GitHub Actions, the report is also automatically appended to the workflow's Step Summary, making it visible directly in the Actions UI without inspecting log output. Optional threshold flags allow the step to fail the build when quality drops below acceptable levels.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Full Report in GitHub Actions (Priority: P1)

A developer has gaze installed in their project and an AI CLI (e.g., claude) available in their CI environment. They want every pull request to include a full gaze quality report — CRAP scores, test quality summary, side effect classification, and a prioritized health assessment — visible directly in the GitHub Actions UI, not buried in log output.

They add a single step to their workflow:

```yaml
- name: Gaze Quality Report
  run: gaze report ./... --ai=claude
```

After the step runs, a formatted markdown report appears in the Actions Step Summary tab, identical in structure and content to what `/gaze` produces in OpenCode.

**Why this priority**: This is the core use case that motivates the entire feature. Every other story builds on this one. Without it, there is no CI report.

**Independent Test**: Can be fully tested by running `gaze report ./... --ai=claude` in a simulated GitHub Actions environment (with `GITHUB_STEP_SUMMARY` set to a temp file) and verifying the output file contains a well-formed markdown report with the expected sections.

**Acceptance Scenarios**:

1. **Given** a Go project with gaze installed and `claude` available on PATH, and `GITHUB_STEP_SUMMARY` set to a file path, **When** `gaze report ./... --ai=claude` is run, **Then** a formatted markdown report is written to that file containing CRAP summary, quality summary, classification summary, and health assessment sections.
2. **Given** the same setup, **When** the command runs, **Then** the same formatted markdown is also printed to stdout.
3. **Given** the same setup with a package pattern that matches multiple packages, **When** `gaze report ./... --ai=claude` is run, **Then** all packages are included in the analysis and the report reflects the aggregate results.

---

### User Story 2 - Build Failure on Quality Regression (Priority: P2)

A developer wants CI to enforce quality standards: if the project's CRAPload or average contract coverage regresses below a configured threshold, the workflow step fails with a non-zero exit code and prints a clear summary to the log. This lets the PR author know their changes degraded quality before merging.

**Why this priority**: Report-only mode has value, but enforcing thresholds is what makes the tool actionable in CI. Without this, developers have no automated safety net against quality regressions.

**Independent Test**: Can be fully tested by running `gaze report ./... --ai=claude --max-crapload=1` against a project with known CRAPload > 1, verifying exit code is 1 and a threshold summary appears on stderr.

**Acceptance Scenarios**:

1. **Given** a project whose CRAPload exceeds `--max-crapload`, **When** `gaze report ./... --ai=claude --max-crapload=5` is run, **Then** the command exits with code 1 and a summary line like `CRAPload: 8/5 (FAIL)` is written to stderr.
2. **Given** a project whose CRAPload is within the limit, **When** the same command is run, **Then** the command exits with code 0 and `CRAPload: 3/5 (PASS)` appears on stderr.
3. **Given** `--min-contract-coverage=70` and a project whose average contract coverage is below 70%, **When** the command runs, **Then** exit code is 1 with an appropriate FAIL summary on stderr.
4. **Given** no threshold flags are provided, **When** the command runs, **Then** it always exits 0 regardless of metric values (report-only mode).
5. **Given** `--max-crapload=0` is explicitly provided and the project has at least one function with any positive CRAP score, **When** `gaze report ./... --ai=claude --max-crapload=0` is run, **Then** the command exits with code 1 and a threshold summary appears on stderr, confirming that zero is treated as a live threshold and not as "disabled".

---

### User Story 3 - Local Development Report Without CI (Priority: P3)

A developer wants to generate the same rich full report locally, outside of OpenCode, without needing to be inside the agent environment. They run `gaze report` from their terminal, piping output wherever they want.

**Why this priority**: Local use is valuable but secondary. The primary target is CI. Local use works as a natural side-effect of the same implementation.

**Independent Test**: Can be fully tested by running `gaze report ./... --ai=claude` in a terminal with no `GITHUB_STEP_SUMMARY` set and verifying that well-formed formatted markdown appears on stdout only.

**Acceptance Scenarios**:

1. **Given** `GITHUB_STEP_SUMMARY` is not set, **When** `gaze report ./... --ai=claude` is run, **Then** the formatted report is written to stdout only (no file written).
2. **Given** a local `.opencode/agents/gaze-reporter.md` file exists, **When** `gaze report` runs, **Then** that file is used as the AI formatting prompt instead of the embedded default.

---

### User Story 4 - Alternative AI CLI (Priority: P4)

A developer uses ollama for local AI or gemini in their CI environment. They want the same `gaze report` command to work with their preferred AI CLI by specifying it via `--ai`.

**Why this priority**: Supporting multiple AI CLIs broadens the user base significantly. The architecture must accommodate this, but the core value is delivered by any single working adapter.

**Independent Test**: Can be fully tested by running `gaze report ./... --ai=ollama --model=llama3.2` and verifying that a well-formed report is produced (requires ollama with the specified model available).

**Acceptance Scenarios**:

1. **Given** ollama is on PATH with a model available, **When** `gaze report ./... --ai=ollama --model=llama3.2` is run, **Then** a formatted report is produced on stdout.
2. **Given** gemini is on PATH, **When** `gaze report ./... --ai=gemini` is run, **Then** a formatted report is produced on stdout.
3. **Given** `--ai=ollama` is specified but `--model` is omitted, **When** the command runs, **Then** it fails immediately with an error: "`--model` is required when using ollama".

---

### Edge Cases

- What happens when `--ai` flag is omitted? The command fails immediately with a clear error message listing valid adapter names, before any analysis is run.
- What happens when the specified AI CLI is not found on PATH? The command fails with a clear error and a suggested install path, before any analysis is run.
- What happens when one of the four analysis sub-commands fails (e.g., no test coverage data)? The report is produced with the available data; missing sections are replaced with a visible warning callout. The command does not fail solely because one analysis step failed.
- What happens when the package pattern matches zero packages? The command fails with a clear error before invoking the AI CLI.
- What happens when `GITHUB_STEP_SUMMARY` is set but the file path is not writable? The command writes to stdout successfully and emits a warning to stderr about the Step Summary write failure; it does not fail entirely.
- What happens when the AI CLI times out or returns an error? The command fails with a clear error that includes the AI CLI's stderr output.
- What happens when the AI CLI exits successfully (exit code 0) but produces empty or clearly unusable output (e.g., only whitespace)? The command fails with exit code 1 and a message identifying the adapter and describing the empty response (e.g., "AI adapter 'claude' returned empty response; check model availability or API quota").
- What happens when `--max-crapload` and `--min-contract-coverage` are both set and both are breached? Both failures are reported on stderr and the command exits 1.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST provide a `report` subcommand that, when invoked, runs all four gaze analysis operations (CRAP scoring, test quality assessment, side effect classification, and documentation scanning) against the specified package pattern and produces a unified formatted report.
- **FR-002**: The `report` subcommand MUST require an `--ai` flag specifying the AI CLI adapter to use (`claude`, `gemini`, or `ollama`). Omitting `--ai` MUST cause an immediate error with a message listing valid values.
- **FR-003**: The tool MUST support a `--model` flag that passes a model name to the active AI adapter. For ollama, `--model` MUST be required; for claude and gemini, it is optional.
- **FR-004**: The tool MUST invoke the specified AI CLI with the gaze analysis data and a formatting prompt, and use the AI CLI's response as the formatted report content.
- **FR-005**: The formatting prompt MUST be loaded from `.opencode/agents/gaze-reporter.md` in the current working directory if that file exists. If it does not exist, the tool MUST fall back to an embedded default prompt.
- **FR-006**: The formatted report MUST be written to stdout.
- **FR-007**: When the environment variable `GITHUB_STEP_SUMMARY` is set to a file path, the tool MUST also append the formatted report to that file.
- **FR-008**: If `GITHUB_STEP_SUMMARY` is set but the target file is not writable, the tool MUST still succeed (writing to stdout) and emit a warning to stderr describing the write failure.
- **FR-009**: The tool MUST support threshold flags: `--max-crapload`, `--max-gaze-crapload`, and `--min-contract-coverage`. When any threshold is breached, the tool MUST exit with a non-zero exit code and write a threshold summary to stderr.
- **FR-010**: When no threshold flags are provided on the command line (flags are absent, not merely set to zero), the tool MUST exit with code 0 regardless of metric values. A threshold value of zero (e.g., `--max-crapload=0`) is a valid threshold meaning "fail if any function exceeds the CRAP threshold" and MUST cause exit code 1 when breached.
- **FR-011**: When one or more analysis operations fail (e.g., no coverage data available), the tool MUST produce a partial report using the data that was successfully gathered. Sections for failed analyses MUST be replaced with a visible warning indicator. The tool MUST NOT fail solely because one analysis step failed.
- **FR-012**: When the specified AI CLI is not found on PATH, the tool MUST fail immediately with a clear error before running any analysis.
- **FR-013**: When the package pattern matches zero packages, the tool MUST fail with a clear error before invoking the AI CLI.
- **FR-014**: The default package pattern when none is specified MUST be `./...`.
- **FR-015**: The tool MUST accept a `--format` flag with values `text` (default) and `json`. In `json` mode, the raw analysis data is written as structured JSON without AI formatting. In `json` mode, the `--ai` flag is not required and AI adapter validation is skipped entirely; the command completes without invoking any AI CLI.
- **FR-016**: When the AI CLI exits with code 0 but produces empty or clearly unusable output (e.g., only whitespace), the tool MUST exit with code 1 and emit a diagnostic message identifying the adapter and describing the empty response.
- **FR-017**: The tool MUST emit brief progress signals to stderr at key phase transitions (e.g., "Analyzing packages...", "Formatting report..."). These signals MUST appear on stderr only; stdout is reserved for the formatted report.

### Key Entities

- **Report Command**: The `gaze report` subcommand that orchestrates the full pipeline. Attributes: package pattern, AI adapter selection, model name, threshold configuration, output targets.
- **AI Adapter**: A named integration that knows how to invoke a specific external AI CLI with a system prompt and data payload. Supported adapters: `claude`, `gemini`, `ollama`.
- **Formatting Prompt**: The instruction set passed to the AI CLI that tells it how to structure and present the analysis data. Loaded from the local agent file or the embedded default.
- **Analysis Payload**: The structured data bundle containing results from all four analysis operations (CRAP, quality, classification, docscan), passed to the AI adapter for formatting.
- **Step Summary**: The GitHub Actions file (identified by `GITHUB_STEP_SUMMARY`) to which the formatted report is appended for display in the Actions UI.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can add a single workflow step to a GitHub Actions workflow and get a full quality report visible in the Step Summary tab with no additional configuration beyond specifying `--ai`.
- **SC-002**: The formatted report produced by `gaze report` contains the same structural sections as the report produced by the `/gaze` OpenCode command: CRAP summary, quality summary, classification summary, and health assessment with prioritized recommendations.
- **SC-003**: When a threshold flag is provided and the measured metric exceeds (or falls below) the threshold, the command exits non-zero within the same wall-clock time as the analysis itself (no additional delay for threshold evaluation).
- **SC-004**: When one of the four analysis operations fails, the command still produces a report for the remaining operations within the same total runtime, rather than aborting entirely.
- **SC-005**: The gaze-owned analysis portion of `gaze report` (excluding the external AI CLI round-trip) completes within 5 minutes for a typical Go project of fewer than 50 packages on a standard CI runner.
- **SC-006**: All three AI adapters (claude, gemini, ollama) produce a report with identical structure, differing only in the AI-generated prose, given the same analysis data and formatting prompt.

## Clarifications

### Session 2026-03-10

- Q: When the AI CLI exits successfully (exit code 0) but produces empty or clearly unusable output, what should gaze report do? → A: Fatal error, exit non-zero. Exit 1 with a diagnostic message identifying the adapter and describing the empty response.
- Q: What is the acceptable wall-clock time bound for the gaze analysis portion of `gaze report` (excluding AI CLI round-trip)? → A: 5 minutes for a typical Go project of fewer than 50 packages on a standard CI runner.
- Q: How does gaze deliver the formatting prompt and analysis data to the AI CLI process? → A: Prompt as a CLI flag specific to the adapter; analysis JSON data written to the process's stdin.
- Q: Should gaze report emit progress signals to stderr while analysis and AI formatting are running? → A: Yes, lightweight signals at key phase transitions (e.g., "Analyzing packages...", "Formatting report...") on stderr only; stdout reserved for the formatted report.
- Q: When --format=json is used, should gaze report still require and validate the --ai flag? → A: No. JSON mode bypasses all AI adapter logic; --ai is not required and AI validation is skipped entirely.

## Assumptions

- Users are responsible for ensuring their chosen AI CLI is installed and authenticated in their CI environment (e.g., API key set, model pulled). The tool does not manage AI CLI installation or authentication.
- The `gaze` binary is already installed in the CI environment (e.g., via Homebrew, `go install`, or pre-built binary download). This feature does not change how gaze itself is distributed or installed.
- The formatting prompt embedded in the binary is the same content as `.opencode/agents/gaze-reporter.md` from the scaffold. It produces the same report format as the `/gaze` OpenCode command.
- `--format=json` outputs raw analysis data only (no AI invocation, no `--ai` flag required). This is useful for downstream tooling that wants to consume structured data without any AI CLI installed. AI formatting is only applied in `text` mode (the default).
- Each AI adapter delivers the formatting prompt as a CLI flag specific to that adapter (e.g., `claude -p "<prompt>"`) and writes the analysis JSON data to the process's stdin. Exact per-adapter flag names are confirmed during planning, but all adapters follow the prompt-as-flag, data-as-stdin contract.

## Out of Scope

- Adding `gaze report` to gaze's own CI workflow (dogfooding) — this is a separate future decision.
- Changes to the `/gaze` OpenCode command or `gaze-reporter` agent.
- Changes to `gaze self-check`.
- Adding new scaffold files via `gaze init`.
- Browser-based or web UI report rendering.
- Sending the report to external services (Slack, email, etc.).
- Support for AI APIs accessed directly via HTTP (without a CLI wrapper) — all AI integration is via external CLI adapters.
