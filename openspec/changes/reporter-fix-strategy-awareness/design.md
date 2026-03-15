## Context

The gaze-reporter agent prompt is an embedded asset at `internal/scaffold/assets/agents/gaze-reporter.md` that gets deployed to `.opencode/agents/gaze-reporter.md` via `gaze init`. The scaffold drift detection test (`TestEmbeddedAssetsMatchSource`) requires both copies to be identical.

PR #49 added `fix_strategy` to the CRAP JSON output but the agent prompt doesn't reference it. The agent needs two additions: awareness of the new JSON fields, and testability assessment guidance for `add_tests` functions.

## Goals / Non-Goals

### Goals
- Agent consumes `fix_strategy_counts` and displays remediation breakdown
- Agent maps each fix strategy to specific recommendation language
- Agent examines source files for `add_tests` functions to detect coupling barriers
- Both embedded and `.opencode/` copies stay in sync

### Non-Goals
- Modifying any Go production code
- Adding static analysis features to the CRAP engine
- Changing the fix_strategy computation logic

## Decisions

### D1: Add Remediation Breakdown as item 8 in CRAP mode

Insert after the GazeCRAPload summary line (item 7). When `fix_strategy_counts` is present in the JSON, show a table or summary with counts per strategy.

### D2: Add Fix Strategy Awareness block in recommendations

After the existing recommendation formatting rules, add a block that maps each strategy to recommendation language. The `add_tests` strategy includes an instruction to use the Read tool for testability assessment.

### D3: Sync .opencode copy

Copy the modified embedded asset to `.opencode/agents/gaze-reporter.md` to satisfy the scaffold drift test.

## Risks / Trade-offs

### R1: Agent may not always read source files successfully
The Read tool instruction for testability assessment is best-effort. If the file can't be read, the agent falls back to generic `add_tests` guidance. No failure mode.
