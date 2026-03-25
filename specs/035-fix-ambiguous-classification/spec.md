# Feature Specification: Fix Ambiguous Classification for Clear Return Types

**Feature Branch**: `035-fix-ambiguous-classification`
**Created**: 2026-03-25
**Status**: Draft
**Input**: GitHub Issue #78 — Classifier: all_effects_ambiguous for functions with clear return types and GoDoc

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Constructor Functions Are Classified as Contractual (Priority: P1)

A developer writes a Go constructor function following the idiomatic `New` / `NewXxx` pattern (e.g., `func New(baseURL string) *Client`). The function has GoDoc describing what it returns. When they run Gaze, the function's return value effect should be classified as contractual, not ambiguous. Currently, the naming convention analyzer contributes zero signal because `"New"` is not in the recognized prefix list, causing functions that use this ubiquitous Go pattern to receive an unnecessarily low confidence score.

**Why this priority**: The `New`/`NewXxx` naming pattern is the single most common Go constructor convention. Missing it affects every Go project that follows standard naming. This is the lowest-effort, highest-impact fix — adding a naming prefix immediately boosts the confidence score above the contractual threshold for P0 effects.

**Independent Test**: Can be tested by classifying a function named `NewClient` with a `ReturnValue` effect and verifying the naming signal contributes a positive weight, pushing the total confidence above the contractual threshold (80).

**Acceptance Scenarios**:

1. **Given** a function named `New` with a `ReturnValue` (P0) effect and exported visibility, **When** Gaze classifies it, **Then** the naming convention analyzer contributes a positive signal and the effect is classified as "contractual" (confidence >= 80).
2. **Given** a function named `NewClient` with a `ReturnValue` (P0) effect, **When** Gaze classifies it, **Then** it receives the same naming boost as `New` and is classified as "contractual".
3. **Given** a function named `newHelper` (unexported) with a `ReturnValue` effect, **When** Gaze classifies it, **Then** the naming signal still fires for the `new` prefix (case-insensitive or lowercase variant), but the overall score may remain ambiguous due to reduced visibility signal.

---

### User Story 2 - GoDoc "Returns" Keyword Boosts All P0 Return-Related Effects (Priority: P2)

A developer writes GoDoc that says "Returns a *Client configured with..." for a function that has both `ReturnValue` and `ErrorReturn` effects detected. Currently, the GoDoc keyword `"returns"` only applies to `ReturnValue` and `ErrorReturn` effect types — but if the function has other P0 effects (e.g., `PointerArgMutation`), those effects receive zero GoDoc signal even though the GoDoc clearly declares the function's contractual purpose. The GoDoc signal should contribute a general positive signal for all effects on a function whose documentation uses contractual language, not just for the specific effect types in the keyword's `impliesFor` list.

**Why this priority**: This is a broader fix than US1 but addresses the same root cause — effects landing in the ambiguous range because signals are too narrowly scoped. Functions with good GoDoc documentation should not have their secondary effects classified as ambiguous when the documentation clearly establishes the function's contractual nature.

**Independent Test**: Can be tested by classifying a function with GoDoc containing "Returns" and a `PointerArgMutation` effect (P0), and verifying the GoDoc signal contributes a positive weight even though "returns" historically only boosted `ReturnValue`/`ErrorReturn`.

**Acceptance Scenarios**:

1. **Given** a function with GoDoc containing the word "returns" and a `PointerArgMutation` (P0) effect, **When** Gaze classifies it, **Then** the GoDoc analyzer contributes a positive signal (the exact weight may be reduced from the full +15 for a non-matching type, but must be > 0).
2. **Given** a function with GoDoc containing both "returns" and "modifies" and effects for both `ReturnValue` and `ReceiverMutation`, **When** Gaze classifies it, **Then** each effect receives the appropriate GoDoc boost from its matching keyword AND a general positive signal from the overall contractual tone of the GoDoc.

---

### User Story 3 - No Regression for Functions Already Classified Correctly (Priority: P3)

Existing functions that are already classified as contractual, incidental, or correctly ambiguous must not change their classification after these scoring adjustments. The changes to naming conventions and GoDoc signal scoping should only promote effects from ambiguous to contractual — never demote effects from contractual to ambiguous, and never promote effects that should remain ambiguous.

