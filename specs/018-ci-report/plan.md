# Implementation Plan: AI-Powered CI Quality Report

**Branch**: `018-ci-report` | **Date**: 2026-03-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/018-ci-report/spec.md`

## Summary

Add a `gaze report` subcommand that orchestrates gaze's four analysis operations, assembles a combined JSON payload, and passes it to a user-selected AI CLI (`claude`, `gemini`, or `ollama`) with a formatting prompt to produce a rich markdown quality report. The report is written to stdout and, when `$GITHUB_STEP_SUMMARY` is set, also appended to the GitHub Actions Step Summary file. Optional threshold flags cause the command to exit non-zero when quality metrics breach configured limits. The implementation introduces a new `internal/aireport` package following the project's established testable CLI pattern.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Cobra (CLI), `exec.Command` (claude/gemini subprocess), `net/http` (ollama HTTP API), `embed.FS` (embedded default prompt), existing internal packages (`crap`, `quality`, `analysis`, `classify`, `docscan`, `loader`, `taxonomy`)
**Storage**: N/A — ephemeral pipeline only; no persistent state introduced
**Testing**: Standard library `testing` package only; `go test -race -count=1`; fake AI CLI binaries in `testdata/` for subprocess adapter tests; fake HTTP server for ollama adapter tests; `FakeAdapter` struct for unit tests
**Target Platform**: darwin/linux amd64/arm64 (same as existing gaze binary)
**Performance Goals**: Analysis phase (excluding AI CLI round-trip) completes within 5 minutes for < 50 packages on a standard CI runner (SC-005)
**Constraints**: No shell interpolation of user-supplied values in subprocess invocations; `exec.Command` args always passed as distinct Go strings; system prompt > 10 KB written to temp file for claude adapter; AI CLI subprocess runs under context timeout (default 10m)
**Scale/Scope**: Single binary command; no new persistent state; new `internal/aireport` package of ~8 source files + ~8 test files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|---|---|---|
| **I. Accuracy** | PASS | `gaze report` reuses existing analysis pipeline (`crap.Analyze`, `quality.Assess`, `analysis.LoadAndAnalyze`, `docscan.Scan`) with no changes to detection logic. No new side effect detection introduced; accuracy claims are unchanged. |
| **II. Minimal Assumptions** | PASS | `--ai` flag is required (no silent auto-detection). Users control which AI CLI is installed and authenticated. Prompt falls back to embedded default when `.opencode/agents/gaze-reporter.md` is absent. The command works with any package pattern already supported by gaze. |
| **III. Actionable Output** | PASS | The formatted report (via AI adapter) produces the same Top 5 Prioritized Recommendations as the `/gaze` agent. `--format=json` provides machine-readable combined payload. Threshold flags + stderr summary make quality regressions immediately actionable in CI. |
| **IV. Testability** | PASS | `AIAdapter` interface enables `FakeAdapter` test double for all unit tests. Fake AI CLI binaries in `testdata/` enable subprocess integration tests without real AI CLIs. `runReport(reportParams)` follows the existing testable CLI pattern. Coverage strategy defined below. |

**Coverage Strategy** (Constitution Principle IV — mandatory):

| Layer | Strategy | Target |
|---|---|---|
| `internal/aireport` unit tests | `FakeAdapter`; `t.Setenv` for env vars; `t.TempDir()` for file paths | ≥ 80% line coverage |
| `internal/aireport` adapter subprocess tests | Fake CLI binaries in `testdata/fake_claude/` and `testdata/fake_gemini/` compiled during tests via `go build`; fake HTTP server for ollama | ≥ 70% line coverage on adapter files |
| `cmd/gaze` command handler | Existing `main_test.go` pattern; `FakeAdapter` injected via `runnerFunc` override | ≥ 75% line coverage on new `report` command code |
| e2e / manual | SC-006 (cross-adapter structural equivalence) verified manually per quickstart.md; SC-005 benchmark run on CI | Manual |
| Coverage ratchet | All new packages added to CI coverage check; regression treated as test failure | Enforced via `go test -coverprofile` |

**Post-design re-check**: PASS — design decisions in research.md and data-model.md do not introduce violations. The `net/http` call for ollama is wrapped in a function injectable for testing (see `OllamaAdapter.httpClient` field override pattern).

## Project Structure

### Documentation (this feature)

```text
specs/018-ci-report/
├── plan.md          # This file
├── research.md      # Phase 0 output — AI adapter CLI research, design decisions
├── data-model.md    # Phase 1 output — types, interfaces, JSON schemas, state transitions
├── quickstart.md    # Phase 1 output — user-facing usage guide + acceptance test map
├── checklists/
│   └── requirements.md
└── tasks.md         # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
  aireport/                     # NEW — AI adapter interface + pipeline runner
    adapter.go                  # AIAdapter interface, NewAdapter factory, FakeAdapter
    adapter_claude.go           # ClaudeAdapter (exec.Command)
    adapter_gemini.go           # GeminiAdapter (exec.Command + GEMINI.md temp dir)
    adapter_ollama.go           # OllamaAdapter (net/http)
    runner.go                   # Run(RunnerOptions) — orchestrates 4-step pipeline
    payload.go                  # ReportPayload, PayloadErrors, ThresholdConfig
    prompt.go                   # LoadPrompt() — local file load + YAML strip
    output.go                   # WriteStepSummary() — GITHUB_STEP_SUMMARY write
    threshold.go                # EvaluateThresholds()
    adapter_test.go             # FakeAdapter contract tests
    adapter_claude_test.go      # ClaudeAdapter unit + subprocess tests
    adapter_gemini_test.go      # GeminiAdapter unit + subprocess tests
    adapter_ollama_test.go      # OllamaAdapter unit + fake HTTP server tests
    runner_test.go              # Run() integration tests with FakeAdapter
    payload_test.go             # ReportPayload JSON round-trip + partial failure
    prompt_test.go              # LoadPrompt frontmatter stripping
    output_test.go              # WriteStepSummary with t.Setenv/t.TempDir
    threshold_test.go           # EvaluateThresholds contract tests
    testdata/
      fake_claude/
        main.go                 # Fake claude binary for subprocess tests
      fake_gemini/
        main.go                 # Fake gemini binary for subprocess tests

