# Quality Include Unexported

## Why

The `quality` command is unusable for `package main` CLI tools because it hardcodes `IncludeUnexported: false` in `analysis.Options` (`cmd/gaze/main.go:692`). Both `analyze` and `crap` support `--include-unexported`, but `quality` — the command that provides contract coverage, gaze's most differentiated metric — does not.

For CLI tools where all functions are unexported (which is every `package main` by definition), this means contract coverage is completely unavailable. A user tested gaze against a 59-function Go CLI tool with 830 lines of tests and `quality` returned "no functions found to analyze."

Additionally, `package main` has no exported API by definition. Gaze should auto-detect this case and include unexported functions automatically, so users don't need to remember the flag.

Closes #70.

## What Changes

1. Add `--include-unexported` flag to `gaze quality` (flag parity with `analyze`)
2. Auto-detect `package main` and include unexported functions by default in that case
3. Thread the option through `gaze report` pipeline (`runQualityStep`, `runQualityForPackage`) and contract coverage builder (`analyzePackageCoverage`) so `gaze report` on `package main` also works

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `gaze quality`: Gains `--include-unexported` flag
- `runQuality` in `cmd/gaze/main.go`: Wires flag to `analysis.Options.IncludeUnexported`
- Auto-detection: When the analyzed package is `package main`, `IncludeUnexported` is set to `true` automatically in all analysis paths

### Removed Capabilities
- None

## Impact

- `cmd/gaze/main.go` — add flag to `qualityParams`, register in `newQualityCmd`, wire in `runQuality`, add auto-detect
- `internal/aireport/runner_steps.go` — thread `IncludeUnexported` through `runQualityStep` and `runQualityForPackage`
- `internal/crap/contract.go` — thread `IncludeUnexported` through `analyzePackageCoverage`
- `internal/analysis/analyzer.go` — no changes (already supports `IncludeUnexported`)

## Constitution Alignment

### I. Autonomous Collaboration — N/A
### II. Composability First — PASS (no new dependencies)
### III. Observable Quality — PASS (enables contract coverage for `package main`, previously invisible)
### IV. Testability — PASS (flag is testable via existing `qualityParams` pattern; auto-detect testable with `package main` fixture)
