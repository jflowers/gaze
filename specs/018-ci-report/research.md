# Research: AI-Powered CI Quality Report (018-ci-report)

**Date**: 2026-03-10
**Branch**: `018-ci-report`
**Spec**: [spec.md](spec.md)

## Research Questions

1. What are the exact CLI invocation patterns for `claude`, `gemini`, and `ollama`?
2. How should the system prompt be delivered to each adapter without shell injection?
3. How does the embedded prompt in the binary relate to `gaze-reporter.md` (spec 016 conflict)?
4. What is the combined JSON schema for `--format=json` mode?
5. What Go subprocess patterns should be used for safety and testability?
6. How should the `--max-crapload=0` (zero as live threshold) be implemented differently from the existing commands?

---

## Decision 1: AI Adapter Invocation Contract

**Decision**: Each adapter uses a different mechanism for system prompt delivery. The `AIAdapter` interface normalizes them behind a common Go interface.

**Rationale**: Research confirmed that the three CLIs have incompatible system prompt mechanisms. A unified interface is required to make the adapter layer testable without real AI CLIs.

### Claude Adapter

| Property | Value |
|---|---|
| Binary | `claude` |
| Headless flag | `-p` (required; without it, claude enters interactive REPL) |
| System prompt | `--system-prompt "<text>"` (replaces default); for prompts > 10 KB use temp file |
| System prompt file | `--system-prompt-file <path>` |
| Data payload | stdin (piped when `-p` is set) |
| Model override | `--model <name>` (default: claude-sonnet current) |
| Auth | `ANTHROPIC_API_KEY` env var or OAuth token in `~/.claude/` |
| JSON output | `--output-format json` returns `{"result": "...", ...}` |
| Exit codes | 0 = success; non-zero = error |

**Invocation pattern** (using `-p` with inline prompt; prompt as argument, data as stdin):
```
claude -p "<prompt>" --system-prompt "<system>" [--model <name>]
```

**Go exec pattern**:
```go
cmd := exec.CommandContext(ctx, "claude",
    "-p", userMessage,
    "--system-prompt", systemPrompt,
)
cmd.Stdin = bytes.NewReader(jsonPayload)
out, err := cmd.Output()
```

**Decision on large prompts**: `gaze-reporter.md` stripped of frontmatter is ~13 KB. This exceeds the 10 KB safe inline argument threshold. The claude adapter MUST write the system prompt to a temp file and use `--system-prompt-file`. The temp file is removed after the subprocess exits.

### Gemini Adapter

| Property | Value |
|---|---|
| Binary | `gemini` (from `@google/gemini-cli` npm package) |
| Headless flag | `-p` (triggers headless mode; also auto-activated when stdin is not a TTY) |
| System prompt | **No dedicated flag.** Best mechanism: write `GEMINI.md` to a temp directory and set `cmd.Dir`. The `GEMINI.md` file is auto-loaded as context on startup. |
| Data payload | stdin (piped when `-p` is set; stdin is appended to `-p` prompt) |
| Model override | `--model <name>` or `-m <name>` (default: `auto` → gemini-2.5-pro) |
| Auth | `GEMINI_API_KEY` env var (for CI); OAuth for interactive use |
| JSON output | `--output-format json` returns `{"response": "...", "stats": {...}}` |
| Exit codes | 0 = success; 1 = error; 42 = bad input; 53 = turn limit |

**Invocation pattern** (system prompt via `GEMINI.md` in temp dir):
```
# Write GEMINI.md to tmpDir, then:
gemini -p "" --output-format json [-m <model>]
# stdin = JSON payload
# cmd.Dir = tmpDir (where GEMINI.md lives)
```

**Implementation note**: The gemini adapter creates a temp directory, writes the system prompt as `GEMINI.md`, sets `cmd.Dir = tempDir`, and pipes the JSON data via stdin with a minimal `-p` trigger prompt. The temp directory is cleaned up after the subprocess exits.

**Alternatives considered**:
- Prepend system prompt to `-p` value: rejected because it conflates system and user prompt, reduces AI formatting quality, and creates very long flag values.
- Use `GOOGLE_API_KEY` + Vertex AI: rejected because it requires additional GCP setup; `GEMINI_API_KEY` is simpler for CI.

### Ollama Adapter