cmd/gaze/
  main.go                       # + reportParams, runReport(), newReportCmd()
  main_test.go                  # + TestSC001..SC004, BenchmarkReportAnalysis
```

**Structure Decision**: New `internal/aireport` package following the project's existing layered architecture. The command handler in `cmd/gaze/main.go` follows the established `XxxParams + runXxx()` pattern. No changes to any existing package except adding `newReportCmd()` to `cmd/gaze/main.go`.

## Phased Implementation

### Phase 1: Foundation — Payload Assembly and `--format=json` Mode

**Goal**: Run the 4-step pipeline and produce `ReportPayload` JSON output. No AI adapter yet.

**Files created**:
- `internal/aireport/payload.go` — `ReportPayload`, `PayloadErrors`, `ThresholdConfig`, `ThresholdResult`
- `internal/aireport/runner.go` — `Run(RunnerOptions)` calling the 4 pipeline steps; `--format=json` output path
- `internal/aireport/threshold.go` — `EvaluateThresholds()`
- `internal/aireport/payload_test.go`
- `internal/aireport/runner_test.go` (partial — json mode only)
- `internal/aireport/threshold_test.go`
- `cmd/gaze/main.go` — `reportParams`, `runReport()`, `newReportCmd()` (format=json only initially)
- `cmd/gaze/main_test.go` — `TestSC004_PartialFailure`, `TestSC003_ThresholdTiming`

**Acceptance tests passing after Phase 1**:
- SC-003 (threshold timing — zero overhead)
- SC-004 (partial failure — continues with warning)
- US2 scenarios 1–5 (threshold evaluation with FakeAdapter returning known payload)

### Phase 2: Prompt Loading

**Goal**: Load system prompt from local file or embedded default, strip YAML frontmatter.

**Files created**:
- `internal/aireport/prompt.go` — `LoadPrompt(workdir string) (string, error)`
- `internal/aireport/prompt_test.go`

**Key behavior**:
- Reads `<workdir>/.opencode/agents/gaze-reporter.md` if exists
- Strips YAML frontmatter (content between first `---` and second `---`)
- Falls back to embedded default from `embed.FS` (the scaffold asset)
- Returns the raw prompt string

**Embed approach**: `prompt.go` uses `//go:embed` to embed the content of `internal/scaffold/assets/agents/gaze-reporter.md` as the default prompt. This shares the single source of truth with the scaffold package.

### Phase 3: AI Adapter Interface + FakeAdapter

**Goal**: Define the `AIAdapter` interface, implement `FakeAdapter`, implement `NewAdapter` factory.

