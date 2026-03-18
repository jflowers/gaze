# Design: Quality Include Unexported

## Context

The `analyze` command wires `--include-unexported` via `analyzeParams.includeUnexported` → `analysis.Options{IncludeUnexported: true}` → `analysis.Analyze` which skips `!fd.Name.IsExported()` filter at `analyzer.go:60`. The `quality` command hardcodes `IncludeUnexported: false` at `main.go:692`. The `aireport` pipeline hardcodes the same at `runner_steps.go:144,199`. The contract coverage builder hardcodes it at `contract.go:155`.

## Goals / Non-Goals

### Goals
- Add `--include-unexported` flag to `gaze quality` matching `analyze`'s pattern
- Auto-detect `package main` and set `IncludeUnexported = true` automatically
- Thread the option through all analysis paths: quality CLI, report pipeline, contract coverage builder
- All three consumers of `analysis.LoadAndAnalyze` consistently support the flag

### Non-Goals
- Adding `--include-unexported` to `gaze crap` (it uses a different analysis path via `crap.Analyze`)
- Changing the `analysis.Options` struct (it already has the field)
- Changing `quality.Options` (the filtering is upstream of quality assessment)

## Decisions

### D1: Auto-detect `package main` via `packages.Load` result

After loading the package with `loader.Load(pattern)`, check `result.Pkg.Name()`. If it equals `"main"`, set `IncludeUnexported = true` regardless of the flag value. This is done in `runQuality` (CLI path), `runQualityForPackage` (report pipeline), and `analyzePackageCoverage` (contract coverage builder).

**Rationale**: A `main` package has no exported API. Requiring `--include-unexported` for every main package analysis is user-hostile. Auto-detection makes gaze work correctly out-of-the-box.

### D2: Flag parity with `analyze`, not `crap`

The `--include-unexported` flag is added to `quality` only, matching `analyze`. The `crap` command does not have this flag because it uses `crap.Analyze` which operates differently (complexity-based, not AST function iteration). The contract coverage builder inside `crap.BuildContractCoverageFunc` will get the auto-detect behavior but not a CLI flag (it's an internal function, not a command).

### D3: Thread through report pipeline via package name check

Rather than adding a CLI flag to `gaze report`, the report pipeline's quality analysis steps (`runQualityStep`, `runQualityForPackage`) will auto-detect `package main` inline. This avoids adding a new field to `RunnerOptions` for a case that should be automatic.

## Risks / Trade-offs

- **Auto-detect may surprise users**: A user analyzing a `package main` will get unexported functions included without asking for it. This is the correct behavior but differs from `analyze`'s default. Mitigated by: logging a message when auto-detect fires.
- **More functions in quality output**: Including unexported functions increases the number of test-target pairs. For large `main` packages this may produce verbose output. Acceptable — more data is better than no data.
