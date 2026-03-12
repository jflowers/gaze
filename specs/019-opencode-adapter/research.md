# Research: OpenCode AI Adapter for gaze report

**Branch**: `019-opencode-adapter`
**Date**: 2026-03-12
**Plan phase**: Phase 0

## Decision 1: Adapter File Structure — New file `adapter_opencode.go`

**Decision**: Create `internal/aireport/adapter_opencode.go` as a standalone file, mirroring the pattern established by `adapter_claude.go` and `adapter_gemini.go`.

**Rationale**: Each existing adapter is a single self-contained file. Following this convention keeps the package uniform and makes code review straightforward — the reviewer sees a single new file with a predictable structure.

**Alternatives considered**:
- Inlining into `adapter.go` — rejected; violates single-responsibility and makes the factory file grow unbounded with each new adapter.
- Splitting into multiple files — rejected; the adapter is simple enough (one struct, two methods) that splitting adds no value.

---

## Decision 2: System Prompt Delivery — Temp dir with `.opencode/agents/gaze-reporter.md`

**Decision**: Create `os.MkdirTemp("", "gaze-opencode-*")`, create `.opencode/agents/` subdirectory inside it, write the system prompt as `.opencode/agents/gaze-reporter.md` at `0600`, pass `--dir <tmpDir>` and `--agent gaze-reporter` to `opencode run`. Clean up via `defer os.RemoveAll(tmpDir)`.

**Rationale**: OpenCode reads its agent system prompt from `.opencode/agents/<name>.md` relative to the directory specified by `--dir`. This is the only mechanism the CLI exposes for injecting a system prompt from outside. The pattern mirrors `GeminiAdapter` (writes `GEMINI.md` to a temp dir and uses `cmd.Dir`), with the difference that OpenCode uses `--dir` flag rather than `cmd.Dir` to specify the project directory — this is preferred because it avoids coupling the adapter's working directory to the temp dir, leaving `cmd.Dir` at the default (module root or working directory of the `gaze` process).

**Alternatives considered**:
- Using `cmd.Dir = tmpDir` — evaluated but rejected. The `gemini` CLI happens to use `cmd.Dir` because it reads system context from the current working directory; `opencode run` uses `--dir` explicitly as a flag, so setting `cmd.Dir` would change the subprocess's working directory without providing the intended benefit.
- Reusing the user's CWD `.opencode/agents/gaze-reporter.md` — rejected; this bypasses the `LoadPrompt` abstraction (which may load an embedded default or a different file), and it would fail for users who haven't run `gaze init`.
- Passing system prompt inline as a positional arg — rejected; the system prompt is ~13 KB and would exceed shell argument length limits; also, opencode has no `--system-prompt-file` equivalent.

**Agent file format**: The file is written with empty YAML frontmatter (`---\n---\n`) prepended to the stripped system prompt body. This ensures opencode's agent parser recognizes the file as a valid agent definition regardless of parser version, while keeping the prompt content intact.

**File permissions**: `0600` on the agent file (owner read/write only); the temp directory itself is `0700` (default from `os.MkdirTemp`). This protects proprietary prompt instructions.

---

## Decision 3: Payload Delivery — stdin pipe

**Decision**: Pipe the JSON analysis payload to the `opencode run` subprocess via `cmd.Stdin`.

**Rationale**: All three existing adapters (claude, gemini, ollama) deliver the payload via stdin. This is consistent, avoids creating an additional temp file, and respects `opencode run`'s behavior of reading stdin as supplemental context when a positional message argument is provided.

**Alternatives considered**:
- `-f <file>` attachment flag — available in `opencode run` but adds complexity (another temp file, another cleanup path) with no benefit. Rejected.
- Positional message argument — OS argument length limits (~128 KB on macOS, ~2 MB on Linux) make this unreliable for large payloads. Rejected.

---

## Decision 4: Invocation Command Line

**Decision**: Invoke `opencode run --dir <tmpDir> --agent gaze-reporter --format default "" [--model <name>]`

**Rationale**:
- `run` — the non-interactive subcommand (equivalent of `claude -p` or `gemini -p`)
- `--dir <tmpDir>` — points opencode at the temp dir containing `.opencode/agents/gaze-reporter.md`
- `--agent gaze-reporter` — selects the agent by name, matching the file basename without extension
- `--format default` — requests plain-text stdout (not NDJSON event stream); avoids parsing complexity; mirrors `ClaudeAdapter` behavior
- `""` — empty string positional message argument, triggering non-interactive headless mode (mirrors `-p ""` for claude/gemini); payload arrives via stdin
- `-m <name>` — optional model flag, appended only when `cfg.Model != ""`

Confirmed by live testing: `opencode run --format default ""` with payload on stdin produces plain markdown to stdout and exits 0.

**Alternatives considered**:
- Omitting the `""` positional arg — untested; the `run` subcommand may wait for interactive input without a message arg. The empty string is the safest signal for headless mode.
- `--format json` — requires NDJSON event stream parsing; higher complexity with no benefit for this use case. Rejected.

---

## Decision 5: Output Handling — Plain text, no parsing

**Decision**: Read raw stdout bytes from the subprocess (bounded by `io.LimitReader` at 64 MiB) and return them as the formatted report string. No JSON parsing required.