**Why this priority**: Non-regression is essential but is a guardrail, not a feature. It has the lowest priority because it's verified by the existing test suite rather than requiring new user-facing capabilities.

**Independent Test**: Can be tested by running the full classification test suite before and after the change, comparing every function's classification label. Any label change is flagged for review.

**Acceptance Scenarios**:

1. **Given** the existing set of test fixtures with known classifications, **When** Gaze runs classification after the scoring changes, **Then** no function that was previously "contractual" becomes "ambiguous" or "incidental".
2. **Given** a function that was correctly "ambiguous" (e.g., a function with no documentation, no naming convention match, and a P2 effect), **When** Gaze runs classification after the changes, **Then** it remains "ambiguous".

---

### Edge Cases

- What happens when a function is named `New` but has no return value (void function)? The naming signal should still fire, but the lack of a `ReturnValue` effect means the contractual naming signal applies to whatever effects are detected (possibly `ReceiverMutation`).
- What happens when GoDoc contains both contractual and incidental keywords (e.g., "Returns a Client and logs the creation event")? The current behavior evaluates incidental keywords first and returns immediately, shadowing the contractual keyword. This specification does not change that behavior — the incidental-first evaluation order is a separate concern from the issues in #78.
- What happens when the function has no GoDoc at all? The GoDoc signal remains zero. The naming convention and visibility signals alone must be sufficient to push P0 effects above the threshold. With the `New` prefix added (naming +10) and standard visibility (+14), a P0 effect would score 75 + 10 + 14 = 99 — contractual.
- What happens when a function named `NewXxx` returns an unexported type? The naming signal fires (+10), but the visibility signal is reduced (no exported return type). For P0: 75 + 10 + 8 = 93 — still contractual.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The naming convention analyzer MUST recognize the `New` prefix (both `"New"` as an exact match and `"New"` as a prefix like `"NewClient"`, `"NewStore"`) as a contractual naming signal, contributing a positive weight to the confidence score.
- **FR-002**: The naming convention analyzer MUST recognize both uppercase (`"New"`) and lowercase (`"new"`) variants of the constructor prefix, consistent with how other prefixes are handled (e.g., `"Get"` and `"get"`).
- **FR-003**: The GoDoc analyzer MUST contribute a reduced positive signal (less than the full keyword-matched weight but greater than zero) when the GoDoc contains contractual keywords even if the specific effect type being classified does not appear in the keyword's `impliesFor` list. This ensures that a function whose documentation uses contractual language benefits all its detected effects, not just the ones that exactly match the keyword.
- **FR-004**: The naming convention and GoDoc scoring changes MUST NOT cause any function that was previously classified as "contractual" to be reclassified as "ambiguous" or "incidental".
- **FR-005**: The naming convention and GoDoc scoring changes MUST NOT cause any function that was previously classified as "incidental" to be reclassified as "contractual" or "ambiguous".
- **FR-006**: For a function named `New` or `NewXxx` with a P0 effect, exported visibility, and no negative signals, the combined score MUST be >= 80 (contractual threshold) after the naming convention change.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Functions following the `New`/`NewXxx` naming pattern with P0 effects and exported visibility achieve "contractual" classification (confidence >= 80) — resolving the 5 functions reported in issue #78.
- **SC-002**: No existing test fixture function changes its classification label (contractual, ambiguous, or incidental) after the scoring changes — zero regression.
- **SC-003**: The naming convention prefix list includes `New`/`new` and their prefixed variants (`NewXxx`, `newXxx`), consistent with how `Get`/`get`, `Set`/`set` etc. are handled.
- **SC-004**: GoDoc keywords contribute a non-zero positive signal to all effects on the same function, not only to the specific effect types in the keyword's `impliesFor` list — reducing the number of "ambiguous" effects on functions with clear contractual documentation.

## Assumptions

- The naming convention weight for the `New` prefix will be the same as other contractual prefixes (currently +10), not a special-case value.
- The reduced GoDoc signal for non-matching effect types will be less than the full keyword weight (+15) to avoid inflating scores for incidental effects. A reasonable default is half the full weight (e.g., +7 or +8).
- The contradiction penalty (-20) behavior is unchanged — if both positive and negative signals are present, the penalty still applies.
- The contractual threshold (80) and incidental threshold (50) are unchanged.
- These changes affect only the `internal/classify/` package — no changes to effect detection, assertion mapping, or CRAP scoring.
