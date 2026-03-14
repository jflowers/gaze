## ADDED Requirements

### Requirement: computeScores function

`computeScores` MUST accept a slice of `gocyclo.Stat`, a coverage map (`map[string]float64`), and `Options`. It MUST return `[]Score` with one entry per non-skipped function.

#### Scenario: Basic CRAP score computation
- **GIVEN** a `gocyclo.Stat` with complexity 5 and a coverage map entry with 80% coverage for the corresponding function
- **WHEN** `computeScores` is called
- **THEN** the returned `Score` MUST have `Complexity: 5`, `LineCoverage: 80.0`, and `CRAP` equal to `Formula(5, 80.0)`

#### Scenario: Test files are skipped
- **GIVEN** a `gocyclo.Stat` with `Pos.Filename` ending in `_test.go`
- **WHEN** `computeScores` is called
- **THEN** no `Score` entry MUST be produced for that stat

#### Scenario: Generated files are skipped when IgnoreGenerated is true
- **GIVEN** `opts.IgnoreGenerated` is true and a `gocyclo.Stat` references a file containing a `// Code generated` marker
- **WHEN** `computeScores` is called
- **THEN** no `Score` entry MUST be produced for that stat

#### Scenario: Generated files are included when IgnoreGenerated is false
- **GIVEN** `opts.IgnoreGenerated` is false and a `gocyclo.Stat` references a generated file
- **WHEN** `computeScores` is called
- **THEN** a `Score` entry MUST be produced for that stat

#### Scenario: Zero coverage when function not in coverage map
- **GIVEN** a `gocyclo.Stat` for a function that has no entry in the coverage map
- **WHEN** `computeScores` is called
- **THEN** the returned `Score` MUST have `LineCoverage: 0.0`

#### Scenario: GazeCRAP computed when ContractCoverageFunc is set
- **GIVEN** `opts.ContractCoverageFunc` returns `(75.0, true)` for a function with complexity 10
- **WHEN** `computeScores` is called
- **THEN** the returned `Score` MUST have non-nil `ContractCoverage`, `GazeCRAP`, and `Quadrant` fields

#### Scenario: GazeCRAP omitted when ContractCoverageFunc is nil
- **GIVEN** `opts.ContractCoverageFunc` is nil
- **WHEN** `computeScores` is called
- **THEN** the returned `Score` MUST have nil `ContractCoverage`, `GazeCRAP`, and `Quadrant` fields

#### Scenario: GazeCRAP omitted when ContractCoverageFunc returns false
- **GIVEN** `opts.ContractCoverageFunc` returns `(0, false)` for a function
- **WHEN** `computeScores` is called
- **THEN** the returned `Score` for that function MUST have nil `ContractCoverage`, `GazeCRAP`, and `Quadrant` fields

## MODIFIED Requirements

### Requirement: Analyze orchestrates via computeScores

Previously: `Analyze` contained inline score computation (Step 5, lines 104-157).

`Analyze` MUST delegate score computation to `computeScores`. The function's external behavior (signature, return type, output) MUST remain identical. The cyclomatic complexity of `Analyze` SHOULD be reduced to 7 or below.

#### Scenario: Analyze produces identical output after decomposition
- **GIVEN** identical inputs to `Analyze` before and after decomposition
- **WHEN** `Analyze` is called
- **THEN** the returned `*Report` MUST contain the same `Scores` and `Summary` as before

## REMOVED Requirements

None.
