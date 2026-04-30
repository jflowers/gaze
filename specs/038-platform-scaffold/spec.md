# Feature Specification: Multi-Platform Scaffold Deployment

**Feature Branch**: `038-platform-scaffold`  
**Created**: 2026-04-30  
**Status**: Draft  
**Input**: User description: "Create an analogous spec from unbound-force/unbound-force PR #144 for gaze to support platforms other than OpenCode"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cursor-Only Project Scaffolding (Priority: P1)

A developer uses Cursor as their primary AI coding
assistant and wants gaze's quality agents and commands
available natively in Cursor. They run `gaze init
--platform cursor` and receive a `.cursor/` directory
containing gaze-reporter, gaze-test-generator, and
reviewer-testing agents, plus commands and reference
files -- all in Cursor's native formats. The gaze
tooling works immediately in Cursor without any
additional configuration.

**Why this priority**: This is the core value
proposition. Without correct Cursor file generation, the
feature has no purpose. Users of Cursor currently cannot
use gaze's quality reporting and test generation agents.

**Independent Test**: Can be fully tested by running
`gaze init --platform cursor` in an empty directory
with a `go.mod` file and verifying all generated files
match Cursor's expected directory structure, frontmatter
schemas, and file extensions.

**Acceptance Scenarios**:

1. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform cursor`,
   **Then** a `.cursor/` directory is created containing
   `agents/` and `commands/` subdirectories with
   correctly formatted files, and no `.opencode/`
   directory is created.

2. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform cursor`,
   **Then** agent files in `.cursor/agents/` have YAML
   frontmatter with `name` (derived from filename) and
   `description` fields, and do not contain OpenCode-
   specific fields (`mode`, `temperature`, `tools`,
   `maxSteps`, `disabled`).

3. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform cursor`,
   **Then** reference files are deployed to
   `.cursor/references/` with no content transformation
   (they are loaded on demand by agents via file read).

4. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform cursor`,
   **Then** command files are deployed to
   `.cursor/commands/` as plain Markdown with no content
   transformation.

5. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform cursor`,
   **Then** the scaffold summary output includes the
   platform name "cursor" confirming which platform was
   deployed to.

---

### User Story 2 - Dual-Platform Project Scaffolding (Priority: P2)

A team uses both OpenCode and Cursor across different
developers. A lead runs `gaze init --platform opencode
--platform cursor` and the project receives both
`.opencode/` and `.cursor/` directories, each containing
platform-native files derived from the same canonical
embedded assets. Both tools can use gaze's quality
reporting in the same project.

**Why this priority**: Multi-platform support is the
strategic differentiator. Without this, users must
choose one platform or manually copy and adapt agent
files for their second tool.

**Independent Test**: Can be tested by running
`gaze init --platform opencode --platform cursor` in an
empty directory and verifying both `.opencode/` and
`.cursor/` directories are created with correct,
platform-appropriate content.

**Acceptance Scenarios**:

1. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform opencode
   --platform cursor`, **Then** both `.opencode/` and
   `.cursor/` directories are created, each containing
   agents, commands, and reference files.

2. **Given** a project with an existing `.opencode/`
   directory from a previous `gaze init`, **When** a
   user runs `gaze init --platform cursor`, **Then** a
   `.cursor/` directory is created alongside the
   existing `.opencode/` directory, and no files within
   the existing `.opencode/` directory are modified,
   created, or deleted.

3. **Given** a dual-platform project, **When** a user
   runs `gaze init --platform opencode --platform
   cursor` again, **Then** tool-owned files in both
   directories are updated if content has changed, and
   user-owned files in both directories are preserved.

4. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform opencode
   --platform cursor`, **Then** the scaffold summary
   output shows per-platform results with created,
   skipped, updated, and overwritten counts for each
   platform.

---

### User Story 3 - Backward-Compatible Default Behavior (Priority: P1)

An existing gaze user runs `gaze init` without any
`--platform` flag. The behavior is identical to today:
files are deployed to `.opencode/` exactly as before.
No `.cursor/` directory is created. No existing
workflows are broken.

**Why this priority**: Breaking backward compatibility
for existing users would be a regression. This must be
guaranteed alongside the new platform support.

**Independent Test**: Can be tested by running
`gaze init` (no `--platform` flag) and verifying the
output is functionally identical to the current behavior.

**Acceptance Scenarios**:

1. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init` with no `--platform`
   flag, **Then** files are deployed to `.opencode/`
   exactly as they are today, and no `.cursor/`
   directory is created.

2. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `gaze init --platform opencode`,
   **Then** the files deployed to `.opencode/` are
   functionally identical to those produced by
   `gaze init` with no `--platform` flag, with the same
   Created/Skipped/Updated/Overwritten results, and no
   `.cursor/` directory is created.

---

### User Story 4 - Force Overwrite Across Platforms (Priority: P3)

A developer wants to reset all gaze scaffold files to
the latest embedded versions, including files they have
customized. They run `gaze init --platform cursor
--force` and all Cursor files are overwritten, including
user-owned agents like `gaze-reporter.md`.

**Why this priority**: `--force` is an existing escape
hatch. Extending it to new platforms is a parity feature
that ensures consistent behavior.

**Independent Test**: Can be tested by running
`gaze init --platform cursor`, modifying a user-owned
agent file, then running `gaze init --platform cursor
--force` and verifying the modification is overwritten.

**Acceptance Scenarios**:

1. **Given** a project with a customized
   `.cursor/agents/gaze-reporter.md`, **When** a user
   runs `gaze init --platform cursor --force`, **Then**
   the customized file is overwritten with the latest
   embedded version.

2. **Given** a dual-platform project, **When** a user
   runs `gaze init --platform opencode --platform cursor
   --force`, **Then** `--force` applies to both
   platforms -- user-owned files in both `.opencode/`
   and `.cursor/` are overwritten.

---

### Edge Cases

- What happens when `--platform` is given an unknown
  value (e.g., `--platform vim`)? System rejects with
  a clear error listing valid platforms.
- What happens when `--platform` is specified multiple
  times with the same value (e.g., `--platform cursor
  --platform cursor`)? Deduplicated silently; files
  deployed once.
- What happens when a `.cursor/agents/` file exists
  with content the user has customized? User-owned
  files are never overwritten (same ownership model as
  `.opencode/`).
- What happens when `--platform cursor --force` is
  specified? All Cursor files are overwritten, including
  user-owned agents.
- What happens when `--platform opencode --platform
  cursor --force` is specified? `--force` applies to
  all selected platforms -- both `.opencode/` and
  `.cursor/` user-owned files are overwritten.
- What happens when `gaze init --platform cursor` is
  run in a directory with no `go.mod`? Same warning
  as today ("no go.mod found") but initialization
  proceeds.
- What happens when `--platform cursor --force` is
  run but no `.cursor/` directory exists yet? Same as
  a fresh init -- all files are created. `--force` is
  a no-op for non-existent files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept a `--platform` flag on
  `gaze init` that takes one or more platform names as a
  repeatable string value.
- **FR-002**: System MUST support two platform values:
  `opencode` (current behavior) and `cursor` (new).
- **FR-003**: System MUST default to `--platform
  opencode` when no `--platform` flag is specified,
  preserving 100% backward compatibility.
- **FR-004**: System MUST reject unknown platform values
  with an error message listing valid platforms.
- **FR-005**: System MUST deduplicate repeated platform
  values silently.
- **FR-006**: System MUST deploy embedded assets to
  all selected platforms in a single initialization
  operation, processing each asset for every applicable
  platform.
