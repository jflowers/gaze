## Context

The gaze-reporter agent prompt (`.opencode/agents/gaze-reporter.md`) instructs an AI model to interpret gaze JSON output and produce a human-readable quality report. The prompt contains a Quick Reference Example with quadrant labels and a reference to `example-report.md` as canonical output format. Side-by-side testing of `gaze report --format=json` (CLI) vs `/gaze` (in-session agent) revealed three scoring divergences caused by ambiguity in the prompt.

Per the proposal's constitution alignment, this change improves Observable Quality by ensuring the agent's rendered scores match the CLI's deterministic output.

## Goals / Non-Goals

### Goals
- Eliminate CRAPload threshold ambiguity by mandating the agent read `crap_threshold` from JSON
- Eliminate contract coverage scope ambiguity by mandating module-wide averages
- Align quadrant descriptions between the prompt's Quick Reference and the example-report reference
- Add an explicit "Scoring Consistency Rules" section to the prompt as a guardrail

### Non-Goals
- Changing the CLI's text report format or thresholds
- Adding new JSON fields to the payload
- Modifying Go code in any package
- Changing how the gaze-reporter agent structures sections or uses emojis

## Decisions

### D1: Add a "Scoring Consistency Rules" section to the agent prompt

Add a new section between the existing output structure instructions and the formatting rules. This section explicitly states:

1. **CRAPload**: Read `crap_threshold` from the CRAP JSON summary. Display as `"N (functions >= threshold T)"` where T comes from the data.
2. **Contract coverage**: Use the module-wide `avg_contract_coverage` from the quality package summary when available. Do not compute a subset average from selected functions.
3. **GazeCRAPload**: Report from `gaze_crapload` in the CRAP JSON summary. When absent, state "N/A" rather than computing a proxy.
4. **Worst offenders**: Use CRAP scores and fix strategies exactly as they appear in the JSON, without re-thresholding.

**Rationale**: The agent is prone to "helpfully" re-interpreting data. Explicit negative constraints ("do not compute a subset average") are more effective than positive instructions alone.

### D2: Align quadrant descriptions in both files

Update the Quick Reference Example in the prompt to use "contract coverage" instead of "coverage" in Q1/Q2/Q4 descriptions. Update `example-report.md` to match.

Final aligned descriptions:
- Q1 — Safe: "Low complexity, high contract coverage"
- Q2 — Complex But Tested: "High complexity, contracts verified"
- Q3 — Needs Tests: "Simple but underspecified"
- Q4 — Dangerous: "Complex AND contracts not adequately verified"

**Rationale**: "Coverage" is ambiguous between line coverage and contract coverage. Gaze's distinguishing feature is contract coverage — the descriptions must reflect that.

### D3: Update example-report.md CRAPload row to show data-driven threshold

Change the example from `"24 (functions >= threshold 15)"` to `"24 (functions >= threshold T)"` with a comment that T comes from `summary.crap_threshold`. This prevents the model from hardcoding 15.

**Rationale**: The example serves as a template. If it shows a literal "15", models will reproduce "15" even when the JSON contains a different threshold.

## Risks / Trade-offs

- **Risk**: The scoring rules section adds ~15 lines to an already large prompt. Accepted because prompt size is less important than output correctness.
- **Risk**: Models may still deviate from explicit instructions. Mitigated by the negative-constraint style ("do NOT compute a subset") and by the example-report reference showing the correct format.
- **Trade-off**: We do not add automated regression testing for agent output consistency. This would require running the agent against fixture data and comparing scores, which is expensive and brittle. The prompt-level guardrail is the pragmatic approach.
