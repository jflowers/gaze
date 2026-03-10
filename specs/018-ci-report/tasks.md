# Tasks: AI-Powered CI Quality Report (Spec 018)

**Branch**: `018-ci-report` | **Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

## Summary

40 tasks across 9 phases (Phases 1–8 implementation, Phase 9 documentation and coverage gate). T-005a added for runner_steps.go step functions. Phases 4–6
(adapter implementations) are sequentially listed but are logically independent; they may be
worked in any order after Phase 3 completes. Each production-code task has a corresponding or
combined test task. Coverage targets: `internal/aireport` ≥ 80% line, adapter files ≥ 70%,
`cmd/gaze` report command ≥ 75%.

## Tasks

---

### Phase 1 — Foundation: Payload Assembly, Runner, Threshold

**Goal**: Run the 4-step pipeline and produce `ReportPayload` JSON output. No AI adapter yet.
Threshold evaluation and `--format=json` output path are fully operational after this phase.

---

- [x] T-001: Define `ReportPayload`, `ReportSummary`, `PayloadErrors`, `ThresholdConfig`, and `ThresholdResult` types
  - **Phase**: 1
  - **Files**: `internal/aireport/payload.go` (create)
  - **Depends on**: none
  - **Acceptance**: `go build ./internal/aireport/...` compiles; exported types visible via `go doc`; `ReportPayload` includes a `Summary ReportSummary` field (populated during pipeline execution) alongside the existing `json.RawMessage` fields; `ReportSummary` holds `CRAPload int`, `GazeCRAPload int`, `AvgContractCoverage int` as typed ints so `EvaluateThresholds` can read threshold-relevant values without unmarshalling raw JSON
  - **Doc Impact**: GoDoc comments on all exported types including `ReportSummary`

- [x] T-002: Implement `ReportPayload` JSON round-trip and partial-failure tests
  - **Phase**: 1
  - **Files**: `internal/aireport/payload_test.go` (create)
  - **Depends on**: T-001
  - **Acceptance**: `go test -race -count=1 ./internal/aireport/...` passes; tests verify full-success and CRAP-failure partial-payload JSON serialisation/deserialisation round-trips; `PayloadErrors` null vs. non-null field handling verified
  - **Doc Impact**: none

- [x] T-003: Implement `EvaluateThresholds()` in `threshold.go`
  - **Phase**: 1
  - **Files**: `internal/aireport/threshold.go` (create)
  - **Depends on**: T-001
  - **Acceptance**: `go build ./internal/aireport/...` compiles; function signature: `func EvaluateThresholds(cfg ThresholdConfig, payload *ReportPayload) ([]ThresholdResult, bool)`; reads `payload.Summary.CRAPload`, `payload.Summary.GazeCRAPload`, and `payload.Summary.AvgContractCoverage` directly (no JSON unmarshal — typed fields from `ReportSummary`); evaluates all three threshold fields independently
  - **Doc Impact**: GoDoc comment on `EvaluateThresholds`

- [x] T-004: Implement `EvaluateThresholds` contract tests
  - **Phase**: 1
  - **Files**: `internal/aireport/threshold_test.go` (create)
  - **Depends on**: T-003
  - **Acceptance**: Tests cover: nil threshold (disabled/pass), `*0` live threshold with zero actual (pass), `*0` live threshold with positive actual (fail), `*5` with actual < limit (pass), `*5` with actual > limit (fail), all three threshold fields independently (`CRAPload`, `GazeCRAPload`, `AvgContractCoverage`), both CRAPload thresholds breached simultaneously; tests construct `ReportPayload` with known `Summary` values; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

- [x] T-005: Implement `Run(RunnerOptions)` in `runner.go` — `--format=json` path only
  - **Phase**: 1
  - **Files**: `internal/aireport/runner.go` (create)
  - **Depends on**: T-001, T-003
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `Run` with `Format="json"` assembles `ReportPayload` from all four analysis steps (using `AnalyzeFunc` override when set), writes JSON to `Stdout`, captures per-step errors into `PayloadErrors`, emits progress signals to `Stderr`, validates non-empty package list (FR-013), returns error when zero packages matched; `--format=json` skips AI adapter entirely (FR-015)
  - **Doc Impact**: GoDoc comment on `Run` and `RunnerOptions`

