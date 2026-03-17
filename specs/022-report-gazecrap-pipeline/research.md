# Research: GazeCRAP Data in Report Pipeline

## R1: Pipeline Step Ordering

**Decision**: Run `buildContractCoverageFunc` before the CRAP step and inject the result into `runCRAPStep` via an expanded parameter.

**Rationale**: The quality step (Step 2) already runs independently and produces per-package contract coverage data. However, it runs *after* the CRAP step has completed. Rather than reordering the existing pipeline steps (which would break the current `pipelineStepFuncs` interface and require reworking the DI test mocks), the fix is to compute the `ContractCoverageFunc` closure *before* the CRAP step, then pass it as an additional parameter. This mirrors what `gaze crap` does in `cmd/gaze/main.go:runCrap`.

**Alternatives considered**:
- *Reorder steps (quality before CRAP)*: Would require restructuring `runProductionPipeline` and its DI interface. The quality step's output format (`qualityStepResult`) doesn't directly produce a `ContractCoverageFunc` — it produces raw JSON. The callback construction logic lives in `buildContractCoverageFunc`, not in the quality step.
- *Merge CRAP and quality into one step*: Violates the existing separation-of-concerns architecture. The CRAP step produces CRAP JSON; the quality step produces quality JSON. They serve different payload sections.
- *Post-process CRAP results with quality data*: Would require re-running `crap.Analyze` or mutating the CRAP JSON after the fact. More complex and fragile.

## R2: Where to Build the ContractCoverageFunc

**Decision**: Call `BuildContractCoverageFunc` inside `runProductionPipeline`, before the CRAP step. The function is already factored for reuse — it takes `(patterns, moduleDir, stderr)` and returns `(func, degradedPkgs)`.

> **Note**: R3 supersedes the original framing of this decision — the function cannot be called from `cmd/gaze/main.go` because it is `package main`. See R3 for the extraction to `internal/crap/contract.go`.

**Rationale**: `buildContractCoverageFunc` is the canonical implementation that `gaze crap` uses. Reusing it ensures SC-002 (exact match between `gaze crap` and `gaze report` GazeCRAPload). The function internally runs the full quality pipeline per-package (analysis → classify → quality.Assess), builds a lookup closure, and tracks SSA-degraded packages.

**Alternatives considered**:
- *Duplicate the logic in `internal/aireport`*: Violates DRY. The function is 90+ lines with non-trivial logic.
- *Extract to a shared package (e.g., `internal/quality` or `internal/crap`)*: Could work but is over-engineering for this change. The function depends on `analysis`, `classify`, `quality`, `config`, and `loader` — moving it requires deciding which package owns the orchestration. Better suited for a future refactor.

## R3: Accessibility of buildContractCoverageFunc

**Decision**: Export `buildContractCoverageFunc` as `BuildContractCoverageFunc` from `cmd/gaze/main.go` is not feasible (it's in `package main`). Instead, extract the function to `internal/crap/contract.go` where it can be imported by both `cmd/gaze/main.go` and `internal/aireport/runner_steps.go`.

**Rationale**: `cmd/gaze/main.go` is `package main` — nothing can import from it. The function must move to an internal package. `internal/crap` is the natural home because the function produces a `crap.ContractCoverageFunc`-typed callback and `crap.ContractCoverageInfo` values. Its dependencies (`analysis`, `classify`, `quality`, `config`, `loader`) are all internal packages already imported by `cmd/gaze`.

**Alternatives considered**:
- *Leave in `cmd/gaze` and duplicate in `internal/aireport`*: Violates FR-006 (same logic) and risks drift.
- *Put in `internal/quality`*: The function's return type is `crap.ContractCoverageInfo` — placing it in `quality` would create a circular dependency (`quality` → `crap` → `quality`).
- *Create a new `internal/bridge` package*: Unnecessary indirection for a single function.

## R4: Pipeline Step Signature Change

**Decision**: Add a `contractCoverageFunc` parameter to `runCRAPStep` and the corresponding `crapStep` field in `pipelineStepFuncs`.

**Rationale**: The `pipelineStepFuncs` struct exists for dependency injection in tests. The CRAP step needs the callback to compute GazeCRAP. Adding it as a parameter (rather than a global or context value) maintains the explicit DI pattern and testability. Existing tests pass `nil` for the callback, which preserves the current no-GazeCRAP behavior in test mocks.

**Alternatives considered**:
- *Use a closure that captures the callback*: Less explicit, harder to test.
- *Pass via an `Options` struct*: Over-engineering for a single parameter addition.

## R5: Quality Step Redundancy

**Decision**: When `ContractCoverageFunc` is built before the CRAP step, the quality analysis runs twice — once inside `buildContractCoverageFunc` (to build the callback) and once in the quality step (to produce quality JSON for the payload). Accept this redundancy.

**Rationale**: The two quality runs serve different purposes. The callback run produces a per-function closure for GazeCRAP scoring. The quality step run produces the quality JSON payload for the AI formatter. Merging them would require significant refactoring of the quality step to both produce JSON and return a closure — complicating the clean step separation. The quality analysis is I/O-bound (package loading, SSA building) so the runtime impact is ~2x for quality, but quality is typically < 30% of total report time.

**Alternatives considered**:
- *Cache quality results from the callback run and reuse in the quality step*: Would require threading a cache through the pipeline. Adds complexity for a moderate performance gain. Can be optimized in a future spec if profiling shows it matters.
- *Run quality only once and derive both outputs*: Requires restructuring the quality step to return both JSON and per-function data. Out of scope for this plumbing fix.

## R6: Coverage Strategy (Constitution IV: Testability)

**Decision**: Unit tests for the extracted `BuildContractCoverageFunc` function, updated tests for `runCRAPStep` with the new parameter, and an integration test verifying GazeCRAP data appears in `gaze report --format=json` output.

**Rationale**: Constitution Principle IV requires coverage strategy for all new code. The extracted function already has indirect coverage via `gaze crap` tests. Direct unit tests with fixture packages ensure the extraction didn't break behavior. The `pipeline_internal_test.go` tests for `runProductionPipeline` need updated mocks for the new `crapStep` signature.

**Test plan**:
- Unit: `TestBuildContractCoverageFunc` variants (already exist in `cmd/gaze/main_test.go`, adapt for new location)
- Unit: Updated `pipeline_internal_test.go` mocks with `ContractCoverageFunc` parameter
- Integration: `TestRunReport_JSONFormat_ValidOutput` asserts `gaze_crapload` is present in JSON when quality succeeds
