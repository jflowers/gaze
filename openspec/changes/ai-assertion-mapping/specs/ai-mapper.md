## ADDED Requirements

### Requirement: AIMapperFunc on quality.Options

`quality.Options` MUST include an `AIMapperFunc` field that, when non-nil, is called as a final fallback for assertions that all mechanical passes fail to map.

#### Scenario: AI mapper is nil (default)
- **GIVEN** `opts.AIMapperFunc` is nil
- **WHEN** an assertion fails all mechanical passes
- **THEN** the assertion MUST be classified as unmapped (identical to current behavior)

#### Scenario: AI mapper returns a match
- **GIVEN** `opts.AIMapperFunc` returns a valid side effect ID
- **WHEN** an assertion fails all mechanical passes
- **THEN** the assertion MUST be mapped to the returned effect at confidence 50

#### Scenario: AI mapper returns no match
- **GIVEN** `opts.AIMapperFunc` returns an empty string
- **WHEN** an assertion fails all mechanical passes
- **THEN** the assertion MUST remain unmapped

#### Scenario: AI mapper returns an error
- **GIVEN** `opts.AIMapperFunc` returns an error
- **WHEN** an assertion fails all mechanical passes
- **THEN** the assertion MUST remain unmapped and a warning SHOULD be logged

### Requirement: AIMapperContext struct

`AIMapperContext` MUST contain sufficient information for an AI model to evaluate whether an assertion verifies a side effect:
- Assertion source code
- Assertion kind
- Test function source code
- Target function qualified name
- List of side effects with type, description, and ID

### Requirement: Confidence 50 for AI-mapped assertions

All AI-mapped assertions MUST have confidence 50, which is lower than any mechanical pass (60-75).

#### Scenario: AI mapping confidence
- **GIVEN** an assertion mapped by `AIMapperFunc`
- **WHEN** the mapping is returned
- **THEN** `Confidence` MUST be 50

## MODIFIED Requirements

### Requirement: MapAssertionsToEffects includes AI fallback

Previously: assertions failing all mechanical passes were classified as unmapped.

`MapAssertionsToEffects` MUST accept an `AIMapperFunc` parameter. When non-nil and mechanical passes fail, it MUST call the AI mapper before classifying as unmapped.

## REMOVED Requirements

None.
