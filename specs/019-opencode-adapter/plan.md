# Implementation Plan: OpenCode AI Adapter for gaze report

**Branch**: `019-opencode-adapter` | **Date**: 2026-03-12 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/019-opencode-adapter/spec.md`

## Summary

Add `opencode` as a fourth AI adapter for `gaze report --ai=opencode`. The implementation follows the established adapter pattern exactly: one new Go source file (`adapter_opencode.go`), one new test file (`adapter_opencode_test.go`), one new fake binary (`testdata/fake_opencode/main.go`), and targeted updates to the adapter factory and CLI wiring. The system prompt is delivered via a temporary directory containing `.opencode/agents/gaze-reporter.md` (Gemini-style), the payload via stdin (universal), and output is read as plain text (Claude-style). No new abstractions, no new interfaces, no new flags beyond what the existing infrastructure already provides.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `os/exec` (subprocess), `os` (temp dir), `path/filepath` (agent file path), `strings` (output trimming), `bytes` (stderr buffer) — all standard library; no new external dependencies
**Storage**: N/A — ephemeral temp dir only; cleaned up via `defer os.RemoveAll`
**Testing**: Standard library `testing` package; fake binary compiled via `go build` at test time; same `testing.Short()` guard as existing adapter tests
**Target Platform**: Linux (CI), macOS (dev); Windows subprocess tests skipped (same as claude/gemini tests)
**Project Type**: Single Go binary CLI (`cmd/gaze/`)
**Performance Goals**: No new performance requirements; adapter is I/O-bound on the `opencode` subprocess (same as claude/gemini); analysis pipeline timeout inherited via `--ai-timeout` (default 10 minutes)
**Constraints**: 64 MiB output cap (existing constant); 512-byte stderr truncation (existing constant); `0600` agent file permissions; `0700` temp directory (OS default from `os.MkdirTemp`)
**Scale/Scope**: Single adapter file (~110 lines); single test file (~200 lines); single fake binary (~70 lines); 3 targeted line-edits in existing files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

### I. Accuracy — PASS

This feature adds a new adapter for formatting reports — it does not alter any side-effect detection, CRAP scoring, quality assessment, or classification logic. All four analysis pipelines are unchanged. The `opencode` adapter operates downstream of all accuracy-sensitive code; its output is a formatted presentation of data that was already correctly computed.

No false positives or false negatives are introduced. The accuracy of gaze's analysis output is unaffected.

### II. Minimal Assumptions — PASS

The feature makes no assumptions about the user's test framework, coding style, or project structure. The only new assumption is that `opencode` binary is on PATH when `--ai=opencode` is specified — this assumption is explicit (FR-007, `ValidateBinary()` pre-flight check, documented in help text and error messages). The `--model` flag remains optional; no model configuration is required.

Existing behavior for `claude`, `gemini`, and `ollama` is entirely unaffected (FR-012).

### III. Actionable Output — PASS

The `opencode` adapter produces the same formatted markdown report as all other adapters, using the same `gaze-reporter` system prompt and the same four-section structure. The output remains both human-readable (stdout) and machine-readable (`--format=json` mode bypasses all adapters, unchanged). Metrics remain comparable across runs because the analysis pipeline is unchanged.

### IV. Testability — PASS

**Coverage strategy**:
- `adapter_opencode.go`: 10 contract-level tests in `adapter_opencode_test.go` covering all observable behaviors (success, empty output, non-zero exit, binary not found, model flag, no model flag, agent file delivery, frontmatter, temp dir cleanup, context cancellation). All tests use the fake binary pattern (no real AI CLI required).
- `adapter.go` (modified): existing `TestNewAdapter_*` tests will exercise the new `"opencode"` case once added to the allowlist.
- `cmd/gaze/main.go` (modified): `TestSC006_CrossAdapterStructure` extended with `"opencode"` entry; uses `FakeAdapter` (no subprocess).
- No new global state introduced; `OpenCodeAdapter` is stateless and testable in isolation.
- `testing.Short()` guards all subprocess-compilation tests (consistent with existing pattern).
- Coverage ratchet: the existing `TestRunSelfCheck` e2e test runs gaze on itself and will exercise the updated allowlist indirectly.

Missing coverage strategy would be CRITICAL per constitution; this plan resolves it explicitly above.

## Project Structure

### Documentation (this feature)

```text
specs/019-opencode-adapter/
├── plan.md          ← this file
├── research.md      ← Phase 0 output (11 decisions documented)
├── data-model.md    ← Phase 1 output (entity model + process flow)
├── quickstart.md    ← Phase 1 output (implementation guide)
└── tasks.md         ← Phase 2 output (/speckit.tasks — not yet created)
```

### Source Code Changes

```text
internal/aireport/
├── adapter.go                          ← MODIFY (allowlist + factory + error string)
├── adapter_opencode.go                 ← CREATE (OpenCodeAdapter: ~110 lines)
├── adapter_opencode_test.go            ← CREATE (10 test cases: ~200 lines)
└── testdata/
    └── fake_opencode/
        └── main.go                     ← CREATE (fake opencode binary: ~80 lines)