**Rationale**: `--format default` produces plain text directly. This mirrors `ClaudeAdapter` (no parsing) and is simpler than `GeminiAdapter` (which parses a JSON envelope). The opencode output is already formatted markdown.

**Alternatives considered**:
- Parsing NDJSON event stream (`--format json`) — adds complexity for text accumulation with no benefit. Ruled out of scope in spec.

---

## Decision 6: Binary Validation — `exec.LookPath("opencode")`

**Decision**: `OpenCodeAdapter` implements `AdapterValidator` via a `ValidateBinary()` method that calls `exec.LookPath("opencode")`. This is called by `ValidateAdapterBinary()` in `runner.go` before the analysis pipeline begins.

**Rationale**: Identical to `ClaudeAdapter` and `GeminiAdapter`. Pre-flight check gives the user an immediate error before multi-minute analysis completes.

---

## Decision 7: Model Flag — Optional, passed as `-m <name>`

**Decision**: When `cfg.Model != ""`, append `-m <name>` to the `opencode run` args. When empty, pass no model flag and let opencode use its configured default.

**Rationale**: OpenCode uses `-m` (short form) for model selection, consistent with other CLIs. Making it optional matches the spec (FR-006) and is consistent with how claude and gemini handle the model flag.

---

## Decision 8: Error Reference Numbers — Use FR-007 (binary not found) and FR-009 (empty output)

**Decision**: Error messages reference spec FRs from spec 019: `FR-007` for binary-not-found and `FR-009` for empty output. This is consistent with how spec 018 FRs are cited in the existing adapters (e.g., `FR-012`, `FR-016` from spec 018).

**Note**: The existing adapters cite `FR-012` and `FR-016` from spec 018. For the opencode adapter, the analogous requirements are `FR-007` (binary check) and `FR-009` (empty output) from spec 019. Error messages in the new adapter will reference these numbers.

---

## Decision 9: Test Strategy — Fake binary pattern, same as claude/gemini

**Decision**: Create `internal/aireport/testdata/fake_opencode/main.go` as a fake `opencode` binary compiled at test time via `go build`. Tests in `adapter_opencode_test.go` use shell wrappers (same technique as claude/gemini tests) to inject `--exit-error` and `--empty-output` flags without modifying the adapter itself.

**Rationale**: Exact same pattern already proven for claude and gemini. Provides deterministic subprocess behavior without real AI CLIs. Guarded by `testing.Short()` (subprocess compilation is slow).

**Test cases**:
1. `TestOpenCodeAdapter_SuccessfulInvocation` — normal happy path; asserts report contains expected text
2. `TestOpenCodeAdapter_AgentFileWrittenToTempDir` — fake_opencode verifies `.opencode/agents/gaze-reporter.md` exists in `--dir`; a passing invocation proves file delivery
3. `TestOpenCodeAdapter_FrontmatterWritten` — fake_opencode reads the agent file and asserts `---` prefix is present
4. `TestOpenCodeAdapter_ModelFlagPassed` — cfg.Model="test-model"; asserts output contains model name
5. `TestOpenCodeAdapter_NoModelFlag_Succeeds` — cfg.Model=""; asserts invocation succeeds (no -m flag passed)
6. `TestOpenCodeAdapter_NotOnPath_ReturnsError` — empty PATH; asserts FR-007 in error message
7. `TestOpenCodeAdapter_NonZeroExit_ReturnsError` — shell wrapper adds --exit-error; asserts error
8. `TestOpenCodeAdapter_EmptyOutput_ReturnsError` — shell wrapper adds --empty-output; asserts FR-009 in error
9. `TestOpenCodeAdapter_TempDirCleanedUp` — counts `gaze-opencode-*` entries before/after; asserts no leak
10. `TestOpenCodeAdapter_ContextCancellation` — pre-cancelled ctx; asserts error returned

**SC-005 extension**: The `TestSC006_CrossAdapterStructure` test in `cmd/gaze/main_test.go` must be extended with an `"opencode"` table entry alongside `"claude"`, `"gemini"`, `"ollama"`.

---

## Decision 10: `adapter.go` and `cmd/gaze/main.go` Changes — Minimal, targeted

**Changes to `internal/aireport/adapter.go`**:
- Add `"opencode": true` to `validAdapters` map
- Add `case "opencode": return &OpenCodeAdapter{config: cfg}, nil` to `NewAdapter()` switch
- Update error string in `NewAdapter()` to list `"opencode"` alongside the others

**Changes to `cmd/gaze/main.go`**:
- Update `--ai` flag help text: `"AI adapter: claude, gemini, ollama, or opencode"`
- Update `--ai` required error message: include `"opencode"` in the list
- Add usage example: `gaze report ./... --ai=opencode`

**No changes** to `runner.go`, `runner_steps.go`, `payload.go`, `threshold.go`, `output.go`, or any other existing files.

---

## Decision 11: `ValidateBinary` Error Reference — `FR-007`

The existing adapters cite `FR-012` from spec 018 in their `ValidateBinary` errors. The OpenCode adapter will cite `FR-007` from spec 019, which is the analogous requirement. The `runner.go` call site (`ValidateAdapterBinary`) is unchanged — it remains adapter-agnostic.
