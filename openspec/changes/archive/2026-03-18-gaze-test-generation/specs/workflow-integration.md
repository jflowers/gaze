# Workflow Integration Spec

## ADDED Requirements

### Requirement: Implementation workflows MUST generate tests after each task

Both `/opsx-apply` and `/speckit.implement` MUST include a test generation step after each implementation task's code changes, before marking the task complete.

#### Scenario: Test generation after task
- **GIVEN** a task that modifies `internal/foo/bar.go`
- **WHEN** the task's code changes are complete
- **THEN** gaze quality analysis runs on `./internal/foo/...`, and if `ContractCoverage.Gaps` exist for new/modified functions, the `gaze-test-generator` agent generates tests

#### Scenario: No gaps detected
- **GIVEN** a task that modifies only test files or files with no gaps
- **WHEN** the test generation step runs
- **THEN** no tests are generated and the task proceeds to completion normally

### Requirement: Test generation MUST be mandatory by default

In mandatory mode (default), task completion is blocked until generated tests compile and pass. In advisory mode, test generation results are shown but the task can be marked complete regardless.

#### Scenario: Mandatory mode blocks on failure
- **GIVEN** mandatory mode is active and a generated test fails
- **WHEN** the implementation workflow tries to mark the task complete
- **THEN** the workflow pauses and reports the test failure, asking the developer to fix the test or the code

#### Scenario: Advisory mode allows skip
- **GIVEN** advisory mode is configured in `.gaze.yaml`
- **WHEN** a generated test fails
- **THEN** the failure is reported but the task is marked complete

### Requirement: Mode MUST be configurable via .gaze.yaml

The test generation mode is read from `.gaze.yaml`:
```yaml
test_generation:
  mode: mandatory  # or "advisory"
```
When the key is absent or `.gaze.yaml` doesn't exist, the default is `mandatory`.

#### Scenario: Config present
- **GIVEN** `.gaze.yaml` contains `test_generation: { mode: advisory }`
- **WHEN** the implementation workflow runs
- **THEN** test generation operates in advisory mode

#### Scenario: Config absent
- **GIVEN** no `.gaze.yaml` exists
- **WHEN** the implementation workflow runs
- **THEN** test generation operates in mandatory mode

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