- **FR-007**: System MUST transform OpenCode agent
  frontmatter to Cursor format by: adding `name`
  (derived from filename stem), preserving
  `description`, and dropping OpenCode-specific fields
  (`mode`, `temperature`, `tools`, `maxSteps`,
  `disabled`).
  [NEEDS CLARIFICATION: The `model` field present in
  some agents and the `agent` field present in some
  commands are not addressed -- should these be dropped
  for Cursor, preserved, or transformed? See HIGH
  finding #1 in review council advisories.]
- **FR-008**: System MUST deploy command files to
  `.cursor/commands/` as plain Markdown with no content
  transformation beyond the version marker insertion.
- **FR-009**: System MUST deploy reference files to
  `.cursor/references/` with no content transformation
  beyond the version marker insertion.
- **FR-010**: System MUST apply the same file ownership
  model to Cursor files: tool-owned files are auto-
  updated on re-init, user-owned files are never
  overwritten unless `--force` is specified.
- **FR-011**: System MUST classify Cursor files using
  the same ownership rules as OpenCode files: the
  existing ownership classification logic determines
  ownership regardless of target platform.
- **FR-012**: System MUST insert the standard version
  marker (`<!-- scaffolded by gaze vX.Y.Z -->`) into
  all files regardless of target platform.
- **FR-013**: System MUST deploy files to the correct
  platform-specific root directory: `.opencode/` for
  OpenCode, `.cursor/` for Cursor.
- **FR-014**: System MUST support the `--force` flag for
  all platforms, overwriting user-owned files when
  specified.
- **FR-015**: System MUST display the selected
  platform(s) in the scaffold summary output so users
  can confirm which platforms were deployed to.
- **FR-016**: System MUST be architecturally extensible
  so that adding a new platform target requires no
  modification to the scaffold engine's core asset-
  walking logic. Each platform's path mapping and
  content transformation MUST be encapsulated
  independently.
- **FR-017**: When multiple platforms are selected,
  the system MUST NOT modify files belonging to a
  platform that was not selected. Running
  `gaze init --platform cursor` on a project with an
  existing `.opencode/` directory MUST leave the
  `.opencode/` files untouched.
- **FR-018**: System MUST report per-platform results
  in the scaffold summary, showing created, skipped,
  updated, and overwritten counts for each selected
  platform.

### Key Entities

- **Platform**: An abstraction representing a target AI
  coding tool. Each platform knows how to map an asset's
  relative path to a platform-specific output directory
  and how to transform content (e.g., frontmatter
  adaptation).
- **Asset**: An embedded file from the scaffold's
  embedded asset store that is deployed to one or more
  platform target directories.
- **File Ownership**: A classification (tool-owned or
  user-owned) that determines whether a file is auto-
  updated on re-init or preserved for user
  customization. Ownership is determined by the asset's
  relative path, independent of the target platform.

## Out of Scope

- Claude Code native `.claude/` directory deployment
  (not applicable; gaze does not manage Claude Code
  configuration)
- Copilot, Windsurf, or other platform support (future
  work via the platform extensibility mechanism)
- MCP configuration translation (gaze does not scaffold
  MCP server configuration; this is an `uf init`
  concern)
- Convention pack or `.mdc` rule translation (gaze does
  not scaffold convention packs; this is an `uf init`
  concern)
- Runtime platform auto-detection (detecting which AI
  tool is currently running)
- Bidirectional sync between platform configurations
- Migration tooling for converting existing platform
  configurations between formats
- `--divisor` mode (gaze does not have a DivisorOnly
  mode)

## Documentation Impact

The following documentation MUST be updated when this
feature is implemented:

- **README.md** — Update `gaze init` command description
  and OpenCode Integration section to mention Cursor
  and multi-platform support.
- **AGENTS.md** — Update Architecture scaffold
  description, Active Technologies, and Recent Changes
  sections.
- **GoDoc comments** — New exported types (platform
  abstraction) and modified function signatures.

A website documentation issue MUST be filed in
`unbound-force/website` per the Website Documentation
Gate before the implementing PR is merged, covering:
- Gaze project page update for Cursor support
- `gaze init` workflow description update for
  `--platform` flag

## Dependencies

- **Spec 005** (gaze-opencode-integration): Defines
  the original `gaze init` subcommand and the scaffold
  package architecture. Note: the scaffold has evolved
  through specs 012 (consolidation), 016 (context
  reduction), and 017 (testing persona) since spec 005.
- **Spec 016** (agent-context-reduction): Defines the
  tool-owned `references/` directory pattern and the
  file ownership classification logic that this spec
  preserves.

### Research References

- unbound-force/unbound-force PR #144 (Spec 035:
  Multi-platform scaffold deployment) informed the
  design. Adapted for gaze's simpler scaffold (8 files
  vs 50, no convention packs, no MCP config).

## Assumptions

- Cursor supports `.cursor/agents/` with Markdown files
  containing YAML frontmatter with `name` and
  `description` fields.
- Cursor supports `.cursor/commands/` with plain
  Markdown command files.
- The `--platform` flag follows established CLI
  conventions for repeatable string slices.
- The `references/` directory pattern (externalized
  context loaded on demand by agents) works the same
  way in Cursor as in OpenCode (agents read files via
  a file-read tool).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can scaffold a Cursor-native gaze
  project with a single command (`gaze init --platform
  cursor`) and have all agent files contain valid
  Cursor-format frontmatter.
- **SC-002**: Users can scaffold a dual-platform gaze
  project (`gaze init --platform opencode --platform
  cursor`) and use gaze's quality agents in both tools
  within the same repository without conflicts.
- **SC-003**: Running `gaze init --platform cursor`
  twice on the same project updates tool-owned files
  without overwriting user-customized agents.
- **SC-004**: All existing `gaze init` behavior is
  preserved when no `--platform` flag is specified
  (100% backward compatibility).
- **SC-005**: The scaffold summary output clearly
  identifies which platform(s) files were deployed to.
- **SC-006**: Adding a new platform target requires
  no changes to the scaffold engine's core asset-
  walking logic (verified by code review of the
  platform extensibility design).
