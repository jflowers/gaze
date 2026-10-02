## Context

`gaze report --ai=opencode` builds a full-fidelity `ReportPayload` (CRAP, Quality, Classify, Docscan as raw JSON plus a `ReportSummary`), then pipes a compacted copy to the AI adapter. The compaction (`CompactForAI()` in `internal/aireport/compact.go`) was introduced by the `report-payload-reduction` change to stay under the AI model's context window.

That change made a now-evidently-wrong trade-off: it kept three **unbounded** arrays that scale with the entire codebase analyzed via `./...` — `crap.scores`, `quality.quality_reports`, and `classify.results` — and stripped the **bounded** worst-offender lists the reporter actually reads. In production this repo now yields a 3.2MB compact payload (~1.05M tokens), which opencode rejects with `prompt is too long: 1057347 tokens > 1000000 maximum`. The `Gaze quality report` CI step (Go 1.24) fails as a result.

This design inverts that trade-off. The compaction remains a one-way projection at the AI serialization boundary; `runJSONPath()` and the canonical `json.Marshal` output are untouched.

## Goals / Non-Goals

### Goals

- Make the compact AI payload fit comfortably under 300KB (~4x margin below the 1M-token window) regardless of codebase size.
- Preserve every field the gaze-reporter prompt reads: summary scalars, quadrant/fix-strategy counts, top-5 worst CRAP/GazeCRAP offenders, top-20 recommended actions, worst-5 coverage tests, classification counts, and coverage gaps with effect type and description.
- Keep `CompactForAI()` a pure projection with no behavior change to analysis, thresholds, or `--format=json`.

### Non-Goals

- Changing the full-fidelity `ReportPayload` structs or the JSON path.
- Changing the gaze-reporter agent prompt (it already references only the preserved fields).
- Filtering at the pipeline-step level (the steps must keep producing full-fidelity data; compaction stays at the serialization boundary).
- Introducing a configurable budget or per-user knobs.

## Decisions

**D1 — Drop `crap.scores`, keep the bounded worst lists.** The reporter's "Top 5 worst CRAP scores" section reads `summary.worst_crap` / `summary.worst_gaze_crap`, and remediation reads `summary.recommended_actions`. These are bounded at 5/5/20 by `buildSummary` (`internal/crap/analyze.go`). The full `scores` array is the second-largest contributor and is never read by the reporter, so it is dropped. `compactCRAPReport.Scores` is removed; `compactCRAPSummary` regains `WorstCRAP`, `WorstGazeCRAP`, and `RecommendedActions`.

**D2 — Bound `quality_reports` to top-N actionable entries, keep `worst_coverage_tests`.** The reporter needs "worst tests by contract coverage" (bottom 5, `WorstCoverageTests`) and the gap details. Full reports array scales with test-target pairs and is dropped from the compact output; we keep at most 50 entries, selected as those with at least one gap, discarded return, ambiguous effect, or unmapped assertion, ordered by ascending contract coverage. `compactPackageSummary` regains `WorstCoverageTests`.

**D3 — Make quality gaps self-contained (chosen over filtering Classify).** The original compaction reduced gaps/discarded/ambiguous effects to bare ID strings, forcing the reporter to cross-reference `classify.results` for descriptions. Since `classify.results` is now dropped, gaps must carry their own description. `compactContractCoverage.Gaps`/`DiscardedReturns` and `compactQualityReport.AmbiguousEffects` become `[]compactSideEffect` (id, type, tier, location, description, target — no classification). This also improves the existing "resilient to nil Classify" behavior, since descriptions no longer depend on Classify at all.

**D4 — Reduce Classify to counts-only.** The reporter's classification distribution is a count summary, already present in the top-level `summary` (`contractual`/`ambiguous`/`incidental`). `compactClassifyField` stops unmarshalling `results` (also saving CPU) and emits `{"version": <v>, "contractual": N, "ambiguous": N, "incidental": N}` sourced from `ReportPayload.Summary`. When `Classify` is nil the field is `null`, preserving step-failure semantics.

**D5 — Strengthen the size-budget test to realistic scale.** The prior 200-function fixture never caught the overflow. The budget test now builds ~2000 functions / ~1500 test-target pairs / ~5000 side effects and asserts `< 300KB`, verifying the invariant holds at the scale that broke CI.

**D6 — Compaction stays at the serialization boundary.** `CompactForAI()` remains the single point of projection; no filtering moves into `runProductionPipeline` or the step functions. This preserves the separation-of-concerns principle documented in the prior change.

## Risks / Trade-offs

- **Risk: The reporter needs a specific per-function score it no longer gets.** Mitigated — the prompt's per-function details (name, CRAP, complexity, coverage, file:line) all come from `worst_crap`/`recommended_actions`, which are preserved. The full `scores` array is never referenced.
- **Risk: A report needs more than 50 quality entries.** The reporter surfaces "top gaps" and "worst tests", both bounded views. 50 actionable entries exceed what a human-facing CI report renders. If future needs arise, the cap is a single constant.
- **Risk: Dropping `classify.results` removes debugging detail (signals).** Accepted — signals were already stripped by the prior change; only label/confidence/reasoning remained, and those are now superseded by the counts plus self-contained gap descriptions.
- **Risk: `worst_crap` entries carry ~14 fields each including per-score cognitive-complexity fields.** Bounded at 5+5+20 entries (~30 objects), negligible vs. the dropped 2000+ arrays.

## Constitution Alignment

This design preserves the constitution alignment established in the proposal: no accuracy change (full JSON and thresholds untouched), fewer assumptions (bounded budget independent of codebase size), stronger actionable output (self-contained gap descriptions), and full testability (unit tests for each projection plus the size-budget and integration tests).
