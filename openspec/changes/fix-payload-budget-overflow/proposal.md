## Why

The `gaze report --ai=opencode` CI step fails on Go 1.24 because the compact AI payload exceeds opencode's 1M-token context window. A production run produced a 3,225,752-byte (~1.05M tokens) compact payload, which opencode rejected with `prompt is too long: 1057347 tokens > 1000000 maximum`, causing `gaze` to fail with "AI adapter returned empty output (FR-016)".

The existing `CompactForAI()` compaction (change `report-payload-reduction`) was calibrated against a synthetic ~200-function payload and stripped the *wrong* fields. It preserved three **unbounded** arrays that scale with the entire codebase analyzed via `./...`:

- `crap.scores` — one entry per analyzed function (~2000+ for this repo)
- `quality.quality_reports` — one entry per test-target pair
- `classify.results` — one entry per detected side effect (the largest contributor)

At the same time it *stripped* exactly the bounded, high-value fields the reporter prompt actually reads: `crap.summary.worst_crap` (top 5), `crap.summary.worst_gaze_crap` (top 5), `crap.summary.recommended_actions` (top 20), and `quality_summary.worst_coverage_tests` (bottom 5).

The codebase has grown well past the ~200-function calibration point (notably the recent cognitive-complexity feature added per-score fields), pushing the compact payload over the token limit. This is a CI-blocking defect that must be fixed so the Go 1.24 job is green again.

## What Changes

Invert the compaction strategy in `internal/aireport/compact.go`: bound the unbounded arrays and preserve the bounded worst-offender lists the reporter needs.

1. **CRAP**: drop the full `scores` array; restore `summary.worst_crap`, `summary.worst_gaze_crap`, and `summary.recommended_actions` (bounded: 5 + 5 + 20 entries).
2. **Quality**: restore `quality_summary.worst_coverage_tests` (bottom 5); bound `quality_reports` to the top-50 actionable entries (those with gaps, ambiguous effects, or unmapped assertions) sorted by lowest contract coverage; replace bare side-effect ID strings with self-contained minimal side-effect objects carrying `id`, `type`, `tier`, `location`, `description`, and `target` so the reporter gets effect type and description without cross-referencing Classify.
3. **Classify**: drop the `results` array entirely; emit a counts-only object (`contractual` / `ambiguous` / `incidental`) sourced from the top-level summary. Classification distribution is already carried by the top-level `summary`.
4. **Size budget**: strengthen the budget test to realistic scale (thousands of functions/effects) and assert the compact output stays under 300KB, independent of total function count.

The `--format=json` path (`runJSONPath()`) and the full `json.Marshal` output are unaffected; this is a one-way projection at the AI serialization boundary only.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `payload-compaction`: the compact AI payload now bounds the `crap.scores`, `quality.quality_reports`, and `classify.results` arrays while preserving the bounded worst-offender lists (`worst_crap`, `worst_gaze_crap`, `recommended_actions`, `worst_coverage_tests`). Quality gap/discarded/ambiguous fields carry self-contained side-effect descriptions instead of bare IDs; Classify output is counts-only.

### Removed Capabilities

None.

## Impact

- **`internal/aireport/compact.go`**: the structs `compactCRAPReport`, `compactCRAPSummary`, `compactQualityReport`, `compactContractCoverage`, and `compactPackageSummary` change shape and `compactClassifyResult` is removed; the builder functions `compactCRAPField`, `compactQualityField`, and `compactClassifyField` are updated accordingly.
- **`internal/aireport/compact_test.go`**: several tests assert the old (inverted) semantics and must be updated; the size-budget test is rescaled.
- **`internal/aireport/payload.go`**: unchanged (full-fidelity structs stay intact).
- **`.opencode/agents/gaze-reporter.md`** (and its embedded-asset copies): verified to reference only the preserved bounded fields; no prompt changes required, but the classification distribution source is confirmed to be the top-level summary counts.
- **CI**: the `Gaze quality report` step (Go 1.24) is unblocked once the payload fits under 300KB.

## Constitution Alignment

Assessed against the Gaze Constitution (`.specify/memory/constitution.md`, v1.3.0).

### I. Accuracy

**Assessment**: PASS

The change does not alter analysis behavior or scoring — it only changes which fields are projected into the AI-facing serialization. Full-fidelity JSON (`--format=json`) and threshold evaluation are untouched, so no accuracy regression is introduced. The preserved worst-offender lists remain populated from the same bounded sources (`buildSummary`, `WorstCoverageTests`), preserving truthful top-5 reporting.

### II. Minimal Assumptions

**Assessment**: PASS

The change makes fewer assumptions about payload size: bounding the arrays means compact output scales with a fixed top-N budget rather than total function count, so the reporter works uniformly on small and large codebases without requiring the user to configure anything.

### III. Actionable Output

**Assessment**: PASS

The reporter's actionable guidance — top-5 worst offenders, fix strategies, coverage gaps with effect type and description, and worst tests — is preserved and, for quality gaps, *strengthened* by making gap descriptions self-contained rather than bare IDs. The counts-only Classify output retains the classification distribution the report displays.

### IV. Testability

**Assessment**: PASS

Every changed projection function (`compactCRAPField`, `compactQualityField`, `compactClassifyField`) is already covered by unit tests in `compact_test.go` using synthetic payloads; these tests are updated to assert the new shape. The strengthened size-budget test verifies the invariant (sub-300KB at realistic scale) in isolation, and the existing `TestRunTextPath_CompactPayloadReceived` integration test confirms the text path delivers the compact bytes. Coverage strategy: unit tests in `internal/aireport/` plus the existing pipeline integration test; no new external services required.
