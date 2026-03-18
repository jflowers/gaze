# Gaze Test Generation

## Why

Gaze identifies quality problems — CRAPload violations, contract coverage gaps, untested functions — but provides no mechanism to fix them. The report's "Top 5 Prioritized Recommendations" are prose suggestions that a developer must manually translate into test code. This creates a gap: gaze can diagnose, but it takes human effort to remediate.

The data to close this gap already exists. Every function in a gaze report carries:
- `FixStrategy` — whether it needs `add_tests`, `add_assertions`, `decompose_and_test`, or `decompose`
- `ContractCoverage.Gaps` — the exact contractual side effects that lack assertions
- `ContractCoverage.GapHints` — Go code snippets for each missing assertion
- Function source code — readable via the file path and line number in the CRAP score

But test generation alone doesn't address all quality gaps. Two other levers exist:

- **Documentation**: Functions whose side effects are classified `ambiguous` (confidence 63-69, just below the 70 contractual threshold) can be pushed to `contractual` by adding GoDoc comments that explicitly describe their observable behavior. This changes quadrant assignments and GazeCRAPload without writing any test code.
- **Assertion mapper visibility**: Existing tests that use helper wrappers (e.g., `result := analyzeFunc(t, ...)`) create depth-2 tracing gaps the mechanical assertion mapper can't resolve. Restructuring assertions to be directly on the target function's return value improves mapping accuracy and contract coverage without changing what's actually tested.

A comprehensive remediation system combines all three: generating new tests, adding documentation, and restructuring existing assertions for visibility — closing the loop from diagnosis to remediation.

There are two modes:

1. **Remediation** (`/gaze fix`) — after a report reveals CRAPload/GazeCRAPload violations, batch-generate tests for the worst offenders to bring the codebase below ratchet thresholds.
2. **Development** (integrated into `/opsx-apply` and `/speckit.implement`) — as new code is written, automatically generate contract-level tests so quality debt never accumulates.

## What Changes

Add a test generation agent, a remediation command, and integration hooks into the implementation workflows.

## Capabilities

### New Capabilities
- `gaze-test-generator` agent: Reads function source + gaze quality data (GapHints, Gaps, FixStrategy, AmbiguousEffects, UnmappedAssertions), performs four actions: generates new test functions (`add_tests`), strengthens existing tests and restructures assertions for mapper visibility (`add_assertions`), adds GoDoc comments to push ambiguous effects above the contractual threshold (`add_docs` — triggered when `ContractCoverageReason` is `all_effects_ambiguous`), and generates test skeletons for functions needing decomposition (`decompose_and_test`)
- `/gaze fix` command: Batch remediation — runs gaze analysis, prioritizes functions by fix strategy, delegates to the test generator agent for each target, verifies generated tests compile and pass. Includes `--strategy` filter and `--top=N` limit.
- Per-task test generation in `/opsx-apply` and `/speckit.implement`: After each implementation task, runs gaze analysis on changed files and generates contract tests for new/modified functions. Mandatory by default, configurable to advisory.

### Modified Capabilities
- `.opencode/skills/openspec-apply-change/SKILL.md`: Gains a test generation step in the per-task implementation loop
- `.opencode/command/speckit.implement.md`: Same per-task hook
- `internal/scaffold/`: Scaffold gains the new agent and command files for `gaze init`

### Removed Capabilities
- None

## Impact

- `.opencode/agents/gaze-test-generator.md` — new agent prompt
- `.opencode/command/gaze-fix.md` — new command
- `.opencode/skills/openspec-apply-change/SKILL.md` — modified (test generation hook)
- `.opencode/command/speckit.implement.md` — modified (test generation hook)
- `internal/scaffold/assets/agents/gaze-test-generator.md` — new scaffold file
- `internal/scaffold/assets/command/gaze-fix.md` — new scaffold file
- `internal/scaffold/scaffold.go` — modified (register new files)

No changes to gaze's Go analysis engine. The test generation is entirely prompt/skill/command engineering using existing `gaze quality --format=json` output.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The test generator agent is a self-describing artifact (markdown prompt) that communicates through well-defined data (gaze JSON output). It produces test files — standard Go artifacts that any developer or CI system can consume independently.

### II. Composability First

**Assessment**: PASS

The `/gaze fix` command is independently usable without the implementation workflow integration. The implementation hooks are additive — they don't break existing `/opsx-apply` or `/speckit.implement` behavior when gaze isn't installed.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observable quality by automating the creation of contract-level tests. Generated tests are verified to compile and pass before being committed. The gaze report itself serves as the provenance record for why each test was generated.

### IV. Testability

**Assessment**: PASS

Generated tests follow the stdlib `testing` convention and are fully runnable with `go test -race -count=1`. The test generation agent's quality criteria are derived from the reviewer-testing agent's rubric, ensuring generated tests meet the same standards as hand-written tests.
