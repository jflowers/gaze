# Divisor Gaze Testing Persona

## Why

The gaze-specific test-quality reviewer (`.opencode/agents/reviewer-testing.md`)
is invisible to unbound-force's Divisor discovery: uf tooling globs
`.opencode/agents/divisor-*.md` and routes reviewers through
`.uf/review-matrix.yaml` / `.uf/reviewer-capabilities.yaml`, so the gaze
testing persona never appears in review councils or dispatch plans. Meanwhile
the generic `divisor-testing` persona ships language-agnostic convention-pack
prompts with no gaze awareness — it cannot interpret CRAP scores, contract
coverage, assertion mappings, quadrants, fix strategies, or the P0-P4 effect
tiers that gaze's report data model provides.

vibe-check solved this exact problem with `divisor-entropy` (commit `3d50414`):
a tool-owned agent in a **new unique divisor slot**, plus entries appended to
both `.uf` config files. This change replicates that precedent for the gaze
testing persona (issue #303).

## What Changes

1. **Agent rename + rewrite** — `git mv .opencode/agents/reviewer-testing.md`
   → `.opencode/agents/divisor-gaze-testing.md`, rewritten as:
   - a **universal core rubric** (language-neutral: CRAP > 30, contract
     coverage, assertion mappings, quadrants, fix strategies, P0-P4 effect
     tiers — the same data model regardless of goprovider vs external
     analyzer), plus
   - a **conditional Go-native section** (stdlib `testing` only,
     `t.Errorf`/`t.Fatalf`, `TestXxx_Description` naming, `testdata/src`
     fixtures, `-race -count=1`, `testing.Short()` guards), and
   - a **conditional external-analyzer section** (non-Go projects defer
     framework specifics to the repo's AGENTS.md / `[PACK]` convention packs;
     universal taxonomy types such as `MonkeyPatch`, `GeneratorYield`, and
     `StreamOutput` map to language constructs).
   Frontmatter: drop the `model:` pin (review-matrix tiers resolve models),
   keep `temperature: 0.1`, convert the `tools:` block to a divisor-style
   `permission:` deny block.
2. **Scaffold mirror + ownership** — mirror the new file to
   `internal/scaffold/assets/agents/divisor-gaze-testing.md` byte-identical
   (File Sync Invariant), delete the old embedded
   `agents/reviewer-testing.md` copy, and add `agents/divisor-gaze-testing.md`
   to `isToolOwned` (tool-owned, overwrite-on-diff — a deliberate ownership
   change from the user-owned `reviewer-testing.md`).
3. **.uf registration in the gaze repo** (entropy precedent) — append
   `{agent: divisor-gaze-testing, capability: review, scopes: [test-quality]}`
   to `.uf/reviewer-capabilities.yaml` (schema-valid: extra entries are
   allowed; only the canonical nine are cross-enforced); swap advisor lists
   in `.uf/review-matrix.yaml` (remove `divisor-testing`, add
   `divisor-gaze-testing` in `code`/`spec`/`feedback`; the `triage` list and
   the generic `divisor-testing` capabilities entry stay, preserving
   dispatch-plugin test-review routing and re-scaffold stability). The
   `uf.review-council.md` persona table is intentionally untouched.
4. **`gaze init` intelligence** — new `internal/scaffold/ufconfig.go` using
   the existing `gopkg.in/yaml.v3` dependency (comment-preserving `Node`
   round-trip): merge the reviewer entry and advisor memberships into
   existing `.uf/*.yaml` idempotently (malformed YAML warns and skips); an
   absent `.uf/` directory skips with a notice and is **never created** (a
   minimal capabilities file missing the canonical nine would fail uf's
   `parseReviewerManifest` KNOWN_REVIEWERS enforcement); remove
   `.opencode/agents/reviewer-testing.md` orphans only when they carry the
   `<!-- scaffolded by gaze ... -->` marker, with a summary notice
   (marker-less, user-authored copies are left alone).
5. **Reference sweep** — update `speckit.testreview.md` (both copies:
   description, Step 3 heading, delegation target, checklist reference),
   `gaze-test-generator.md` (both copies: rubric reference),
   `docs/guides/opencode-integration.md` (file table + tool-owned list,
   fixing the stale ownership wording), `AGENTS.md` (embedded-file list), and
   the scaffold test expectations (`expected` paths, `isToolOwned` cases) in
   `internal/scaffold/scaffold_test.go` and `cmd/gaze/main_test.go`.
6. **Verification + governance** — CI parity from
   `.github/workflows/test.yml`, `TestEmbeddedAssetsMatchSource` green,
   review council APPROVE, website issue in `unbound-force/website` (the
   `gaze init` behavior change is user-facing), PR.

## Capabilities

### New Capabilities

- `testing-persona-agent`: the `divisor-gaze-testing` reviewer persona — a
  universal gaze test-quality rubric with conditional Go-native and
  external-analyzer sections, occupying a new unique divisor slot and
  deployed tool-owned via `gaze init`
- `uf-review-registration`: `gaze init` .uf review-config registration —
  idempotent, comment-preserving merge into existing `.uf/*.yaml`, never
  creating `.uf` files, plus gaze-marked `reviewer-testing.md` orphan cleanup

### Modified Capabilities

- None. (The scaffolded file set changes — `reviewer-testing.md` retired,
  `divisor-gaze-testing.md` added — but the scaffold's testing-persona
  behavior was never specced in OpenSpec; it is captured by the new
  capabilities above rather than a modification.)

### Removed Capabilities

- None

## Impact

**Files added**

- `.opencode/agents/divisor-gaze-testing.md` (renamed from
  `reviewer-testing.md`, rewritten)
- `internal/scaffold/assets/agents/divisor-gaze-testing.md` (byte-identical
  mirror)
- `internal/scaffold/ufconfig.go` (+ `ufconfig_test.go`)

**Files modified**

- `internal/scaffold/scaffold.go` — `isToolOwned` gains
  `agents/divisor-gaze-testing.md`; `Run` invokes the .uf registration and
  orphan cleanup
- `.uf/reviewer-capabilities.yaml`, `.uf/review-matrix.yaml` — persona
  registration and advisor swap (gaze repo only)
- `.opencode/commands/speckit.testreview.md` +
  `internal/scaffold/assets/commands/speckit.testreview.md` — delegation
  target and prose references
- `.opencode/agents/gaze-test-generator.md` +
  `internal/scaffold/assets/agents/gaze-test-generator.md` — rubric reference
- `docs/guides/opencode-integration.md` — file table, tool-owned list, a
  note documenting the new `.uf` registration and orphan-cleanup behavior,
  and the `gaze init` sample-output transcript (renamed `created:` entry;
  8→9 file-count reconciliation with the asset manifest, including
  `dcp.jsonc`)
- `docs/reference/cli/init.md` — canonical `gaze init` CLI reference:
  Scaffolded Files, File Ownership Model, and Examples sections updated
  for the rename and the new `.uf` registration / orphan-cleanup behavior
- `AGENTS.md` — embedded-file list (File Sync Invariants) and Recent Changes
- `internal/scaffold/scaffold_test.go`, `cmd/gaze/main_test.go` — expected
  paths and `isToolOwned` cases

**Files deleted**

- `internal/scaffold/assets/agents/reviewer-testing.md` (superseded mirror)

**Behavioral impact**

- `gaze init` on a project with existing `.uf/` config registers the persona
  idempotently; on a project without `.uf/` it scaffolds the agent file,
  prints a notice, and creates no `.uf` files
- Existing target projects: gaze-marked `reviewer-testing.md` orphans are
  removed on re-init; marker-less copies are untouched
- `/speckit.testreview` delegates to `divisor-gaze-testing` (spec review mode)

**Known limitation (documented)**: the review-dispatch plugin's test-review
mode special-cases the generic `divisor-testing`. `/speckit.testreview`
direct Task-tool delegation moves to the new agent, but dispatch-mode test
review stays on the generic tester.

**Website Documentation Gate**: applies — `gaze init` behavior changes
user-visibly (persona registration in `.uf` config, orphan removal). A
`unbound-force/website` issue must be created before the implementing PR
merges.

**Blog coverage**: noted, not skipped — this change meets the significance
threshold for the org website's gaze blog channel (a new divisor persona
slot joining review councils, a universal test-quality rubric extending
gaze review to external-analyzer projects, and new `gaze init`
registration intelligence). The website issue filed for the Documentation
Gate flags the blog opportunity for the content team; authoring the post
itself stays in the content pipeline, out of this change's scope.

