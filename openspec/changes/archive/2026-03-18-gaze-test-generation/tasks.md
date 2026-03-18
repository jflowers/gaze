# Tasks: Gaze Test Generation

## Phase 1: Test Generator Agent

- [x] 1.1 Create `.opencode/agents/gaze-test-generator.md` agent prompt. The agent must: accept a function's source code + `ContractCoverage.Gaps` + `GapHints` + `FixStrategy` + `AmbiguousEffects` + `UnmappedAssertions` + `ContractCoverageReason` + `EffectConfidenceRange` as input; generate complete Go test functions using stdlib `testing` only; follow `TestXxx_Description` naming; use `t.Errorf`/`t.Fatalf` for assertions; handle five actions: `add_tests` (full test), `add_assertions` (strengthen existing test + restructure helper-wrapped assertions for mapper visibility where `UnmappedReason` is `helper_param` or `inline_call`), `add_docs` (add/improve GoDoc comments when `ContractCoverageReason` is `all_effects_ambiguous` and confidence is 58-69, to push classifier confidence above 70), `decompose_and_test` (skeleton with TODOs), and `decompose` (skip with explanation). Include convention detection rules: read existing `*_test.go` to match package declaration style, import patterns, naming patterns, table-driven test variable names, and error assertion style. Embed quality criteria from the reviewer-testing agent rubric (assertion depth, test isolation, contract coverage focus).
- [x] 1.2 Test the agent manually: run `gaze quality --format=json ./internal/config` to get gap data, then invoke the agent with a function that has gaps. Verify the generated test compiles and passes.

## Phase 2: /gaze fix Command

- [x] 2.1 Create `.opencode/command/gaze-fix.md` command file. The command must: accept `[package-pattern]` (default `./...`), optional `--strategy=add_tests|add_assertions|decompose_and_test`, optional `--top=N`, optional `--dry-run`; run `gaze crap --format=json` and `gaze quality --format=json` to collect scores and gaps; filter to actionable fix strategies (exclude `decompose`); sort by priority (add_tests first by CRAP desc, then add_assertions, then decompose_and_test); for each target, read the function source + existing test file, delegate to the `gaze-test-generator` agent; after all generation, run `go build` and `go test -race -count=1 -run <generated-test-names>` to verify; report summary (N generated, M compile, K pass).
- [x] 2.2 Test the command manually: run `/gaze fix --top=3 ./internal/config` and verify tests are generated, compile, and pass.

## Phase 3: Implementation Workflow Integration

- [x] 3.1 Modify `.opencode/skills/openspec-apply-change/SKILL.md` to add a test generation step in the per-task loop (step 6). After making code changes and before marking `[x]`: identify changed `.go` files via `git diff --name-only`; for each non-test `.go` file, run `gaze quality --format=json` on its package; if `ContractCoverage.Gaps` exist for new/modified functions, invoke `gaze-test-generator` agent; verify generated tests compile and pass. In mandatory mode (default), block task completion until tests pass. In advisory mode, show results but allow marking `[x]`. Read mode from `.gaze.yaml` `test_generation.mode` key (default: `mandatory`). Add a note that if `gaze` binary is not available, the step is silently skipped (composability — gaze is optional).
- [x] 3.2 Modify `.opencode/command/speckit.implement.md` to add the same test generation step in the per-task execution loop (step 8). Same behavior as the `/opsx-apply` integration: identify changed files, run quality analysis, generate tests, verify, respect mode config.
- [x] 3.3 Test the integration: create a small change via `/opsx-propose`, implement a task that adds a new Go function, and verify the test generation hook fires and produces a test.

## Phase 4: Scaffold Integration

- [x] 4.1 Copy `.opencode/agents/gaze-test-generator.md` to `internal/scaffold/assets/agents/gaze-test-generator.md`.
- [x] 4.2 Copy `.opencode/command/gaze-fix.md` to `internal/scaffold/assets/command/gaze-fix.md`.
- [x] 4.3 Update `internal/scaffold/scaffold.go` to register the two new files in the asset list. The agent file is tool-owned (overwrite on diff). The command file is tool-owned (overwrite on diff).
- [x] 4.4 Run `go build ./cmd/gaze` to verify embedded assets compile.
- [x] 4.5 Run `gaze init --force` in a temp directory and verify both new files are scaffolded.

## Phase 5: Verification

- [x] 5.1 Run `go test -race -count=1 -short ./internal/scaffold/...` to verify scaffold tests pass (including any prompt drift tests).
- [x] 5.2 Run `/gaze fix --top=3 --dry-run ./...` on the gaze codebase to verify the full pipeline works end-to-end without writing files.
- [x] 5.3 Run `/gaze fix --top=1 ./internal/config` to generate a real test, verify it compiles and passes, then revert the generated test (this is a smoke test, not a permanent change).