- [x] T-005a: Implement the four analysis step functions in `runner_steps.go`
  - **Phase**: 1
  - **Files**: `internal/aireport/runner_steps.go` (create/complete stub)
  - **Depends on**: T-001
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `runCRAPStep`, `runQualityStep`, `runClassifyStep`, `runDocscanStep` are each implemented; each function accepts a context, package pattern, loader, and returns its result plus an error; `Run()` in `runner.go` calls these via the `AnalyzeFunc` hook or directly; all four populate fields on `ReportPayload` and the new `payload.Summary` struct; each step's error is captured in `PayloadErrors` (partial failure per FR-011, not abort); `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: GoDoc comments on all four exported step functions

- [x] T-006: Implement `runner_test.go` — `--format=json` and partial-failure scenarios
  - **Phase**: 1
  - **Files**: `internal/aireport/runner_test.go` (create)
  - **Depends on**: T-005
  - **Acceptance**: Tests cover: `--format=json` writes valid `ReportPayload` JSON to stdout; CRAP step failure yields partial payload with non-null `errors.crap`; zero-package pattern returns error before analysis; progress signals appear on stderr; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

- [x] T-007: Add `reportParams`, `runReport()`, and `newReportCmd()` skeleton to `cmd/gaze/main.go` — `--format=json` mode only
  - **Phase**: 1
  - **Files**: `cmd/gaze/main.go` (modify)
  - **Depends on**: T-005
  - **Acceptance**: `go build ./cmd/gaze` compiles; `gaze report --help` shows the subcommand with `--format`, `--ai`, `--model`, `--ai-timeout`, `--max-crapload`, `--max-gaze-crapload`, `--min-contract-coverage` flags; `gaze report ./... --format=json` runs the JSON path (AI adapter not yet wired); threshold `*int` flags use `cmd.Flags().Changed()` pattern per research.md Decision 5
  - **Doc Impact**: GoDoc comments on `reportParams` and `runReport`

- [x] T-008: Implement `TestSC003_ThresholdTiming` and `TestSC004_PartialFailure` in `cmd/gaze/main_test.go`
  - **Phase**: 1
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-007
  - **Acceptance**: `TestSC003_ThresholdTiming` verifies threshold evaluation adds < 1 ms to total runtime; `TestSC004_PartialFailure` injects a failing `AnalyzeFunc` for the CRAP step via `runnerFunc`, verifies report is produced with `> ⚠️` warning and command exits 0; `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

---

### Phase 2 — Prompt Loading

**Goal**: Load system prompt from local file or embedded default; strip YAML frontmatter.

---

- [x] T-009: Implement `LoadPrompt()` in `prompt.go`
  - **Phase**: 2
  - **Files**: `internal/aireport/prompt.go` (create)
  - **Depends on**: none (can proceed in parallel with Phase 1)
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `LoadPrompt(workdir)` reads `.opencode/agents/gaze-reporter.md` when it exists and strips YAML frontmatter (content between opening `---` and closing `---`); falls back to embedded default from `internal/scaffold/assets/agents/gaze-reporter.md` via `//go:embed`; embedded default is the same file used by the scaffold package (single source of truth)
  - **Doc Impact**: GoDoc comment on `LoadPrompt`

- [x] T-010: Implement `LoadPrompt` tests in `prompt_test.go`
  - **Phase**: 2
  - **Files**: `internal/aireport/prompt_test.go` (create)
  - **Depends on**: T-009
  - **Acceptance**: Tests cover: local file present with YAML frontmatter → stripped content returned; local file present without frontmatter → content returned as-is; local file absent → embedded default returned (non-empty string); embedded default content does not begin with `---`; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

---

### Phase 3 — AI Adapter Interface and FakeAdapter

