# Data Model: AI-Powered CI Quality Report (018-ci-report)

**Date**: 2026-03-10
**Branch**: `018-ci-report`
**Spec**: [spec.md](spec.md) | **Research**: [research.md](research.md)

---

## Package: `internal/aireport`

New package. All production types and logic for `gaze report` live here. The `cmd/gaze/main.go` command handler delegates to this package via `reportParams` + `runReport()`.

---

## Primary Types

### `AIAdapter` (interface)

```go
// AIAdapter formats an analysis payload using an external AI CLI or API.
// Implementations must be safe to call with a context that may be cancelled.
type AIAdapter interface {
    // Format invokes the AI integration with the given system prompt and
    // JSON payload (from payload io.Reader), returning the formatted
    // markdown report or an error.
    Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error)
}
```

**Implementations** (all in `internal/aireport/adapter_*.go`):

| Type | Adapter name | Transport |
|---|---|---|
| `ClaudeAdapter` | `"claude"` | `exec.Command("claude", "-p", ..., "--system-prompt-file", ...)` |
| `GeminiAdapter` | `"gemini"` | `exec.Command("gemini", "-p", "", "--output-format", "json")` with `GEMINI.md` temp dir |
| `OllamaAdapter` | `"ollama"` | `net/http` POST to `http://localhost:11434/api/generate` |

**Test double**:
```go
// FakeAdapter is an AIAdapter for use in tests.
type FakeAdapter struct {
    Response string
    Err      error
    // Calls records invocations for assertion in tests.
    Calls []FakeAdapterCall
}

type FakeAdapterCall struct {
    SystemPrompt string
    Payload      []byte
}
```

---

### `AdapterConfig`

```go
// AdapterConfig holds the user-specified AI adapter configuration.
type AdapterConfig struct {
    // Name is the adapter identifier: "claude", "gemini", or "ollama".
    Name string

    // Model is the model name to use. Required for ollama; optional for
    // claude and gemini (uses each CLI's default when empty).
    Model string

    // Timeout is the maximum duration to wait for the AI adapter to respond.
    // Applied to the subprocess or HTTP request context.
    // Default: 10 minutes.
    Timeout time.Duration

    // OllamaHost overrides the default ollama server URL.
    // Reads from OLLAMA_HOST env var when empty.
    OllamaHost string
}
```

---

### `ReportPayload`

```go
// ReportPayload is the combined analysis data passed to the AI adapter
// and written to stdout in --format=json mode.
type ReportPayload struct {
    // CRAP holds the raw JSON from gaze crap --format=json.
    // Nil when the CRAP analysis step failed.
    CRAP json.RawMessage `json:"crap"`

    // Quality holds the raw JSON from gaze quality --format=json.
    // Nil when the quality analysis step failed.
    Quality json.RawMessage `json:"quality"`

    // Classify holds the raw JSON from gaze analyze --classify --format=json.
    // Nil when the classification step failed.
    Classify json.RawMessage `json:"classify"`

    // Docscan holds the raw JSON from gaze docscan ([]docscan.DocumentFile).
    // Nil when the docscan step failed.
    Docscan json.RawMessage `json:"docscan"`

    // Errors records step-level failures. A nil value means the step
    // succeeded. A non-nil value is the error message string.
    Errors PayloadErrors `json:"errors"`
}

// PayloadErrors records per-step failure messages.
type PayloadErrors struct {
    CRAP     *string `json:"crap"`
    Quality  *string `json:"quality"`
    Classify *string `json:"classify"`
    Docscan  *string `json:"docscan"`
}
```

**Wire JSON example** (full success):
```json
{
  "crap":     { "scores": [...], "summary": {...} },
  "quality":  { "quality_reports": [...], "quality_summary": {...} },
  "classify": { "version": "1.2.3", "results": [...] },
  "docscan":  [ { "path": "README.md", "content": "...", "priority": 2 } ],
  "errors":   { "crap": null, "quality": null, "classify": null, "docscan": null }
}
```