cmd/gaze/
├── main.go                             ← MODIFY (--ai help, error msg, usage examples)
└── main_test.go                        ← MODIFY (TestSC006: add "opencode" to loop)
```

**Structure Decision**: Single Go project, standard library only. Follows the existing `internal/aireport/adapter_*.go` pattern exactly. No new packages, no new directories beyond `testdata/fake_opencode/`.

## Complexity Tracking

No constitution violations to justify. This feature is additive only.

## Implementation Design

### `OpenCodeAdapter` Invocation

```
opencode run \
  --dir <tmpDir> \
  --agent gaze-reporter \
  --format default \
  "" \
  [-m <model>]
  stdin: <JSON analysis payload>
  stdout: plain-text markdown report
```

- `<tmpDir>` = `os.MkdirTemp("", "gaze-opencode-*")`
- Agent file = `<tmpDir>/.opencode/agents/gaze-reporter.md`
- Agent file content = `"---\n---\n" + systemPrompt`
- `""` = empty positional message arg (headless trigger, mirrors `-p ""` for claude/gemini)
- `--format default` = plain-text stdout (no NDJSON parsing)
- `-m <model>` = optional; appended only when `cfg.Model != ""`

### Temp Directory Structure

```text
<tmpDir>/                         ← os.MkdirTemp("", "gaze-opencode-*"), 0700
└── .opencode/
    └── agents/
        └── gaze-reporter.md      ← 0600; "---\n---\n" + systemPrompt
```

`opencode run --dir <tmpDir>` resolves the agent at `<tmpDir>/.opencode/agents/gaze-reporter.md`.

### Error References

| Condition | Error string | Spec FR |
|-----------|-------------|---------|
| Binary not found (ValidateBinary) | `"opencode not found on PATH (FR-007): %w"` | FR-007 |
| Binary not found (Format defense) | `"opencode not found on PATH (FR-007): %w"` | FR-007 |
| Temp dir creation failure | `"creating temp dir for opencode agent: %w"` | FR-003 |
| Agents subdir creation failure | `"creating .opencode/agents dir: %w"` | FR-003 |
| Agent file write failure | `"writing agent file: %w"` | FR-003 |
| Stdout pipe creation failure | `"creating stdout pipe for opencode: %w"` | FR-010 |
| Subprocess start failure | `"starting opencode: %w"` | FR-002 |
| Non-zero exit | `"opencode exited with error: %w\nstderr: %s"` | FR-008 |
| Read error | `"reading opencode output: %w"` | FR-010 |
| Empty output | `"opencode returned empty output (FR-009): ensure the opencode CLI is working correctly"` | FR-009 |

### `fake_opencode` Binary Design

Accepts flags:
- `run` — expected subcommand as first positional arg (ignored after presence check)
- `--dir <path>` — verifies `.opencode/agents/gaze-reporter.md` exists there
- `--agent <name>` — ignored after verification
- `--format <value>` — ignored
- `-m <model>` — echoed in output when present
- `--exit-error` — exits 1 with stderr message
- `--empty-output` — writes nothing to stdout

On success: reads stdin, writes `"# Fake OpenCode Report\n\n🔍 CRAP Analysis\n\n📊 Quality\n\n🧪 Classification\n\n🏥 Health\n"` to stdout.

### Changes to `adapter.go`

```go
// validAdapters: add "opencode"
var validAdapters = map[string]bool{
    "claude":   true,
    "gemini":   true,
    "ollama":   true,
    "opencode": true,
}

