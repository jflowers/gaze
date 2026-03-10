# Implementation Plan: AI-Powered CI Quality Report

**Branch**: `018-ci-report` | **Date**: 2026-03-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/018-ci-report/spec.md`

## Summary

Add a `gaze report` subcommand that orchestrates gaze's four analysis operations
(CRAP, quality, classification, docscan), assembles a combined JSON payload, and
pipes it to a user-specified external AI CLI (`claude`, `gemini`, or `ollama`)
with a formatting prompt derived from `gaze-reporter.md`. The formatted markdown
report is written to stdout and optionally appended to `$GITHUB_STEP_SUMMARY`
for visibility in the GitHub Actions UI. Optional threshold flags allow the step
to fail the build when CRAPload or contract coverage regresses below configured
limits.

The core production logic lives in the new `internal/aireport` package. The CLI
layer in `cmd/gaze` adds `reportParams`, `runReport()`, and `newReportCmd()` following
the existing testable-CLI pattern. All three AI adapters use distinct subprocess or
HTTP invocation strategies: claude via `exec.Command` + temp file for the system
prompt, gemini via `exec.Command` + `GEMINI.md` in a temp directory, and ollama
via `net/http` POST to `/api/generate`.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- `exec.Command` (standard library) — claude and gemini subprocess invocation
- `net/http` (standard library) — ollama HTTP API adapter
- `embed.FS` (standard library) — embedded default formatting prompt
- `github.com/spf13/cobra` (existing) — CLI command registration
- Existing internal packages: `crap`, `quality`, `analysis`, `classify`, `docscan`,
  `loader`, `taxonomy`, `config`, `report`, `scaffold`

**Storage**: Filesystem only — temp files for system prompt delivery (removed after
subprocess exits); `$GITHUB_STEP_SUMMARY` append-write; no persistent state.

**Testing**: Standard library `testing` package; `httptest.NewServer` for ollama;
`testdata/fake_claude/main.go` and `testdata/fake_gemini/main.go` fake binaries
compiled at test time for subprocess adapter tests.

**Target Platform**: Linux (CI), macOS (local). The binary already distributes for
both via GoReleaser. No platform-specific behavior introduced.

**Project Type**: Single binary CLI (existing pattern).

**Performance Goals**:
- SC-005: Gaze-owned analysis phase (excluding AI CLI round-trip) completes within
  5 minutes for a project with fewer than 50 packages on a standard CI runner.
- Threshold evaluation adds < 1 ms to total runtime (SC-003).

**Constraints**:
- No shell interpolation of user-supplied values in `exec.Command` args (FR-012
  security constraint; args must be passed as separate Go strings).
- `GITHUB_STEP_SUMMARY` write failure must not abort the command (FR-008).
- Partial pipeline failure must not abort the command (FR-011).
- AI adapter output must be non-empty/non-whitespace or command fails (FR-016).
- `--ai` flag required in `text` mode; skipped entirely in `json` mode (FR-015).
- `--model` required for ollama; optional for claude and gemini (FR-003).

**Scale/Scope**:
- Typical target: < 50 packages, standard CI runner (2–4 cores, 8 GB RAM).
- New package: `internal/aireport` (~9 files, ~500 LOC production, ~600 LOC tests).
- New code in `cmd/gaze`: ~200 LOC production, ~400 LOC tests.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Accuracy

**PASS**. The feature does not modify any analysis engine. It assembles existing
analysis outputs into a combined payload. Accuracy of the underlying CRAP,
quality, classification, and docscan analyses is unchanged. The AI formatting
layer is additive — it formats data already produced by the verified analysis
engines. No new false-positive or false-negative risk is introduced.

### II. Minimal Assumptions

**PASS**. The feature requires no annotation or restructuring of user code.
The `--ai` flag makes the adapter choice explicit. The `--model` requirement
for ollama is enforced with a clear error (no silent default). The embedded
default prompt works without any local configuration (`gaze init` not
required). The `--format=json` path requires no AI CLI at all.

### III. Actionable Output

**PASS**. The formatted report produced by `gaze report` contains the same
structural sections as the `/gaze` OpenCode command: CRAP summary, quality
summary, classification summary, and prioritized health assessment (SC-002).
Threshold failure output on stderr is explicit: `CRAPload: 13/10 (FAIL)`.
The `--format=json` mode provides machine-readable output for downstream
tooling. Progress signals (FR-017) guide the user during long-running analysis.

### IV. Testability

**PASS** — with the following coverage strategy:

| Layer | Test type | Coverage target |
|---|---|---|
| `internal/aireport` overall | Unit (standard `testing`) | ≥ 80% line |
| `adapter_claude.go` | Unit + fake subprocess | ≥ 70% line |
| `adapter_gemini.go` | Unit + fake subprocess | ≥ 70% line |
| `adapter_ollama.go` | Unit + `httptest.Server` | ≥ 70% line |
| `cmd/gaze` report command | Unit + `FakeAdapter` | ≥ 75% line |

Every function in `internal/aireport` is independently testable:
- `Run()` accepts `AnalyzeFunc` injection to bypass the real pipeline.
- All three adapters accept injectable transport (`execFunc` or `httpClient`).
- `WriteStepSummary` uses `t.TempDir()` + `t.Setenv` — no global state.
- `EvaluateThresholds` is a pure function of `ThresholdConfig` and `ReportPayload`.
- `LoadPrompt` is testable via temp directories and embedded content.

Coverage ratchets are enforced via T-036 (post-implementation coverage gate).
SC-006 (cross-adapter structural equivalence) is designated **manual verification**
per the checklist — automated CI covers all other success criteria.

**Pre-implementation constitution check violations**: None. All four principles
pass. No complexity violations or waivers required.

**Post-design re-check (after Phase 1)**: No new violations identified. The
data model (`data-model.md`) confirms all types are independently testable and
the dependency injection pattern is applied consistently across all new code.

## Project Structure

### Documentation (this feature)

```text
specs/018-ci-report/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output — all NEEDS CLARIFICATION resolved
├── data-model.md        # Phase 1 output — types, source layout, state transitions
├── quickstart.md        # Phase 1 output — GitHub Actions examples, flags reference
├── checklists/
│   └── requirements.md  # Spec quality checklist (pre-implementation gate)
└── tasks.md             # Phase 2 output — 39 tasks across 9 phases
```

### Source Code (repository root)

```text
internal/
└── aireport/                         (new package)
    ├── adapter.go                    # AIAdapter interface, NewAdapter factory, FakeAdapter
    ├── adapter_claude.go             # ClaudeAdapter — exec.Command + temp file
    ├── adapter_gemini.go             # GeminiAdapter — exec.Command + GEMINI.md temp dir
    ├── adapter_ollama.go             # OllamaAdapter — net/http POST /api/generate
    ├── runner.go                     # Run() — 4-step pipeline + AI formatting
    ├── runner_steps.go               # runCRAPStep, runQualityStep, runClassifyStep, runDocscanStep
    ├── payload.go                    # ReportPayload, PayloadErrors, ThresholdConfig, ThresholdResult
    ├── prompt.go                     # LoadPrompt() — local file load + frontmatter strip
    ├── output.go                     # WriteStepSummary() — GITHUB_STEP_SUMMARY write
    ├── threshold.go                  # EvaluateThresholds()
    ├── adapter_test.go               # FakeAdapter contract tests
    ├── adapter_claude_test.go        # ClaudeAdapter tests (fake subprocess)
    ├── adapter_gemini_test.go        # GeminiAdapter tests (fake subprocess)
    ├── adapter_ollama_test.go        # OllamaAdapter tests (httptest.Server)
    ├── runner_test.go                # Run() tests (FakeAdapter + AnalyzeFunc injection)
    ├── payload_test.go               # ReportPayload JSON round-trip tests
    ├── prompt_test.go                # LoadPrompt() tests
    ├── output_test.go                # WriteStepSummary() tests
    ├── threshold_test.go             # EvaluateThresholds() contract tests
    └── testdata/
        ├── fake_claude/
        │   └── main.go               # Fake claude binary for subprocess tests
        └── fake_gemini/
            └── main.go               # Fake gemini binary for subprocess tests

