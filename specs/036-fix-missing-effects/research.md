# Research: Fix Missing Effect Detection

**Date**: 2026-03-25
**Branch**: `036-fix-missing-effects`

## R1: AST-Based Mutation Detection Patterns

**Decision**: Walk the function body AST for `*ast.AssignStmt` nodes where the LHS root identifier matches the receiver name, and for `*ast.CallExpr` nodes where the function is a method call on a receiver field.

**Rationale**: The SSA-based detector looks for `*ssa.Store` instructions that trace back to receiver or pointer parameters. The AST equivalent is checking whether an assignment's LHS or a call expression's target involves the receiver. This covers the most common mutation patterns:

| Pattern | AST Node | Example |
|---------|----------|---------|
| Direct field assignment | `AssignStmt{LHS: SelectorExpr{X: Ident(recv), Sel: field}}` | `si.pages = newPages` |
| Map index assignment | `AssignStmt{LHS: IndexExpr{X: SelectorExpr{X: Ident(recv), Sel: field}}}` | `si.index[key] = value` |
| Method call on field | `CallExpr{Fun: SelectorExpr{X: SelectorExpr{X: Ident(recv), Sel: field}, Sel: method}}` | `si.index.Delete(key)` |
| Append to slice field | `AssignStmt{LHS: SelectorExpr{X: Ident(recv), Sel: field}, RHS: CallExpr(append)}` | `si.items = append(si.items, x)` |

**Alternatives considered**:
- **Type-based detection** (check if the receiver is `*T` and declare a mutation effect unconditionally): Rejected — too many false positives. Read-only methods on pointer receivers are common.
- **Control flow analysis** (walk if/switch/for to find mutations in branches): Deferred — the basic flat AST walk is sufficient for the v1 fallback. SSA provides branch-aware analysis when available.
- **Interface method set analysis** (check if the receiver implements a mutating interface): Deferred — requires type checker integration beyond what the AST walk provides.

## R2: Receiver Name Resolution

**Decision**: Extract the receiver name from `fd.Recv.List[0].Names[0].Name`. Return nil (no effects) if the receiver is unnamed.

**Rationale**: The receiver name is needed to match assignments — e.g., in `func (si *SearchIndex) ReindexPage(...)`, we need `"si"` to recognize `si.field = value` as a receiver mutation. If the receiver is unnamed (`func (*SearchIndex) ReindexPage(...)`), we cannot trace assignments and must bail out. This is rare in practice — the Go style guide recommends always naming receivers.

**Alternatives considered**:
- **Synthesize a name for unnamed receivers**: Rejected — would require modifying the AST, and unnamed receivers almost always indicate the method doesn't use the receiver (read-only or no-op).

## R3: Pointer Receiver Check

**Decision**: Check `fd.Recv.List[0].Type` for `*ast.StarExpr`. If it's not a star expression, the receiver is a value type and mutations are not observable — return nil (FR-004).

**Rationale**: Value receivers are copies. Assigning to `s.field = value` on a value receiver modifies the copy, not the original. This is not an observable side effect from the caller's perspective. The SSA-based detector also only emits `ReceiverMutation` for pointer receivers.

## R4: `no_test_coverage` Coverage Reason

**Decision**: In `BuildContractCoverageFunc`, maintain a parallel set of "functions with detected effects" from the analysis results. When the coverage map lookup returns `ok=false` for a function, check the effects set. If the function has effects: `Reason: "no_test_coverage"`. If no effects: `Reason: "no_effects_detected"`.

**Rationale**: The current code returns default zero values when a function is not in the coverage map, which the CRAP pipeline interprets as "no data." Distinguishing "has effects but no test" from "genuinely no effects" gives users actionable information: the former needs a test, the latter needs investigation into why effects weren't detected.

**Alternatives considered**:
- **Add all functions to the coverage map regardless of test linkage**: Rejected — the coverage map is built from `quality.Assess` output, which only produces reports for functions linked to tests. Changing this would require restructuring the quality pipeline.
- **Report all untested functions separately**: Considered — but the existing pipeline already reports all functions in CRAP output. The `no_test_coverage` reason just provides a better diagnostic for why contract coverage is 0%.

## R5: AST Fallback Annotation

**Decision**: Append `" (AST fallback)"` to the `Description` field of AST-detected mutation effects.

**Rationale**: This is visible in both JSON and text output without any schema changes. Users and agents can see that the effect was detected via the lower-fidelity fallback. If SSA succeeds (FR-005), the fallback is not used and the description is the standard SSA-detected text.

**Alternatives considered**:
- **New `Source` field on `SideEffect`**: Rejected — would require schema changes to `taxonomy.SideEffect` and updates to all JSON consumers.
- **New `Metadata` map on `SideEffect`**: Rejected — same schema change concern.
- **Log-only (no annotation on the effect)**: Rejected — logs are transient and not visible in report output. The user needs to see which effects are approximate.

## R6: Pointer Argument Mutation Detection

**Decision**: For the initial AST fallback, detect pointer argument mutations using the same AST walking approach as receiver mutations — find assignments where the LHS root identifier matches a pointer parameter name.

**Rationale**: The same patterns apply: `param.field = value`, `param.index[key] = value`, `param.Method()`. The SSA-based detector handles both `ReceiverMutation` and `PointerArgMutation` in the same `detectMutations` function, and the AST fallback should mirror this scope.

**Risk**: Pointer parameters are more likely to be used in read-only fashion than receivers (e.g., `page *Page` passed for reading, not writing). However, FR-003 and FR-009 constrain the fallback to only detect actual assignments and method calls, not read access.
