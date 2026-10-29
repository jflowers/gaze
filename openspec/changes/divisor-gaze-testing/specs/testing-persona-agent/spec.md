## ADDED Requirements

### Requirement: Divisor slot occupancy

The gaze testing persona SHALL occupy the new unique divisor slot
`divisor-gaze-testing`, materialized as `.opencode/agents/divisor-gaze-testing.md`.
It MUST NOT replace, modify, or take over the uf-owned generic
`divisor-testing` persona.

#### Scenario: Renamed agent file exists

- **GIVEN** the gaze repository after this change
- **WHEN** the `.opencode/agents/` directory is listed
- **THEN** `divisor-gaze-testing.md` MUST exist
- **AND** `reviewer-testing.md` MUST NOT exist

#### Scenario: Discovery by uf divisor globs

- **GIVEN** uf tooling that discovers reviewers by globbing
  `.opencode/agents/divisor-*.md`
- **WHEN** reviewer discovery runs against the gaze repository
- **THEN** `divisor-gaze-testing` MUST be discovered as an available reviewer

### Requirement: Universal core rubric

The persona's audit rubric MUST be language-neutral and grounded in gaze's
report data model: CRAP threshold (> 30), contract coverage, assertion
mappings, quadrants, fix strategies, and P0-P4 effect tiers. The rubric MUST
apply identically whether the underlying data originates from the Go
provider or an external analyzer, because both produce the same report data
model.

#### Scenario: Rubric is driven by gaze report data

- **GIVEN** a review with gaze quality data available (CRAP scores, contract
  coverage, assertion mappings, quadrants, fix strategies, effect tiers)
- **WHEN** the persona audits test quality
- **THEN** findings MUST reference those metrics rather than generic
  coverage platitudes
- **AND** high-complexity under-tested functions (CRAP > 30) MUST be
  prioritized over simple accessors

#### Scenario: Provider-agnostic application

- **GIVEN** a non-Go project analyzed through an external analyzer
- **WHEN** the persona audits test quality
- **THEN** the core rubric MUST still apply (same metrics, same thresholds)

### Requirement: Conditional Go-native section

The persona MUST include a Go-native conventions section that applies only
when the project under review is Go: standard library `testing` package only
(no testify, gomega, or other external assertion libraries),
`t.Errorf`/`t.Fatalf` direct assertions, `TestXxx_Description` test naming,
`testdata/src` fixtures, `-race -count=1` execution compatibility, and
`testing.Short()` guards for slow tests.

#### Scenario: Go project review

- **GIVEN** a Go project under review
- **WHEN** the persona audits test code
- **THEN** it MUST apply the Go-native conventions
- **AND** MUST flag third-party assertion libraries, missing
  `TestXxx_Description` naming, and race-detector-incompatible tests

#### Scenario: Non-Go project review

- **GIVEN** a non-Go project under review
- **WHEN** the persona audits test code
- **THEN** the Go-native conventions MUST NOT be applied

### Requirement: Conditional external-analyzer section

The persona MUST include an external-analyzer section for non-Go projects
that defers test-framework specifics to the repository's `AGENTS.md` and
`[PACK]` convention packs, and maps universal taxonomy side-effect types
(e.g., `MonkeyPatch`, `GeneratorYield`, `StreamOutput`) to the language
constructs they represent.

#### Scenario: Framework specifics deferred

- **GIVEN** a non-Go project with an `AGENTS.md` describing its test
  framework
- **WHEN** the persona audits test code
- **THEN** framework-specific judgments MUST follow the repo's `AGENTS.md` /
  convention packs rather than assumed conventions

#### Scenario: Universal taxonomy mapping

- **GIVEN** quality data whose side-effect types are universal taxonomy
  entries
- **WHEN** the persona evaluates assertion coverage of those effects
- **THEN** each type MUST be interpreted as its language-construct meaning
  (e.g., `MonkeyPatch` as runtime replacement of functions/methods)

### Requirement: Frontmatter contract

The agent file's YAML frontmatter MUST NOT pin a `model:` (review-matrix
tiers resolve models), MUST set `temperature: 0.1`, and MUST express tool
restrictions as a divisor-style `permission:` deny block instead of the
legacy `tools:` block.

#### Scenario: Frontmatter shape

- **GIVEN** `.opencode/agents/divisor-gaze-testing.md`
- **WHEN** its frontmatter is parsed
- **THEN** `model:` MUST be absent
- **AND** `temperature` MUST equal `0.1`
- **AND** a `permission:` block MUST deny write/edit/bash capabilities

### Requirement: Byte-identical tool-owned scaffold mirror

The embedded copy at
`internal/scaffold/assets/agents/divisor-gaze-testing.md` MUST be
byte-identical to `.opencode/agents/divisor-gaze-testing.md` (File Sync
Invariant, enforced by `TestEmbeddedAssetsMatchSource`), the legacy embedded
`agents/reviewer-testing.md` copy MUST be deleted, and
`agents/divisor-gaze-testing.md` MUST be classified tool-owned in
`isToolOwned` (overwrite-on-diff, even without `--force`).

#### Scenario: Asset manifest after rename

- **GIVEN** the scaffold's embedded assets
- **WHEN** the asset manifest is enumerated
- **THEN** it MUST contain exactly 9 files including
  `agents/divisor-gaze-testing.md`
- **AND** MUST NOT contain `agents/reviewer-testing.md`

#### Scenario: Overwrite-on-diff for the tool-owned persona

- **GIVEN** a target project whose `.opencode/agents/divisor-gaze-testing.md`
  differs from the embedded version
- **WHEN** `gaze init` runs without `--force`
- **THEN** the file MUST be updated to the embedded version
- **AND** the action MUST be reported as `updated` in the summary

#### Scenario: Drift detection stays green

- **GIVEN** the mirrored pair of agent files
- **WHEN** `TestEmbeddedAssetsMatchSource` runs
- **THEN** the embedded copy MUST be byte-identical to the `.opencode/`
  source

### Requirement: Gaze repository .uf registration

The gaze repository's `.uf` review config MUST register
`divisor-gaze-testing`: the capabilities file gains a `test-quality` review
entry, and the matrix advisor lists swap `divisor-testing` for
`divisor-gaze-testing` in the `code`, `spec`, and `feedback` modes only.
Everything else — the `triage` advisor list, the generic `divisor-testing`
capabilities entry, and the `uf.review-council.md` persona table — MUST
remain unchanged.

#### Scenario: Capabilities entry appended

- **GIVEN** `.uf/reviewer-capabilities.yaml`
- **WHEN** the registration is applied
- **THEN** a `divisor-gaze-testing` reviewer entry with capability `review`
  and scopes `[test-quality]` MUST exist alongside the unchanged
  `divisor-testing` entry

#### Scenario: Advisor swap in three modes

- **GIVEN** `.uf/review-matrix.yaml`
- **WHEN** the registration is applied
- **THEN** the `code`, `spec`, and `feedback` advisor lists MUST contain
  `divisor-gaze-testing` and MUST NOT contain `divisor-testing`
- **AND** the `triage` advisor list MUST still contain `divisor-testing`

#### Scenario: Review council inclusion

- **GIVEN** the registered persona
- **WHEN** a review council or dispatch plan is produced for the gaze
  repository
- **THEN** `divisor-gaze-testing` MUST participate as the test-quality
  reviewer