| Property | Value |
|---|---|
| Binary | `ollama` (server + CLI pair) |
| Headless mode | `ollama run <model>` reads from stdin when stdin is not a TTY |
| System prompt | **No CLI flag.** Only via: (a) Modelfile `SYSTEM` directive, (b) HTTP API `system` field |
| Data payload | stdin for `ollama run`; `prompt` field for HTTP API |
| Model override | First positional argument to `ollama run <model>` (required; no default) |
| Auth | None (local); `OLLAMA_HOST` env var for remote server |
| JSON output | CLI: none (streaming text). HTTP API: `{"stream": false}` returns JSON |
| Exit codes | 0 = success; non-zero = server down or error |

**Decision**: The ollama adapter uses the **HTTP REST API** (`/api/generate`) rather than `exec.Command("ollama", "run", ...)`. This is the correct approach because:
1. The HTTP API supports `"system"` field — the only way to deliver a proper system prompt
2. The HTTP API returns structured JSON with `"response"` field (no stream parsing needed)
3. The HTTP API is faster (no process startup overhead)
4. The CLI produces streaming token-by-token output that is harder to parse

**HTTP API call pattern**:
```go
type ollamaRequest struct {
    Model  string `json:"model"`
    System string `json:"system"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
}
// POST to http://localhost:11434/api/generate (or OLLAMA_HOST)
// Response: {"model":"...","response":"...","done":true,...}
```

**The ollama adapter does NOT use `exec.Command`** — it uses `net/http`. The `AIAdapter` interface's `Format` method signature is network-transport-agnostic, so this works cleanly.

**Alternatives considered**:
- Use `ollama run` via exec: rejected because no system prompt flag, streaming output is complex to parse, and requires full stream buffering.
- Use Modelfile-based approach: rejected because it requires pre-creating a model, which is not feasible in a one-shot CI step.

---

## Decision 2: System Prompt Content (spec 016 Conflict Resolution)

**Decision**: The binary embeds a **stripped version** of `gaze-reporter.md` — the YAML frontmatter (lines 1–14, between `---` delimiters) is removed before embedding. The stripped content is the system prompt delivered to AI CLIs.

**Rationale**: Spec 016 (agent-context-reduction) modified `gaze-reporter.md` to reference external files (`.opencode/references/example-report.md`, `.opencode/references/doc-scoring-model.md`) via the OpenCode Read tool. This tool infrastructure is unavailable when invoking an external AI CLI as a subprocess.

**Resolution**:
- The embedded default prompt in the binary is the full content of `gaze-reporter.md` **after** stripping the YAML frontmatter block (the block between the opening `---` and closing `---` at lines 1–14).
- The stripped prompt is self-contained: it contains all the formatting instructions, emoji vocabulary, and the canonical example — everything the AI needs.
- The reference file loading instructions in the prompt body (e.g., "Use the Read tool to load `.opencode/references/example-report.md`") are OpenCode-specific. When this prompt is used by an external AI CLI, the AI will not have tool access and will ignore those instructions. The formatting rules themselves are sufficient.
- The local file override (FR-005: load `.opencode/agents/gaze-reporter.md` if present) reads the file as-is and strips the YAML frontmatter using the same stripping logic before passing to the AI CLI.

**Stripping logic**: Remove content from the first line if it is `---` through the next `---` line (inclusive). This is a one-pass string operation on the raw file bytes.

**Alternatives considered**:
- Embed a separate, manually-maintained `report-prompt.md` distinct from `gaze-reporter.md`: rejected because it creates two sources of truth for the same formatting instructions, which drift over time.
- Inline the reference files at embed time: rejected because the reference files are tool-owned and change independently; inlining them creates coupling.
- Use the prompt as-is (with frontmatter): rejected because the YAML frontmatter would be passed as system prompt content to the AI CLI, confusing the model.

---

## Decision 3: `AIAdapter` Interface Design

**Decision**: A Go interface with a single method, accepting a context (for timeout), system prompt string, and data payload as `io.Reader`.

```go
// AIAdapter formats an analysis payload using an external AI CLI or API.
// The system prompt contains the formatting instructions; the payload
// contains the structured JSON data to be formatted.
type AIAdapter interface {
    Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error)
}
```

**Rationale**: A single-method interface is the smallest testable surface. The `io.Reader` payload allows large JSON payloads to be streamed without loading them into memory twice. The `context.Context` enables the AI CLI timeout (SC-005 / FR from council review) to be enforced per-invocation.

**Fake adapter for testing**:
```go
// fakeAdapter is a test double for AIAdapter.
type fakeAdapter struct {
    response string
    err      error
}
func (f *fakeAdapter) Format(_ context.Context, _, _ string, _ io.Reader) (string, error) {
    return f.response, f.err
}
```

This enables unit testing of all pipeline logic (threshold evaluation, Step Summary write, partial-failure handling, progress signals) without any real AI CLI.

---

## Decision 4: Combined `--format=json` Payload Schema

**Decision**: `gaze report --format=json` outputs a single JSON object with four top-level keys, one per analysis operation. Each value uses the exact same schema as the corresponding sub-command's `--format=json` output, with one normalization: docscan's bare array is wrapped in a key.

```json
{
  "crap": { /* crap.Report: {"scores": [...], "summary": {...}} */ },
  "quality": { /* qualityOutput: {"quality_reports": [...], "quality_summary": {...}} */ },
  "classify": { /* report.JSONReport: {"version": "...", "results": [...]} */ },
  "docscan": [ /* []docscan.DocumentFile: [{path, content, priority}] */ ],
  "errors": {
    "crap":     null,
    "quality":  null,
    "classify": null,
    "docscan":  null
  }
}
```

**The `errors` field** is always present and contains either `null` (success) or an error message string for each pipeline step. This enables `--format=json` consumers to detect partial failures programmatically.

**Rationale**: Using the existing sub-command schemas requires zero new serialization logic — each pipeline step writes into a `bytes.Buffer` via the existing `WriteJSON` functions, then the buffer contents are decoded into `json.RawMessage` fields for the combined envelope. No new struct definitions are needed for the individual pipeline sections.

**Combined envelope Go type**:
```go
// ReportPayload is the JSON output for --format=json mode.
// Each field contains the raw JSON from the corresponding analysis step.
type ReportPayload struct {
    CRAP     json.RawMessage `json:"crap"`
    Quality  json.RawMessage `json:"quality"`
    Classify json.RawMessage `json:"classify"`
    Docscan  json.RawMessage `json:"docscan"`
    Errors   PayloadErrors   `json:"errors"`
}

