# Tasks: Report Actionability

## 1. Add metric definitions to agent prompt

- [x] 1.1 In `.opencode/agents/gaze-reporter.md`, add a "Metric Definitions" subsection after the existing 5 Scoring Consistency Rules (after rule 5 about quadrant counts). The subsection should define CRAPload (line coverage, `summary.crapload`), GazeCRAPload (contract coverage, `summary.gaze_crapload`, NOT the Q4 count), and quadrant counts (`summary.quadrant_counts`, Q4 is one component not a synonym for GazeCRAPload).

## 2. Sync prompt copies

- [x] 2.1 Copy `.opencode/agents/gaze-reporter.md` to `internal/scaffold/assets/agents/gaze-reporter.md`.
- [x] 2.2 Copy `.opencode/agents/gaze-reporter.md` to `internal/aireport/assets/agents/gaze-reporter.md`.

## 3. Verification

- [x] 3.1 Run `go build ./cmd/gaze` to verify embedded assets compile.
- [x] 3.2 Diff all three copies to confirm they are identical.
