# Feature Specification: OpenCode AI Adapter for gaze report

**Feature Branch**: `019-opencode-adapter`
**Created**: 2026-03-12
**Status**: Implemented
**Input**: User description: "Support opencode as a gaze report --ai provider"

## Overview

Go developers who use OpenCode as their AI agent runtime want to generate full `gaze report` quality reports using the same `opencode` binary they already have installed — without needing a separate `claude` or `gemini` CLI. Today, `gaze report --ai` only supports `claude`, `gemini`, and `ollama`. OpenCode users who do not have those CLIs installed are excluded from CI report generation.

This feature adds `opencode` as a fourth AI adapter for `gaze report --ai=opencode`, following the same integration patterns as the three existing adapters.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Report with OpenCode (Priority: P1)

A developer uses OpenCode as their primary AI agent environment and has the `opencode` binary installed. They want to add `gaze report` to their CI pipeline without installing an additional AI CLI (`claude`, `gemini`, or `ollama`) that they don't otherwise use.

They add a single step to their workflow:

```yaml
- name: Gaze Quality Report
  run: gaze report ./... --ai=opencode
```

After the step runs, a formatted markdown quality report appears on stdout (and in the GitHub Actions Step Summary if `GITHUB_STEP_SUMMARY` is set), identical in structure to reports produced by the other AI adapters.

**Why this priority**: This is the entire purpose of the feature. Without it there is nothing. Every other story depends on this one being viable.

**Independent Test**: Can be fully tested by running `gaze report ./... --ai=opencode` with the `opencode` binary on PATH and verifying that a well-formed markdown report containing CRAP summary, quality summary, classification summary, and health assessment sections is written to stdout.

**Acceptance Scenarios**:

1. **Given** a Go project with gaze installed and `opencode` available on PATH, **When** `gaze report ./... --ai=opencode` is run, **Then** a formatted markdown report is written to stdout containing all four analysis sections.
2. **Given** the same setup with `GITHUB_STEP_SUMMARY` set to a file path, **When** the command runs, **Then** the same formatted markdown is also appended to that file.
3. **Given** `opencode` is available on PATH and `--model claude-3-5-sonnet` is specified, **When** `gaze report ./... --ai=opencode --model=claude-3-5-sonnet` is run, **Then** the specified model is passed to the opencode subprocess and a formatted report is produced.
4. **Given** `opencode` is available on PATH and `--model` is omitted, **When** `gaze report ./... --ai=opencode` is run, **Then** the command succeeds using opencode's own configured default model.

---

### User Story 2 - Consistent Behavior with Other Adapters (Priority: P2)

A developer switching from `--ai=claude` to `--ai=opencode` expects the same behavior: the same error messages, the same threshold flags, the same output format, the same partial-failure handling. No opencode-specific ceremony is required.

**Why this priority**: Behavioral parity with existing adapters is what makes the feature trustworthy. An adapter that behaves differently in edge cases — errors, empty output, binary not found — creates a confusing and inconsistent user experience.

**Independent Test**: Can be fully tested by verifying that `--ai=opencode` produces the same structural report sections as `--ai=claude` when given the same analysis data, and that all existing error-handling behaviors (binary not found, empty output, non-zero exit) work identically.

**Acceptance Scenarios**:

1. **Given** `opencode` is not installed on PATH, **When** `gaze report ./... --ai=opencode` is run, **Then** the command fails immediately with a clear error message before running any analysis — identical behavior to when `claude` or `gemini` is missing.
2. **Given** `opencode` is on PATH but returns empty output, **When** `gaze report ./... --ai=opencode` is run, **Then** the command exits non-zero with a diagnostic message identifying `opencode` as the adapter and describing the empty response.
3. **Given** `opencode` is on PATH but exits with a non-zero exit code, **When** `gaze report ./... --ai=opencode` is run, **Then** the command fails with a clear error that includes the opencode process's stderr output (truncated to prevent secret leakage).
4. **Given** `--ai=opencode` is used with `--format=json`, **When** the command runs, **Then** the opencode binary is never invoked and the raw JSON payload is written to stdout — consistent with `--format=json` behavior for all other adapters.
5. **Given** the same analysis payload is given to all four adapters via `gaze report`, **When** each adapter produces its report, **Then** all four reports contain the same structural sections (CRAP summary, quality summary, classification summary, health assessment), differing only in AI-generated prose.