**Goal**: Define the `AIAdapter` interface; implement `FakeAdapter`; implement `NewAdapter` factory with allowlist validation.

---

- [x] T-011: Define `AIAdapter` interface, `AdapterConfig`, `FakeAdapter`, and `NewAdapter` factory in `adapter.go`
  - **Phase**: 3
  - **Files**: `internal/aireport/adapter.go` (create)
  - **Depends on**: T-001
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `AIAdapter` interface has exactly `Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error)`; `NewAdapter` validates name against `{"claude", "gemini", "ollama"}` and returns error for unknown names (FR-002); `FakeAdapter` is exported, has `Response`, `Err`, and `Calls []FakeAdapterCall` fields; `FakeAdapter` implements `AIAdapter`
  - **Doc Impact**: GoDoc comments on `AIAdapter`, `AdapterConfig`, `FakeAdapter`, `NewAdapter`

- [x] T-012: Implement `FakeAdapter` contract tests in `adapter_test.go`
  - **Phase**: 3
  - **Files**: `internal/aireport/adapter_test.go` (create)
  - **Depends on**: T-011
  - **Acceptance**: Tests verify: `FakeAdapter` returns configured `Response` and `Err`; `Calls` slice grows with each `Format` invocation; `FakeAdapterCall` records `SystemPrompt` and full `Payload` bytes; `NewAdapter` rejects unknown name with descriptive error listing valid values; `NewAdapter` accepts all three valid names without error; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

---

### Phase 4 — Claude Adapter

**Goal**: Implement `ClaudeAdapter` using `exec.Command`; no shell interpolation; system prompt via temp file.

---

- [x] T-013: Implement fake claude binary in `testdata/fake_claude/main.go`
  - **Phase**: 4
  - **Files**: `internal/aireport/testdata/fake_claude/main.go` (create)
  - **Depends on**: none
  - **Acceptance**: `go build ./internal/aireport/testdata/fake_claude` compiles; binary reads `--system-prompt-file` flag path and stdin; writes a canned markdown response to stdout; exits 0 on normal invocation; exits non-zero when invoked with `--exit-error` flag (to test error path); writes nothing to stdout when invoked with `--empty-output` flag (to test FR-016 empty output path)
  - **Doc Impact**: none

- [x] T-014: Implement `ClaudeAdapter` in `adapter_claude.go`
  - **Phase**: 4
  - **Files**: `internal/aireport/adapter_claude.go` (create)
  - **Depends on**: T-011, T-013
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `ClaudeAdapter.Format` checks `claude` on PATH via `exec.LookPath` (FR-012); writes system prompt to `os.CreateTemp` file; invokes `exec.CommandContext(ctx, "claude", "-p", "", "--system-prompt-file", tmpPath [, "--model", model])` with args as distinct Go strings (no shell interpolation, no `sh -c`); pipes `ReportPayload` JSON to stdin; reads stdout as formatted report; removes temp file in defer; returns error when AI output is empty/whitespace (FR-016); context timeout enforced
  - **Doc Impact**: GoDoc comment on `ClaudeAdapter`

- [x] T-015: Implement `ClaudeAdapter` unit and subprocess tests in `adapter_claude_test.go`
  - **Phase**: 4
  - **Files**: `internal/aireport/adapter_claude_test.go` (create)
  - **Depends on**: T-014
  - **Acceptance**: Tests compile fake claude binary from `testdata/fake_claude` in `TestMain` or per-test helper using `exec.Command("go", "build", ...)`; test cases cover: successful invocation returns canned markdown; `exec.LookPath` failure returns error before analysis; non-zero subprocess exit returns error with stderr captured; empty output returns FR-016 diagnostic error; `--model` is passed when set; temp file is cleaned up; context cancellation kills subprocess; adapter files achieve ≥ 70% line coverage; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

---

### Phase 5 — Gemini Adapter

**Goal**: Implement `GeminiAdapter` using `exec.Command` with `GEMINI.md` in a temp directory.

---