**Files created**:
- `internal/aireport/adapter.go` — `AIAdapter` interface, `FakeAdapter`, `AdapterConfig`, `NewAdapter(cfg AdapterConfig) (AIAdapter, error)`
- `internal/aireport/adapter_test.go` — `FakeAdapter` contract tests

**`NewAdapter` validates**: adapter name is in allowlist `{"claude", "gemini", "ollama"}`. Returns error if name is unknown (satisfies FR-002 early validation).

### Phase 4: Claude Adapter

**Goal**: Implement `ClaudeAdapter` using `exec.Command`.

**Files created**:
- `internal/aireport/adapter_claude.go` — `ClaudeAdapter`
- `internal/aireport/testdata/fake_claude/main.go` — fake binary
- `internal/aireport/adapter_claude_test.go`

**Invocation details**:
- Check `claude` on PATH via `exec.LookPath` before running analysis (FR-012)
- Write system prompt to `os.CreateTemp` file; pass path via `--system-prompt-file`
- Positional prompt for user message: `-p ""` (empty; data is in stdin)
- Pipe `ReportPayload` JSON bytes to stdin
- Read stdout as the formatted report
- Remove temp file in defer
- Context timeout enforced via `exec.CommandContext`

**Fake claude binary** (in `testdata/`): a small Go program that reads its flags and stdin, then writes a canned markdown response to stdout. Built via `go build` in test setup using `TestMain` or per-test helper.

### Phase 5: Gemini Adapter

**Goal**: Implement `GeminiAdapter` using `exec.Command` + `GEMINI.md` temp directory.

**Files created**:
- `internal/aireport/adapter_gemini.go` — `GeminiAdapter`
- `internal/aireport/testdata/fake_gemini/main.go` — fake binary
- `internal/aireport/adapter_gemini_test.go`

**Invocation details**:
- Check `gemini` on PATH via `exec.LookPath`
- Create temp directory via `os.MkdirTemp`
- Write system prompt content to `<tmpDir>/GEMINI.md`
- Set `cmd.Dir = tmpDir`
- Invoke: `gemini -p "" --output-format json [-m <model>]`
- Pipe JSON payload to stdin
- Parse `--output-format json` response: extract `response` field string
- Remove temp directory in defer

### Phase 6: Ollama Adapter

**Goal**: Implement `OllamaAdapter` using `net/http` (not `exec.Command`).

**Files created**:
- `internal/aireport/adapter_ollama.go` — `OllamaAdapter`
- `internal/aireport/adapter_ollama_test.go` — uses `httptest.NewServer`

**Invocation details**:
- Read `OLLAMA_HOST` env var (default: `http://localhost:11434`)
- `--model` is required for ollama; validate before analysis (FR-003)
- POST to `{host}/api/generate` with JSON body:
  ```json
  {"model": "<model>", "system": "<systemPrompt>", "prompt": "<payloadJSON>", "stream": false}
  ```
- Extract `response` field from JSON response
- Context timeout via `req.WithContext(ctx)`
- `httpClient` field on `OllamaAdapter` allows injection of `*http.Client` for tests

### Phase 7: Output and Step Summary

**Goal**: Implement `WriteStepSummary` and progress signal emission.

**Files created**:
- `internal/aireport/output.go` — `WriteStepSummary(path, content string) error`
- `internal/aireport/output_test.go`

**Step Summary write behavior**:
- `os.Lstat(path)` — must not fail or return a non-regular-file
- Validate path is absolute
- Open with `os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)`
- Write formatted report content
- On any validation or write error: emit warning to stderr, return nil (not an error — FR-008)

**Progress signals** (emitted to `stderr` by `runner.go`):
```
Analyzing packages... (CRAP)
Analyzing packages... (Quality)
Analyzing packages... (Classification)
Scanning documentation...
Formatting report...
Writing Step Summary...
```

### Phase 8: Integration — Wire `newReportCmd()` and Acceptance Tests

**Goal**: Register command, wire all phases, add SC-001/SC-002 acceptance tests.

**Files modified**:
- `cmd/gaze/main.go` — add `root.AddCommand(newReportCmd())`
- `cmd/gaze/main_test.go` — `TestSC001_GithubActionsReport`, `TestSC002_ReportStructure`, `BenchmarkReportAnalysis`

**TestSC001**: sets `t.Setenv("GITHUB_STEP_SUMMARY", tmpFile)`, uses `FakeAdapter` via `runnerFunc` override, verifies step summary file contains required section markers.

**TestSC002**: uses `FakeAdapter` returning a known markdown response, verifies stdout contains `🔍`, `📊`, `🧪`, `🏥` section markers.

