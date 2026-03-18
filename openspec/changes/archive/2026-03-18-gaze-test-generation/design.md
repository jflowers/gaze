# Design: Gaze Test Generation

## Context

Gaze already produces machine-readable data about what tests are missing:
- `crap.Score.FixStrategy` — one of `add_tests`, `add_assertions`, `decompose_and_test`, `decompose`
- `taxonomy.ContractCoverage.Gaps` — list of contractual `SideEffect`s that lack assertions
- `taxonomy.ContractCoverage.GapHints` — Go code snippet per gap (parallel array)
- `taxonomy.ContractCoverage.DiscardedReturns` + `DiscardedReturnHints` — return values explicitly ignored with `_`

The `GapHints` are currently passive — displayed as suggestions in the quality text report. This change makes them active inputs to an AI agent that generates complete test functions.

## Goals / Non-Goals

### Goals
- Generate complete, compilable, runnable test functions from gaze quality data
- Prioritize by fix strategy: `add_tests` (easiest wins) → `add_assertions` → `decompose_and_test` (skeleton)
- Integrate into implementation workflows so tests are generated as code is written
- Work on any Go project gaze can analyze (not just gaze itself)
- Append generated tests to existing `*_test.go` files (not separate files)
- Follow the target project's testing conventions (stdlib `testing`, naming patterns)

### Non-Goals
- Generating tests for `decompose`-strategy functions (refactoring needed first, not tests)
- Replacing hand-written tests (generated tests complement, not replace)
- Adding new Go code to gaze's analysis engine (this is purely prompt engineering)
- Supporting non-Go projects

## Decisions

### D1: Agent-based test generation (not CLI subcommand)

Test generation is an AI agent prompt (`.opencode/agents/gaze-test-generator.md`) invoked via OpenCode, not a `gaze test-gen` CLI subcommand. This is because:
- Test generation requires reading source code, understanding context, and making judgment calls — LLM territory
- The GapHints provide the assertion skeleton but the agent needs to construct the test setup (imports, struct initialization, mock data)
- A CLI subcommand would need to embed an LLM or produce low-quality template tests

The agent receives gaze JSON data + function source code and produces complete test functions.

### D2: Three-tier fix strategy handling

| Strategy | Agent behavior |
|----------|----------------|
| `add_tests` | Generate complete test function: setup → call → assertions from GapHints. Target has 0% line coverage so the test must exercise the function end-to-end. |
| `add_assertions` | Read existing test functions for the target, identify where assertions are missing (using Gaps data), add specific assertions from GapHints into the existing test. Also restructure existing assertions that use helper wrappers (identified by `UnmappedReason: helper_param` or `inline_call`) so the assertion mapper can trace them to the target function's side effects. The function already has line coverage; it needs contract-level verification and mapper visibility. |
| `add_docs` | Triggered when `ContractCoverageReason` is `all_effects_ambiguous` and `EffectConfidenceRange` shows confidence just below 70. Add or improve GoDoc comments on the function that explicitly describe the observable side effects (return values, mutations, I/O). This pushes the classifier confidence above 70, flipping effects from `ambiguous` to `contractual` and improving quadrant classification without any test code changes. |
| `decompose_and_test` | Generate a test skeleton with TODOs for each gap. The function needs refactoring before proper tests are possible, but the skeleton documents what tests should exist after refactoring. |
| `decompose` | Skip — no test can fix a pure complexity problem. The agent notes this in its output. |

**Note on `add_docs`**: This is not a fix strategy assigned by `crap.Analyze` — it's detected by the agent when it sees Q3 functions whose `ContractCoverageReason` is `all_effects_ambiguous`. The agent checks: if all effects are ambiguous AND confidence is 58-69 (close to the 70 threshold), adding GoDoc is more effective than adding tests because the issue is classification, not coverage.

### D3: `/gaze fix` command workflow

```
/gaze fix [package-pattern]
/gaze fix --strategy=add_tests   # filter to specific strategy
/gaze fix --top=5                # limit to top N by CRAP score
/gaze fix --dry-run              # show what would be generated, don't write
```

Steps:
1. Run `gaze crap --format=json [pattern]` → get scores + fix strategies
2. Run `gaze quality --format=json [pattern]` → get gaps + hints per function
3. Filter to functions with actionable fix strategies (exclude `decompose`)
4. Sort by priority: `add_tests` first, then `add_assertions`, then `decompose_and_test`
5. For each target function:
   a. Read the function's source code via file path + line number
   b. Read the existing `*_test.go` if present
   c. Construct a prompt with: source, gaps, hints, existing tests, project conventions
   d. Generate test code
   e. Append to the appropriate `*_test.go` file
6. Run `go build [pattern]` to verify compilation
7. Run `go test -race -count=1 -run [generated-test-names] [pattern]` to verify tests pass
8. Report: N tests generated, M compile, K pass, coverage delta

### D4: Implementation workflow integration

Add a hook in both `/opsx-apply` (SKILL.md step 6) and `/speckit.implement` (step 8) that runs after each implementation task:

```
After code changes, before marking [x]:
1. Identify changed .go files (git diff --name-only)
2. For each non-test .go file with changes:
   a. Run gaze quality --format=json on its package
   b. Check for ContractCoverage.Gaps on new/modified functions
   c. If gaps exist, invoke gaze-test-generator agent
   d. Verify generated tests compile and pass
3. If mandatory mode (default): block task completion until tests pass
   If advisory mode: show results, allow skipping
4. Mark task [x]
```

The mode is controlled by checking for a `.gaze.yaml` config key:
```yaml
test_generation:
  mode: mandatory  # or "advisory"
```

When `.gaze.yaml` doesn't exist or doesn't have this key, default to `mandatory`.

### D5: Test file placement

Generated tests append to the existing `*_test.go` file for the target function's source file. If no `*_test.go` exists, create one with the standard package declaration and imports.

The agent uses the same package declaration as the existing test file (either `package foo` or `package foo_test`). If creating a new file, use `package foo_test` (external test package) by default for exported functions, `package foo` for unexported functions.

### D6: Convention detection

The agent reads the target project's existing test files to detect conventions:
- Package declaration style (`package foo` vs `package foo_test`)
- Import patterns (which testing helpers are used)
- Naming patterns (is `TestXxx_Description` or `TestXxxDescription` used?)
- Table-driven test style (is `tt` or `tc` used for the loop variable?)
- Error assertion style (`if err != nil { t.Fatal(err) }` vs `t.Fatalf("unexpected error: %v", err)`)

If no existing tests exist, fall back to gaze's own conventions from AGENTS.md.

## Risks / Trade-offs

- **Generated test quality**: LLM-generated tests may have subtle issues (wrong expected values, missing edge cases). Mitigated by: compilation check, test execution, and the reviewer-testing agent as a quality gate.
- **Test maintenance burden**: Generated tests that are too tightly coupled to implementation create maintenance debt. Mitigated by: the agent focuses on contractual side effects (observable behavior), not implementation details.
- **Mandatory mode friction**: Blocking task completion on test generation could slow down implementation. Mitigated by: configurable to advisory mode, and the agent only generates tests for functions with gaps (not every function).
- **GapHints accuracy**: The hints are mechanical templates, not always directly pasteable. The agent needs to adapt them to the actual function signature and test context. This is an LLM strength.
