## ADDED Requirements

### Requirement: Idempotent .uf registration merge

When the target project contains `.uf/reviewer-capabilities.yaml` and
`.uf/review-matrix.yaml`, `gaze init` MUST merge the persona registration
into them: append the reviewer entry
`{agent: divisor-gaze-testing, capability: review, scopes: [test-quality]}`
to the capabilities file if absent, and in the matrix file remove
`divisor-testing` from and add `divisor-gaze-testing` to the `code`, `spec`,
and `feedback` advisor lists. The merge MUST be idempotent — a second run
MUST produce no changes.

#### Scenario: Existing .uf config is merged

- **GIVEN** a target project with `.uf/reviewer-capabilities.yaml` and
  `.uf/review-matrix.yaml` that do not yet reference `divisor-gaze-testing`
- **WHEN** `gaze init` runs
- **THEN** the capabilities file MUST contain the `divisor-gaze-testing`
  reviewer entry
- **AND** the matrix advisor lists for `code`, `spec`, and `feedback` MUST
  contain `divisor-gaze-testing` and MUST NOT contain `divisor-testing`
- **AND** the `triage` advisor list MUST be unchanged

#### Scenario: Idempotent second run

- **GIVEN** a target project already registered by a previous `gaze init`
- **WHEN** `gaze init` runs again
- **THEN** both `.uf` files MUST be byte-identical to their state before the
  second run (no rewrite, no duplicate entries)

### Requirement: Comment-preserving YAML round-trip

The .uf merge MUST preserve existing comments in the merged `.uf/*.yaml`
files, including the `# scaffolded by uf` scaffold marker and inline
comments, when re-encoding the documents.

#### Scenario: Comments survive the merge

- **GIVEN** a `.uf/reviewer-capabilities.yaml` containing header comments and
  a trailing `# scaffolded by uf` marker
- **WHEN** `gaze init` merges the persona entry
- **THEN** all pre-existing comments MUST remain in the written file

### Requirement: Never create .uf files

When the target project has no `.uf/` directory, `gaze init` MUST NOT create
any `.uf/` files; it MUST scaffold the agent file and print a notice that
.uf registration was skipped. This guard exists because a minimal
capabilities file missing the canonical nine reviewers fails uf's
`parseReviewerManifest` KNOWN_REVIEWERS enforcement.

#### Scenario: Absent .uf directory

- **GIVEN** a target project without a `.uf/` directory
- **WHEN** `gaze init` runs
- **THEN** `.opencode/agents/divisor-gaze-testing.md` MUST be scaffolded
- **AND** no `.uf/` directory or files MUST be created
- **AND** the summary MUST print a notice that .uf registration was skipped

### Requirement: Malformed YAML degrades gracefully

When an existing `.uf/*.yaml` file cannot be parsed, `gaze init` MUST warn
and skip the registration for that file rather than failing the init or
overwriting the malformed file.

#### Scenario: Malformed capabilities file

- **GIVEN** a target project whose `.uf/reviewer-capabilities.yaml` is not
  valid YAML
- **WHEN** `gaze init` runs
- **THEN** a warning MUST be printed
- **AND** the malformed file MUST be left unchanged on disk
- **AND** `gaze init` MUST still succeed (exit 0, remaining scaffolding
  completes)

### Requirement: Gaze-marked orphan cleanup

`gaze init` MUST remove a legacy `.opencode/agents/reviewer-testing.md` in
the target project only when the file carries a
`<!-- scaffolded by gaze ... -->` version marker, reporting the removal in
the summary. Files without the marker (user-authored or independently
maintained copies) MUST be left untouched.

#### Scenario: Marker-bearing orphan removed

- **GIVEN** a target project containing `.opencode/agents/reviewer-testing.md`
  with a `<!-- scaffolded by gaze ... -->` marker
- **WHEN** `gaze init` runs
- **THEN** the file MUST be deleted
- **AND** the summary MUST report the removal

#### Scenario: Marker-less copy untouched

- **GIVEN** a target project containing `.opencode/agents/reviewer-testing.md`
  without any gaze version marker
- **WHEN** `gaze init` runs
- **THEN** the file MUST be left unchanged
- **AND** no error or removal MUST be reported for it

#### Scenario: No orphan present

- **GIVEN** a target project with no `.opencode/agents/reviewer-testing.md`
- **WHEN** `gaze init` runs
- **THEN** orphan cleanup MUST be a silent no-op
