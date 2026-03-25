# Research: Fix Ambiguous Classification for Clear Return Types

**Date**: 2026-03-25
**Branch**: `035-fix-ambiguous-classification`

## R1: Naming Convention — "New" Prefix Inclusion

**Decision**: Add `"New"` (uppercase only) to `contractualPrefixes` in `naming.go` with `impliesFor: [ReturnValue, ErrorReturn]`.

**Rationale**: `New`/`NewXxx` is the most common Go constructor naming convention (Effective Go, standard library). It currently contributes zero signal, leaving P0 effects on constructor functions in the ambiguous range. Adding it with `ReturnValue`/`ErrorReturn` implications matches the semantic intent: constructors create and return new values. The uppercase-only variant is consistent with how other contractual prefixes are defined (e.g., `"Get"`, `"Build"`, `"Parse"` — all uppercase).

**Alternatives considered**:
- Adding both `"New"` and `"new"`: Rejected for consistency — no other contractual prefix has a lowercase variant. Unexported constructors (`newHelper`) receive reduced visibility signal anyway, making them appropriately harder to classify as contractual.
- Adding `"New"` with `nil` impliesFor (all effect types): Rejected — constructors primarily produce return values, not mutations. Using `nil` would over-boost mutation effects on constructor functions.
- Adding `"Create"` as well: Deferred — `"Create"` is less idiomatic in Go than in other languages. Can be added later if evidence supports it.

## R2: GoDoc — Reduced Signal for Non-Matching Effect Types

**Decision**: When a contractual keyword is found in GoDoc but the effect type does not match the keyword's `impliesFor` list, return a reduced positive signal of +5 (one-third of the full +15 weight).

**Rationale**: The presence of any contractual keyword in GoDoc establishes that the function has an intentionally observable side effect. Even if the specific keyword doesn't match the effect type being classified, the GoDoc's contractual tone is a positive signal. A reduced weight (+5) reflects the lower certainty compared to a direct type match (+15), while avoiding the current zero-signal outcome that leaves effects ambiguous on well-documented functions.

**Arithmetic validation**:
- P0 effect (base 75) + visibility (+14) + reduced GoDoc (+5) = 94 → contractual ✓
- P0 effect (base 75) + reduced GoDoc (+5) + no visibility = 80 → contractual (borderline) ✓
- P1 effect (base 60) + visibility (+14) + reduced GoDoc (+5) = 79 → ambiguous (appropriate — P1 effects need stronger evidence) ✓
- P2 effect (base 50) + visibility (+14) + reduced GoDoc (+5) = 69 → ambiguous (appropriate) ✓
- P0 effect (base 75) + reduced GoDoc (+5) + incidental naming (-10) = 70, then contradiction (-20) = 50 → incidental/ambiguous boundary (appropriate — mixed signals should be ambiguous) ✓

**Alternatives considered**:
- Full +15 weight for all effect types (remove type matching): Rejected — this would over-boost incidental effects. A function that "returns a value" and also has a `GlobalMutation` P1 effect should not get the full +15 for the mutation.
- Half weight (+7 or +8): Considered. +7 would still produce contractual results for P0 with visibility (75 + 14 + 7 = 96). However, +5 provides a more conservative boost and better distinguishes direct-match (+15) from indirect-match (+5) in the reasoning output. The gap between 5 and 15 is clearer than between 7 and 15.
- Zero weight (status quo): Rejected — this is the root cause of issue #78. Functions with clear GoDoc using contractual language should not have their secondary effects stuck at zero GoDoc signal.

## R3: Interaction Between Changes

**Decision**: The naming convention change (R1) and GoDoc change (R2) are independent and additive. They do not interact in any way that creates unexpected scoring behavior.

**Arithmetic for the reported issue #78 functions after both changes**:

For `func New(baseURL, authToken string) *Client` with `ReturnValue` (P0):
- Base: 50 + 25 (P0) = 75
- Naming: +10 (`"New"` prefix matches `ReturnValue`)
- Visibility: +14 (exported function +8, exported return type +6)
- GoDoc "returns": +15 (direct match for `ReturnValue`)
- Total: 114 → clamped to 100 → contractual ✓

For the same function with a hypothetical `PointerArgMutation` (P0):
- Base: 50 + 25 (P0) = 75
- Naming: +0 (`"New"` prefix does not imply `PointerArgMutation`)
- Visibility: +14
- GoDoc "returns": +5 (reduced signal — keyword found but type doesn't match)
- Total: 94 → contractual ✓

## R4: Non-Regression Strategy

**Decision**: Use the existing `TestSC001_MechanicalContractualAccuracy` acceptance test as the primary non-regression gate. This test loads real fixtures and verifies classification accuracy against known-good labels.

**Rationale**: The test already covers the critical path — it would catch any demotion from contractual to ambiguous. Adding a specific regression test for the `"New"` prefix provides targeted coverage for the new behavior.

**Risk assessment**: The only functions at risk of regression are those with incidental naming that currently receive zero naming signal but would now receive zero naming signal (unchanged) — and functions with GoDoc that currently receive zero GoDoc signal for non-matching types but would now receive +5 (only upward, never downward). Neither change can reduce an existing score.
