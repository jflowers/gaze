## Context

The three subprocess AI adapters (`ClaudeAdapter`, `GeminiAdapter`, `OpenCodeAdapter`) each implement a `Format(ctx, systemPrompt, payload) (string, error)` method that follows the same 7-phase pattern:

1. Binary resolution via `exec.LookPath`
2. Temp directory + system prompt file creation
3. Argument construction
4. Subprocess execution (`exec.CommandContext`, `StdoutPipe`, `Start`)
5. Output capture (`io.ReadAll` with `LimitReader`)
6. Error handling with stderr truncation
7. Output parsing + empty check

Phases 1, 4, 5, and 6 are structurally identical across all three adapters. Phases 2, 3, and 7 are adapter-specific. The duplication inflates each method to 50-75 lines with 11-12 branches.

## Goals / Non-Goals

### Goals
- Extract shared subprocess execution into `runSubprocess` helper
- Reduce each adapter `Format` method's complexity from 11-12 to ~5-6
- Make subprocess execution testable in isolation
- Eliminate duplicated code across the three adapters
- Preserve all existing adapter tests (zero regressions)

### Non-Goals
- Changing the `AIAdapter` interface
- Modifying `OllamaAdapter` (HTTP-based, shares no code with subprocess adapters)
- Changing output formats, error messages, or observable behavior
- Consolidating temp dir creation (adapter-specific layouts differ too much)
- Refactoring test helpers or fake binaries

## Decisions

### D1: `runSubprocess` signature

```go
func runSubprocess(
    ctx context.Context,
    binaryName string,
    args []string,
    cmdDir string,
    payload io.Reader,
) ([]byte, error)
```

- `binaryName`: Passed to `exec.LookPath` then `exec.CommandContext`. Each adapter passes its binary name (`"claude"`, `"gemini"`, `"opencode"`).
- `args`: Pre-built by the adapter. The helper doesn't know about adapter-specific flags.
- `cmdDir`: If non-empty, sets `cmd.Dir`. Gemini needs this (sets `cmd.Dir = tmpDir`); Claude and OpenCode pass `""`.
- `payload`: Set as `cmd.Stdin`. All three adapters pass the JSON payload reader.
- Returns `([]byte, error)`: Raw bytes from stdout. Callers convert to string (Claude/OpenCode) or unmarshal JSON (Gemini).

**Rationale**: This signature captures exactly the shared behavior. Binary lookup, pipe setup, output limiting, stderr capture, and error formatting are all inside. Adapter-specific concerns (temp dirs, args, output parsing) stay outside.

### D2: Place `runSubprocess` in `adapter.go`

Co-located with the `AIAdapter` interface, the adapter constants (`maxAdapterOutputBytes`, `maxAdapterStderrBytes`), and the `NewAdapter` factory. This keeps the shared helper near the code that defines the contract it serves.

**Rationale**: The helper is package-private and serves only the subprocess adapters. Placing it in a new file would fragment the small adapter infrastructure unnecessarily.

### D3: Binary lookup inside `runSubprocess`

`exec.LookPath` is part of the subprocess lifecycle — it validates the binary exists before attempting execution. Including it in `runSubprocess` eliminates 3 lines from each adapter and keeps the "find binary → run binary → capture output" sequence atomic.

**Rationale**: All three adapters perform the lookup identically. Separating it would require callers to do `path, err := exec.LookPath(name)` then pass `path` — more boilerplate for no benefit.

### D4: Temp dir creation stays in each adapter

The temp directory structure is adapter-specific:
- Claude: `tmpDir/prompt.md` (single file)
- Gemini: `tmpDir/GEMINI.md` (single file, `cmd.Dir = tmpDir`)
- OpenCode: `tmpDir/.opencode/agents/gaze-reporter.md` (nested structure with YAML frontmatter)

Abstracting this into the shared helper would require a complex "prompt delivery strategy" pattern that adds more complexity than it removes.

**Rationale**: The temp dir creation is 5-8 lines per adapter with different file layouts and conventions. Keeping it adapter-specific is clearer than a strategy interface.

### D5: Error messages preserve adapter identity

The error messages from `runSubprocess` include the `binaryName` parameter so errors still say "claude exited with error" or "gemini not found on PATH" rather than a generic "subprocess failed". This maintains diagnostic quality.

**Rationale**: Users need to know which binary failed. The FR references (FR-012, FR-007, FR-016, FR-009) stay in each adapter's code since they're adapter-specific requirements.

## Risks / Trade-offs

### R1: Minimal risk — pure refactoring

This is a pure internal refactoring with zero API changes. The existing fake-binary test infrastructure exercises the full subprocess lifecycle end-to-end, so regressions will be caught.

### R2: FR references shift location

The FR-specific error messages (e.g., "claude not found on PATH (FR-012)") currently live inline in each `Format` method. After decomposition, the LookPath error comes from `runSubprocess` with a generic message. Each adapter wraps the `runSubprocess` error or the LookPath is still in `runSubprocess` with the `binaryName` providing context. The FR references for empty-output checks stay in each adapter since they differ (FR-016 vs FR-009).

### R3: Testing `runSubprocess` requires a real binary

The function calls `exec.LookPath` and `exec.CommandContext`, which require a real executable on the filesystem. Tests can use `go build` to compile a minimal test binary (similar to the existing fake binary pattern) or use a shell command like `echo` or `cat` for basic tests.
