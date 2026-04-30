# Specification Quality Checklist: Multi-Platform Scaffold Deployment

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-04-30
**Updated**: 2026-04-30 (post-review-council iteration 1)
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [ ] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified
- [x] Documentation impact identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- Adapted from unbound-force/unbound-force PR #144 (Spec 035:
  Multi-Platform Scaffold Deployment) for gaze's simpler scaffold
  system (8 embedded files vs 50, no convention packs, no MCP
  configuration, no DivisorOnly mode).
- FR-007 references specific frontmatter field names (`mode`,
  `temperature`, `tools`, etc.) which are input/output contracts
  necessary for testability, not implementation prescriptions.
- FR-011 was rewritten to remove the `isToolOwned()` function name
  reference per review council feedback.
- Dependencies section updated to note scaffold evolution through
  specs 012, 016, and 017 since the original spec 005.
- Documentation Impact section added per Curator review.
- Website documentation gate requirement added per AGENTS.md.
- User Story 3 Independent Test softened from "byte-identical" to
  "functionally identical" per Architect review.
- US3-AS2 expanded with specific equivalence criteria per Tester review.
- Acceptance scenarios added for FR-015 (US1-AS5) and FR-018 (US2-AS4)
  per Tester/Adversary review.
- Edge case added for `--force` with non-existent platform directory
  per SRE review.
- Extensibility assumption (line 329 in original) removed as redundant
  with FR-016/SC-006 per Architect review.
- Key Entities "Asset" rewritten to remove `internal/scaffold/assets/`
  path reference per Adversary review.

## Outstanding HIGH/CRITICAL Findings (Human Decision Required)

The following findings from the review council require human
judgment. They were NOT auto-fixed:

1. **CRITICAL (Adversary, Tester)**: Missing coverage strategy
   section. Constitution Principle IV requires coverage strategy
   in the plan. The spec should signal testing expectations to
   the planner. Recommendation: Add a coverage strategy note or
   defer to plan.md with an explicit acknowledgment.

2. **HIGH (Adversary, Tester, SRE, Scribe)**: FR-007 frontmatter
   transformation contract is incomplete. The `model` field
   (present in `reviewer-testing.md`) and `agent` field (present
   in `gaze.md` commands) are not addressed. The `name`
   derivation rule is unspecified (e.g., `gaze-reporter.md` →
   `name: gaze-reporter` or `name: Gaze Reporter`?). The drop
   list may be incomplete. Recommendation: Define a complete
   field disposition table or specify "preserve only `name` and
   `description`; drop all other frontmatter fields."

3. **HIGH (Adversary, SRE)**: Cursor format assumptions are
   unvalidated. No Cursor documentation source is cited. No
   degradation strategy if Cursor changes its format. The Scribe
   notes that Cursor may use `.cursor/prompts/` rather than
   `.cursor/commands/`. Recommendation: Create a `research.md`
   artifact validating Cursor's directory structure and
   frontmatter schema before planning.

4. **HIGH (Tester, Adversary)**: FR-016/SC-006 extensibility
   requirement is not objectively testable. "Verified by code
   review" is subjective. Recommendation: Either rewrite as a
   testable requirement (e.g., "a mock platform can be added
   without modifying existing code") or move to architectural
   constraints.

5. **HIGH (Tester)**: Missing cross-platform ownership acceptance
   scenarios. No scenario verifies tool-owned overwrite-on-diff
   behavior per platform, force flag isolation to selected
   platforms only, or ownership classification consistency across
   platforms.

6. **HIGH (SRE)**: Drift detection test
   (`TestEmbeddedAssetsMatchSource`) not addressed. The spec
   must define how the existing drift test is preserved for
   OpenCode and whether a new Cursor-specific validation test
   is needed.
