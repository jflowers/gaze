## ADDED Requirements

### Requirement: gaze-test-generator MUST support verify action

The gaze-test-generator agent MUST support a `verify` action that reruns `gaze quality --format=json` on the target package after test generation and reports the contract coverage improvement. The action SHOULD compare before/after coverage percentages and report the delta.

#### Scenario: Verify shows coverage improvement
- **GIVEN** a package with 25% contract coverage and the test generator has added new assertions
- **WHEN** the agent runs the `verify` action
- **THEN** the agent runs `gaze quality --format=json <package>`, compares the new coverage percentage to the baseline, and reports the delta (e.g., "Contract coverage: 25% → 67% (+42%)")

#### Scenario: Verify shows no improvement
- **GIVEN** a package where the generated tests do not improve contract coverage (e.g., assertions on local variables, not side effects)
- **WHEN** the agent runs the `verify` action
- **THEN** the agent reports that coverage did not improve and suggests reviewing the generated assertions for mapping to the function's side effects

#### Scenario: Verify handles missing baseline
- **GIVEN** a package with no prior quality data (first run)
- **WHEN** the agent runs the `verify` action
- **THEN** the agent reports the absolute coverage percentage without a delta comparison

## MODIFIED Requirements

(none)

## REMOVED Requirements

(none)
