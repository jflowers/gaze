## ADDED Requirements

### Requirement: CRAPload threshold MUST be data-driven

The gaze-reporter agent prompt MUST instruct the agent to read the CRAPload threshold value from `summary.crap_threshold` in the CRAP JSON data. The agent MUST NOT hardcode a threshold value. The CRAPload line in the output MUST use the format `"N (functions >= threshold T)"` where T is the value from the JSON.

#### Scenario: Agent renders CRAPload with threshold from data
- **GIVEN** a CRAP JSON payload with `summary.crap_threshold = 15` and `summary.crapload = 31`
- **WHEN** the agent renders the CRAP Summary table
- **THEN** the CRAPload row reads `"31 (functions >= threshold 15)"`

#### Scenario: Agent renders CRAPload with non-default threshold
- **GIVEN** a CRAP JSON payload with `summary.crap_threshold = 30` and `summary.crapload = 82`
- **WHEN** the agent renders the CRAP Summary table
- **THEN** the CRAPload row reads `"82 (functions >= threshold 30)"`

### Requirement: Contract coverage MUST use module-wide average

The gaze-reporter agent prompt MUST instruct the agent to report the module-wide average contract coverage from the quality package summary. The agent MUST NOT compute a subset average from selected functions or packages. When module-level quality data returns zero tests, the agent SHOULD note this limitation rather than substituting a subset metric.

#### Scenario: Agent reports module-wide contract coverage
- **GIVEN** a quality JSON payload with package summary `avg_contract_coverage = 32`
- **WHEN** the agent renders the Health Assessment scorecard
- **THEN** the Contract Coverage dimension shows a grade based on 32%, not a higher subset value

#### Scenario: Agent handles zero-test module-level results
- **GIVEN** a quality JSON payload where module-level analysis returns 0 tests
- **WHEN** the agent renders the Quality Summary
- **THEN** the agent notes the limitation and does NOT substitute a favorable subset average as the headline metric

### Requirement: Scoring Consistency Rules section MUST exist in prompt

The gaze-reporter agent prompt MUST contain a "Scoring Consistency Rules" section that explicitly states:
1. CRAPload threshold comes from JSON data, not the prompt
2. Contract coverage uses module-wide averages
3. GazeCRAPload comes from the JSON `gaze_crapload` field
4. Worst offender scores and fix strategies are rendered verbatim from JSON

#### Scenario: Prompt contains scoring rules section
- **GIVEN** the gaze-reporter agent prompt file
- **WHEN** the file is read
- **THEN** it contains a section titled "Scoring Consistency Rules" with instructions for CRAPload, contract coverage, GazeCRAPload, and worst offenders

## MODIFIED Requirements

### Requirement: Quadrant descriptions MUST use "contract coverage" terminology

Previously: The Quick Reference Example in the agent prompt used "high coverage" (Q1), "covered" (Q2), and "untested" (Q4), which conflate line coverage with contract coverage.

The prompt's Quick Reference Example and the `example-report.md` reference file MUST use consistent quadrant descriptions that specify "contract coverage" rather than generic "coverage":
- Q1 — Safe: "Low complexity, high contract coverage"
- Q2 — Complex But Tested: "High complexity, contracts verified"
- Q3 — Needs Tests: "Simple but underspecified"
- Q4 — Dangerous: "Complex AND contracts not adequately verified"

#### Scenario: Quadrant labels are consistent between prompt and reference
- **GIVEN** the gaze-reporter agent prompt and the example-report.md reference file
- **WHEN** both files are read
- **THEN** the quadrant descriptions in both files match exactly

### Requirement: Example report CRAPload row MUST show parameterized threshold

Previously: The example-report.md showed `"24 (functions >= threshold 15)"` with a literal 15.

The example report MUST show a parameterized threshold that signals the value comes from data, not a constant. The row SHOULD use `"N (functions >= threshold T)"` or include a comment/note indicating T is read from `summary.crap_threshold`.

#### Scenario: Example report does not hardcode threshold
- **GIVEN** the example-report.md reference file
- **WHEN** the CRAPload row is read
- **THEN** the threshold value is presented as data-driven, not as the literal "15"

## REMOVED Requirements

None.
