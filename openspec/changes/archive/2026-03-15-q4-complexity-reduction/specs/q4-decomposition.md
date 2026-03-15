## ADDED Requirements

None — all new functions are unexported helpers. No new public API.

## MODIFIED Requirements

### Requirement: All Q4 functions reduced below complexity 10

Each of the five Q4 Dangerous functions MUST be decomposed such that its cyclomatic complexity is reduced to 10 or below, as measured by `gocyclo`.

#### Scenario: scaffold.Run complexity reduced
- **GIVEN** the decomposed `scaffold.Run` function
- **WHEN** `gocyclo` is run on `internal/scaffold/scaffold.go`
- **THEN** `scaffold.Run` MUST report complexity 10 or below

#### Scenario: crap.WriteText complexity reduced
- **GIVEN** the decomposed `crap.WriteText` function
- **WHEN** `gocyclo` is run on `internal/crap/report.go`
- **THEN** `crap.WriteText` MUST report complexity 10 or below

#### Scenario: classify.ComputeScore complexity reduced
- **GIVEN** the decomposed `classify.ComputeScore` function
- **WHEN** `gocyclo` is run on `internal/classify/score.go`
- **THEN** `classify.ComputeScore` MUST report complexity 10 or below

#### Scenario: aireport.Run complexity reduced
- **GIVEN** the decomposed `aireport.Run` function
- **WHEN** `gocyclo` is run on `internal/aireport/runner.go`
- **THEN** `aireport.Run` MUST report complexity 10 or below

#### Scenario: OllamaAdapter.Format complexity reduced
- **GIVEN** the decomposed `(*OllamaAdapter).Format` function
- **WHEN** `gocyclo` is run on `internal/aireport/adapter_ollama.go`
- **THEN** `(*OllamaAdapter).Format` MUST report complexity 10 or below

### Requirement: Behavioral equivalence preserved

Each decomposed function MUST produce identical output and behavior as the original. Existing tests MUST pass without modification.

#### Scenario: All existing tests pass
- **GIVEN** the decomposed codebase
- **WHEN** `go test -race -count=1 -short ./...` is run
- **THEN** all tests MUST pass (no regressions)

## REMOVED Requirements

None.