// PayloadErrors records any pipeline step failures.
// A nil value means the step succeeded.
type PayloadErrors struct {
    CRAP     *string `json:"crap"`
    Quality  *string `json:"quality"`
    Classify *string `json:"classify"`
    Docscan  *string `json:"docscan"`
}
```

**Alternatives considered**:
- Define new typed structs re-describing each sub-command's schema: rejected — creates duplication and potential schema drift.
- Omit failed steps entirely: rejected — consumers cannot distinguish "step not run" from "step failed". The `errors` field makes partial failure explicit.

---

## Decision 5: Threshold Flag Semantics — Zero as Live Threshold

**Decision**: Threshold flags use `*int` (pointer-to-int) rather than `int` with a zero sentinel. A nil pointer means "not provided"; a non-nil pointer (including `*0`) means "threshold is active."

**Rationale**: The existing `gaze crap` and `gaze self-check` commands use `int` with `> 0` checks to mean "threshold active." This means `--max-crapload=0` is silently ignored. The spec (FR-010) requires zero to be a valid live threshold for `gaze report`. Using `*int` flags avoids this ambiguity cleanly without changing existing command behavior.

**Cobra flag pattern for pointer-to-int**:
```go
// In reportParams:
maxCrapload     *int
maxGazeCrapload *int
minContractCoverage *int