- [x] T-016: Implement fake gemini binary in `testdata/fake_gemini/main.go`
  - **Phase**: 5
  - **Files**: `internal/aireport/testdata/fake_gemini/main.go` (create)
  - **Depends on**: none
  - **Acceptance**: `go build ./internal/aireport/testdata/fake_gemini` compiles; binary checks that `GEMINI.md` exists in its working directory; reads stdin; writes a canned `{"response": "# Fake Report\n..."}` JSON to stdout; exits 0 on normal invocation; exits non-zero when invoked with `--exit-error`; writes `{"response": ""}` when invoked with `--empty-output`
  - **Doc Impact**: none

- [x] T-017: Implement `GeminiAdapter` in `adapter_gemini.go`
  - **Phase**: 5
  - **Files**: `internal/aireport/adapter_gemini.go` (create)
  - **Depends on**: T-011, T-016
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `GeminiAdapter.Format` checks `gemini` on PATH via `exec.LookPath` (FR-012); creates temp directory via `os.MkdirTemp`; writes system prompt to `<tmpDir>/GEMINI.md`; sets `cmd.Dir = tmpDir`; invokes `exec.CommandContext(ctx, "gemini", "-p", "", "--output-format", "json" [, "-m", model])` as distinct Go strings; pipes JSON payload to stdin; parses `--output-format json` response and extracts `response` field; removes temp directory in defer; returns FR-016 error on empty/whitespace response; context timeout enforced
  - **Doc Impact**: GoDoc comment on `GeminiAdapter`

- [x] T-018: Implement `GeminiAdapter` unit and subprocess tests in `adapter_gemini_test.go`
  - **Phase**: 5
  - **Files**: `internal/aireport/adapter_gemini_test.go` (create)
  - **Depends on**: T-017
  - **Acceptance**: Tests compile fake gemini binary in `TestMain` or per-test helper; test cases cover: successful invocation returns extracted `response` string; `GEMINI.md` written to temp dir (verified via fake binary check); `exec.LookPath` failure returns error; non-zero exit returns error; empty `response` field returns FR-016 error; `--model`/`-m` flag passed when set; temp directory cleaned up after subprocess exits; context cancellation kills subprocess; adapter files achieve ≥ 70% line coverage; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

---

### Phase 6 — Ollama Adapter

**Goal**: Implement `OllamaAdapter` using `net/http` POST to `/api/generate`; injectable `http.Client` for testing.

---

- [x] T-019: Implement `OllamaAdapter` in `adapter_ollama.go`
  - **Phase**: 6
  - **Files**: `internal/aireport/adapter_ollama.go` (create)
  - **Depends on**: T-011
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `OllamaAdapter.Format` validates `--model` is non-empty (FR-003); reads `OLLAMA_HOST` env var (default `http://localhost:11434`); POSTs `{"model": "<model>", "system": "<systemPrompt>", "prompt": "<payloadJSON>", "stream": false}` to `{host}/api/generate`; extracts `response` field from JSON response body; uses `req.WithContext(ctx)` for timeout; `httpClient` field on `OllamaAdapter` allows test injection of `*http.Client`; returns FR-016 error when `response` is empty/whitespace; returns error when `--model` is absent (per FR-003: required for ollama)
  - **Doc Impact**: GoDoc comment on `OllamaAdapter`

- [x] T-020: Implement `OllamaAdapter` tests using `httptest.NewServer` in `adapter_ollama_test.go`
  - **Phase**: 6
  - **Files**: `internal/aireport/adapter_ollama_test.go` (create)
  - **Depends on**: T-019
  - **Acceptance**: Tests use `httptest.NewServer` with injected `httpClient`; test cases cover: successful POST returns extracted `response`; missing `--model` returns immediate error before HTTP call; HTTP 500 response returns error; empty `response` field returns FR-016 error; `OLLAMA_HOST` env var overrides default host; context cancellation aborts in-flight request; request body contains correct `model`, `system`, `prompt`, `stream: false` fields; adapter files achieve ≥ 70% line coverage; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