---

### Edge Cases

- What happens when `opencode` is not on PATH? The command fails immediately with a clear error before any analysis runs (pre-flight binary check).
- What happens when `opencode` exits 0 but produces empty or whitespace-only output? The command exits non-zero with a diagnostic message identifying the adapter and describing the empty response.
- What happens when `opencode` exits with a non-zero code? The command fails with an error wrapping the exit error and including a truncated (max 512 bytes) snippet of the subprocess stderr.
- What happens when `--model` is omitted for `opencode`? The command succeeds; opencode uses its own configured default model. No error is produced.
- What happens when the system prompt is very long? The system prompt is written to a temporary file on disk (not passed as a CLI argument) to avoid OS argument length limits. The temp file is removed after the subprocess exits.
- What happens when the temp directory cannot be created? The command fails immediately with a clear error before invoking opencode.
- What happens when `--ai=opencode` is specified and `opencode` exists on PATH but the agent file write fails (e.g., disk full)? The command fails immediately with a clear error before invoking opencode.
- What happens when `--ai` error message lists valid adapters? The message MUST include `opencode` alongside `claude`, `gemini`, and `ollama`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST accept `opencode` as a valid value for the `--ai` flag in `gaze report`.
- **FR-002**: When `--ai=opencode` is specified, the tool MUST invoke the `opencode run` subcommand to format the analysis payload, using the same system prompt loaded by the existing `LoadPrompt` mechanism. The subprocess MUST be governed by the same context deadline as all other adapters (the `--ai-timeout` flag, default 10 minutes); no separate timeout flag is introduced.
- **FR-003**: The system prompt MUST be delivered to the `opencode run` subprocess by writing it as an agent definition file in a temporary directory, then passing the path to that directory and the agent name as arguments to the subprocess. The agent definition file MUST be prefixed with empty YAML frontmatter (`---\n---\n`) before the prompt body to ensure opencode recognizes it as a valid agent file. The temporary directory MUST be removed after the subprocess exits.
- **FR-004**: The analysis payload MUST be delivered to the `opencode run` subprocess via its standard input.
- **FR-005**: The `opencode run` subprocess MUST be invoked with flags that produce plain-text output to stdout (not a structured event stream). The adapter MUST read that plain-text output as the formatted report.
- **FR-006**: `--model` MUST be optional for the `opencode` adapter. When provided, the model name MUST be passed to the `opencode run` subprocess. When omitted, opencode uses its own configured default and the adapter MUST NOT pass any model flag.
- **FR-007**: When the `opencode` binary is not found on PATH, the tool MUST fail immediately with a clear error before running any analysis. This check MUST be performed as a pre-flight validation before the analysis pipeline begins.
- **FR-008**: When `opencode run` exits with a non-zero exit code, the tool MUST fail with an error that includes the exit error and a truncated snippet of the subprocess stderr output (maximum 512 bytes) to aid debugging while preventing secret leakage.
- **FR-009**: When `opencode run` exits with code 0 but produces empty or whitespace-only output, the tool MUST exit non-zero with a diagnostic message identifying `opencode` as the adapter.
- **FR-010**: The tool MUST cap the amount of output read from the `opencode run` subprocess at 64 MiB to prevent out-of-memory conditions on unexpectedly large responses.
- **FR-011**: The `--ai` flag help text, the error message for an omitted `--ai` flag in text mode, and the command usage examples MUST all be updated to include `opencode` alongside the existing adapter names.
- **FR-012**: All existing adapter behaviors (claude, gemini, ollama) MUST continue to work without regression.

### Key Entities

