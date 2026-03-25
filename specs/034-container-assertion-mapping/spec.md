# Feature Specification: Container Type Assertion Mapping

**Feature Branch**: `034-container-assertion-mapping`
**Created**: 2026-03-24
**Status**: Draft
**Input**: GitHub Issue #77 — Classifier: MCP tool functions with *mcp.CallToolResult return type score Q3 despite strong assertions

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Assertions on Unwrapped Container Fields Count Toward Contract Coverage (Priority: P1)

A developer writes a Go function that returns a generic container type (e.g., a third-party SDK result wrapper, an HTTP response, a raw JSON message). Their tests call the function, extract the wrapped payload (e.g., by type-asserting a field and unmarshalling JSON), and assert on the resulting values. When they run Gaze, contract coverage should reflect these assertions as covering the function's return value effect — the same way direct variable assertions work today.

**Why this priority**: This is the core problem. 25 functions in the reference project hit this pattern, making up the majority of the GazeCRAPload. Without this, contract coverage is systematically understated for any project that wraps results in generic containers.

**Independent Test**: Can be fully tested by creating a Go test fixture with a function returning a wrapper type, a test that unmarshals and asserts on fields of the wrapper, and verifying that the assertion mapping connects the field-level assertions back to the ReturnValue side effect.

**Acceptance Scenarios**:

1. **Given** a function that returns a generic container type wrapping a JSON body, **When** a test assigns the return value, unmarshals a field of the container into a map, and asserts on a key of that map, **Then** the assertion is mapped to the function's ReturnValue side effect with confidence >= 55.
2. **Given** a function returning a wrapper type and a test with 4 assertions on unwrapped fields, **When** Gaze computes contract coverage, **Then** contract coverage is > 0% (the assertions are not all unmapped).
3. **Given** a function returning a typed struct (not a container), **When** Gaze runs assertion mapping, **Then** behavior is identical to today — no regression in existing mapping accuracy.

---

### User Story 2 - Multi-Step Unwrap Chains Are Traced (Priority: P2)

A developer's test accesses the return value through multiple intermediate steps before reaching the assertion target. For example: assign the return value, access a field of a field, type-assert to a concrete type, extract a string property, unmarshal into a Go value, then assert on fields of that Go value. The mapping should trace through the full chain, not just one level of field access.

**Why this priority**: The real-world test pattern involves 3-4 intermediate steps. A solution that only handles one level of indirection would miss the actual use case.

**Independent Test**: Can be tested with a fixture that chains type assertion, field access, and unmarshal before asserting, and verifying the assertion is mapped.

**Acceptance Scenarios**:

1. **Given** a test that chains field access, type assertion, and unmarshal before asserting, **When** Gaze maps assertions, **Then** assertions on the unmarshal output variable are mapped to the ReturnValue effect.
2. **Given** an unwrap chain with 4 intermediate steps, **When** Gaze maps assertions, **Then** the mapping succeeds (the chain depth does not cause a failure).

---

### User Story 3 - Container Mapping Works Alongside Existing Mapping Passes (Priority: P3)

The new container-aware mapping integrates with Gaze's existing multi-pass assertion mapping pipeline (direct identity, indirect root, helper bridge, inline call, AI mapper) without disrupting their behavior or confidence levels. The container mapping is a new pass in the pipeline with its own confidence tier.

**Why this priority**: The mapping pipeline is carefully layered with decreasing confidence levels. The new pass must fit into this architecture cleanly. It should only activate when the higher-confidence passes fail, and should not interfere with assertions that already map successfully.

**Independent Test**: Can be tested by running the full mapping pipeline on existing test fixtures and verifying that all previously-mapped assertions retain their original confidence and effect IDs.

**Acceptance Scenarios**:

1. **Given** an assertion that maps via direct identity (confidence 75), **When** Gaze runs the full pipeline including the new container pass, **Then** the assertion retains confidence 75 (the container pass does not override it).
2. **Given** an unmapped assertion where only the container pass can match, **When** the pipeline runs, **Then** the assertion is mapped via the container pass with a confidence level lower than the inline call pass (60) but higher than or equal to the AI mapper (50).

---

### Edge Cases

