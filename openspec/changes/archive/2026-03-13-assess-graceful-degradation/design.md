## Context

`quality.Assess` (`internal/quality/quality.go:54`) is the entry point for the test quality pipeline. It takes classified analysis results and a loaded test package, builds SSA, infers targets via call graph analysis, maps assertions to side effects, and returns `QualityReport` entries and a `PackageSummary`.

When `BuildTestSSA` fails (line 92-95), `Assess` returns `(nil, nil, error)` — a hard failure with zero results. Three callers handle this:

1. `runQuality` (`cmd/gaze/main.go:877`) — wraps and returns the error, CLI exits non-zero.
2. `analyzePackageCoverage` (`cmd/gaze/main.go:593`) — logs at debug level, returns nil silently.
3. `runQualityForPackage` (`internal/aireport/runner_steps.go:139`) — returns nil silently.

The analysis pipeline (`analysis.Analyze`) already handles SSA failures gracefully — when `BuildSSA` returns nil, mutation analysis is skipped and AST-based results are still returned. This design brings the quality pipeline into alignment.

## Goals / Non-Goals

### Goals
- `quality.Assess` returns partial (degraded) results instead of an error when `BuildTestSSA` fails
- Degraded results include test function enumeration, assertion detection, and detection confidence (all AST-only)
- A machine-readable `SSADegraded` field on `PackageSummary` signals partial results to consumers
- All three callers updated to handle degraded results instead of nil/error
- `QualitySchema` updated to include the new field
- `gaze quality` CLI exits 0 with a warning on SSA failure, printing available results

### Non-Goals
- Name-based target inference fallback (e.g., `TestFoo` → `Foo`) — this introduces a second inference path with different accuracy characteristics and warrants its own spec
- Changes to `BuildTestSSA` error/panic behavior — that was addressed in spec 021
- Changes to `analysis.BuildSSA` or the analysis pipeline — already handles degradation correctly
- Support for partial SSA (e.g., SSA built for some packages but not others within a single `Assess` call) — `Assess` operates on a single package

## Decisions

### D1: Degraded reports omit target and mapping data

Without SSA, `InferTargets` (which walks SSA basic blocks) and `MapAssertionsToEffects` (which traces SSA data flow) cannot operate. Degraded reports will have:

- `TestFunction` / `TestLocation` — populated (AST)
- `AssertionDetectionConfidence` — populated (AST via `DetectAssertions`)
- `TargetFunction` — zero-valued `FunctionTarget{}` (no inference possible)
- `ContractCoverage` — zero-valued (no mapping possible)
- `OverSpecification` — zero-valued
- `UnmappedAssertions` / `AmbiguousEffects` — nil

**Rationale**: Returning honestly zero-valued data with `SSADegraded: true` is more useful than returning nothing. Consumers can enumerate which tests exist and how many assertions they contain. The `SSADegraded` flag prevents consumers from treating zero coverage as a real measurement. This aligns with Observable Quality — the output is self-describing about its limitations.

### D2: SSADegraded on PackageSummary only, not on individual QualityReport

SSA failure is package-scoped — if `BuildTestSSA` fails, it fails for the entire package. Every report in the package is degraded. Putting the flag on `PackageSummary` avoids redundant per-report fields and keeps the data model clean.

**Rationale**: Composability — the per-report struct stays focused on test-target pair metrics. Package-level metadata belongs on the package-level summary.

### D3: Assess returns nil error on SSA failure

When SSA fails, `Assess` returns `(reports, summary, nil)` — not an error. The degradation is communicated through `summary.SSADegraded = true` and the warning on `opts.Stderr`. This means callers don't need to change their error-handling semantics (except `runQuality`, which currently propagates the error as a CLI exit code).

**Rationale**: An SSA failure is not a user error or a bug in the input — it's a toolchain limitation. Treating it as an error forces callers to decide between ignoring the error (losing data) or surfacing it (blocking the user). Treating it as degraded output lets callers display what they have.

### D4: Degraded loop produces one report per test function

In the normal path, the loop is `for _, tf := range testFuncs { for _, target := range InferTargets(...) { ... } }` — one report per test-target pair. In degraded mode, there are no inferred targets, so the loop produces one report per test function with a zero-valued target.

**Rationale**: This preserves the count of test functions in `PackageSummary.TotalTests`, which is still useful for understanding test density even without SSA.

### D5: JSON Schema adds ssa_degraded as optional boolean

The `ssa_degraded` field is added to the `PackageSummary` schema definition as an optional boolean (not in `required`). This maintains backward compatibility — existing consumers that don't know about the field will ignore it per standard JSON behavior.

**Rationale**: Additive schema change that doesn't break existing consumers. Aligns with Observable Quality — the schema documents what the field means.

## Risks / Trade-offs

### R1: Degraded reports may confuse consumers that don't check SSADegraded

Consumers that aggregate `ContractCoverage.Percentage` without checking `SSADegraded` will incorrectly include 0% coverage packages in their averages. Mitigation: the field is documented in the JSON Schema and the warning message explains the situation.

### R2: gaze quality exit code semantics change

Currently `gaze quality` exits non-zero on SSA failure. After this change, it exits 0 with a warning. CI pipelines that relied on the exit code to detect SSA failures will silently pass. Mitigation: this is the desired behavior — SSA failure is not a test quality problem. Users who need to detect degradation can check the JSON output for `ssa_degraded: true`.

### R3: No target inference in degraded mode limits usefulness

Without targets, degraded reports can't tell users "TestFoo exercises Foo". The reports only say "TestFoo exists and has N assertions". This is honestly limited — but strictly better than returning nothing. A name-based fallback (non-goal D1) could improve this in a future spec.
