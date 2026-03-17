## Context

`runSubprocess` in `internal/aireport/adapter.go` currently returns `([]byte, error)` — only stdout bytes. Stderr is captured in a `bytes.Buffer` but only included in the error message when the subprocess exits non-zero. When a subprocess exits 0 with empty stdout, each adapter (opencode, claude, gemini) reports a generic "returned empty output" error with no diagnostic context.

Real-world failure: opencode 1.2.27 exits 0 with empty stdout when authentication fails, writing "Error: Invalid API key." to stderr. Gaze reports "opencode returned empty output (FR-009)" — the user has no way to diagnose the cause without manually running opencode.

## Goals / Non-Goals

### Goals
- Surface subprocess stderr in adapter error messages when stdout is empty
- Maintain existing error behavior for non-zero exit codes (unchanged)
- Keep stderr truncation at `maxAdapterStderrBytes` (512 bytes) for all paths

### Non-Goals
- Changing how non-zero exit errors are formatted (already includes stderr)
- Logging stderr on successful runs with non-empty stdout
- Handling OllamaAdapter (uses HTTP, not subprocess)

## Decisions

### D1: Return stderr alongside stdout from runSubprocess

Change `runSubprocess` return from `([]byte, error)` to `([]byte, []byte, error)` — returning `(stdout, stderr, error)`. On non-zero exit, the function still returns an error with stderr embedded (existing behavior). On success, both stdout and stderr bytes are returned to the caller.

**Rationale**: This is the minimal change that gives each adapter access to stderr for diagnostic messages. The alternative — having `runSubprocess` itself detect empty stdout and include stderr — would couple the subprocess helper to adapter-specific error formatting (FR-009 vs FR-016 references).

### D2: Include stderr in empty-output errors only when stderr is non-empty

Each adapter's empty-output check appends truncated stderr to the error message only when stderr contains non-whitespace content. When stderr is empty, the existing error message is preserved exactly.

Format: `"opencode returned empty output (FR-009): ensure the opencode CLI is working correctly\nstderr: Error: Invalid API key."` — stderr appended after a newline, matching the format already used in non-zero exit errors.

**Rationale**: Consistent with existing error formatting in `runSubprocess` for non-zero exits. Avoids cluttering the error when stderr is empty.

### D3: Truncate stderr at maxAdapterStderrBytes in callers

Each adapter truncates stderr at 512 bytes before including it in the error, using the same `maxAdapterStderrBytes` constant and truncation pattern already used in `runSubprocess`. This prevents leaking secrets from AI CLI output.

**Rationale**: Reuses the existing security-motivated truncation constant. The truncation logic is simple enough to inline in each adapter (3 lines) rather than extracting a helper.

## Risks / Trade-offs

- **API surface change**: `runSubprocess` return signature changes from 2 to 3 values. All callers (3 adapters + tests) must be updated. This is a small, contained blast radius.
- **Stderr may contain ANSI codes**: AI CLIs (especially opencode) write ANSI escape sequences to stderr. These will appear in error messages. This is acceptable — the error is diagnostic, not user-facing prose.
- **Stderr may contain sensitive information**: The existing `maxAdapterStderrBytes` truncation limits exposure. The same risk exists for non-zero exit errors today.
