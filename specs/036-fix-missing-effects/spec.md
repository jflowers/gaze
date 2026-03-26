# Feature Specification: Fix Missing Effect Detection

**Feature Branch**: `036-fix-missing-effects`
**Created**: 2026-03-25
**Status**: Draft
**Input**: GitHub Issue #79 — Classifier: no_effects_detected for functions with observable side effects

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Receiver Mutation Detection Degrades Gracefully with AST Fallback (Priority: P1)

A developer writes a void method that mutates its receiver's fields (e.g., `func (si *SearchIndex) ReindexPage(page *Page)`). The method modifies maps, slices, or fields on the receiver struct. Currently, mutation detection relies entirely on SSA (Static Single Assignment) analysis. When SSA construction fails for a package — which happens silently when the dependency graph triggers a builder panic — the analysis returns zero effects for the function. The developer sees `no_effects_detected` despite clear mutations.

The system should detect receiver mutations using AST analysis as a fallback when SSA is unavailable, rather than silently reporting zero effects.

**Why this priority**: This is the highest-impact issue — 3 of the 4 affected functions are void methods with receiver mutations. SSA failure is the root cause, and these functions have no detectable effects without mutation analysis. An AST-based fallback ensures mutation detection works even when SSA degrades.

**Independent Test**: Can be tested by analyzing a function with a pointer receiver that assigns to a receiver field (e.g., `si.pages[id] = page`), with SSA intentionally disabled, and verifying that a `ReceiverMutation` effect is detected via the AST fallback.

**Acceptance Scenarios**:

1. **Given** a void method with a pointer receiver that assigns to a field of the receiver, **When** SSA construction fails and the analysis runs, **Then** at least one `ReceiverMutation` effect is detected via the AST fallback.
2. **Given** a void method that modifies a map on the receiver (e.g., `si.index[key] = value`), **When** SSA is unavailable, **Then** the AST fallback detects the mutation as a `ReceiverMutation` effect.
3. **Given** a function where SSA succeeds normally, **When** the analysis runs, **Then** the SSA-based mutation detection takes precedence and the AST fallback is not used (no duplicate effects).

---

### User Story 2 - Return Value Detection Includes All Exported Functions (Priority: P2)

A developer writes an exported function with a non-error return value (e.g., `func StripBullet(s string) string`). When they run Gaze, the function's return value should be detected as a side effect. Currently, the analysis engine correctly detects return values via AST analysis — but if the function is not linked to any test by the target inference step, it does not appear in the contract coverage report at all, making it invisible to the quality pipeline.

The system should surface functions with detected effects in the quality report even when test-target inference cannot find a matching test.

**Why this priority**: This addresses a pipeline visibility gap rather than a detection gap. The effect IS detected — the problem is downstream in target inference. However, it causes the same user-visible symptom (`no_effects_detected` or missing from report entirely), so it belongs in this fix.

**Independent Test**: Can be tested by analyzing a function that returns a non-error value and verifying a `ReturnValue` effect is detected regardless of whether a test calls it.

**Acceptance Scenarios**:

1. **Given** an exported function with a `string` return type, **When** the analysis engine runs on its package, **Then** a `ReturnValue` effect is detected with type `ReturnValue` and tier P0.
2. **Given** a function with a detected `ReturnValue` effect but no test calling it, **When** the contract coverage report is generated, **Then** the function appears in the report with reason `no_test_coverage` (not `no_effects_detected`), distinguishing "no test" from "no effects."

---

### User Story 3 - Diagnostic Visibility for SSA Degradation (Priority: P3)

When SSA construction fails for a package, the developer has no way to know that mutation detection was silently skipped. The system should surface SSA degradation diagnostics so developers understand why their functions report zero effects and can take action (e.g., updating dependencies, reporting upstream bugs).

**Why this priority**: Diagnostic visibility is important for trust and debugging, but does not directly fix the detection gap. US1 (AST fallback) provides the actual fix; US3 ensures users understand when the fallback was used.

**Independent Test**: Can be tested by analyzing a package where SSA construction fails and verifying that the output includes a diagnostic message identifying the degraded package.

**Acceptance Scenarios**:

1. **Given** a package where SSA construction fails, **When** the analysis runs, **Then** the output includes a diagnostic indicating which package experienced SSA degradation.
2. **Given** a function analyzed with the AST fallback (because SSA failed), **When** the effects are reported, **Then** each AST-detected mutation effect includes a flag or annotation indicating it was detected via the fallback method rather than SSA.

