## Context

Gaze's test-quality reviewer lives at `.opencode/agents/reviewer-testing.md`
(spec 017): a user-owned, skip-if-present scaffold asset with a `model:` pin
and a legacy `tools:` frontmatter block. Because its filename lacks the
`divisor-` prefix, uf's reviewer discovery (which globs
`.opencode/agents/divisor-*.md` and routes through
`.uf/review-matrix.yaml` / `.uf/reviewer-capabilities.yaml`) cannot see it —
the persona never joins review councils or dispatch plans. The generic
uf-owned `divisor-testing` persona fills the slot but has no gaze awareness.

The repository's own `.uf/` files (git-tracked) already demonstrate the
coexistence pattern: `divisor-entropy` occupies a unique slot with entries in
both config files. Issue #303 locks the design: replicate that precedent for
a new `divisor-gaze-testing` slot, rewrite the persona as a universal rubric
with conditional sections, and teach `gaze init` to register the persona in
existing target projects' `.uf` config.

Constitution alignment (from the proposal): Minimal Assumptions drives the
universal-plus-conditional content split and the never-create `.uf` guard;
Testability drives the table-driven `ufconfig` coverage and the byte-identity
invariants; Actionable Output drives metric-citing findings and init summary
notices. Accuracy is N/A (no analysis-engine changes).

## Goals / Non-Goals

### Goals

- `divisor-gaze-testing` discovered by uf divisor globs and participating in
  gaze review councils (advisor for `code`/`spec`/`feedback`)
- A persona rubric grounded in gaze's report data model that works for both
  Go-native and external-analyzer projects
- `gaze init` registers the persona in existing `.uf/*.yaml` idempotently,
  preserving comments, never creating `.uf` files
- Legacy gaze-marked `reviewer-testing.md` orphans removed on re-init;
  marker-less copies untouched
- All byte-identity invariants and CI checks green

### Non-Goals

- Taking over or modifying the uf-owned generic `divisor-testing` persona
  (its file, its capabilities entry, its `triage` advisor membership, or the
  `uf.review-council.md` persona table)
- Changing the review-dispatch plugin's test-review routing (it special-cases
  generic `divisor-testing` — documented known limitation)
- Changing scaffold behavior for any other asset (marker logic, ownership of
  other files, summary format beyond the new notices)
- Retroactively updating completed spec artifacts under `specs/`
  (Spec Artifact Immutability — including `specs/017-testing-persona/`, which
  references the old name as a point-in-time record)
- Creating `.uf/` files in projects that don't already use uf

## Decisions

### D1: New unique divisor slot, advisor swap for coexistence

The persona moves to `.opencode/agents/divisor-gaze-testing.md` — a new
unique slot — and the gaze repo's `.uf/review-matrix.yaml` advisor lists swap
`divisor-testing` → `divisor-gaze-testing` in `code`, `spec`, and `feedback`
only. The generic `divisor-testing` capabilities entry and `triage` advisor
membership stay.

Rejected alternatives:

- **Take over the `divisor-testing` slot**: the slot is uf-owned; uf
  re-scaffolding would clobber the gaze-aware rewrite, and the dispatch
  plugin's test-review mode special-cases the generic name. A takeover
  couples gaze to uf's deployment cadence (Composability First: the persona
  must remain independently deployable).
- **Two separate personas (Go + universal)**: doubles maintenance and splits
  the single test-quality scope across two reviewers.

Following the `divisor-entropy` precedent (vibe-check commit `3d50414`):
unique slot + entries appended to both `.uf` files. The
`uf.review-council.md` persona table is intentionally untouched — it
documents only the canonical nine personas and handles unknown personas
generically.

### D2: Universal core rubric with conditional sections

The rewritten persona keeps the proven dual-mode architecture (Code Review
Mode default, Spec Review Mode), the finding Output Format, and the Decision
Criteria section, and follows the divisor persona structural pattern (Source
Documents including `[PACK]` convention packs and the optional Dewey
knowledge-graph step, an Out-of-Scope section delegating other dimensions to
their owning personas). The audit checklists reorganize into:

1. **Universal core** — gaze report data model: CRAP > 30 prioritization,
   contract coverage, assertion mappings, quadrants, fix strategies, P0-P4
   effect tiers. Identical for goprovider and external-analyzer data (same
   report schema).
