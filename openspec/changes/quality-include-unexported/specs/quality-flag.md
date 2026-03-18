# Quality Include Unexported Spec

## ADDED Requirements

### Requirement: `gaze quality` MUST support `--include-unexported` flag

The `quality` subcommand MUST accept `--include-unexported` to include unexported functions in analysis, matching the `analyze` command's flag behavior.

#### Scenario: Flag enables unexported function analysis
- **GIVEN** a `package main` tool with only unexported functions
- **WHEN** the user runs `gaze quality --include-unexported ./cmd/tool/`
- **THEN** quality reports are produced for the unexported functions

#### Scenario: Without flag, exported-only is default
- **GIVEN** a library package with both exported and unexported functions
- **WHEN** the user runs `gaze quality ./pkg/lib/`
- **THEN** only exported functions appear in quality reports

### Requirement: `package main` MUST auto-detect and include unexported

When the analyzed package is `package main`, gaze MUST automatically include unexported functions regardless of whether `--include-unexported` is specified. A `main` package has no exported API by definition.

#### Scenario: Auto-detect for package main
- **GIVEN** a `package main` tool
- **WHEN** the user runs `gaze quality ./cmd/tool/` (no `--include-unexported`)
- **THEN** unexported functions are included automatically

#### Scenario: Auto-detect logs a message
- **GIVEN** a `package main` tool
- **WHEN** auto-detection fires
- **THEN** a log message indicates "package main detected, including unexported functions"

### Requirement: Report pipeline MUST auto-detect `package main`

The `gaze report` pipeline (`runQualityStep`, `runQualityForPackage`) and contract coverage builder (`analyzePackageCoverage`) MUST auto-detect `package main` and include unexported functions.

#### Scenario: gaze report on package main
- **GIVEN** a `package main` tool
- **WHEN** the user runs `gaze report ./cmd/tool/`
- **THEN** quality analysis includes unexported functions in the report payload

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