---

### Edge Cases

- What happens when a function has both return values AND receiver mutations, and SSA fails? The return values are detected via AST (existing behavior), and the receiver mutations should be detected via the AST fallback (new behavior). Both sets of effects should appear in the results.
- What happens when the AST fallback detects a "mutation" that is actually a local variable assignment (not a receiver field)? The fallback must distinguish between receiver field assignments (`si.field = value`) and local variable assignments (`x := value`) to avoid false positives.
- What happens when a method uses a value receiver (not a pointer receiver)? Value receiver mutations are not observable by the caller, so the AST fallback should NOT detect them as `ReceiverMutation` effects.
- What happens when a function calls a method on the receiver (e.g., `si.index.Set(key, value)`) rather than assigning directly? The AST fallback should detect method calls on receiver fields as potential mutations, since method calls on pointer-typed fields may mutate shared state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When SSA construction fails for a package, the analysis engine MUST fall back to AST-based mutation detection for functions in that package, rather than silently reporting zero mutation effects.
- **FR-002**: The AST-based mutation fallback MUST detect direct field assignments on pointer receiver parameters (e.g., `si.field = value`, `si.index[key] = value`) as `ReceiverMutation` effects.
- **FR-003**: The AST-based mutation fallback MUST NOT detect assignments to local variables or parameters (non-receiver) as `ReceiverMutation` effects.
- **FR-004**: The AST-based mutation fallback MUST NOT detect mutations on value receivers (non-pointer) as `ReceiverMutation` effects, since these are not observable by the caller.
- **FR-005**: When SSA succeeds, SSA-based mutation detection MUST take precedence over the AST fallback. The fallback MUST NOT produce duplicate effects alongside SSA results.
- **FR-006**: Functions with detected effects that are not linked to any test by target inference MUST appear in the contract coverage report with a reason that distinguishes "no test" from "no effects" (e.g., `no_test_coverage` vs `no_effects_detected`).
- **FR-007**: When SSA construction fails, the analysis output MUST include a diagnostic identifying which package experienced degradation, consistent with the existing `SSADegradedPackages` reporting mechanism.
- **FR-008**: The AST fallback MUST detect method calls on receiver fields (e.g., `si.map.Delete(key)`) as potential `ReceiverMutation` effects, since method calls on pointer-typed receiver fields may mutate shared state.
- **FR-009**: The AST-based mutation detection MUST NOT introduce false positives for read-only field access (e.g., `_ = si.field`, `fmt.Println(si.field)`) — only assignments and method calls on receiver fields count as mutations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Void methods with pointer receivers that assign to receiver fields detect at least one `ReceiverMutation` effect even when SSA is unavailable — resolving 3 of the 4 functions reported in issue #79.
- **SC-002**: Exported functions with non-error return values detect a `ReturnValue` effect — confirmed for `StripBullet` and similar simple functions.
- **SC-003**: No existing function that previously had detected effects loses any effects after this change — zero regression on effect detection.
- **SC-004**: When SSA fails and the AST fallback is used, the analysis output includes a diagnostic identifying the degraded package.

## Assumptions

- The AST-based mutation fallback will produce lower-fidelity results than SSA-based detection (it cannot trace data flow or alias analysis), but detecting direct field assignments and method calls on receiver fields covers the most common mutation patterns.
- The `ReceiverMutation` effects detected by the AST fallback will carry the same type and tier (P0) as SSA-detected mutations, so they flow through classification and contract coverage identically.
- The existing `SSADegradedPackages` field in the CRAP summary and `SSADegraded` field in quality reports provide the infrastructure for FR-007 diagnostics — this feature extends their use, not introduces a new mechanism.
- The `no_test_coverage` reason (FR-006) is a new reason string that complements the existing `no_effects_detected` and `all_effects_ambiguous` reasons in the contract coverage pipeline.
- The AST fallback for method calls on receiver fields (FR-008) may produce false positives for read-only method calls (e.g., `si.map.Get(key)` where `Get` does not mutate). This is an acceptable trade-off for the AST fallback — SSA analysis provides higher fidelity when available, and the AST fallback provides "better than nothing" coverage when SSA fails.