---

### Phase 7 — Output and Step Summary

**Goal**: Implement `WriteStepSummary` with path validation; wire progress signals into `runner.go`.

---

- [x] T-021: Implement `WriteStepSummary()` in `output.go`
  - **Phase**: 7
  - **Files**: `internal/aireport/output.go` (create)
  - **Depends on**: none (can proceed in parallel with Phases 4–6)
  - **Acceptance**: `go build ./internal/aireport/...` compiles; `WriteStepSummary(path, content string, stderr io.Writer) error` validates path is absolute; opens file with `os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW, 0644)` — `O_NOFOLLOW` atomically refuses to follow symlinks at open time, eliminating the TOCTOU race between any existence check and the open call; if open fails because path is a symlink (`ELOOP` / `syscall.ELOOP`), emits warning to stderr and returns nil; on any other write failure emits warning to `stderr` and returns `nil` (FR-008); the `os.Lstat` pre-check may be retained as an informational step but `O_NOFOLLOW` is the authoritative symlink guard
  - **Doc Impact**: GoDoc comment on `WriteStepSummary`

- [x] T-022: Implement `WriteStepSummary` tests in `output_test.go`
  - **Phase**: 7
  - **Files**: `internal/aireport/output_test.go` (create)
  - **Depends on**: T-021
  - **Acceptance**: Tests use `t.TempDir()` and `t.Setenv`; test cases cover: empty path → warning emitted, returns nil; relative path → warning emitted, returns nil; valid absolute path to existing file → content appended; valid absolute path to non-existent file in writable dir → file created with content; unwritable path (e.g., `/dev/full` or chmod 000 parent) → warning emitted, returns nil; symlink path → `O_NOFOLLOW` causes open to fail with `ELOOP`; warning emitted, returns nil (symlinks are rejected atomically — no TOCTOU window); `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

- [x] T-023: Wire progress signals into `runner.go` (`--format=text` path)
  - **Phase**: 7
  - **Files**: `internal/aireport/runner.go` (modify)
  - **Depends on**: T-005, T-011
  - **Acceptance**: `Run` emits the following signals to `RunnerOptions.Stderr` at the correct phase transitions (FR-017): `"Analyzing packages... (CRAP)\n"`, `"Analyzing packages... (Quality)\n"`, `"Analyzing packages... (Classification)\n"`, `"Scanning documentation...\n"`, `"Formatting report...\n"`, `"Writing Step Summary...\n"` (last only when `StepSummaryPath` is non-empty); signals go to stderr only, never to stdout; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

- [x] T-024: Implement `--format=text` path in `runner.go` — AI formatting and Step Summary write
  - **Phase**: 7
  - **Files**: `internal/aireport/runner.go` (modify)
  - **Depends on**: T-005, T-011, T-021, T-023
  - **Acceptance**: `Run` with `Format="text"` calls `Adapter.Format(ctx, systemPrompt, payloadReader)`; checks output for empty/whitespace and returns FR-016 error; writes formatted markdown to `Stdout`; calls `WriteStepSummary` when `StepSummaryPath` is non-empty; AI adapter timeout governed by `AdapterCfg.Timeout` via `context.WithTimeout`; `go test -race -count=1 ./internal/aireport/...` passes with updated `runner_test.go`
  - **Doc Impact**: none

- [x] T-025: Update `runner_test.go` to cover `--format=text` path and Step Summary write
  - **Phase**: 7
  - **Files**: `internal/aireport/runner_test.go` (modify)
  - **Depends on**: T-024, T-012
  - **Acceptance**: New test cases cover: `FakeAdapter` returning known markdown is written to stdout; `FakeAdapter` returning empty string triggers FR-016 exit; `StepSummaryPath` set to temp file → file receives content; `StepSummaryPath` set to unwritable path → warning emitted, stdout still receives report, function returns nil; progress signals appear on stderr for `--format=text`; `go test -race -count=1 ./internal/aireport/...` passes
  - **Doc Impact**: none

---

### Phase 8 — Integration: Wire `newReportCmd()` and Acceptance Tests

**Goal**: Register `gaze report` command, wire all phases together, add SC-001/SC-002 acceptance tests.

---

- [x] T-026: Wire `newReportCmd()` into Cobra root command and complete `runReport()` with full AI adapter path
  - **Phase**: 8
  - **Files**: `cmd/gaze/main.go` (modify)
  - **Depends on**: T-007, T-009, T-011, T-014, T-017, T-019, T-024
  - **Acceptance**: `root.AddCommand(newReportCmd())` added; `gaze report ./... --ai=claude` compiles and reaches the AI invocation; `runReport` validates `--ai` present when `--format=text` (FR-002); calls `aireport.NewAdapter` with allowlist enforcement; calls `aireport.LoadPrompt(workDir)`; calls `aireport.Run(RunnerOptions{...})` with all fields wired from `reportParams`; `--format=json` path skips all AI validation (FR-015); threshold flags use `cmd.Flags().Changed("max-crapload")` pattern for `*int` semantics (FR-010); `--ai-timeout` default 10m; `--min-contract-coverage` missing `--model` for ollama returns immediate error (FR-003); `gaze report --help` output is accurate; `go build ./cmd/gaze` succeeds
  - **Doc Impact**: GoDoc comments on `newReportCmd`, `runReport`, `reportParams`

- [x] T-027: Implement `TestSC001_GithubActionsReport` in `cmd/gaze/main_test.go`
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026, T-012
  - **Acceptance**: Test sets `t.Setenv("GITHUB_STEP_SUMMARY", tmpFile)` and injects `FakeAdapter` via `runnerFunc` override returning a markdown report with `🔍`, `📊`, `🧪`, `🏥` section markers; verifies `tmpFile` is non-empty and contains all four markers after the command runs; verifies command exits 0; `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

