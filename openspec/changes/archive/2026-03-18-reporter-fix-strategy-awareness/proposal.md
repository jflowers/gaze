## Why

The gaze-reporter agent prompt produces scoring that diverges from the CLI's own CRAP/quality output in three measurable ways:

1. **CRAPload threshold mismatch**: The agent example report uses "threshold 15" but does not enforce that the threshold shown comes from the JSON data's `crap_threshold` field. The agent has been observed using different threshold interpretations (15 vs 30), producing CRAPload counts that differ from the CLI (82 vs 31 for the same data).

2. **Contract coverage cherry-picking**: The agent reports "81.9% average among GazeCRAP-analyzed functions" while the CLI's module-wide quality analysis reports 32.1%. The agent selectively reports from a favorable subset rather than the full `avg_contract_coverage` from the quality data.

3. **Quadrant meaning inconsistency**: The Quick Reference Example in the prompt says "high coverage" and "untested" while the example-report.md reference file says "high contract coverage" and "contracts not adequately verified". The distinction between line coverage and contract coverage is a core gaze concept that must be precise.

These inconsistencies mean a CI report and an in-session /gaze report produce different grades and recommendations from identical underlying data.

## What Changes

Update the gaze-reporter agent prompt and example-report reference to:
- Mandate reading `crap_threshold` from JSON data rather than hardcoding a value
- Mandate using full module-wide contract coverage, not a subset
- Align quadrant descriptions between the Quick Reference Example and the reference file
- Add explicit instructions against cherry-picking favorable subsets for summary metrics

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `gaze-reporter agent prompt`: Adds explicit scoring consistency rules that bind the agent to use the same thresholds and scope as the CLI text report
- `example-report.md reference`: Aligns quadrant descriptions with the prompt's Quick Reference Example

### Removed Capabilities
- None

## Impact

- `.opencode/agents/gaze-reporter.md` — prompt text changes (scoring consistency section)
- `.opencode/references/example-report.md` — quadrant description alignment
- `internal/scaffold/assets/agents/gaze-reporter.md` — embedded copy of agent prompt
- `internal/scaffold/assets/references/example-report.md` — embedded copy of reference

No Go code changes. No test changes. No CLI behavior changes. The underlying analysis engine is correct — only the AI's interpretation instructions change.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The agent prompt is a self-describing artifact that communicates scoring rules to the AI model. This change tightens those rules to reduce ambiguity, improving the artifact's quality as a communication medium.

### II. Composability First

**Assessment**: N/A

No dependencies introduced. The agent prompt remains a standalone markdown file consumed by any OpenCode-compatible runtime.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observable quality by ensuring the AI-generated report produces the same numeric scores as the CLI's machine-parseable JSON/text output. Before this change, the agent could produce metrics that contradicted the CLI — undermining provenance and trust.

### IV. Testability

**Assessment**: N/A

The change is to markdown prompt text. The gaze-reporter agent's output is validated by human review, not automated tests. The underlying scoring functions (`WriteText`, `EvaluateThresholds`) are already unit-tested and unaffected.