2. **Conditional Go-native section** — stdlib `testing` only,
   `t.Errorf`/`t.Fatalf`, `TestXxx_Description`, `testdata/src` fixtures,
   `-race -count=1`, `testing.Short()` guards. The prompt instructs the
   persona to apply it only when the project under review is Go (go.mod
   present or AGENTS.md declares Go).
3. **Conditional external-analyzer section** — defers framework specifics to
   the repo's `AGENTS.md` / `[PACK]` convention packs; maps universal
   taxonomy types (`MonkeyPatch`, `GeneratorYield`, `StreamOutput`, ...) to
   language constructs.

Rationale: Minimal Assumptions — non-Go consumers must not inherit Go
conventions, and Go consumers must not lose them. The alternative (a fully
generic rubric) was rejected because it would discard the Go conventions that
gaze's own dogfooding depends on.

### D3: Frontmatter conversion

Drop `model:` (review-matrix tiers resolve models at dispatch time; a pin
would fight the matrix), keep `temperature: 0.1`, and replace the legacy
`tools:` block with a divisor-style `permission:` deny block
(`edit: deny`, `bash: deny`, `webfetch: deny` — matching the pattern of
uf's `divisor-testing.md`). The persona remains read-only; only the
frontmatter dialect changes.

### D4: Tool-owned classification (deliberate ownership change)

`agents/divisor-gaze-testing.md` is added to `isToolOwned` —
overwrite-on-diff without `--force` — replacing the user-owned
skip-if-present behavior of `reviewer-testing.md`. The embedded mirror at
`internal/scaffold/assets/agents/divisor-gaze-testing.md` is byte-identical
(File Sync Invariant, `TestEmbeddedAssetsMatchSource`); the legacy embedded
copy is deleted; the asset manifest stays at 9 files.

Rationale: the persona is gaze-managed machinery whose rubric must track
gaze's data model (same reasoning as `gaze-test-generator.md` and the
`divisor-entropy` precedent). Trade-off: users who customized the persona
lose that customization path on re-init — accepted, because a stale rubric
referencing a superseded file/slot has little preservation value.

### D5: `ufconfig.go` — comment-preserving `yaml.Node` merge