// NewAdapter: add case and update error string
case "opencode":
    return &OpenCodeAdapter{config: cfg}, nil
// error: "must be one of \"claude\", \"gemini\", \"ollama\", or \"opencode\""
```

### Changes to `cmd/gaze/main.go`

Three targeted string updates:
1. `--ai` flag description: `"AI adapter: claude, gemini, ollama, or opencode"`
2. Required-flag error: `"--ai is required in text mode: must be one of \"claude\", \"gemini\", \"ollama\", or \"opencode\""`
3. Usage examples block: add `gaze report ./... --ai=opencode` and `gaze report ./... --ai=opencode --model=claude-3-5-sonnet`

### Changes to `cmd/gaze/main_test.go`

One line change:
```go
// Before:
for _, adapterName := range []string{"claude", "gemini", "ollama"} {
// After:
for _, adapterName := range []string{"claude", "gemini", "ollama", "opencode"} {
```

## Coverage Strategy

| Component | Test type | Guard | Target |
|-----------|-----------|-------|--------|
| `OpenCodeAdapter.ValidateBinary()` | Unit (no subprocess) | None | binary-not-found path |
| `OpenCodeAdapter.Format()` — success | Integration (fake binary) | `testing.Short()` | happy path, agent file, frontmatter, stdin |
| `OpenCodeAdapter.Format()` — model flag | Integration (fake binary) | `testing.Short()` | `-m` appended when set |
| `OpenCodeAdapter.Format()` — no model | Integration (fake binary) | `testing.Short()` | no `-m` when not set |
| `OpenCodeAdapter.Format()` — not on PATH | Unit (no subprocess) | None | FR-007 error string |
| `OpenCodeAdapter.Format()` — non-zero exit | Integration (shell wrapper) | `testing.Short()` | exit error path |
| `OpenCodeAdapter.Format()` — empty output | Integration (shell wrapper) | `testing.Short()` | FR-009 error string |
| `OpenCodeAdapter.Format()` — temp dir cleanup | Integration (fake binary) | `testing.Short()` | no temp dir leak |
| `OpenCodeAdapter.Format()` — ctx cancelled | Integration (fake binary) | `testing.Short()` | context deadline respected |
| `NewAdapter("opencode")` | Unit | None | via existing adapter_test.go or new test |
| `TestSC006_CrossAdapterStructure` opencode | Unit (FakeAdapter) | None | structural parity |

## CI Parity Gate

Before marking implementation complete, run:

```bash
go build ./...
go test -race -count=1 -short ./...
```

These are the exact commands from `.github/workflows/test.yml` (unit + integration suite). The e2e suite (`TestRunSelfCheck`) is unchanged and does not need local re-execution for this feature.

## Post-Design Constitution Re-check

All four principles re-evaluated after Phase 1 design:

- **Accuracy**: Unchanged — adapter is downstream of all analysis code. PASS.
- **Minimal Assumptions**: `--dir` and `--agent` flags are stable `opencode run` flags (confirmed by live testing). `--format default` is the documented default. All assumptions explicit. PASS.
- **Actionable Output**: Same four-section report structure. Same `--format=json` raw mode. PASS.
- **Testability**: 10 test cases specified with observable behavior contracts. No implementation details leaked into tests. `testing.Short()` guard consistent with existing pattern. Coverage strategy fully documented above. PASS.