- **OpenCode Adapter**: The new AI adapter that wraps the `opencode run` subprocess. Attributes: binary path (resolved at runtime), model name (optional), system prompt (written to a temp agent file), analysis payload (delivered via stdin), plain-text output (read from stdout).
- **Agent Definition File**: The file written by the adapter to a temporary directory that delivers the system prompt to the `opencode run` subprocess via opencode's native agent file convention.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can run `gaze report ./... --ai=opencode` and receive a complete, structured quality report in stdout without installing any AI CLI other than `opencode`.
- **SC-002**: The report produced by `--ai=opencode` contains the same four structural sections as reports produced by `--ai=claude`, `--ai=gemini`, and `--ai=ollama`, given the same analysis input.
- **SC-003**: When `opencode` is not on PATH, the command fails immediately with a clear error message — no analysis pipeline steps are executed before the failure.
- **SC-004**: All pre-existing `gaze report` tests continue to pass without modification; the `opencode` adapter introduces no regressions.
- **SC-005**: The `opencode` adapter has automated test coverage for: normal success, empty-output error, non-zero-exit error, binary-not-found error, and model-flag presence/absence — using the same fake-binary test pattern as the existing claude and gemini adapters.

## Clarifications

### Session 2026-03-12 (clarify pass)

- Q: Should the `opencode` adapter respect the existing `--ai-timeout` context deadline, or use a separate timeout? → A: Inherit the existing `--ai-timeout` context (same `ctx` passed to `Format()`), identical to all other adapters. No new timeout flag.
- Q: When writing the system prompt to the temporary agent definition file, should the adapter add minimal YAML frontmatter? → A: Prepend empty YAML frontmatter (`---\n---\n`) before the prompt body so opencode recognizes the file as a valid agent definition regardless of version-specific parsing behavior.
- Q: Should an empty `""` positional message argument be passed to `opencode run`, or should stdin alone be sufficient? → A: Pass `""` as the positional message argument, mirroring the `-p ""` pattern used by claude and gemini to trigger non-interactive headless operation while the JSON payload is delivered via stdin.

### Session 2026-03-12 (pre-spec design decisions)

- Q: How should the system prompt be delivered to `opencode run`? → A: Written as an agent definition file in a temporary directory. The subprocess is passed the path to that temp directory and the agent name as arguments. This mirrors the Gemini adapter's approach of writing `GEMINI.md` to a temp dir and setting the working directory.
- Q: How should the analysis payload be delivered? → A: Via stdin, same as all other adapters.
- Q: Should `--model` be required for `opencode`? → A: No. Optional, same as `claude` and `gemini`. When absent, opencode uses its own configured default.
- Q: What output format should be requested from `opencode run`? → A: Plain-text (the default output format). This avoids NDJSON event-stream parsing complexity and mirrors the Claude adapter's behavior.
- Q: Should `--model` be required for `opencode`? → A: No. Optional, same as `claude` and `gemini`.
- Q: What positional message argument should be passed to `opencode run`? → A: An empty string `""`, mirroring the `-p ""` pattern used by `claude` and `gemini` to signal non-interactive/headless operation while payload is read from stdin.

## Assumptions

- Users are responsible for ensuring `opencode` is installed, authenticated, and has a working default model configured (or provide `--model` explicitly). The adapter does not manage opencode installation, authentication, or model selection.
- The `opencode run` subcommand reads from stdin when an empty `""` positional message argument is provided; passing `""` triggers non-interactive headless operation (mirroring `-p ""` for claude and gemini) and the JSON payload is delivered via stdin.
- The plain-text output mode of `opencode run` produces output on stdout that is suitable as a formatted markdown report without further parsing.
- The formatting prompt embedded in the binary (or loaded from `.opencode/agents/gaze-reporter.md` in the working directory) is compatible with opencode's agent execution model. Read-tool instructions in the prompt body are opencode-native and will execute normally when `opencode run` processes the agent file.
- The temp directory and agent file write pattern (mirroring the Gemini adapter) is sufficient to deliver the system prompt to opencode without requiring any changes to how gaze loads or stores the prompt string.

## Out of Scope

- Parsing the NDJSON event stream format from `opencode run --format json`.
- Making `--model` required for the `opencode` adapter.
- Adding opencode-specific flags (e.g., `--thinking`, `--variant`, `--share`).
- Changes to the `gaze-reporter` agent prompt or `.opencode/` scaffold files.
- Changes to existing adapters (claude, gemini, ollama).
- Adding `gaze report --ai=opencode` to gaze's own CI workflow (dogfooding).
- Support for attaching to a running opencode server (`--attach`).