- [x] T-028: Implement `TestSC002_ReportStructure` and `TestSC006_CrossAdapterStructure` in `cmd/gaze/main_test.go`
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026, T-012
  - **Acceptance**: `TestSC002_ReportStructure` — injects `FakeAdapter` returning known markdown with `🔍`, `📊`, `🧪`, `🏥` section markers; verifies all four markers appear in stdout in order; verifies no extraneous content (threshold failure, error messages) on stderr. `TestSC006_CrossAdapterStructure` — table-driven test with adapter names `{"claude", "gemini", "ollama"}`; for each adapter name, injects `FakeAdapter` with identical payload and canned response; asserts all four emoji markers (`🔍`, `📊`, `🧪`, `🏥`) appear in the same order in stdout for every adapter name; this verifies structural equivalence without requiring real AI CLIs (SC-006 automated regression gate). `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

- [x] T-029: Implement `TestSC005_AnalysisPerformance` in `cmd/gaze/main_test.go` (SC-005)
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026
  - **Acceptance**: `TestSC005_AnalysisPerformance` is guarded by `testing.Short()` (skipped in standard `-short` suite); runs the real analysis pipeline on `./...` (the gaze module itself) with `context.WithTimeout(ctx, 5*time.Minute)`; injects `FakeAdapter` so only the analysis phase (CRAP, quality, classify, docscan) is timed — AI round-trip excluded; asserts the pipeline completes without a timeout error (`ctx.Err() == nil`); if timeout fires, test fails with a clear message: "analysis phase exceeded 5-minute SC-005 limit"; `go test -race -count=1 -run TestSC005_AnalysisPerformance -timeout 30m ./cmd/gaze/...` passes; `go test -race -count=1 -short ./cmd/gaze/...` skips it
  - **Doc Impact**: none

- [x] T-030: Validate `--format=json` mode end-to-end through `cmd/gaze` command layer
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026
  - **Acceptance**: Test runs `runReport` with `format="json"` and `AnalyzeFunc` stub returning a known payload; verifies stdout is valid JSON parseable as `ReportPayload`; verifies `--ai` flag is not required in this mode; verifies `errors` field has expected null/non-null structure; `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

