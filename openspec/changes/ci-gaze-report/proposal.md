# CI Gaze Report

## Why

The `gaze report` command produces GazeCRAP scores, quadrant distribution, and quality metrics — but it only runs when a downstream project (like `get-out`) invokes it in their own CI. Gaze's own CI has no quality gate using its own tool. Running `gaze report` in gaze's CI ensures the tool eats its own dog food and catches regressions in the report pipeline itself.

The report should run on only one Go version (1.24) to avoid redundant analysis and keep CI runtime reasonable. The unit+integration test step already generates coverage data that can be reused via `--coverprofile`.

## What Changes

Add a `gaze report` step to the unit+integration CI job, conditional on Go 1.24 only. The step reuses the coverage profile from the test step to avoid a second `go test` run, and enforces quality thresholds via `--max-crapload` and `--max-gaze-crapload`.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `.github/workflows/test.yml`: Unit+Integration job gains a `Gaze quality report` step that runs `gaze report` with `--coverprofile` and threshold flags, conditional on `matrix.go-version == '1.24'`

### Removed Capabilities
- None

## Impact

- `.github/workflows/test.yml` — modified (add step + adjust test step to emit coverprofile on Go 1.24)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies CI configuration only. No artifact communication patterns are affected.

### II. Composability First

**Assessment**: PASS

No new dependencies introduced. `gaze report` is run from the built binary already available in the job. The `--coverprofile` flag reuses data from the test step.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observable quality by running gaze against itself in CI, producing machine-parseable quality metrics and enforcing thresholds. Failures surface as CI check failures with actionable error messages.

### IV. Testability

**Assessment**: N/A

No Go code changes. The workflow change is tested by CI execution itself.
