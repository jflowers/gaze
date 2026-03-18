## Why

When an AI adapter subprocess (opencode, claude, gemini) exits with code 0 but produces empty stdout, gaze reports a generic error like "opencode returned empty output (FR-009)" with no diagnostic context. The actual cause — visible in the subprocess's stderr (e.g., "Error: Invalid API key") — is silently discarded because `runSubprocess` only captures stderr for non-zero exits.

This was observed in CI where opencode 1.2.27 exited 0 with empty stdout due to an authentication failure. The user received no actionable information about the root cause.

## What Changes

Surface subprocess stderr in adapter empty-output error messages. When a subprocess exits 0 but stdout is empty, include the captured stderr (truncated to `maxAdapterStderrBytes`) in the FR-009/FR-016 error returned by each adapter.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `runSubprocess`: Returns stderr bytes alongside stdout bytes on success, enabling callers to include stderr in diagnostic messages
- `OpenCodeAdapter.Format`: FR-009 error message includes subprocess stderr when available
- `ClaudeAdapter.Format`: FR-016 error message includes subprocess stderr when available
- `GeminiAdapter.Format`: FR-016 error message includes subprocess stderr when available

### Removed Capabilities
- None

## Impact

- `internal/aireport/adapter.go`: `runSubprocess` return signature changes to include stderr bytes
- `internal/aireport/adapter_opencode.go`: Empty-output error includes stderr context
- `internal/aireport/adapter_claude.go`: Empty-output error includes stderr context
- `internal/aireport/adapter_gemini.go`: Empty-output error includes stderr context
- `internal/aireport/adapter_opencode_test.go`: Tests updated for new return signature
- `internal/aireport/adapter_claude_test.go`: Tests updated for new return signature
- `internal/aireport/adapter_gemini_test.go`: Tests updated for new return signature
- `internal/aireport/subprocess_test.go`: Tests updated for new return signature

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change is internal to gaze's error reporting. No artifact communication patterns are affected.

### II. Composability First

**Assessment**: PASS

No new dependencies introduced. The change modifies internal error message formatting within existing adapter functions.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observability — surfacing stderr in error messages gives users machine-readable and human-readable diagnostic context instead of a generic "empty output" message.

### IV. Testability

**Assessment**: PASS

The `runSubprocess` function is already unit-tested. The return signature change is testable via existing fake binary patterns. Each adapter's empty-output error path is covered by existing tests that will be updated.
