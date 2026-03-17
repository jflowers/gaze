# Data Model: GazeCRAP Data in Report Pipeline

## Modified Entities

### pipelineStepFuncs (internal/aireport/runner.go)

The `crapStep` field signature expands to accept a `ContractCoverageFunc` parameter.

| Field | Current Type | New Type |
|-------|-------------|----------|
| `crapStep` | `func([]string, string, string, io.Writer) (*crapStepResult, error)` | `func([]string, string, string, io.Writer, func(string, string) (crap.ContractCoverageInfo, bool)) (*crapStepResult, error)` |

The new parameter is the contract coverage callback. When `nil`, GazeCRAP fields are omitted (existing behavior).

### runCRAPStep (internal/aireport/runner_steps.go)

| Parameter | Type | Description |
|-----------|------|-------------|
| `patterns` | `[]string` | Package patterns |
| `moduleDir` | `string` | Module root directory |
| `coverProfile` | `string` | Path to coverage profile |
| `stderr` | `io.Writer` | Warning output |
| `contractCoverageFunc` | `func(string, string) (crap.ContractCoverageInfo, bool)` | Contract coverage callback (new) |

When `contractCoverageFunc` is non-nil, it is set on `crap.Options.ContractCoverageFunc` before calling `crap.Analyze`.

### crapStepResult (internal/aireport/runner_steps.go)

No changes. The `GazeCRAPload int` field already exists but has been always-zero. After this change, it will contain the actual GazeCRAPload value.

## New Entities

### BuildContractCoverageFunc (internal/crap/contract.go)

Extracted from `cmd/gaze/main.go:buildContractCoverageFunc`. Public function.

| Parameter | Type | Description |
|-----------|------|-------------|
| `patterns` | `[]string` | Package patterns to analyze |
| `moduleDir` | `string` | Module root directory |
| `stderr` | `io.Writer` | Warning output for SSA degradation messages |

| Return | Type | Description |
|--------|------|-------------|
| `func` | `func(pkg, function string) (ContractCoverageInfo, bool)` | Lookup closure, or nil if all packages failed |
| `degradedPkgs` | `[]string` | SSA-degraded package paths |

Dependencies: `analysis`, `classify`, `quality`, `config`, `loader`, `taxonomy` (all existing internal packages).

## Unchanged Entities

- `crap.Options` — already has `ContractCoverageFunc` field
- `crap.ContractCoverageInfo` — already defined
- `crap.Score` — per-function fields: `gaze_crap` (JSON tag, `*float64`) and `quadrant` (JSON tag, `*string`)
- `crap.Summary` — summary fields: `quadrant_counts` (JSON tag, `map[string]int`), `gaze_crapload` (JSON tag, `*int`)
- `ReportSummary` — already has `GazeCRAPload int` field. Note: when `BuildContractCoverageFunc` returns nil (all packages degraded), `crapStepResult.GazeCRAPload` will be 0 (same as "no Q4 functions"). The `SSADegradedPackages` field is the only indicator that GazeCRAP was unavailable vs. genuinely zero.
- `ReportPayload` — no changes needed; CRAP JSON already flows through `payload.CRAP`

## Data Flow (after change)

```
runProductionPipeline
  │
  ├── BuildContractCoverageFunc(patterns, moduleDir, stderr)
  │     └── returns (ccFunc, degradedPkgs)
  │
  ├── Step 1: runCRAPStep(patterns, moduleDir, coverProfile, stderr, ccFunc)
  │     └── sets opts.ContractCoverageFunc = ccFunc
  │     └── crap.Analyze → Report with GazeCRAP scores, quadrants, gaze_crapload
  │
  ├── Step 2: runQualityStep(patterns, moduleDir, stderr)
  │     └── produces quality JSON for AI payload
  │
  ├── Step 3: runClassifyStep(patterns, moduleDir)
  │     └── produces classification JSON
  │
  └── Step 4: runDocscanStep(moduleDir)
        └── produces docscan JSON
```