**Wire JSON example** (partial failure — CRAP step failed):
```json
{
  "crap":     null,
  "quality":  { "quality_reports": [...], "quality_summary": {...} },
  "classify": { "version": "1.2.3", "results": [...] },
  "docscan":  [ { "path": "README.md", "content": "...", "priority": 2 } ],
  "errors":   {
    "crap":     "coverage profile generation failed: no test files found in ./...",
    "quality":  null,
    "classify": null,
    "docscan":  null
  }
}
```

---

### `ThresholdConfig`

```go
// ThresholdConfig holds the CI gate thresholds for gaze report.
// A nil field means "not provided on command line" — threshold is disabled.
// A non-nil field with value 0 means "fail if any function exceeds threshold"
// (zero is a valid live threshold).
type ThresholdConfig struct {
    MaxCrapload         *int
    MaxGazeCrapload     *int
    MinContractCoverage *int
}

// ThresholdResult records the evaluation outcome for one threshold.
type ThresholdResult struct {
    Name    string // e.g. "CRAPload", "GazeCRAPload", "AvgContractCoverage"
    Actual  int
    Limit   int
    Passed  bool
}
```

**Threshold evaluation** (used by `cmd/gaze/main.go` after `runReport` returns):
```go
// EvaluateThresholds checks ThresholdConfig against ReportPayload summary data.
// Returns a slice of results (one per non-nil threshold) and whether all passed.
func EvaluateThresholds(cfg ThresholdConfig, payload *ReportPayload) ([]ThresholdResult, bool)
```

---

### `RunnerOptions`

```go
// RunnerOptions configures the report pipeline runner.
type RunnerOptions struct {
    // Patterns is the package pattern(s) to analyze (e.g., "./...").
    Patterns []string

    // ModuleDir is the root of the Go module being analyzed.
    ModuleDir string

    // Adapter is the AI adapter to use for formatting.
    Adapter AIAdapter

    // AdapterCfg holds adapter-specific configuration.
    AdapterCfg AdapterConfig

    // SystemPrompt is the formatting instructions for the AI adapter.
    // Loaded from .opencode/agents/gaze-reporter.md (stripped of YAML
    // frontmatter) or the embedded default prompt.
    SystemPrompt string

    // Format is "text" (default) or "json".
    Format string

    // Stdout receives the formatted report (text mode) or combined JSON
    // payload (json mode).
    Stdout io.Writer

    // Stderr receives progress signals, threshold summaries, and warnings.
    Stderr io.Writer

    // Thresholds holds the CI gate configuration.
    Thresholds ThresholdConfig

    // StepSummaryPath is the value of $GITHUB_STEP_SUMMARY, if set.
    // Empty means Step Summary output is disabled.
    StepSummaryPath string

    // AnalyzeFunc overrides the analysis pipeline for testing.
    // When nil, the production pipeline is called.
    AnalyzeFunc func(patterns []string, moduleDir string) (*ReportPayload, error)
}
```

---

## Package: `cmd/gaze` — `reportParams`

```go
// reportParams holds the parsed flags for the report command.
// Follows the existing testable CLI pattern (see crapParams, qualityParams).
type reportParams struct {
    patterns            []string
    format              string
    adapterName         string
    modelName           string
    aiTimeout           time.Duration
    maxCraploadSet      bool
    maxCrapload         int
    maxGazeCraploadSet  bool
    maxGazeCrapload     int
    minContractCovSet   bool
    minContractCoverage int
    stdout              io.Writer
    stderr              io.Writer

    // runnerFunc overrides aireport.Run for testing.
    runnerFunc func(aireport.RunnerOptions) error
}
```

---

## Source File Layout