cmd/gaze/
├── main.go                           # + root.AddCommand(newReportCmd())
│                                     # + reportParams, runReport(), newReportCmd()
└── main_test.go                      # + TestSC001..SC004, T-027..T-035 tests
```

**Structure Decision**: Single project layout following the existing `cmd/gaze` +
`internal/` pattern. The new `internal/aireport` package is a peer of the existing
`internal/crap`, `internal/quality`, `internal/classify`, etc. The `cmd/gaze`
command layer remains the sole entry point, with all business logic in `internal/`.

Note: Several files under `internal/aireport/` already exist as scaffolding from
earlier planning sessions (`adapter.go`, `adapter_claude.go`, `adapter_gemini.go`,
`adapter_ollama.go`, `runner.go`, `runner_steps.go`, `payload.go`, `output.go`,
`threshold.go`). The adapter `Format` methods are unimplemented stubs. Implementation
tasks (T-001 through T-035) will complete and test these files.

## Complexity Tracking

No Constitution Check violations requiring justification. The implementation
follows existing project patterns throughout:

- `internal/aireport` mirrors the peer package structure (`internal/crap`,
  `internal/quality`, etc.).
- `cmd/gaze` command registration follows the existing `newXxxCmd()` + `runXxx()`
  + `xxxParams` pattern already present for `crap`, `quality`, `analyze`, `docscan`.
- Dependency injection (via `AnalyzeFunc`, `httpClient`, `execFunc`) follows the
  existing `analyzeFunc` and `coverageFunc` injection pattern in `crapParams`.
- The three AI adapters all implement the same `AIAdapter` interface — no
  conditional dispatch in the runner itself.

The only non-trivial design choice is the gemini `GEMINI.md` temp-directory
workaround (research.md R-002). This is required by the gemini CLI design and
is documented, isolated to `GeminiAdapter`, and testable via the fake binary.
