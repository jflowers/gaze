# Research: Container Type Assertion Mapping

**Date**: 2026-03-24
**Branch**: `034-container-assertion-mapping`

## R1: Forward Data Flow Tracing via Go AST

**Decision**: Use `ast.Inspect` with `Pos()` comparison for forward tracing, combined with `types.Info.Uses`/`types.Info.Defs` for variable identity.

**Rationale**: The existing mapping pipeline uses backward tracing (from assertions to the target call). The container unwrap pass requires forward tracing (from the target call's return variable through intermediate assignments to the assertion). Go AST guarantees `BlockStmt.List` ordering matches source order. Using `Pos()` comparison is safer than index-based comparison because the target call may be nested inside a `t.Run` closure.

**Alternatives considered**:
- SSA-based forward tracing: More precise but unnecessary complexity for sequential assignment patterns. SSA would handle phi nodes and control flow, but the container-unmarshal-assert pattern is always sequential within a single test function body.
- Index-based `Body.List` iteration: Simpler but breaks when statements are nested in closures or sub-blocks.

## R2: Transformation Signature Detection

**Decision**: Use `types.Info.Types[callExpr.Fun].Type` to get the function signature, then inspect parameter types for the structural pattern (byte-like input + pointer destination).

**Rationale**: The codebase already uses `info.Types[expr]` extensively for type inspection (mutation.go, p1effects.go, p2effects.go). The `types.Signature` from a `CallExpr.Fun` provides parameter types via `Params()`, enabling structural matching without function name hardcoding.

**Key patterns**:
- Byte-like parameter: `*types.Slice` with `Elem() == types.Byte`, or `*types.Basic` with `Kind() == types.String`
- Pointer destination: `*types.Pointer` (direct check)
- io.Reader: Method set check for `Read([]byte) (int, error)` via `types.NewMethodSet`
- Method calls: `sig.Recv()` is separate from `sig.Params()`; receiver is not in the parameter list
- Variadic: `sig.Variadic()` true means last `Params()` entry is `[]T`

**Alternatives considered**:
- Hardcoded function list (`json.Unmarshal`, `xml.Unmarshal`, etc.): Simpler but violates FR-001/FR-008 (structural, not type-specific). Would miss third-party unmarshal functions, proto decoders, custom deserializers.
- Interface-based detection (require `encoding.BinaryUnmarshaler`): Too narrow; `json.Unmarshal` is a package-level function, not a method on the destination.

## R3: Variable Identity Across Assignments

**Decision**: Use `types.Object` pointer identity for tracking variables through assignment chains.

**Rationale**: Go's type checker assigns a unique `types.Object` (specifically `*types.Var`) to each variable declaration. All references to that variable (via `TypesInfo.Uses`) return the same pointer. This is how the existing mapping pipeline confirms variable identity (matchAssertionToEffect Pass 1 at mapping.go:797).

**Key implementation detail**: For `json.Unmarshal(body, &data)`, the `&data` is an `*ast.UnaryExpr` with `Op == token.AND` and `X` being the `*ast.Ident` for `data`. Must unwrap the `&` operator to reach the identifier and look up `info.Uses[ident]`.

**Gotcha**: Variable shadowing creates a new `types.Object` for the inner declaration. Not a concern for the container pattern since the unmarshal destination and the assertion reference the same declaration scope.

## R4: Reusable Infrastructure

The following existing functions can be reused or adapted:

| Component | Location | Reuse |
|-----------|----------|-------|
| `findAssignLHS` | `mapping.go:508` | Pattern for AST assignment walking; adapt for forward direction |
| `containsPos` | `mapping.go:548` | Position range checking for "after" filtering |
| `resolveExprRoot` | `mapping.go:674` | Unwinding `data.Field` / `data["key"]` to root ident for RHS checking |
| `mapAssignLHSToEffects` | `mapping.go:336` | TypesInfo.Defs/Uses lookup pattern for assignment LHS |
| Match Pass 1 | `mapping.go:762` | ast.Inspect + info.Uses identity checking pattern |

No new external dependencies required. All infrastructure uses standard library `go/ast`, `go/types`, and existing `golang.org/x/tools` imports.

## R5: Chain Depth and the MCP Test Pattern

**Decision**: Support tracing through up to 6 intermediate steps, covering the full MCP pattern with margin.

The real-world MCP test pattern has 4 steps:
1. `result := target(ctx, req)` — assign return value
2. `content := result.Content[0].(mcp.TextContent)` — field access + index + type assertion
3. `json.Unmarshal([]byte(content.Text), &data)` — field access + type conversion + transformation
4. `data["status"]` — map index access in assertion

Steps 2-3 may span multiple AST statements or collapse into fewer. A depth limit of 6 provides margin for more complex unwrap chains without risk of exponential blowup.

**Rationale**: The existing inline call pass has no depth limit (it walks the entire expression tree). The AI mapper has no depth limit either. A depth limit of 6 provides safety bounds while covering all realistic patterns. Each step is O(n) over the function's statement count, so total cost is O(6n) — negligible for test functions.

## R6: Error Exclusion from Transformation Calls

**Decision**: When a transformation call returns `(T, error)` and the test assigns both results, the error variable is explicitly excluded from the tracked variable set.

**Implementation**: After detecting a transformation call at position P, find the enclosing assignment statement. If the assignment has multiple LHS expressions (e.g., `err := json.Unmarshal(body, &data)`), only the pointer destination variable becomes the next tracked link. The error LHS variable is not added to the tracked set. This is determined by the parameter index — the pointer destination parameter's position maps to the destination variable, not the error return.

Note: `json.Unmarshal` returns only `error`, not `(T, error)`. The destination is a call argument (pointer), not a return value. The error exclusion applies to the function's return value (the error), not its arguments. The pointer destination variable is identified from the call arguments (the `&data` argument at the pointer parameter position), not from the assignment LHS.
