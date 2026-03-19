## ADDED Requirements

### Requirement: ContractCoverageReason on Score

`Score` MUST include `ContractCoverageReason *string` (JSON `contract_coverage_reason,omitempty`) explaining why contract coverage is what it is.

#### Scenario: All effects ambiguous
- **GIVEN** a function where all side effects are classified "ambiguous"
- **WHEN** `computeScores` builds the score
- **THEN** `ContractCoverageReason` MUST be `"all_effects_ambiguous"`

#### Scenario: No effects detected
- **GIVEN** a function with zero side effects
- **WHEN** `computeScores` builds the score
- **THEN** `ContractCoverageReason` MUST be `"no_effects_detected"`

#### Scenario: Normal coverage
- **GIVEN** a function with contractual effects and some coverage
- **WHEN** `computeScores` builds the score
- **THEN** `ContractCoverageReason` MUST be nil

### Requirement: EffectConfidenceRange on Score

When `ContractCoverageReason` is `"all_effects_ambiguous"`, `Score` MUST include `EffectConfidenceRange [2]int` (JSON `effect_confidence_range,omitempty`) with min and max classification confidence values.

## MODIFIED Requirements

### Requirement: ContractCoverageFunc returns richer data

Previously: `func(pkg, fn string) (float64, bool)`

`ContractCoverageFunc` MUST return `(ContractCoverageInfo, bool)` carrying the percentage, reason, and confidence range.

## REMOVED Requirements

None.
