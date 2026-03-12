# Implementation Plan: Pass Pre-Generated Coverage Profile to gaze report

**Branch**: `020-report-coverprofile` | **Date**: 2026-03-12 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/020-report-coverprofile/spec.md`

## Summary

Add a `--coverprofile <path>` flag to `gaze report` so users can supply a pre-generated Go coverage profile instead of triggering an internal `go test` run. The underlying `crap.Analyze` already supports accepting a pre-existing profile via `Options.CoverProfile` — this plan wires that field through the CLI layer (`newReportCmd` → `reportParams` → `runReport` → `RunnerOptions` → `runCRAPStep`) with minimal additions. No new packages, no new dependencies, no behavior change for users who omit the flag.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Cobra (CLI flag registration) — existing; no new dependencies
**Storage**: N/A — the profile path is a read-only input consumed transiently during analysis
**Testing**: Standard library `testing` package; `go test -race -count=1`
**Target Platform**: Linux, macOS, Windows (wherever `gaze report` runs today)
**Project Type**: Single binary CLI (`cmd/gaze/`)
**Performance Goals**: No regression — supplying `--coverprofile` skips the `go test` subprocess, so the CRAP step is strictly faster
**Constraints**: Zero behavior change when `--coverprofile` is omitted (FR-003, SC-005)
**Scale/Scope**: 5 files modified, ~40 lines of production code, ~120 lines of test code

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Accuracy — PASS

The flag passes the user-supplied profile directly to `crap.Analyze`, which uses the same parser (`ParseCoverProfile`) as the internal generation path. CRAP scores are determined solely by the profile content — identical profile, identical scores. No new analysis logic is introduced. SC-004 (score determinism) is verifiable by test.

### II. Minimal Assumptions — PASS

The flag is optional and additive. When omitted, behavior is byte-for-byte identical to the current release. No annotation, restructuring, or setup change is required of the user. The only assumption introduced is that the user's supplied profile is in standard Go coverage format — which is explicit in the flag description (FR-007) and is the only format `go test -coverprofile` produces.

### III. Actionable Output — PASS

Error messages for invalid paths are derived from `crap.Analyze`'s existing descriptive errors (`"cover profile %q: ..."`, `"cover profile %q is a directory..."`). No new opaque error strings are introduced. The `--help` text (FR-007) and README example (FR-009) guide the user toward the correct invocation.

### IV. Testability — PASS

Coverage strategy:
- `runCRAPStep` signature change: unit test in `internal/aireport/runner_steps_test.go` — call with a real profile, assert `crapStepResult` non-nil
- CLI flag plumbing: 4 contract tests in `cmd/gaze/main_test.go` — valid path (spy `runnerFunc`), nonexistent path, directory path, unparseable content
- Subprocess elimination (SC-001): verified via spy `runnerFunc` (the `reportParams.runnerFunc` override) that captures `RunnerOptions` and asserts `opts.CoverProfile == suppliedPath`; `runnerFunc` intercepts before `aireport.Run` so no subprocess is exercised; no `testing.Short()` guard needed on T012
- Tests exercising real package loading (T013) guarded by `testing.Short()`

No coverage ratchet regression is expected; the new code paths are fully covered by the above tests.

## Project Structure

### Documentation (this feature)

```text
specs/020-report-coverprofile/
├── plan.md              # This file
├── research.md          # Phase 0 output ✓
├── data-model.md        # Phase 1 output ✓
├── quickstart.md        # Phase 1 output ✓
├── checklists/
│   └── requirements.md  # Quality gate ✓
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code Changes

```text
cmd/gaze/
└── main.go              # +coverProfile field on reportParams
                         # +coverProfile var in newReportCmd
                         # +--coverprofile flag registration
                         # +coverProfile in reportParams{} literal
                         # +CoverProfile in RunnerOptions{} literal
                         # +--coverprofile example in Long usage string

cmd/gaze/
└── main_test.go         # +TestRunReport_CoverProfile_ValidPath
                         # +TestRunReport_CoverProfile_NonexistentPath
                         # +TestRunReport_CoverProfile_DirectoryPath

internal/aireport/
└── runner.go            # +CoverProfile string field on RunnerOptions

internal/aireport/
└── runner_steps.go      # +coverProfile string param on runCRAPStep
                         # +opts.CoverProfile = coverProfile in runCRAPStep
                         # +pass coverProfile in runProductionPipeline call

internal/aireport/
└── runner_steps_test.go # +TestRunCRAPStep_WithCoverProfile

README.md                # +CI example with --coverprofile (FR-009)
```

No new files. No new packages. No new external dependencies.

## Phase 0: Research

**Status**: Complete. See [research.md](research.md).

Key decisions:
1. Thread as `string` parameter — no new structs needed
2. `runCRAPStep` gains `coverProfile string` as third positional parameter
3. Validation lives only in `crap.Analyze` — no duplication
4. `AnalyzeFunc` test override is unaffected
5. 4 targeted tests across 2 files cover all new code paths

## Phase 1: Design & Contracts

**Status**: Complete. See [data-model.md](data-model.md) and [quickstart.md](quickstart.md).

### Changed surfaces (in dependency order)

#### 1. `internal/aireport/runner_steps.go` — `runCRAPStep`