## Constitution Alignment

Assessed against the Gaze project constitution v1.3.0
(`.specify/memory/constitution.md`).

### I. Accuracy

**Assessment**: N/A

Agent prompt content and scaffold tooling only. No change to gaze's
side-effect detection, scoring, or report engines; the rubric consumes the
existing (already regression-tested) report data model verbatim.

### II. Minimal Assumptions

**Assessment**: PASS

The universal core rubric keeps all Go specifics inside a conditional
section, so non-Go consumers of the scaffolded agent are not forced into Go
conventions — external-analyzer projects defer framework specifics to their
own AGENTS.md / convention packs. `gaze init` never creates `.uf` files and
never touches marker-less (user-authored or heavily customized)
`reviewer-testing.md` copies; its two contractual assumptions — the
`.uf/*.yaml` schema and the gaze version marker — are explicit and enforced
by tests.

### III. Actionable Output

**Assessment**: PASS

The rubric's findings cite specific gaze metrics (CRAP score vs the > 30
threshold, contract coverage percentages, quadrant placement, fix-strategy
labels, unasserted effect tiers) rather than aggregate scores, and the init
summary reports each `.uf` registration and orphan-removal action taken,
keeping output comparable across runs.

### IV. Testability

**Assessment**: PASS

`ufconfig.go` is a pure file-in/file-out collaborator invoked with injected
`Options` (TargetDir, Stdout) — table-driven tests cover every branch:
absent `.uf/`, present, no-op, malformed YAML, comment preservation,
idempotent second run, and orphan removal (marker-bearing + marker-less).
The byte-identity mirror invariant (`TestEmbeddedAssetsMatchSource`) and CI
parity (`go build ./...`, `go test -race -count=1 -short ...`,
`golangci-lint run`, per `.github/workflows/test.yml` and MegaLinter) gate
the change end to end.
