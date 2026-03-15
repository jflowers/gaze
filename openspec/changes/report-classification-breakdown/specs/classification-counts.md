## ADDED Requirements

### Requirement: CountLabels helper

`classify.CountLabels` MUST accept `[]taxonomy.AnalysisResult` and return counts of contractual, ambiguous, and incidental side effects.

#### Scenario: Mixed classification labels
- **GIVEN** results with 5 contractual, 3 ambiguous, and 1 incidental side effects
- **WHEN** `CountLabels` is called
- **THEN** it MUST return (5, 3, 1)

#### Scenario: No classified effects
- **GIVEN** results where all side effects have nil Classification
- **WHEN** `CountLabels` is called
- **THEN** it MUST return (0, 0, 0)

### Requirement: Classification counts on ReportSummary

`ReportSummary` MUST include `Contractual`, `Ambiguous`, and `Incidental` int fields populated from the classify step.

#### Scenario: Counts propagated to ReportSummary
- **GIVEN** `runClassifyStep` processes results with 10 contractual and 5 ambiguous effects
- **WHEN** `runProductionPipeline` completes
- **THEN** `payload.Summary.Contractual` MUST be 10 and `payload.Summary.Ambiguous` MUST be 5

## MODIFIED Requirements

### Requirement: runClassifyStep returns typed result

Previously: `runClassifyStep` returned `(json.RawMessage, error)`.

`runClassifyStep` MUST return `(*classifyStepResult, error)` containing the raw JSON plus classification label counts.

## REMOVED Requirements

None.
