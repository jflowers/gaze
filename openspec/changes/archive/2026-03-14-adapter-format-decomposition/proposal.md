## Why

The three subprocess AI adapter `Format` methods (`ClaudeAdapter`, `GeminiAdapter`, `OpenCodeAdapter`) have the worst CRAP scores in the project: 104.2, 125.3, and 126.6 respectively. Each has cyclomatic complexity 11-12 with under 10% line coverage. All three are Q4 Dangerous in the GazeCRAP quadrant.

The high complexity comes from four duplicated code blocks that appear in all three methods:
1. Binary lookup via `exec.LookPath` (3 lines each)
2. Temp directory creation with deferred cleanup (4-5 lines each)
3. Subprocess pipe setup: `StdoutPipe` + stderr buffer + `Start` + `LimitReader` + `Wait` (12 lines each)
4. Error handling with stderr truncation (7-10 lines each)

These blocks are structurally identical across all three adapters but are interleaved with adapter-specific logic (system prompt delivery, arg construction, output parsing), making each method 50-75 lines with 11-12 branches.

## What Changes

Extract a shared `runSubprocess` helper that encapsulates the common subprocess execution pattern (blocks 1, 3, 4 above). Each adapter's `Format` method reduces to:
1. Create temp dir and write system prompt file (adapter-specific layout)
2. Build args (adapter-specific flags)
3. Call `runSubprocess(ctx, binary, args, cmdDir, payload)`
4. Parse output (raw string or JSON unmarshal — adapter-specific)
5. Empty-output check

Temp directory creation (block 2) stays in each adapter because the directory structure differs: Claude writes a single `prompt.md`, Gemini writes `GEMINI.md` and sets `cmd.Dir`, OpenCode creates nested `.opencode/agents/gaze-reporter.md` with YAML frontmatter.

## Capabilities

### New Capabilities
- `runSubprocess`: Unexported function in `adapter.go` that handles binary lookup, subprocess pipe setup with stdout/stderr capture, output size limiting (`maxAdapterOutputBytes`), and error formatting with stderr truncation (`maxAdapterStderrBytes`). Returns `([]byte, error)`.

### Modified Capabilities
- `ClaudeAdapter.Format`: Simplified to adapter-specific setup + `runSubprocess` call + empty check. Complexity reduced from 11 to ~5.
- `GeminiAdapter.Format`: Same simplification + JSON unmarshal. Complexity reduced from 12 to ~6.
- `OpenCodeAdapter.Format`: Same simplification. Complexity reduced from 12 to ~5.

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/aireport/adapter.go` | Add `runSubprocess` helper |
| `internal/aireport/adapter_claude.go` | Simplify `Format` to use `runSubprocess` |
| `internal/aireport/adapter_gemini.go` | Same |
| `internal/aireport/adapter_opencode.go` | Same |
| `internal/aireport/adapter_subprocess_test.go` | New tests for `runSubprocess` |
| `AGENTS.md` | Update Recent Changes |

No changes to `OllamaAdapter` (HTTP-based, architecturally different). No changes to the `AIAdapter` interface, `runner.go`, or any callers. No output format changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No changes to artifact formats or cross-hero interfaces. The refactoring is internal to the `aireport` package. The `AIAdapter` interface is unchanged. All JSON output and CLI behavior remain identical.

### II. Composability First

**Assessment**: PASS

Each adapter remains independently functional. The shared `runSubprocess` helper is a package-private implementation detail — no new external dependencies or cross-package coupling. The `AIAdapter` interface contract is unchanged.

### III. Observable Quality

**Assessment**: PASS

No changes to output formats, JSON schemas, or report structure. The refactoring improves internal code quality (lower complexity, eliminated duplication) without altering any observable outputs.

### IV. Testability

**Assessment**: PASS

The extracted `runSubprocess` function is independently testable with a simple test binary. The existing fake binary test infrastructure (`testdata/fake_claude/`, etc.) continues to work unchanged. The decomposition makes the subprocess execution contract testable in isolation, separate from each adapter's prompt delivery and output parsing logic.
