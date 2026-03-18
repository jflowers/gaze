# Design: Report Actionability

## Context

The gaze-reporter agent prompt has Scoring Consistency Rules that instruct the agent to read metrics from JSON fields. However, the rules only say *where* to read values, not *what they mean*. The agent sees `gaze_crapload=24` and `quadrant_counts.Q4_Dangerous=0` and incorrectly reports "GazeCRAPload is 0" because it assumes GazeCRAPload = Q4 count.

## Goals / Non-Goals

### Goals
- Define GazeCRAPload, CRAPload, and quadrant counts as distinct metrics in the prompt
- Add explicit negative constraint preventing GazeCRAPload/Q4 conflation
- Keep the definitions concise (the prompt is already large)

### Non-Goals
- Changing JSON field names or adding new fields
- Modifying the CLI text report format
- Adding automated tests for agent output accuracy

## Decisions

### D1: Add "Metric Definitions" subsection to Scoring Consistency Rules

Add 3 definitions immediately after the existing 5 rules:

```
### Metric Definitions (read carefully)

- **CRAPload**: Count of functions with CRAP score >= `crap_threshold`.
  CRAP uses line coverage. Read from `summary.crapload`.
- **GazeCRAPload**: Count of functions with GazeCRAP score >=
  `gaze_crap_threshold`. GazeCRAP uses contract coverage (stronger
  signal). Read from `summary.gaze_crapload`. This is NOT the Q4
  count — Q3 functions with low contract coverage also contribute.
- **Quadrant counts**: Distribution of functions across Q1-Q4 based
  on complexity AND contract coverage. Read from
  `summary.quadrant_counts`. Q4 count is one component of the
  quadrant distribution, not a synonym for GazeCRAPload.
```

**Rationale**: The agent needs to understand the semantic difference, not just the field location. The "NOT the Q4 count" negative constraint is the most important line — it directly prevents the observed misinterpretation.

### D2: Keep all three prompt copies in sync

All three copies (`.opencode/agents/`, `internal/scaffold/assets/agents/`, `internal/aireport/assets/agents/`) must be updated identically. The `prompt_test.go` drift test enforces this for `scaffold` vs `aireport`.

## Risks / Trade-offs

- **Prompt length increase**: ~8 lines added. The prompt is already 464 lines; this is <2% increase. Acceptable.
- **Agent may still misinterpret**: Prompt constraints are probabilistic, not deterministic. The negative constraint ("NOT the Q4 count") is the strongest available signal.
