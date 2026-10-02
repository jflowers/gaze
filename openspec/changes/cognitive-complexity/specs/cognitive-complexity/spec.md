## ADDED Requirements

### Requirement: Cognitive Complexity Computation

The system MUST compute per-function cognitive complexity following the SonarSource cognitive complexity specification, adapted for Go AST.

#### Scenario: Simple if statement
- **GIVEN** a function containing a single `if` statement at nesting level 0
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1 (increment for `if`, no nesting penalty)

#### Scenario: Nested if statement
- **GIVEN** a function containing an `if` statement nested inside another `if` statement
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 3 (1 for outer `if` + 1 for inner `if` + 1 nesting penalty)

#### Scenario: else if chain
- **GIVEN** a function containing `if` / `else if` / `else if` / `else`
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 3 (1 for `if` + 1 for first `else if` + 1 for second `else if`; `else` does not increment per Sonar spec when it is a flat chain)

#### Scenario: for loop
- **GIVEN** a function containing a `for` loop at nesting level 0
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1

#### Scenario: range loop
- **GIVEN** a function containing a `for...range` loop at nesting level 0
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1

#### Scenario: switch statement
- **GIVEN** a function containing a `switch` statement at nesting level 0
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1 (increment for `switch`; individual `case` labels do NOT increment)

#### Scenario: type switch statement
- **GIVEN** a function containing a `type switch` statement
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1

#### Scenario: logical operator sequence
- **GIVEN** a function containing `a && b && c` (sequence of same logical operator)
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1 (one increment per mixed-logical-operator sequence)

#### Scenario: mixed logical operators
- **GIVEN** a function containing `a && b || c`
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 2 (one increment for `&&` sequence + one increment for `||` sequence)

#### Scenario: goto statement
- **GIVEN** a function containing a `goto` statement
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1

#### Scenario: recursion
- **GIVEN** a function `Foo` that calls itself recursively
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST include 1 increment for the recursive call

#### Scenario: deeply nested structure
- **GIVEN** a function with `for` > `if` > `switch` (3 levels of nesting)
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 6 (1 for `for` + 1+1 for `if` + 1+2 for `switch`)

#### Scenario: select statement
- **GIVEN** a function containing a `select` statement
- **WHEN** cognitive complexity is computed
- **THEN** the cognitive complexity MUST be 1 (treated like `switch`)

### Requirement: Cognitive Complexity Provider Interface

The system MUST provide a `CognitiveComplexityProvider` interface following the existing provider pattern, enabling language-specific implementations and external analyzer integration.

#### Scenario: Go provider returns correct values
- **GIVEN** a `GoCognitiveComplexityProvider` and a Go package with functions of known cognitive complexity
- **WHEN** `Analyze` is called with the package patterns
- **THEN** the result MUST contain per-function cognitive complexity values matching the expected values

#### Scenario: external analyzer protocol
- **GIVEN** an external analyzer that implements the `cognitive_complexity` JSON-RPC method
- **WHEN** the system requests cognitive complexity data
- **THEN** the system MUST use the external analyzer's cognitive complexity values

#### Scenario: external analyzer without cognitive complexity
- **GIVEN** an external analyzer that does NOT implement `cognitive_complexity`
- **WHEN** the system requests cognitive complexity data
- **THEN** the system MUST gracefully degrade (cognitive complexity fields remain nil) without error

### Requirement: JSON Output for Cognitive Complexity

The system MUST include `cognitive_complexity` per function in JSON output.

#### Scenario: JSON includes cognitive complexity
- **GIVEN** a function with cognitive complexity 5
- **WHEN** JSON output is produced
- **THEN** the function's JSON representation MUST include `"cognitive_complexity": 5`

#### Scenario: JSON omits when unavailable
- **GIVEN** a function where cognitive complexity was not computed (e.g., external analyzer without support)
- **WHEN** JSON output is produced
- **THEN** the function's JSON representation MUST omit the `cognitive_complexity` field

### Requirement: CI Gate Flag

The system MUST support a `--max-cognitive-complexity=<N>` flag on `gaze analyze` and `gaze crap` commands. On `gaze analyze` the flag is a gate only: cognitive complexity is computed for the gate and reported on stderr, but is not added to the side-effect analysis output. On `gaze crap` and `gaze report` cognitive complexity is computed and reported in the score output regardless of the flag.

#### Scenario: all functions within threshold
- **GIVEN** `--max-cognitive-complexity=15` and all functions have cognitive complexity <= 15
- **WHEN** the command runs
- **THEN** the command MUST exit with code 0

#### Scenario: function exceeds threshold
- **GIVEN** `--max-cognitive-complexity=15` and at least one function has cognitive complexity > 15
- **WHEN** the command runs
- **THEN** the command MUST exit with code 1

#### Scenario: flag not provided
- **GIVEN** no `--max-cognitive-complexity` flag
- **WHEN** the command runs
- **THEN** the command MUST NOT fail on account of cognitive complexity. `gaze crap` and `gaze report` MUST still compute and report cognitive complexity; `gaze analyze` MUST NOT compute cognitive complexity (the gate is skipped).
