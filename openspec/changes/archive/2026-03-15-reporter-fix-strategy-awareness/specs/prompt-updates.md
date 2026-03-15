## ADDED Requirements

### Requirement: Remediation Breakdown in CRAP mode

When `fix_strategy_counts` is present in the CRAP JSON summary, the agent MUST display a remediation breakdown showing counts per strategy.

### Requirement: Fix Strategy Awareness in recommendations

The agent MUST map each `fix_strategy` value to specific recommendation language. For `add_tests` functions, the agent SHOULD examine the source file signature and recommend interface extraction when concrete external types are detected.

## MODIFIED Requirements

### Requirement: Scaffold drift parity

Both `internal/scaffold/assets/agents/gaze-reporter.md` and `.opencode/agents/gaze-reporter.md` MUST have identical content.

## REMOVED Requirements

None.
