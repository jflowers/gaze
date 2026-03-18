# /gaze fix Command Spec

## ADDED Requirements

### Requirement: Command MUST run gaze analysis and generate tests for actionable functions

The `/gaze fix` command MUST run gaze CRAP and quality analysis, filter to functions with actionable fix strategies, and delegate to the `gaze-test-generator` agent for each target.

#### Scenario: Basic invocation
- **GIVEN** a Go project with CRAPload violations
- **WHEN** the user runs `/gaze fix ./...`
- **THEN** gaze analysis runs, functions with `add_tests`/`add_assertions`/`decompose_and_test` strategies are identified, and tests are generated for each

#### Scenario: Strategy filter
- **GIVEN** a project with mixed fix strategies
- **WHEN** the user runs `/gaze fix --strategy=add_tests ./...`
- **THEN** only functions with `add_tests` strategy are processed

#### Scenario: Top-N limit
- **GIVEN** a project with 20 CRAPload functions
- **WHEN** the user runs `/gaze fix --top=5 ./...`
- **THEN** only the 5 worst CRAP-score functions are processed

### Requirement: Command MUST verify generated tests

After generating tests, the command MUST run `go build` and `go test` to verify the generated code compiles and passes. Failures MUST be reported with context.

#### Scenario: Compilation failure
- **GIVEN** a generated test that fails to compile
- **WHEN** `go build` runs
- **THEN** the command reports the compilation error and which test function failed

#### Scenario: Test failure
- **GIVEN** a generated test that compiles but fails
- **WHEN** `go test` runs
- **THEN** the command reports the test failure and suggests the assertion may need adjustment

### Requirement: Command MUST prioritize by fix strategy

Functions MUST be processed in this order: `add_tests` first, then `add_assertions`, then `decompose_and_test`. Within each strategy group, sort by CRAP score descending (worst first).

#### Scenario: Priority ordering
- **GIVEN** functions A (add_assertions, CRAP 50), B (add_tests, CRAP 30), C (add_tests, CRAP 90)
- **WHEN** the command processes them
- **THEN** the order is C, B, A (add_tests first sorted by CRAP, then add_assertions)
