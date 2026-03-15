## Why

PR #49 added `fix_strategy` and `fix_strategy_counts` fields to the CRAP JSON output, but the gaze-reporter agent prompt doesn't know about them. The agent produces reports without referencing remediation strategies, missing the opportunity to provide actionable guidance.

Additionally, issue #50 identified that functions needing architecture changes (interface extraction, DI) before testing can't be detected mechanically. Rather than adding complex static analysis to the CRAP engine, the AI agent can assess testability at report time by examining function signatures via the Read tool.

## What Changes

Update the gaze-reporter agent prompt (`internal/scaffold/assets/agents/gaze-reporter.md`) to:

1. Consume `fix_strategy_counts` from the CRAP JSON summary and display a "Remediation Breakdown" in reports.
2. Map each `fix_strategy` value to specific recommendation language in the Top 5 Recommendations section.
3. For `add_tests` functions, instruct the agent to examine the source file signature and recommend interface extraction when concrete external types are detected.

## Capabilities

### New Capabilities
- Remediation Breakdown display in CRAP mode reports when `fix_strategy_counts` is present.
- Fix Strategy Awareness in recommendations: each strategy maps to specific action guidance.
- Testability assessment: agent reads source files for `add_tests` functions and detects coupling to concrete external types.

### Modified Capabilities
- gaze-reporter agent prompt: two new sections added (Remediation Breakdown instruction, Fix Strategy Awareness block).

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/scaffold/assets/agents/gaze-reporter.md` | Add remediation breakdown and fix strategy awareness sections |
| `.opencode/agents/gaze-reporter.md` | Sync with embedded asset (scaffold drift detection requires parity) |

## Constitution Alignment

### I. Accuracy — PASS
No changes to analysis logic. The agent prompt consumes existing accurate data.

### II. Minimal Assumptions — PASS
No new requirements on user codebases. The agent reads source files that are already available.

### III. Actionable Output — PASS
This is the primary motivation. The prompt changes make the agent's recommendations more actionable by mapping fix strategies to specific guidance.

### IV. Testability — PASS
No production code changes. The scaffold drift test ensures embedded and deployed copies stay in sync.
