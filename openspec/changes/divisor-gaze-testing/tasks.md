<!--
  Phases mirror issue #303's locked implementation plan (Phases 1-5 with
  acceptance criteria). Design decisions D1-D8 in design.md govern the
  details; the issue is the locked statement of intent.
-->

## 1. Agent Rename + Rewrite (Phase 1)

- [x] 1.1 `git mv .opencode/agents/reviewer-testing.md .opencode/agents/divisor-gaze-testing.md` (preserve file history).
- [x] 1.2 Convert the frontmatter of `.opencode/agents/divisor-gaze-testing.md`: drop the `model:` pin (review-matrix tiers resolve models), keep `temperature: 0.1`, replace the `tools:` block with a divisor-style `permission:` deny block (`edit: deny`, `bash: deny`, `webfetch: deny`).
- [x] 1.3 Rewrite the body following the divisor persona structure (Role; Source Documents including `[PACK]` convention packs and the optional Dewey knowledge-graph step; dual-mode architecture with Code Review Mode default and Spec Review Mode; Out-of-Scope section delegating other dimensions to their owning personas; Output Format; Decision Criteria) with a **universal core rubric**: CRAP > 30 prioritization, contract coverage, assertion mappings, quadrants, fix strategies, P0-P4 effect tiers — driven by gaze report data (identical for goprovider and external-analyzer sources).
- [x] 1.4 Add the **conditional Go-native section**: stdlib `testing` only (no testify/gomega), `t.Errorf`/`t.Fatalf` direct assertions, `TestXxx_Description` naming, `testdata/src` fixtures, `-race -count=1` compatibility, `testing.Short()` guards — applied only when the project under review is Go (go.mod present or AGENTS.md declares Go).
- [x] 1.5 Add the **conditional external-analyzer section**: non-Go projects defer framework specifics to the repo's `AGENTS.md` / `[PACK]` convention packs; universal taxonomy types (`MonkeyPatch`, `GeneratorYield`, `StreamOutput`, ...) map to their language constructs.
- [x] 1.6 Mirror to `internal/scaffold/assets/agents/divisor-gaze-testing.md` byte-identical (File Sync Invariant) and delete the legacy `internal/scaffold/assets/agents/reviewer-testing.md`; asset manifest stays at 9 files.
- [x] 1.7 [P] Add `agents/divisor-gaze-testing.md` to `isToolOwned` in `internal/scaffold/scaffold.go` (tool-owned, overwrite-on-diff — deliberate change from the user-owned `reviewer-testing.md`).

## 2. .uf Registration in the Gaze Repo (Phase 2)

Note: `uf.review-council.md` persona table is intentionally untouched — it
documents only the canonical nine personas and handles unknown personas
generically.

- [x] 2.1 [P] Append the reviewer entry `{agent: divisor-gaze-testing, capability: review, scopes: [test-quality]}` to `.uf/reviewer-capabilities.yaml`, keeping the generic `divisor-testing` entry and all comments (including the `# scaffolded by uf` marker).
- [x] 2.2 [P] In `.uf/review-matrix.yaml`, remove `divisor-testing` and add `divisor-gaze-testing` in the `code`, `spec`, and `feedback` advisor lists; leave the `triage` list, `always:` overrides, and all other keys untouched.

## 3. gaze init Intelligence (Phase 3)

- [x] 3.1 Extract `isMarkerLine(line string) bool` from `removeExistingMarker` in `internal/scaffold/scaffold.go` and refactor `removeExistingMarker` to use it, so marker detection and removal share one predicate.
- [x] 3.2 Create `internal/scaffold/ufconfig.go` with `syncUFReviewConfig(opts Options)` using `gopkg.in/yaml.v3` `yaml.Node` round-trip: absent `.uf/` directory → print a notice, never create files; `reviewer-capabilities.yaml` → append the `divisor-gaze-testing` entry if no entry with that `agent` exists; `review-matrix.yaml` → remove `divisor-testing` / ensure `divisor-gaze-testing` in `advisor.code`, `advisor.spec`, `advisor.feedback` (triage untouched); re-encode with `SetIndent(2)`; write only when the document changed (idempotent no-op never rewrites); malformed YAML or missing expected keys → warn to stdout and skip that file without failing init.
- [x] 3.3 Add orphan cleanup to `internal/scaffold/ufconfig.go`: remove `.opencode/agents/reviewer-testing.md` only when its content carries a `<!-- scaffolded by gaze ... -->` marker (via `isMarkerLine`); record removals in a new `Result.Removed []string` field and report them in `printSummary`; marker-less copies are left untouched.
- [x] 3.4 Wire `syncUFReviewConfig` and the orphan cleanup into `Run()` (`internal/scaffold/scaffold.go`) after the asset walk, so the thin `runInit` wrapper in `cmd/gaze/main.go` needs no changes.
- [x] 3.5 Create `internal/scaffold/ufconfig_test.go` — table-driven coverage of every branch: absent `.uf/` (notice printed, no files created), present config (merged correctly), no-op (already registered), idempotent second run (byte-identical, no rewrite), malformed YAML (warn + skip + file unchanged on disk + init succeeds), comment preservation (including the `# scaffolded by uf` marker), orphan removal (marker-bearing removed and reported; marker-less untouched; absent is a silent no-op).

