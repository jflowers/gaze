# Implementation Plan: Fix Ambiguous Classification for Clear Return Types

**Branch**: `035-fix-ambiguous-classification` | **Date**: 2026-03-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/035-fix-ambiguous-classification/spec.md`

## Summary

Fix the classification scoring engine so that exported Go constructor functions (`New`/`NewXxx`) and functions with GoDoc contractual keywords receive appropriate confidence scores, preventing false "ambiguous" classifications for P0 effects. Two changes: (1) add `"New"` to the naming convention prefix list, and (2) introduce a reduced GoDoc signal for non-matching effect types. Both changes are confined to `internal/classify/`.

## Technical Context

**Language/Version**: Go 1.25+ (module minimum per go.mod directive)
**Primary Dependencies**: `go/ast`, `go/types`, `golang.org/x/tools/go/packages` (all existing; no new dependencies)
**Storage**: N/A — no persistence changes
**Testing**: Standard library `testing` package; existing classification acceptance tests (`TestSC001`–`TestSC006`), unit tests for each signal analyzer, integration tests with real fixtures
**Target Platform**: Cross-platform CLI (darwin/linux x amd64/arm64)
**Project Type**: Single Go module
**Performance Goals**: Static analysis — no runtime performance constraints. Signal analyzers operate O(1) per effect; no algorithmic change.
**Constraints**: Changes limited to `internal/classify/` package. No new external dependencies. Contractual threshold (80) and incidental threshold (50) unchanged. Existing classifications must not regress.
**Scale/Scope**: Affects 3 files in `internal/classify/` (naming.go, godoc.go, and their tests). No cross-package impacts.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Accuracy — PASS

- The changes directly reduce false negatives: constructor functions with clear return types that were incorrectly classified as "ambiguous" will now be classified as "contractual".
- FR-004 and FR-005 ensure no false positives are introduced (no existing contractual or incidental classification is altered).
- SC-002 enforces zero regression on existing test fixtures.
- The `"New"` prefix is the most common Go constructor naming convention — adding it to the prefix list corrects a clear detection gap.

### II. Minimal Assumptions — PASS

- The `"New"` prefix follows established Go conventions (Effective Go, standard library patterns). Adding it does not impose new assumptions on user code.
- The GoDoc reduced signal applies the same structural analysis already in place — it broadens the scope of existing keyword matching without requiring new user annotations.
- No source annotation or restructuring required.

### III. Actionable Output — PASS

- Functions that were previously reported as "ambiguous" with zero contract coverage will now be "contractual" with measurable contract coverage, directly improving the actionability of quality reports.
- The confidence score change is visible in JSON output, allowing users to understand why the classification changed.

### IV. Testability — PASS

- Both changes are independently testable via the existing signal analyzer unit test infrastructure.
- The `"New"` prefix can be validated with a single test case in `TestNamingSignal_ContractualPrefixes`.
- The reduced GoDoc signal can be validated with test cases in `TestAnalyzeGodocSignal_ContractualKeywords`.
- **Coverage strategy**: Unit tests for each changed function, non-regression via `TestSC001_MechanicalContractualAccuracy`, new acceptance test for the `New` prefix pattern.

## Project Structure

### Documentation (this feature)

```text
specs/035-fix-ambiguous-classification/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
internal/classify/
├── naming.go            # Add "New"/"new" to contractualPrefixes
├── godoc.go             # Add reduced signal for non-matching effect types
├── classify_test.go     # Add test cases for New prefix and reduced GoDoc signal
├── godoc_test.go        # Add test cases for reduced GoDoc signal
└── (all other files unchanged)
```

**Structure Decision**: All changes are within `internal/classify/`. No new files created. The naming.go change is a 1-line addition to the prefix list. The godoc.go change is ~5 lines of logic for the reduced signal path.

## Design

### Change 1: Add "New" to Naming Convention Prefixes

In `naming.go`, add to the `contractualPrefixes` slice:

```go
{"New", []taxonomy.SideEffectType{taxonomy.ReturnValue, taxonomy.ErrorReturn}},
```

This follows the exact same structure as existing entries like `{"Get", ...}` and `{"Build", ...}`. The `impliesFor` list includes `ReturnValue` and `ErrorReturn` — the two effects a constructor function produces.

The matching mechanism (`strings.HasPrefix`) already handles both exact match (`"New"`) and prefix variants (`"NewClient"`, `"NewStore"`). The lowercase variant (`"new"`) is handled by the existing incidental/contractual prefix check flow at lines 66-74/77-98, which checks both uppercase and lowercase variants via the prefix list and `strings.HasPrefix`.

Wait — reviewing the code more carefully: the `contractualPrefixes` only contains uppercase entries, and the match at line 78 uses `strings.HasPrefix(funcName, cp.prefix)`. For unexported functions like `newHelper`, the match would fail because `"n" != "N"`. To support FR-002, I need to also add `{"new", ...}` as a separate entry, consistent with how `"log"/"Log"` are handled in `incidentalPrefixes`.

However, looking at the incidental prefixes more closely: they have both `"log"` and `"Log"` as separate entries (lines 39-44). But the contractual prefixes only have uppercase entries (e.g., `"Get"` but not `"get"`). This means unexported functions like `getConfig` already receive zero naming signal from the contractual prefix list. This is by design — unexported functions are less likely to be contractual API surface.

For FR-002, the spec says to recognize `"new"` "consistent with how other prefixes are handled." Since other contractual prefixes only handle uppercase, adding lowercase would be inconsistent. The spec's acceptance scenario for `newHelper` already acknowledges that "the overall score may remain ambiguous due to reduced visibility signal" — so not matching the lowercase variant is acceptable behavior.

**Decision**: Add only `"New"` (uppercase) to `contractualPrefixes`, consistent with the existing pattern. This satisfies FR-001, FR-002 (consistency), and FR-006.

### Change 2: Reduced GoDoc Signal for Non-Matching Effect Types

In `godoc.go`, when a contractual keyword is found in the GoDoc but the effect type does not match any entry in the keyword's `impliesFor` list, return a reduced positive signal instead of zero.

Current flow (lines 59-72):
1. For each contractual keyword, check if `docText` contains the keyword
2. If found, check if `effectType` matches any `impliesFor` entry
3. If match: return weight +15
4. If no match: continue to next keyword (eventually return zero)

New flow:
1. Same as above for steps 1-3
2. If a contractual keyword is found but no `impliesFor` entry matches: track that a contractual keyword was found
3. After exhausting all keywords: if a contractual keyword was found but never matched the effect type, return a reduced weight (+5)

The reduced weight of +5 is chosen because:
- It is less than the full keyword weight (+15), satisfying FR-003
- It is greater than zero, satisfying FR-003
- It is less than half the full weight (+7.5), providing a conservative boost
- For a P0 effect (base 75) with visibility (+14) and this reduced GoDoc (+5), the total is 94 — contractual
- For a P2 effect (base 50) with no other signals and this reduced GoDoc (+5), the total is 55 — still ambiguous (appropriate, because a P2 effect should not become contractual from GoDoc alone)

**No contradiction risk**: The reduced GoDoc signal is always positive (+5), so it cannot trigger the contradiction penalty unless combined with a separate negative signal (e.g., incidental naming). This is the same behavior as the full +15 weight.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