- What happens when the test unmarshals into a struct (not a map)? The mapping should still trace through the unmarshal call and recognize assertions on struct fields.
- What happens when the unwrap step fails at runtime (e.g., nil pointer, bad type assertion)? The mapping operates on static analysis, not runtime values. If the AST contains the unwrap pattern, it should be traced regardless of runtime behavior.
- What happens when multiple return values are unwrapped separately? Each unwrap chain should be traced independently to its respective side effect.
- What happens when the container is passed to a helper function before being unwrapped? This should be handled by the existing helper bridge pass, not the container pass. The two passes should not conflict.
- What happens when the function returns a non-container type but the test still unmarshals it? The container pass should still trace through the unmarshal — the pattern is not specific to any particular return type.
- What happens when the test checks the error return of the unmarshal call? The error assertion is not mapped to the ReturnValue effect — it validates the unmarshal operation, not the target function's behavioral contract.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The assertion mapping pipeline MUST detect when a test assigns the return value of a target function to a variable, then accesses a field or element of that variable, and the accessed value flows into a transformation call, and the output of that transformation is subsequently used in an assertion. A transformation call is recognized structurally by its signature pattern (a function accepting a byte-like input and a pointer destination argument), not by a hardcoded list of function names.
- **FR-002**: When the pattern in FR-001 is detected, the mapping MUST connect the assertion to the original function's ReturnValue side effect.
- **FR-003**: The container mapping pass MUST be inserted into the existing multi-pass pipeline at a confidence level between the inline call pass (60) and the AI mapper pass (50) — specifically at confidence 55.
- **FR-004**: The container mapping MUST trace through at least 4 intermediate steps between the target function call and the assertion (field access, type assertion, index access, unmarshal call).
- **FR-005**: The container mapping MUST NOT alter the confidence or mapping of assertions that are already matched by higher-confidence passes (direct identity at 75, indirect root at 65, helper bridge at 70, inline call at 60).
- **FR-006**: The container mapping MUST handle unmarshal into both map-typed and struct-typed destinations.
- **FR-007**: The assertion mapping accuracy ratchet (currently 85.0%) MUST NOT regress after this change.
- **FR-008**: The container mapping MUST work with any return type — it is not specific to any particular wrapper type. The pattern is structural: return value to field/element access to transformation to assertion.
- **FR-009**: The container mapping MUST NOT map assertions on the error return of a transformation call (e.g., the error from an unmarshal operation) to the target function's ReturnValue effect. Only assertions on the destination value of the transformation are mapped.

### Key Entities

- **Unwrap Chain**: A sequence of operations (field access, index access, type assertion, function call) that connects a variable holding a function's return value to a variable used in an assertion expression. Each link in the chain is a data-flow step.
- **Transformation Call**: A function call in the unwrap chain that converts one representation to another (e.g., unmarshal converts bytes to a Go value). Recognized structurally by signature pattern — any function accepting a byte-like input and a pointer destination argument — rather than by name. The mapping must bridge across these calls by connecting the input argument (derived from the return value) to the output argument (used in assertions).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Functions returning generic container types with test assertions on unwrapped fields achieve > 0% contract coverage (currently 0% for the 25 affected functions in the reference project).
- **SC-002**: The assertion mapping accuracy ratchet floor holds at >= 85.0% after the change (no regression on existing test fixtures).
- **SC-003**: At least 50% of previously-unmapped assertions that follow the container-unwrap-assert pattern are mapped by the new pass (measured against a test fixture representing the container test pattern).
- **SC-004**: No existing mapped assertions change their confidence level or effect ID after the change (zero regression on direct, indirect, helper bridge, and inline call passes).

## Clarifications

### Session 2026-03-24

- Q: Which transformation functions should the container pass recognize — a hardcoded list or a structural pattern? → A: Structural signature pattern (any function accepting byte-like input and pointer destination), not a hardcoded function list.
- Q: Should unmarshal error assertions be mapped to the ReturnValue effect? → A: No — exclude them. The error belongs to the unmarshal operation, not the target function's contract.

## Assumptions

- The container-unwrap-assert pattern is structural and can be detected via static AST analysis without requiring type-specific knowledge of particular container types.
- The transformation function (e.g., unmarshal) takes a pointer argument as the destination; the mapping traces from that pointer argument's base variable to subsequent assertions.
- The existing test fixture suite in Gaze's own test data is representative enough to validate the non-regression requirement (SC-002, SC-004). A new fixture will be added to validate the container pattern specifically.
- The confidence level of 55 for the container pass appropriately reflects the additional indirection uncertainty compared to direct (75) and indirect (65) mapping, while providing stronger evidence than AI-assisted mapping (50).
