# CI Workflow Requirements

## ADDED Requirements

### Requirement: Unit+Integration job MUST run gaze report on Go 1.24

The unit+integration CI job MUST include a step that runs `gaze report` with text output and quality thresholds, conditional on `matrix.go-version == '1.24'`.

#### Scenario: Go 1.24 job runs gaze report
- **GIVEN** a push or PR to `main` triggers the unit+integration job
- **WHEN** the Go 1.24 matrix entry runs
- **THEN** a "Gaze quality report" step executes `gaze report ./... --coverprofile=coverage.out --max-crapload=16 --max-gaze-crapload=5`

#### Scenario: Go 1.25 job does NOT run gaze report
- **GIVEN** the same push/PR triggers the Go 1.25 matrix entry
- **WHEN** the Go 1.25 job completes
- **THEN** no "Gaze quality report" step executes

### Requirement: Test step MUST emit coverprofile

The test step in the unit+integration job MUST produce a `coverage.out` file by adding `-coverprofile=coverage.out` to the `go test` command. This allows `gaze report` to reuse the coverage data.

#### Scenario: Coverage file produced
- **GIVEN** the test step runs
- **WHEN** `go test -race -count=1 -short -coverprofile=coverage.out ./...` completes
- **THEN** a `coverage.out` file exists in the workspace

### Requirement: Threshold failures MUST fail the CI job

When `gaze report` exits non-zero (threshold exceeded), the CI job MUST fail.

#### Scenario: CRAPload exceeds threshold
- **GIVEN** `--max-crapload=16` is set
- **WHEN** the project has more than 16 functions with CRAP >= threshold
- **THEN** the step fails and the CI job reports failure

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
