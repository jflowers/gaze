## 1. Add Scoring Consistency Rules to agent prompt

- [x] 1.1 Add a "Scoring Consistency Rules" section to `.opencode/agents/gaze-reporter.md` after the output structure instructions. The section MUST contain explicit rules for: CRAPload threshold (read from `summary.crap_threshold`), contract coverage (use module-wide average), GazeCRAPload (read from `summary.gaze_crapload`), worst offenders (render verbatim from JSON). Include negative constraints: "do NOT hardcode threshold values", "do NOT compute subset averages".
- [x] 1.2 Copy the updated agent prompt to `internal/scaffold/assets/agents/gaze-reporter.md` (the embedded scaffold copy).

## 2. Align quadrant descriptions

- [x] 2.1 Update the Quick Reference Example in `.opencode/agents/gaze-reporter.md` to use "contract coverage" terminology: Q1 "Low complexity, high contract coverage", Q2 "High complexity, contracts verified", Q4 "Complex AND contracts not adequately verified". Q3 "Simple but underspecified" stays unchanged.
- [x] 2.2 Update `.opencode/references/example-report.md` quadrant descriptions to match the prompt's Quick Reference (ensure exact same wording).
- [x] 2.3 Copy the updated example-report to `internal/scaffold/assets/references/example-report.md`.

## 3. Parameterize example CRAPload threshold

- [x] 3.1 In `.opencode/references/example-report.md`, change the CRAPload row from `"24 (functions >= threshold 15)"` to a format that signals the threshold is data-driven (e.g., annotate that T comes from `summary.crap_threshold`).
- [x] 3.2 Copy the updated example-report to `internal/scaffold/assets/references/example-report.md`.

## 4. Verification

- [x] 4.1 Run `go build ./cmd/gaze` to verify the embedded assets compile.
- [x] 4.2 Verify constitution alignment: Observable Quality is satisfied — the agent prompt now contains explicit rules ensuring its output scores match the CLI's deterministic output.
- [x] 4.3 Diff `.opencode/agents/gaze-reporter.md` against `internal/scaffold/assets/agents/gaze-reporter.md` to confirm they are identical.
- [x] 4.4 Diff `.opencode/references/example-report.md` against `internal/scaffold/assets/references/example-report.md` to confirm they are identical.
