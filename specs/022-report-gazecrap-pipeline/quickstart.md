# Quickstart: GazeCRAP Data in Report Pipeline

## What This Changes

The `gaze report` command currently produces CRAP scores based only on line coverage. GazeCRAP (which uses contract coverage — a stronger signal) is computed by `gaze crap` but missing from `gaze report`. After this change, `gaze report` produces GazeCRAP scores, quadrant distribution, and GazeCRAPload in its output.

## Before (current behavior)

```bash
$ gaze report --format=json --coverprofile=coverage.out ./...
# JSON output has gaze_crap: null, quadrant: null
# Summary has gaze_crapload: null, quadrant_counts: {}
# --max-gaze-crapload threshold always passes (no data to evaluate)
```

## After (new behavior)

```bash
$ gaze report --format=json --coverprofile=coverage.out ./...
# JSON output has gaze_crap: 12.5, quadrant: "Q1_Safe"
# Summary has gaze_crapload: 4, quadrant_counts: {Q1_Safe: 200, Q2_ComplexButTested: 10, Q3_NeedsTests: 5, Q4_Dangerous: 4}
# --max-gaze-crapload=3 would now FAIL because gaze_crapload=4
```

## Implementation Steps (high level)

1. Extract `buildContractCoverageFunc` from `cmd/gaze/main.go` to `internal/crap/contract.go` as `BuildContractCoverageFunc`
2. Update `cmd/gaze/main.go` to call the new location
3. Expand `runCRAPStep` signature to accept a `ContractCoverageFunc` parameter
4. Update `runProductionPipeline` to call `BuildContractCoverageFunc` before the CRAP step and pass the result
5. Update `pipelineStepFuncs` and all test mocks for the new signature
6. Add/update tests verifying GazeCRAP data appears in report JSON output

## Verification

```bash
# Build
go build ./cmd/gaze

# Unit tests
go test -race -count=1 -short ./internal/crap/... ./internal/aireport/... ./cmd/gaze/...

# Integration: verify GazeCRAP data in report JSON
go test -race -count=1 -short -run TestRunReport_JSONFormat ./cmd/gaze/...
```
