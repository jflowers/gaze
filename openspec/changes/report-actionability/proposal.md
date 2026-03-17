# Report Actionability

## Why

Side-by-side testing of `gaze report --ai=opencode` (CI) vs the `/gaze` agent command (OpenCode TUI) revealed the agent misreports GazeCRAPload as 0 when the actual value is 24. The agent confused GazeCRAPload (functions with GazeCRAP >= threshold) with Q4 count (functions in the Dangerous quadrant). These are different metrics: a simple function with 0% contract coverage is Q3 (Needs Tests) but still has GazeCRAP >= 15, contributing to GazeCRAPload.

The Scoring Consistency Rules added in `reporter-fix-strategy-awareness` tell the agent to read `gaze_crapload` from the JSON, but don't explain what the metric means. Without understanding the definition, the agent substitutes its own interpretation.

## What Changes

1. **Agent prompt**: Add a "Metric Definitions" section to the Scoring Consistency Rules that defines GazeCRAPload, CRAPload, and quadrant counts as distinct metrics with different semantics.
2. **Agent prompt**: Add an explicit negative constraint: "GazeCRAPload is NOT the Q4 count."

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `gaze-reporter agent prompt`: Adds metric definitions that distinguish GazeCRAPload from Q4 count, preventing the agent from conflating the two

### Removed Capabilities
- None

## Impact

- `.opencode/agents/gaze-reporter.md` — prompt text changes (metric definitions)
- `internal/scaffold/assets/agents/gaze-reporter.md` — embedded scaffold copy
- `internal/aireport/assets/agents/gaze-reporter.md` — aireport embedded copy

No Go code changes. No CLI behavior changes. The underlying analysis engine and JSON output are correct — only the AI's interpretation instructions need clarification.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

The agent prompt is a self-describing artifact. No runtime coupling is affected.

### II. Composability First

**Assessment**: N/A

No dependencies introduced. The agent prompt remains a standalone markdown file.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observable quality by ensuring the AI-generated report accurately reflects the machine-parseable JSON output. Before this change, the agent could report GazeCRAPload=0 when the JSON contains GazeCRAPload=24 — a provenance violation.

### IV. Testability

**Assessment**: N/A

Markdown prompt changes. The agent's output accuracy is validated by comparing CLI and TUI reports against the same input data.
