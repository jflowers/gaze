## ADDED Requirements

### Requirement: pipelineStepFuncs injection struct

`pipelineStepFuncs` MUST be an unexported struct with four function fields corresponding to the four analysis step functions. Each field MUST default to the real step function when nil.

#### Scenario: Zero-valued struct defaults to real functions
- **GIVEN** a `pipelineStepFuncs{}` (all nil fields)
- **WHEN** `runProductionPipeline` is called with this struct
- **THEN** it MUST call the real `runCRAPStep`, `runQualityStep`, `runClassifyStep`, and `runDocscanStep` functions

#### Scenario: Injected step function is called instead of real
- **GIVEN** a `pipelineStepFuncs` with `crapStep` set to a fake function
- **WHEN** `runProductionPipeline` is called
- **THEN** the fake `crapStep` MUST be called instead of `runCRAPStep`

## MODIFIED Requirements

### Requirement: runProductionPipeline accepts injectable step functions

Previously: `runProductionPipeline` called the four step functions directly with no injection point.

`runProductionPipeline` MUST accept a `pipelineStepFuncs` parameter. When step function fields are nil, it MUST default to the real implementations. The function's external behavior (return type, error semantics, partial-failure handling) MUST remain identical.

#### Scenario: All steps succeed
- **GIVEN** all four step functions return success
- **WHEN** `runProductionPipeline` is called
- **THEN** the returned `ReportPayload` MUST have non-nil CRAP, Quality, Classify, and Docscan fields and nil error fields

#### Scenario: One step fails, others succeed
- **GIVEN** `crapStep` returns an error and the other three return success
- **WHEN** `runProductionPipeline` is called
- **THEN** `payload.Errors.CRAP` MUST be non-nil, `payload.CRAP` MUST be nil, and the other three sections MUST be populated normally

#### Scenario: Multiple steps fail
- **GIVEN** `crapStep` and `qualityStep` both return errors
- **WHEN** `runProductionPipeline` is called
- **THEN** `payload.Errors.CRAP` and `payload.Errors.Quality` MUST both be non-nil, and the other two sections MUST be populated normally

#### Scenario: Empty patterns
- **GIVEN** an empty patterns slice
- **WHEN** `runProductionPipeline` is called
- **THEN** it MUST return an error before calling any step functions

#### Scenario: CRAP step populates summary fields
- **GIVEN** `crapStep` returns a result with CRAPload=5 and GazeCRAPload=3
- **WHEN** `runProductionPipeline` is called
- **THEN** `payload.Summary.CRAPload` MUST be 5 and `payload.Summary.GazeCRAPload` MUST be 3

#### Scenario: Quality step populates summary fields
- **GIVEN** `qualityStep` returns a result with AvgContractCoverage=85
- **WHEN** `runProductionPipeline` is called
- **THEN** `payload.Summary.AvgContractCoverage` MUST be 85

## REMOVED Requirements

None.