## 4. Reference Sweep (Phase 4)

- [x] 4.1 [P] Update `.opencode/commands/speckit.testreview.md` and its byte-identical mirror `internal/scaffold/assets/commands/speckit.testreview.md`: frontmatter description, Step 3 heading, Task-tool delegation target, and checklist reference → `divisor-gaze-testing`.
- [x] 4.2 [P] Update `.opencode/agents/gaze-test-generator.md` and its mirror `internal/scaffold/assets/agents/gaze-test-generator.md`: quality-criteria rubric reference → the `divisor-gaze-testing` agent rubric.
- [x] 4.3 [P] Update `docs/guides/opencode-integration.md`: file-table row for `agents/divisor-gaze-testing.md` with tool-owned ownership (fixing the stale "Tool-owned (via `gaze-fix.md` command)" wording on the old row), the tool-owned file list, and a short section documenting the new `.uf` registration and orphan-cleanup behavior of `gaze init`.
- [x] 4.4 [P] Update `AGENTS.md` File Sync Invariants embedded-file list: `agents/reviewer-testing.md` → `agents/divisor-gaze-testing.md`.
- [x] 4.5 [P] Update `internal/scaffold/scaffold_test.go`: `TestRun_CreatesFiles` expected path, `TestAssetPaths_Returns9Files` manifest map key, ownership counts/comments in `TestRun_SkipsExisting` and `TestRun_OverwriteOnDiff_ToolOwned` (now 2 user-owned + 7 tool-owned), and `TestIsToolOwned` cases (`agents/divisor-gaze-testing.md` → true; retire the `agents/reviewer-testing.md` false case or keep it as a negative regression guard).
- [x] 4.6 [P] Update `cmd/gaze/main_test.go` `runInit` expected-paths list: `.opencode/agents/reviewer-testing.md` → `.opencode/agents/divisor-gaze-testing.md`.

## 5. Verification + Governance (Phase 5)

- [x] 5.1 CI parity — derive exact commands from `.github/workflows/test.yml` and MegaLinter (per the AGENTS.md CI Parity Gate) and run them locally: `go build ./...`; `go test -race -count=1 -short -timeout 15m ./...`; `go test -race -count=1 -run 'TestRunSelfCheck' -timeout 30m ./cmd/gaze/...`; `golangci-lint run`. All green, including `TestEmbeddedAssetsMatchSource`.
- [x] 5.2 Constitution alignment verification (per the proposal's assessments): Minimal Assumptions — never-create `.uf` guard and marker-less-copy preservation confirmed by tests; Actionable Output — init summary reports registration and removal actions; Testability — `ufconfig_test.go` branch coverage complete; Accuracy — N/A (no analysis-engine changes).
- [x] 5.3 Acceptance verification against issue #303: `divisor-gaze-testing` discovered by uf divisor globs and included in gaze review councils; registration idempotent with comments and the `# scaffolded by uf` marker preserved; no-`.uf` project scaffolds the agent and prints a notice without creating `.uf` files; gaze-marked `reviewer-testing.md` orphans removed on re-init with customized (marker-less) copies untouched; all byte-identity invariants and CI checks pass.
- [x] 5.4 Run `/uf.review-council` and resolve all REQUEST CHANGES findings until all reviewers return APPROVE.
- [x] 5.5 [P] Create the website documentation issue in `unbound-force/website` (user-facing `gaze init` behavior change: persona registration in existing `.uf` config, orphan cleanup) — required before the PR merges.
- [x] 5.6 Update `AGENTS.md` Recent Changes with a `divisor-gaze-testing` entry documenting the rename, registration, and init intelligence. PR submission follows `/uf.finale` after council APPROVE (`Closes #303` via `originating_issue: 303` in `.openspec.yaml`); minimal to no code changes between APPROVE and PR.