```go
// Before
func runCRAPStep(patterns []string, moduleDir string, stderr io.Writer) (*crapStepResult, error) {
    opts := crap.DefaultOptions()
    opts.Stderr = stderr
    ...
}

// After
func runCRAPStep(patterns []string, moduleDir string, coverProfile string, stderr io.Writer) (*crapStepResult, error) {
    opts := crap.DefaultOptions()
    opts.CoverProfile = coverProfile
    opts.Stderr = stderr
    ...
}
```

Call site in `runProductionPipeline` updated to pass `""` (empty = generate internally) — backward-compatible once `RunnerOptions.CoverProfile` is threaded through.

#### 2. `internal/aireport/runner.go` — `RunnerOptions`

```go
// New field added after StepSummaryPath:
// CoverProfile is the path to a pre-generated Go coverage profile.
// When non-empty, the CRAP analysis step uses this file directly
// instead of spawning go test internally (FR-001, FR-002).
// Empty string means "generate internally" (default behavior, FR-003).
CoverProfile string
```

`runProductionPipeline` call updated to `runCRAPStep(patterns, moduleDir, opts.CoverProfile, opts.Stderr)`.

#### 3. `cmd/gaze/main.go` — `reportParams` and `newReportCmd`

```go
// reportParams gets one new field:
coverProfile string

// newReportCmd gets one new local var:
var coverProfile string

// One new flag registration:
cmd.Flags().StringVar(&coverProfile, "coverprofile", "",
    "path to a pre-generated coverage profile (skips internal go test run)")

// reportParams literal gains:
coverProfile: coverProfile,

// RunnerOptions literal gains:
CoverProfile: p.coverProfile,

// Long usage string gains a --coverprofile example.
```

### Test contracts

#### `TestRunCRAPStep_WithCoverProfile` (internal/aireport/runner_steps_test.go)

```text
Setup:    Generate coverage.out via go test -coverprofile on testdata package
          (or use an existing profile from testdata/)
Action:   Call runCRAPStep([]string{"./..."}, moduleDir, profilePath, io.Discard)
Assert:   result != nil; result.CRAPload >= 0 (not an error)
Guard:    testing.Short()
```

#### `TestRunReport_CoverProfile_ValidPath` (cmd/gaze/main_test.go)

```text
Setup:    Write a minimal valid coverage profile to t.TempDir()
          Use a spy runnerFunc (reportParams.runnerFunc override) that captures
          the RunnerOptions passed to it and returns nil; also counts calls.
Action:   Call runReport(reportParams{..., coverProfile: profilePath, format: "json",
          runnerFunc: spy.fn})
Assert:   No error returned
          spy.capturedOpts.CoverProfile == profilePath   (wiring verified)
          spy.callCount == 1                             (SC-001: no double call)
Guard:    None (no real package loading; runnerFunc spy bypasses aireport.Run)
Note:     runnerFunc intercepts at the cmd/gaze layer (before aireport.Run).
          It proves reportParams→RunnerOptions wiring only. The runCRAPStep
          wiring (RunnerOptions→crap.Options.CoverProfile) is covered by T013.
```

#### `TestRunReport_CoverProfile_NonexistentPath` (cmd/gaze/main_test.go)

```text
Setup:    coverProfile = filepath.Join(t.TempDir(), "nonexistent.out")
Action:   Call runReport(reportParams{..., coverProfile: coverProfile, format: "json"})
Assert:   err != nil; err.Error() contains the path; err.Error() contains "no such file"
Guard:    None (no subprocess)
```

#### `TestRunReport_CoverProfile_DirectoryPath` (cmd/gaze/main_test.go)

```text
Setup:    coverProfile = t.TempDir() (a directory)
Action:   Call runReport(reportParams{..., coverProfile: coverProfile, format: "json"})
Assert:   err != nil; err.Error() contains "directory"
Guard:    None (no subprocess)
```

### `--help` output (FR-007)

The new flag appears under `gaze report --help` as:

```
      --coverprofile string   path to a pre-generated coverage profile (skips internal go test run)
```

### README addition (FR-009)

A new subsection under the `gaze report` documentation section:

```markdown
#### Using a pre-generated coverage profile

If your CI workflow already runs `go test -coverprofile`, pass that profile
to `gaze report` to avoid running tests twice:

```bash
go test -race -count=1 -coverprofile=coverage.out ./...
gaze report ./... --ai=claude --coverprofile=coverage.out
```
```

## Complexity Tracking

No constitution violations. No complexity justification required.

## Post-Design Constitution Re-Check

All four principles remain PASS after Phase 1 design:

- **Accuracy**: No change to analysis logic; profile parsing unchanged
- **Minimal Assumptions**: Flag is optional; empty = existing behavior
- **Actionable Output**: Error messages come from existing descriptive paths in `crap.Analyze`; `--help` and README guide correct usage
- **Testability**: 4 CLI tests in `cmd/gaze/main_test.go` + 1 step test in `runner_steps_test.go`; T012 uses `runnerFunc` spy (intercepts at cmd/gaze layer, before `aireport.Run`) — no `testing.Short()` guard needed; T013 uses real `runCRAPStep` and guards with `testing.Short()`; T013 covers the `runCRAPStep→crap.Options.CoverProfile` wiring leg that T012's spy cannot reach; no shared mutable state in tests
