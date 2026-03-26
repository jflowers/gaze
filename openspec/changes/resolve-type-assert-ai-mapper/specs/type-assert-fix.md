## ADDED Requirements

### Requirement: resolveExprRoot MUST handle TypeAssertExpr

The `resolveExprRoot` function in the assertion mapping pipeline MUST recursively resolve through `*ast.TypeAssertExpr` nodes by extracting the inner expression (`e.X`), consistent with how it handles `SelectorExpr`, `IndexExpr`, `StarExpr`, and `ParenExpr`.

#### Scenario: Type assertion in container unwrap chain
- **GIVEN** a test that assigns `text := result.Content[0].(TextContent).Text` where `Content` is an interface slice and `result` holds the target function's return value
- **WHEN** the container unwrap pass traces the data flow chain
- **THEN** the `resolveExprRoot` fallback resolves through the `TypeAssertExpr` to find `result` as the root identifier, enabling the chain to continue tracing

#### Scenario: Type assertion does not break existing mappings
- **GIVEN** an assertion that maps via direct identity (confidence 75) on an expression that does not contain a type assertion
- **WHEN** the mapping pipeline runs with the TypeAssertExpr fix
- **THEN** the assertion retains its original confidence and effect ID (zero regression)

### Requirement: Container unwrap fixture MUST cover type assertion patterns

The `containerunwrap` test fixture MUST include a variant that uses an interface slice with a type assertion, mirroring the real-world MCP SDK pattern where `Content` is `[]Content` (interface) and tests use `result.Content[0].(TextContent).Text`.

#### Scenario: Interface slice with type assertion produces mapped assertion
- **GIVEN** a test fixture function returning a struct with an interface slice field, and a test that type-asserts an element, extracts a field, unmarshals, and asserts on the result
- **WHEN** the assertion mapping pipeline runs on this fixture
- **THEN** the assertion on the unmarshal output is mapped to the `ReturnValue` effect at confidence 55
