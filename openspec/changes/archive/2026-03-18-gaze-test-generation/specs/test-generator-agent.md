# Test Generator Agent Spec

## ADDED Requirements

### Requirement: Agent MUST generate compilable, runnable Go test functions

The `gaze-test-generator` agent MUST produce complete Go test functions that compile with `go build` and pass with `go test -race -count=1`. Generated tests MUST NOT require manual editing to compile.

#### Scenario: Generated test compiles
- **GIVEN** a function `Add(a, b int) int` with a `ReturnValue` gap and hint `got := target(); // assert got == expected`
- **WHEN** the agent generates a test
- **THEN** the test function compiles with `go build ./...`

#### Scenario: Generated test passes
- **GIVEN** the same function with correct implementation
- **WHEN** the generated test runs
- **THEN** `go test -race -count=1 -run TestAdd ./...` exits 0

### Requirement: Agent MUST handle all actionable fix strategies plus doc improvement

The agent MUST handle `add_tests`, `add_assertions`, `add_docs`, and `decompose_and_test` strategies. It MUST skip `decompose`-only functions with an explanation.

#### Scenario: add_tests generates full test
- **GIVEN** a function with `fix_strategy: add_tests` and 0% line coverage
- **WHEN** the agent generates tests
- **THEN** a complete test function with setup, call, and assertions is produced

#### Scenario: add_assertions strengthens existing test and improves mapper visibility
- **GIVEN** a function with `fix_strategy: add_assertions` and existing tests with `UnmappedReason: helper_param`
- **WHEN** the agent processes the function
- **THEN** new assertions are added based on GapHints AND existing helper-wrapped assertions are restructured so the mapper can trace them to the target function's side effects

#### Scenario: add_docs pushes ambiguous effects to contractual
- **GIVEN** a Q3 function with `ContractCoverageReason: all_effects_ambiguous` and `EffectConfidenceRange: [63, 69]`
- **WHEN** the agent processes the function
- **THEN** GoDoc comments are added or improved on the function describing its observable side effects, pushing classifier confidence above 70

#### Scenario: add_docs not applied when confidence is too low
- **GIVEN** a function with ambiguous effects at confidence 30-50
- **WHEN** the agent evaluates the function
- **THEN** the agent applies `add_tests` or `add_assertions` instead (GoDoc alone won't push confidence above 70 when it's far below)

#### Scenario: decompose_and_test generates skeleton
- **GIVEN** a function with `fix_strategy: decompose_and_test`
- **WHEN** the agent processes the function
- **THEN** a test skeleton with TODO comments for each gap is produced

#### Scenario: decompose is skipped
- **GIVEN** a function with `fix_strategy: decompose`
- **WHEN** the agent evaluates the function
- **THEN** the agent reports "skipped — needs decomposition, not tests" and moves on

### Requirement: Agent MUST follow target project testing conventions

Generated tests MUST use Go stdlib `testing` package only (no testify, gomega). Test functions MUST follow `TestXxx_Description` naming. Assertions MUST use `t.Errorf`/`t.Fatalf` directly.

#### Scenario: Convention compliance
- **GIVEN** any target function in any Go project
- **WHEN** the agent generates a test
- **THEN** the test uses only `import "testing"`, names the function `TestXxx_Something`, and uses `t.Errorf` or `t.Fatalf` for assertions

### Requirement: Agent MUST append to existing test files

Generated tests MUST be appended to the existing `*_test.go` file for the target function's source file. If no test file exists, the agent MUST create one with the appropriate package declaration.

#### Scenario: Append to existing file
- **GIVEN** a function in `foo.go` with an existing `foo_test.go`
- **WHEN** the agent generates a test
- **THEN** the new test function is appended to `foo_test.go`

#### Scenario: Create new test file
- **GIVEN** a function in `bar.go` with no `bar_test.go`
- **WHEN** the agent generates a test
- **THEN** a new `bar_test.go` is created with correct package declaration and imports
