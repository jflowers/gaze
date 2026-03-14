## ADDED Requirements

### Requirement: runSubprocess helper

`runSubprocess` MUST accept a context, binary name, args, optional working directory, and a payload reader. It MUST return the subprocess stdout as `[]byte` and an error.

`runSubprocess` MUST resolve the binary via `exec.LookPath` before execution. If the binary is not found, it MUST return an error that includes the binary name.

`runSubprocess` MUST cap stdout reads at `maxAdapterOutputBytes` (64 MiB). It MUST cap stderr included in error messages at `maxAdapterStderrBytes` (512 bytes).

`runSubprocess` MUST set `cmd.Dir` when `cmdDir` is non-empty. When `cmdDir` is empty, `cmd.Dir` MUST NOT be set.

`runSubprocess` MUST set `cmd.Stdin` to the provided payload reader.

#### Scenario: Successful subprocess execution
- **GIVEN** a binary that exits 0 and writes output to stdout
- **WHEN** `runSubprocess` is called with the binary name and args
- **THEN** the returned `[]byte` MUST contain the stdout output and error MUST be nil

#### Scenario: Binary not found
- **GIVEN** a binary name that does not exist on PATH
- **WHEN** `runSubprocess` is called
- **THEN** the returned error MUST contain the binary name and indicate it was not found

#### Scenario: Subprocess exits non-zero
- **GIVEN** a binary that exits with a non-zero status and writes to stderr
- **WHEN** `runSubprocess` is called
- **THEN** the returned error MUST include stderr content, truncated to `maxAdapterStderrBytes` if longer

#### Scenario: Working directory is set
- **GIVEN** `cmdDir` is set to a valid directory path
- **WHEN** `runSubprocess` is called
- **THEN** the subprocess MUST execute with its working directory set to `cmdDir`

#### Scenario: Context cancellation
- **GIVEN** a context that is cancelled while the subprocess is running
- **WHEN** `runSubprocess` is called with that context
- **THEN** the subprocess MUST be terminated and an error MUST be returned

## MODIFIED Requirements

### Requirement: ClaudeAdapter.Format uses runSubprocess

Previously: `ClaudeAdapter.Format` contained inline binary lookup, subprocess pipe setup, output capture, and error handling.

`ClaudeAdapter.Format` MUST delegate subprocess execution to `runSubprocess`. The method MUST retain adapter-specific logic: temp dir creation with `prompt.md`, arg construction with `--system-prompt-file`, and empty-output check with FR-016 reference.

#### Scenario: Claude produces identical output
- **GIVEN** identical inputs to `ClaudeAdapter.Format` before and after decomposition
- **WHEN** `Format` is called
- **THEN** the returned string and error MUST be identical

### Requirement: GeminiAdapter.Format uses runSubprocess

Previously: `GeminiAdapter.Format` contained inline binary lookup, subprocess pipe setup, output capture, and error handling.

`GeminiAdapter.Format` MUST delegate subprocess execution to `runSubprocess` with `cmdDir` set to the temp directory. The method MUST retain adapter-specific logic: temp dir creation with `GEMINI.md`, arg construction with `--output-format json`, JSON response parsing, and empty-output check.

#### Scenario: Gemini produces identical output
- **GIVEN** identical inputs to `GeminiAdapter.Format` before and after decomposition
- **WHEN** `Format` is called
- **THEN** the returned string and error MUST be identical

### Requirement: OpenCodeAdapter.Format uses runSubprocess

Previously: `OpenCodeAdapter.Format` contained inline binary lookup, subprocess pipe setup, output capture, and error handling.

`OpenCodeAdapter.Format` MUST delegate subprocess execution to `runSubprocess`. The method MUST retain adapter-specific logic: temp dir creation with nested `.opencode/agents/gaze-reporter.md` and YAML frontmatter, arg construction with `--dir`, `--agent`, `--format default`, and empty-output check with FR-009 reference.

#### Scenario: OpenCode produces identical output
- **GIVEN** identical inputs to `OpenCodeAdapter.Format` before and after decomposition
- **WHEN** `Format` is called
- **THEN** the returned string and error MUST be identical

## REMOVED Requirements

None.