A new `internal/scaffold/ufconfig.go` (invoked from `scaffold.Run` after the
asset walk, so `runInit`'s thin wrapper needs no changes) provides:

- **`syncUFReviewConfig(opts Options)`** — registration. If `.uf/` is
  absent: print a notice, create nothing, return (a minimal capabilities
  file missing the canonical nine fails uf's `parseReviewerManifest`
  KNOWN_REVIEWERS enforcement — Minimal Assumptions: never manufacture a
  config the consumer would reject). Otherwise, per file, independently:
  - `reviewer-capabilities.yaml`: decode into `yaml.Node`; locate the
    `reviewers` sequence; append
    `{agent: divisor-gaze-testing, capability: review, scopes: [test-quality]}`
    if no entry with that `agent` exists.
  - `review-matrix.yaml`: locate `advisor.code`/`advisor.spec`/
    `advisor.feedback`; remove `divisor-testing` and ensure
    `divisor-gaze-testing` in each; leave `triage` and every other node
    untouched (unknown fields pass through the `Node` round-trip verbatim).
  - Re-encode with `yaml.Encoder` `SetIndent(2)` to match the existing file
    style; write only when the document actually changed (no-op runs never
    rewrite — byte-stable idempotency).
  - Malformed YAML or a missing expected key: warn to `opts.Stdout` and skip
    that file; `gaze init` still succeeds (graceful degradation, mirroring
    the go.mod warning pattern).
- **Orphan cleanup** — remove `.opencode/agents/reviewer-testing.md` only
  when its content carries a `<!-- scaffolded by gaze ... -->` marker, reusing
  the marker-line predicate already implemented by `removeExistingMarker`
  (extract a shared `isMarkerLine(line string) bool` so detection and
  removal cannot drift). Removed paths are recorded in a new
  `Result.Removed []string` field so the summary and tests can assert on
  them; marker-less copies are left alone.

Rejected alternatives:

- **Regex or string-splice YAML editing**: fragile against formatting
  variance; `yaml.Node` round-trips comments and unknown fields correctly.
- **Struct decode/re-encode** (`map`/`struct`): silently drops comments —
  unacceptable, both `.uf` files carry operational comments and the
  `# scaffolded by uf` marker.
- **Creating `.uf/` when absent**: produces a config uf itself rejects.

`gopkg.in/yaml.v3` is already a module-level dependency (`internal/config`),
so `go.mod` gains no entries — but `internal/scaffold` currently has zero
`yaml.v3` usage, so `ufconfig.go` adds that package's first
`gopkg.in/yaml.v3` import edge (package efferent coupling ΔCe +1; the
projected instability shift is far below every vibe-check gate).

### D6: Gaze repo `.uf` registration is a direct git edit

The repo's own `.uf/reviewer-capabilities.yaml` / `.uf/review-matrix.yaml`
(both git-tracked) are edited directly in the PR to the exact end-state
`syncUFReviewConfig` produces in target projects — `gaze init` is not run
against the gaze repo itself (repo `.opencode/` sources are deliberately
marker-free; markers are target-project artifacts).

### D7: Reference sweep in lockstep

Both copies (`.opencode/` source + `internal/scaffold/assets/` mirror) of
`speckit.testreview.md` (description, Step 3 heading, delegation target,
checklist reference) and `gaze-test-generator.md` (rubric reference) are
updated together — byte-identity is test-enforced.
`docs/guides/opencode-integration.md` updates the file table (the
`divisor-gaze-testing.md` row states tool-owned ownership, fixing the stale
"Tool-owned (via `gaze-fix.md` command)" wording on the old row), the
tool-owned list, and documents the new `.uf` registration + orphan-cleanup
behavior (Documentation Validation Gate: user-facing `gaze init` change).
`AGENTS.md` updates the embedded-file list. Test expectations
(`scaffold_test.go` expected paths, `TestIsToolOwned` cases,
`main_test.go` expected paths) follow the rename.

### D8: Verification derives exact commands from CI

Per the AGENTS.md CI Parity Gate, Phase 5 derives commands from
`.github/workflows/test.yml` / MegaLinter rather than the issue's shorthand:
`go build ./...`, `go test -race -count=1 -short -timeout 15m ./...`, the
e2e self-check leg
(`go test -race -count=1 -run 'TestRunSelfCheck' -timeout 30m ./cmd/gaze/...`),
and `golangci-lint run`. `TestEmbeddedAssetsMatchSource` must be green.
Governance: review council APPROVE, `unbound-force/website` issue before
merge (user-facing `gaze init` change), PR closes #303
(`originating_issue: 303` in `.openspec.yaml`).

## Risks / Trade-offs

- **Ownership change removes the customization path** (D4): users who edited
  the persona get it overwritten on re-init. Accepted per the locked design;
  the file is tool-managed rubric machinery, and the same run deploys the
  replacement.
- **Marker heuristic is provenance, not modification detection** (D5): a
  scaffolded-then-edited `reviewer-testing.md` still carries the marker and
  is removed. The marker is the only reliable provenance signal; a
  content-diff heuristic would keep stale orphans whenever users tweaked
  anything. Documented so users know to preserve customizations before
  upgrading.
- **`yaml.v3` comment repositioning**: foot/inline comments attached to the
  last node of a sequence can shift position on re-encode. Mitigated by the
  comment-preservation test pinning expectations and by the no-write-on-no-
  change rule (registered projects are never rewritten).
- **Two test-quality reviewers in capabilities** (`divisor-testing` +
  `divisor-gaze-testing`, both scoped `test-quality`): intentional
  coexistence. Advisor lists disambiguate for `code`/`spec`/`feedback`;
  `triage` keeps the generic tester only; `uf.review-council.md` handles
  unknown personas generically. Risk of double-dispatch exists only if a
  future uf version routes capabilities directly — accepted and documented.
- **Dispatch-plugin test review stays generic** (known limitation): the
  plugin special-cases `divisor-testing`; only `/speckit.testreview`'s
  direct Task-tool delegation moves to the new agent.
- **`.uf` schema evolution** (review-matrix `version: 3` may change): the
  merge is field-targeted and `Node`-transparent, so unknown fields survive;
  a future renaming of `advisor` would degrade to warn-and-skip, not corrupt.