- [x] T-031: Validate threshold enforcement end-to-end through `cmd/gaze` command layer (US2 scenarios 1–7)
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026, T-012
  - **Acceptance**: Seven sub-tests (matching US2 scenarios 1–7): (1) CRAPload > max-crapload → exit 1, `(FAIL)` on stderr; (2) CRAPload ≤ max-crapload → exit 0, `(PASS)` on stderr; (3) avg contract coverage < min → exit 1 with FAIL; (4) no threshold flags → exit 0 regardless; (5) `--max-crapload=0` with positive actual → exit 1 (zero is live threshold); (6) GazeCRAPload > `--max-gaze-crapload` → exit 1, `GazeCRAPload: X/Y (FAIL)` on stderr; (7) `--max-gaze-crapload=0` with positive actual → exit 1 (zero is live threshold for GazeCRAPload); all use `FakeAdapter` and `AnalyzeFunc` stub with known `payload.Summary` values; `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

- [x] T-032: Validate error path: missing `--ai` flag returns descriptive error without running analysis
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026
  - **Acceptance**: Test invokes `runReport` with empty `adapterName` and `format="text"`; verifies error returned before any analysis step runs (AnalyzeFunc not called); error message lists valid adapter names `{"claude", "gemini", "ollama"}` (FR-002); `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

- [x] T-033: Validate error path: unknown `--ai` value returns descriptive error without running analysis
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026
  - **Acceptance**: Test invokes `runReport` with `adapterName="badai"`; verifies `NewAdapter` returns error; no analysis runs; error message includes allowlist; `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

- [x] T-034: Validate error path: `--ai=ollama` without `--model` returns error immediately
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026
  - **Acceptance**: Test invokes `runReport` with `adapterName="ollama"` and empty `modelName`; verifies error returned with message `"--model is required when using ollama"` before analysis runs (FR-003); `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

- [x] T-035: Validate error path: `GITHUB_STEP_SUMMARY` set to unwritable path — stdout succeeds, warning on stderr
  - **Phase**: 8
  - **Files**: `cmd/gaze/main_test.go` (modify)
  - **Depends on**: T-026, T-012
  - **Acceptance**: Test sets `GITHUB_STEP_SUMMARY` to an invalid path (e.g., `"/nonexistent/dir/summary.md"`); injects `FakeAdapter`; verifies command exits 0; verifies stdout contains the report; verifies stderr contains a warning about Step Summary write failure; FR-008 contract confirmed; `go test -race -count=1 -short ./cmd/gaze/...` passes
  - **Doc Impact**: none

---

### Phase 9 — Documentation, Spec Commit, and Coverage Verification

**Goal**: Ensure all documentation is updated, coverage targets are met, and spec artifacts are committed before implementation begins.

---

- [x] T-036: Verify coverage targets for all new packages
  - **Phase**: 9 (post-implementation gate)
  - **Files**: none (test execution only)
  - **Depends on**: T-001 through T-035
  - **Acceptance**: `go test -race -count=1 -coverprofile=coverage.out ./internal/aireport/... ./cmd/gaze/...` followed by `go tool cover -func=coverage.out`; `internal/aireport` overall ≥ 80% line coverage; `adapter_claude.go`, `adapter_gemini.go`, `adapter_ollama.go` each ≥ 70% line coverage; new `report` command code in `cmd/gaze/main.go` ≥ 75% line coverage; any shortfall must be addressed with targeted tests before marking this task complete
  - **Doc Impact**: none

- [x] T-037: Update `README.md` with `gaze report` subcommand documentation
  - **Phase**: 9
  - **Files**: `README.md` (modify)
  - **Depends on**: T-026
  - **Acceptance**: README contains: `gaze report` in the commands table or subcommand list; at minimum one example of `gaze report ./... --ai=claude`; flags reference (`--ai`, `--model`, `--ai-timeout`, `--format`, `--max-crapload`, `--max-gaze-crapload`, `--min-contract-coverage`); GitHub Actions usage example; note about `GITHUB_STEP_SUMMARY` integration; `gaze report --format=json` example; `go build ./cmd/gaze` still passes after README change
  - **Doc Impact**: README.md

