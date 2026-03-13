## Why

`quality.Assess` treats `BuildTestSSA` failures as hard errors, returning `(nil, nil, error)` with zero results. This causes three user-visible problems:

- **`gaze quality <pkg>`** exits non-zero with an error message — no report produced at all.
- **`gaze report`** silently drops the package from the quality section — no warning emitted.
- **`gaze crap`** silently loses contract coverage data for the package.

The analysis pipeline (`analysis.Analyze`) already handles this gracefully — when `BuildSSA` returns nil, mutation analysis is skipped but AST-based results are still returned. The quality pipeline should follow the same pattern.

This was identified during spec 021 (SSA panic recovery) and explicitly deferred as out-of-scope (FR-006 prohibited caller changes). Tracked as GitHub issue #30.

## What Changes

Refactor `quality.Assess` to continue past `BuildTestSSA` failures instead of returning an error:

1. When SSA fails, skip SSA-dependent analysis (target inference via call graph, assertion-to-effect mapping).
2. Still return results computable without SSA (test function enumeration, assertion detection, detection confidence).
3. Signal degradation via a new `SSADegraded bool` field on `PackageSummary` so callers and downstream consumers know results are partial.
4. Emit a warning to `opts.Stderr` so users see the degradation in terminal output.
5. Update all three callers to handle degraded results instead of nil/error.

## Capabilities

### New Capabilities
- `SSADegraded` field on `PackageSummary`: Machine-readable boolean indicating quality results are partial because SSA construction failed. Defaults to `false` — backward compatible.

### Modified Capabilities
- `quality.Assess`: Returns partial results with zero-valued contract coverage and over-specification instead of a hard error when `BuildTestSSA` fails. Test function enumeration and assertion detection confidence are still populated. Returns `nil` error on SSA failure (degradation is not an error).
- `gaze quality` CLI: Exits 0 with a warning when SSA fails, printing available partial results instead of an error message.
- `gaze report` quality step: Includes degraded package results in the quality section instead of silently dropping them.
- `gaze crap` contract coverage: Logs a warning when SSA fails for a package instead of silently producing nil.

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/quality/quality.go` | Refactor `Assess` to handle `BuildTestSSA` error as degradation |
| `internal/taxonomy/types.go` | Add `SSADegraded bool` to `PackageSummary` |
| `internal/report/schema.go` | Add `ssa_degraded` to `QualitySchema` `PackageSummary` definition |
| `cmd/gaze/main.go` (~877) | `runQuality`: handle partial results, exit 0 with warning |
| `cmd/gaze/main.go` (~593) | `analyzePackageCoverage`: surface warning for degraded results |
| `internal/aireport/runner_steps.go` (~139) | `runQualityForPackage`: return degraded results instead of nil |
| `internal/quality/quality_test.go` | New tests for degraded path |
| `internal/report/report_test.go` | Update schema validation tests |

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change modifies `quality.Assess`'s error contract and adds a machine-readable `SSADegraded` field to `PackageSummary`. All outputs remain self-describing JSON with provenance metadata. No cross-hero coupling is introduced — `quality.Assess` remains a standalone function with the same signature.

### II. Composability First

**Assessment**: PASS

`quality.Assess` remains independently callable with the same function signature. The change is backward-compatible: callers that already handle nil/error gracefully will now receive non-nil partial results, which is strictly better. The `SSADegraded` field is additive and defaults to `false`.

### III. Observable Quality

**Assessment**: PASS

The `SSADegraded` field provides explicit machine-parseable indication of partial results in JSON output. The JSON Schema is updated to document the new field. Consumers can programmatically distinguish full-fidelity results from degraded ones rather than having to infer partiality from missing packages.

### IV. Testability

**Assessment**: PASS

The degraded path is testable in isolation — construct a scenario where `BuildTestSSA` fails and verify `Assess` returns non-nil reports with `SSADegraded: true` on the summary. No external services required. The `safeSSABuild` helper from spec 021 already provides the mechanism to trigger SSA failures in a controlled way.