**SC-006 verification** (manual, per quickstart.md): run `gaze report ./... --ai=claude`, `--ai=gemini`, `--ai=ollama` on the gaze module itself; verify all four section markers present in each output.

## Requirement-to-Component Mapping

| FR | Component | Phase |
|---|---|---|
| FR-001 | `runner.go` — `Run()` orchestrates 4 steps | 1 |
| FR-002 | `adapter.go` — `NewAdapter` validates allowlist; `newReportCmd` checks `--ai` present | 3 |
| FR-003 | `adapter_ollama.go` — validates `--model` required; `newReportCmd` for claude/gemini optional | 6 |
| FR-004 | `adapter_*.go` — `Format()` invokes AI CLI/API | 4–6 |
| FR-005 | `prompt.go` — `LoadPrompt()` | 2 |
| FR-006 | `runner.go` — writes to `Stdout` | 1 |
| FR-007 | `output.go` — `WriteStepSummary()` | 7 |
| FR-008 | `output.go` — warn on write failure, return nil | 7 |
| FR-009 | `threshold.go` — `EvaluateThresholds()` | 1 |
| FR-010 | `reportParams` — `*int` flags + `cmd.Flags().Changed()` | 8 |
| FR-011 | `runner.go` — per-step error capture into `PayloadErrors` | 1 |
| FR-012 | `adapter_claude.go`, `adapter_gemini.go` — `exec.LookPath` before analysis | 4–5 |
| FR-013 | `runner.go` — validate non-empty package list after load | 1 |
| FR-014 | `newReportCmd` — default pattern `./...` | 8 |
| FR-015 | `runner.go` — `--format=json` path skips AI | 1 |
| FR-016 | `runner.go` — empty output check after `Format()` | 1 |
| FR-017 | `runner.go` — progress signals to `Stderr` | 7 |

## Security Constraints (Council Finding — Required)

All AI CLI subprocess invocations **MUST** satisfy:

1. **No shell interpolation**: all arguments to `exec.Command` are separate Go strings, never concatenated into a shell command string. The `"sh"`, `"-c"` pattern is prohibited.
2. **Allowlist validation**: `--ai` flag value is validated against the exact set `{"claude", "gemini", "ollama"}` before any subprocess is spawned. Unknown values produce an immediate error.
3. **System prompt delivery**: system prompt content is never passed as an inline flag value if it may exceed safe argument length. The claude adapter MUST use `--system-prompt-file <tmpfile>`. The gemini adapter MUST use `GEMINI.md` written to a temp directory.
4. **Path validation**: `GITHUB_STEP_SUMMARY` path MUST be validated via `os.Lstat` before writing. Non-absolute paths or non-regular-file results produce a stderr warning and skip the write.
5. **Timeout**: every AI adapter invocation (subprocess or HTTP) runs under `context.WithTimeout`. Default: 10 minutes. Configurable via `--ai-timeout`.

## Constitution Check (Post-Design)

| Principle | Status | Evidence |
|---|---|---|
| **I. Accuracy** | PASS | No new detection logic. Reuses `crap.Analyze`, `quality.Assess`, `analysis.LoadAndAnalyze`, `docscan.Scan` without modification. |
| **II. Minimal Assumptions** | PASS | `--ai` required (no silent auto-detection). `--model` required for ollama only. `GITHUB_STEP_SUMMARY` detected from environment, not assumed. |
| **III. Actionable Output** | PASS | AI-formatted report includes prioritized recommendations per spec 011 voice standard. `--format=json` is machine-readable. Threshold flags + stderr summary are actionable CI signals. |
| **IV. Testability** | PASS | `AIAdapter` interface + `FakeAdapter` enables unit testing all pipeline logic without real AI CLIs. Fake subprocess binaries enable subprocess contract testing. `runReport(reportParams)` is fully unit-testable. Coverage targets defined per layer. |

## Complexity Tracking

No constitution violations requiring justification.

| Addition | Justification |
|---|---|
| New `internal/aireport` package | Required to keep AI adapter logic isolated and testable; cannot be in `cmd/gaze` (violates single-responsibility) |
| `net/http` in `OllamaAdapter` | Ollama CLI has no system prompt flag; HTTP API is the only clean interface. `http.Client` is injectable for testing. |
| Temp file / temp dir for claude/gemini | System prompt is > 10 KB; inline CLI arg risks OS arg length limits. Temp file is cleaned up in defer. No persistent state introduced. |