- [x] T-038: Update `AGENTS.md` Recent Changes and Active Technologies sections
  - **Phase**: 9
  - **Files**: `AGENTS.md` (modify)
  - **Depends on**: T-037
  - **Acceptance**: `AGENTS.md` Recent Changes entry for `018-ci-report` describes the new `gaze report` subcommand, `internal/aireport` package, and the three AI adapter integrations; Active Technologies entry reflects `net/http` for ollama and `exec.Command` for claude/gemini; Architecture section updated with `internal/aireport/` package description; `go build ./cmd/gaze` still passes
  - **Doc Impact**: AGENTS.md

- [x] T-039: Commit spec artifacts and push branch before implementation begins (Spec Commit Gate)
  - **Phase**: 9 (pre-implementation gate — must be done BEFORE T-001 through T-036 if not already done)
  - **Files**: `specs/018-ci-report/tasks.md` (this file), all other spec artifacts under `specs/018-ci-report/`
  - **Depends on**: none (this task must precede all implementation tasks)
  - **Acceptance**: `git log --oneline` shows a commit containing `specs/018-ci-report/tasks.md` and all other spec artifacts pushed to remote `018-ci-report` branch before any production code is written; commit message follows conventional commits format (e.g., `docs: add tasks.md for spec 018-ci-report`)
  - **Doc Impact**: spec artifacts only (no production code)

---

## Dependency Graph (summary)

```
T-039 (spec commit gate) ──────────────────────────────────────► all implementation tasks

Phase 1 (foundation):
T-001 ──► T-002 (payload tests)
T-001 ──► T-003 ──► T-004 (threshold tests)
T-001 ──► T-005a (runner_steps.go)
T-001 + T-003 ──► T-005 ──► T-006 (runner json tests)
T-005 + T-005a ──► T-007 ──► T-008

Phase 2 (prompt loading):
T-009 ──► T-010                               (independent of Phase 1)

Phase 3 (adapter interface):
T-001 ──► T-011 ──► T-012

Phase 4 (claude):
T-013 (fake binary, independent)
T-011 + T-013 ──► T-014 ──► T-015

Phase 5 (gemini):
T-016 (fake binary, independent)
T-011 + T-016 ──► T-017 ──► T-018

Phase 6 (ollama):
T-011 ──► T-019 ──► T-020

Phase 7 (output + text path):
T-021 ──► T-022                               (independent of Phases 4–6)
T-005 + T-011 ──► T-023
T-005 + T-011 + T-021 + T-023 ──► T-024
T-024 + T-012 ──► T-025

Phase 8 (integration):
T-007 + T-009 + T-011 + T-014 + T-017 + T-019 + T-024 ──► T-026
T-026 + T-012 ──► T-027, T-028, T-031, T-032, T-033, T-034, T-035
T-026 ──► T-029 (TestSC005_AnalysisPerformance), T-030

Phase 9 (docs + coverage gate):
T-001..T-035 ──► T-036
T-026 ──► T-037
T-037 ──► T-038
```

## Coverage Summary

| Layer | Target | Task(s) |
|---|---|---|
| `internal/aireport` unit | ≥ 80% line | T-002, T-004, T-005a, T-006, T-010, T-012, T-022, T-025 |
| `adapter_claude.go` | ≥ 70% line | T-015 |
| `adapter_gemini.go` | ≥ 70% line | T-018 |
| `adapter_ollama.go` | ≥ 70% line | T-020 |
| `cmd/gaze` report code | ≥ 75% line | T-008, T-027–T-035 |
| Coverage gate | enforced | T-036 |
| SC-005 performance | automated (timeout) | T-029 (`TestSC005_AnalysisPerformance`) |
| SC-006 cross-adapter structure | automated (`FakeAdapter`) | T-028 (`TestSC006_CrossAdapterStructure`) |
