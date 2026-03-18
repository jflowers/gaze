## ADDED Requirements

### Requirement: runSubprocess MUST return stderr on success

`runSubprocess` MUST return captured stderr bytes alongside stdout bytes when the subprocess exits with code 0. The return signature MUST be `(stdout []byte, stderr []byte, err error)`.

#### Scenario: Subprocess exits 0 with stderr output
- **GIVEN** a subprocess that writes "warning: something" to stderr and "OK" to stdout
- **WHEN** the subprocess exits with code 0
- **THEN** `runSubprocess` returns `([]byte("OK"), []byte("warning: something"), nil)`

#### Scenario: Subprocess exits 0 with empty stderr
- **GIVEN** a subprocess that writes "OK" to stdout and nothing to stderr
- **WHEN** the subprocess exits with code 0
- **THEN** `runSubprocess` returns `([]byte("OK"), []byte(""), nil)`

#### Scenario: Subprocess exits non-zero (unchanged)
- **GIVEN** a subprocess that exits with code 1 and writes "error" to stderr
- **WHEN** `runSubprocess` processes the result
- **THEN** `runSubprocess` returns `(nil, nil, error)` with stderr embedded in the error message (existing behavior preserved)

## MODIFIED Requirements

### Requirement: Adapter empty-output errors MUST include stderr context

Previously: Adapters returned generic errors like `"opencode returned empty output (FR-009): ensure the opencode CLI is working correctly"` with no subprocess diagnostic context.

Each adapter's empty-output error MUST include truncated stderr (at `maxAdapterStderrBytes`) when stderr contains non-whitespace content. The stderr MUST be appended after a newline, using the format: `"<existing message>\nstderr: <truncated stderr>"`.

#### Scenario: OpenCode exits 0 with empty stdout and stderr containing error
- **GIVEN** opencode exits with code 0, empty stdout, and stderr "Error: Invalid API key."
- **WHEN** `OpenCodeAdapter.Format` detects empty output
- **THEN** the error message includes `"FR-009"` AND contains `"stderr: Error: Invalid API key."`

#### Scenario: Claude exits 0 with empty stdout and stderr containing error
- **GIVEN** claude exits with code 0, empty stdout, and stderr "Authentication failed"
- **WHEN** `ClaudeAdapter.Format` detects empty output
- **THEN** the error message includes `"FR-016"` AND contains `"stderr: Authentication failed"`

#### Scenario: Adapter exits 0 with empty stdout and empty stderr
- **GIVEN** an adapter subprocess exits with code 0, empty stdout, and empty stderr
- **WHEN** the adapter detects empty output
- **THEN** the error message is unchanged from existing behavior (no stderr suffix appended)

## REMOVED Requirements

None.