// Cobra does not have IntVarP for *int, so use a local int + set explicitly:
var maxCrapload int
var maxCraploadSet bool
cmd.Flags().IntVar(&maxCrapload, "max-crapload", 0, "Fail if CRAPload exceeds N")
// Post-parse: if cmd.Flags().Changed("max-crapload") { p.maxCrapload = &maxCrapload }
```

The `cmd.Flags().Changed("flag-name")` API returns true if the flag was explicitly set on the command line, regardless of its value. This correctly distinguishes absent from zero.

**Alternatives considered**:
- Use `-1` as sentinel (not provided): rejected — users could accidentally pass `-1` meaning "no threshold" when they mean a tight threshold.
- Use `int` with `> 0` guard (existing pattern): rejected — breaks FR-010 requirement that zero is a valid threshold.

---

## Decision 6: Partial Report Failure Format

**Decision**: When a pipeline step fails, the analysis payload for that step is replaced with a structured error marker. The AI formatting prompt is still invoked with the partial payload. The warning format in the formatted report is the `> ⚠️` blockquote style from spec 011.

**Payload error representation** (for AI adapter):
```json
{
  "crap": null,
  "quality": {"quality_reports": [], "quality_summary": null},
  "classify": {"version": "dev", "results": []},
  "docscan": [],
  "errors": {
    "crap": "coverage profile generation failed: no test files found",
    "quality": null,
    "classify": null,
    "docscan": null
  }
}
```

**The formatting prompt instructs the AI**: When a section's data is null or empty AND the corresponding error string is non-null, emit a `> ⚠️ [Section] analysis unavailable: [error]` blockquote instead of that section's table.

**Exit code for partial failure**: Exit 0 unless a threshold is also breached. The partial report with warnings is a valid report. This aligns with FR-011: "The tool MUST NOT fail solely because one analysis step failed."

---

## Decision 7: `GITHUB_STEP_SUMMARY` Path Validation

**Decision**: Before writing to `GITHUB_STEP_SUMMARY`, validate that the path:
1. Is an absolute path (starts with `/`)
2. Is opened with `syscall.O_NOFOLLOW` to atomically refuse symlink following at the OS level

**Implementation**: Open the file with `os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW, 0644)`. If `O_NOFOLLOW` causes the open to fail because the path is a symlink (`ELOOP`), emit a warning to stderr and return nil. If the open fails for any other reason (unwritable, parent doesn't exist), emit a warning and return nil. Do not abort the command — Step Summary write failure is non-fatal per FR-008.

**Why `O_NOFOLLOW` instead of `Lstat`**: A prior design used `os.Lstat` followed by `os.OpenFile`, but this has a TOCTOU race — a symlink created between the check and the open would be followed by `OpenFile`. `O_NOFOLLOW` is an atomic guard at the `open(2)` syscall level: if the path names a symlink at open time, the kernel returns `ELOOP` without following it, eliminating the race window entirely.

**Platform note**: `O_NOFOLLOW` is available on Linux and macOS (Darwin). It is part of POSIX and the gaze release targets both platforms via GoReleaser.

**Rationale**: Satisfies the council's HIGH finding about the TOCTOU symlink write vulnerability. The `GITHUB_STEP_SUMMARY` is set by GitHub Actions and is always a regular file, so in normal use `O_NOFOLLOW` has no observable effect. It protects against misconfigured or adversarial environments.

---

## Decision 8: AI CLI Timeout

**Decision**: Add `--ai-timeout` flag (default: `10m`) to `gaze report`. The AI CLI subprocess (or HTTP request for ollama) runs under a `context.WithTimeout` derived from this value. When the timeout is exceeded, the subprocess is killed and the command exits 1 with a message like: `AI adapter 'claude' timed out after 10m0s`.

**Rationale**: Satisfies the council's HIGH finding about unbounded AI CLI subprocess execution. The 10-minute default is generous enough for large projects while preventing indefinite CI pipeline blocking.

---

## Summary of Resolved NEEDS CLARIFICATION Items

| Item | Resolution |
|---|---|
| Claude exact flag for system prompt | `--system-prompt-file <tmpfile>` (prompt > 10 KB) or `--system-prompt "<text>"` |
| Gemini system prompt mechanism | `GEMINI.md` in temp dir with `cmd.Dir` set |
| Ollama invocation | HTTP REST API (`/api/generate`) not `exec.Command` |
| `gaze-reporter.md` frontmatter conflict | Strip YAML frontmatter before using as system prompt |
| `--format=json` combined schema | Four-key envelope (`crap`, `quality`, `classify`, `docscan`) + `errors` |
| Zero-value threshold | `*int` flags + `cmd.Flags().Changed()` |
| Partial failure format | `null` section data + `errors` field; AI prompt handles warnings |
| `GITHUB_STEP_SUMMARY` path safety | `os.Lstat` validation; warn + skip on failure |
| AI CLI timeout | `--ai-timeout` flag (default 10m) via `context.WithTimeout` |
