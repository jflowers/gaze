# Metric Definitions Spec

## ADDED Requirements

### Requirement: Agent prompt MUST define GazeCRAPload as distinct from Q4 count

The gaze-reporter agent prompt MUST contain a "Metric Definitions" subsection within the Scoring Consistency Rules that defines GazeCRAPload as the count of functions with GazeCRAP >= threshold, explicitly stating it is NOT the Q4 count.

#### Scenario: Agent reports GazeCRAPload correctly
- **GIVEN** a JSON payload with `summary.gaze_crapload = 24` and `summary.quadrant_counts.Q4_Dangerous = 0`
- **WHEN** the agent renders the GazeCRAPload metric
- **THEN** the agent reports GazeCRAPload as 24 (not 0)

#### Scenario: Agent distinguishes GazeCRAPload from Q4 in narrative
- **GIVEN** GazeCRAPload = 24 and Q4 = 0
- **WHEN** the agent writes the Health Assessment
- **THEN** the agent does not state "GazeCRAPload is 0" or equate GazeCRAPload with Q4 count

### Requirement: Metric definitions MUST cover CRAPload, GazeCRAPload, and quadrant counts

The definitions section MUST define all three metrics with their source JSON fields, the coverage type they use (line vs contract), and what threshold they reference.

#### Scenario: Prompt contains all three definitions
- **GIVEN** the gaze-reporter agent prompt
- **WHEN** the Scoring Consistency Rules section is read
- **THEN** it contains definitions for CRAPload, GazeCRAPload, and quadrant counts

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