```
internal/aireport/
  adapter.go              AIAdapter interface + NewAdapter factory + FakeAdapter
  adapter_claude.go       ClaudeAdapter — exec.Command with temp file for system prompt
  adapter_gemini.go       GeminiAdapter — exec.Command with GEMINI.md in temp dir
  adapter_ollama.go       OllamaAdapter — net/http POST to /api/generate
  runner.go               Run() orchestrates the 4-step pipeline + AI formatting
  payload.go              ReportPayload, PayloadErrors, ThresholdConfig, ThresholdResult
  prompt.go               LoadPrompt() — local file load + YAML frontmatter stripping
  output.go               WriteStepSummary() — GITHUB_STEP_SUMMARY write with validation
  threshold.go            EvaluateThresholds()
  adapter_test.go         FakeAdapter contract tests
  adapter_claude_test.go  ClaudeAdapter unit tests (fake process via testdata/)
  adapter_gemini_test.go  GeminiAdapter unit tests (fake process via testdata/)
  adapter_ollama_test.go  OllamaAdapter unit tests (fake HTTP server)
  runner_test.go          Run() integration tests using FakeAdapter
  payload_test.go         ReportPayload JSON round-trip + partial failure tests
  prompt_test.go          LoadPrompt() with frontmatter strip tests
  output_test.go          WriteStepSummary() with t.Setenv + writable/unwritable tests
  threshold_test.go       EvaluateThresholds() contract tests
  testdata/
    fake_claude/main.go   Fake claude binary for integration tests
    fake_gemini/main.go   Fake gemini binary for integration tests

cmd/gaze/
  main.go                 + reportParams, runReport(), newReportCmd()
  main_test.go            + TestSC001_GithubActionsReport, TestSC002_ReportStructure,
                            TestSC003_ThresholdTiming, TestSC004_PartialFailure,
                            TestSC005_AnalysisBenchmark (BenchmarkReportAnalysis)
```

---

## JSON Schemas Referenced

`gaze report --format=json` output reuses the existing sub-command schemas verbatim:

| Key | Schema source |
|---|---|
| `crap` | `internal/crap` — `crap.Report` struct with JSON tags |
| `quality` | `internal/quality` — `qualityOutput` wrapper struct |
| `classify` | `internal/report` — `report.JSONReport` struct |
| `docscan` | `internal/docscan` — `[]docscan.DocumentFile` bare array |
| `errors` | New: `PayloadErrors` in `internal/aireport/payload.go` |

The combined envelope is defined by `ReportPayload` in `internal/aireport/payload.go`.

---

## State Transitions

```
gaze report ./... --ai=claude
         │
         ▼
[1] Validate --ai flag → resolve AdapterConfig
         │ error: --ai missing or unknown → exit 1 (before analysis)
         ▼
[2] Validate AI CLI availability (FR-012)
         │ error: binary not on PATH → exit 1 (before analysis)  
         │ (skipped in --format=json mode per FR-015)
         ▼
[3] Load formatting prompt (FR-005)
         │ .opencode/agents/gaze-reporter.md (strip frontmatter)
         │ or embedded default
         ▼
[4] Run 4-step analysis pipeline (FR-001, FR-011)
         │ crap.Analyze → ReportPayload.CRAP
         │ quality.Assess → ReportPayload.Quality
         │ analysis.LoadAndAnalyze + classify → ReportPayload.Classify
         │ docscan.Scan → ReportPayload.Docscan
         │ each step: failure → null field + PayloadErrors entry (partial report continues)
         ▼
[5a] --format=json: write ReportPayload JSON to stdout → done (no AI)
[5b] --format=text:
         │ Emit "Formatting report..." to stderr (FR-017)
         │ adapter.Format(ctx, systemPrompt, payloadReader) → markdown string
         │ error: AI CLI not found → exit 1
         │ error: AI CLI exits non-zero → exit 1
         │ error: AI output empty/whitespace → exit 1 (FR-016)
         ▼
[6] Write markdown to stdout (FR-006)
         ▼
[7] Write to GITHUB_STEP_SUMMARY if set (FR-007/FR-008)
         │ validate path → warn + skip if invalid (not exit 1)
         ▼
[8] Evaluate thresholds (FR-009/FR-010)
         │ EvaluateThresholds(cfg, payload) → []ThresholdResult
         │ print threshold summary to stderr
         │ any FAIL → exit 1
         ▼
exit 0
```
