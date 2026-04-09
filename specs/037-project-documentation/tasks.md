# Tasks: Project Documentation

**Input**: Design documents from `/specs/037-project-documentation/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation. This is a documentation-only spec — all tasks create Markdown files in `docs/`. No Go code, no tests, no build changes.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths and source files to read

---

## Phase 1: Setup (Directory Structure)

**Purpose**: Create the `docs/` directory tree so all subsequent tasks can write files into the correct locations.

- [x] T001 Create `docs/` directory structure per Diataxis framework: `docs/getting-started/`, `docs/concepts/`, `docs/reference/cli/`, `docs/guides/`, `docs/architecture/`, `docs/porting/` (FR-001, FR-003 — all files will be self-contained plain Markdown with no build step)

**Checkpoint**: All directories exist. No content files yet.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create `docs/index.md` and `docs/reference/glossary.md` — these are referenced by ALL other documentation pages and MUST be complete first.

**CRITICAL**: No user story work can begin until this phase is complete. Every doc page links to the glossary on first use of domain terms (R6), and the index is the navigation hub (FR-002).

- [x] T002 Create `docs/reference/glossary.md` — canonical definitions for all domain terms (FR-015, SC-010). Terms: CRAPload, GazeCRAPload, GazeCRAP, CRAP, quadrant names (Q1–Q4), contract coverage, over-specification, assertion mapping, fix strategy labels, tier names (P0–P4), classification labels, signal names, side effect, behavioral contract, Diataxis. **Source**: all concept docs (forward-looking), README, spec artifacts, `internal/crap/crap.go`, `internal/taxonomy/types.go`, `internal/classify/score.go`
- [x] T003 Create `docs/index.md` — navigation hub with audience routing (FR-002). Sections: "New to Gaze?", "Understanding the Model", "CLI Reference", "Integrating with CI", "Contributing", "Porting to Other Languages". Include complete table of contents with links to all planned pages. **Source**: plan.md project structure, data-model.md section taxonomy

**Checkpoint**: Glossary and index exist. All subsequent pages can link to glossary terms and be listed in the index.

---

## Phase 3: User Story 1 — Power User Learns Core Concepts (Priority: P1)

**Goal**: A power user who has run `gaze analyze` can understand what the output means — side effect types, classification labels, confidence scores, quadrants, and fix strategies.

**Independent Test**: A new user reads only the concepts section and then correctly interprets a `gaze crap` output — identifying which functions are dangerous, why, and what fix strategy applies.

### Implementation for User Story 1

- [x] T004 [P] [US1] Create `docs/concepts/side-effects.md` — all 37 effect types across 5 tiers (P0–P4) with definitions, examples, and detection status (FR-007, SC-003, FR-028). **Source**: `internal/taxonomy/types.go` (all `SideEffectType` constants and `Tier` constants), spec-001. Mark P3/P4 types as "Defined — detection not yet implemented."
- [x] T005 [P] [US1] Create `docs/concepts/classification.md` — the five signal analyzers (interface, visibility, caller, naming, godoc), confidence scoring formula, tier-based boosts (P0 +25, P1 +10), three classification labels (contractual ≥75, ambiguous 50–74, incidental <50), threshold semantics (FR-008). **Source**: `internal/classify/score.go` (`tierBoost`, `accumulateSignals`, `classifyLabel`), `internal/classify/interface.go`, `internal/classify/visibility.go`, `internal/classify/callers.go`, `internal/classify/naming.go`, `internal/classify/godoc.go`, spec-002
- [x] T006 [P] [US1] Create `docs/concepts/scoring.md` — CRAP formula, GazeCRAP formula, four quadrants (Q1 Safe, Q2 Complex But Tested, Q3 Simple But Underspecified, Q4 Dangerous), fix strategies (decompose, add_tests, add_assertions, decompose_and_test), CRAPload, GazeCRAPload, worked examples (FR-009, SC-004). **Source**: `internal/crap/crap.go` (`Formula`, `GazeCRAPFormula`, `ClassifyQuadrant`, `FixStrategy` constants), `internal/crap/analyze.go` (`assignFixStrategy`), spec-004
- [x] T007 [P] [US1] Create `docs/concepts/quality.md` — test-target pairing, assertion detection, four mechanical mapping passes (direct identity, indirect root, helper bridge, inline call), contract coverage, over-specification (FR-010). **Source**: `internal/quality/quality.go`, `internal/quality/assertion.go`, `internal/quality/mapping.go` (the four passes), `internal/quality/pairing.go`, spec-003
- [x] T008 [P] [US1] Create `docs/concepts/analysis-pipeline.md` — how AST and SSA analysis work together, what each phase detects (returns/sentinels via AST, mutations via SSA), how they compose into the full analysis result (FR-011). **Source**: `internal/analysis/returns.go`, `internal/analysis/mutation.go`, `internal/analysis/effects.go`, `internal/loader/loader.go`, spec-001

**Checkpoint**: All 5 concept pages complete. A power user can understand Gaze's model by reading these pages. Verify: all 37 effect types listed (SC-003), formulas match source code (SC-004), glossary terms linked on first use.

---

## Phase 4: User Story 2 — New User Gets Started Quickly (Priority: P1)

**Goal**: A developer who has never used Gaze can install it and produce meaningful output from `analyze`, `crap`, and `quality` on their own Go project within 10 minutes.

**Independent Test**: Time a developer following only the quickstart on a fresh Go project — they complete all steps and can describe what Gaze told them.

### Implementation for User Story 2

- [x] T009 [P] [US2] Create `docs/getting-started/installation.md` — Homebrew (`brew install unbound-force/tap/gaze`), `go install`, build-from-source, platform notes for macOS/Linux (FR-004). **Source**: README.md §Installation, `.goreleaser.yaml`, `go.mod`
- [x] T010 [P] [US2] Create `docs/getting-started/concepts.md` — why line coverage is insufficient, what contract-level analysis adds, mental model overview before tool usage (FR-006). **Source**: README.md introduction, spec-001, spec-003. Link to `docs/concepts/` for deeper dives.
- [x] T011 [US2] Create `docs/getting-started/quickstart.md` — guided walkthrough: install → `gaze analyze ./...` → `gaze crap` → `gaze quality` on user's own project, with expected output and interpretation guidance (FR-005, SC-001). **Source**: README.md §Commands, actual CLI output from running commands. Link to concept pages for "learn more" and CLI reference pages for "all flags."

**Checkpoint**: All 3 getting-started pages complete. A new user can go from zero to meaningful output. Verify: quickstart is completable in <10 minutes (SC-001), installation covers all methods.

---

## Phase 5: User Story 3 — Developer Configures CI Quality Gates (Priority: P2)

**Goal**: A team lead can integrate Gaze into their CI pipeline with threshold enforcement, coverage profile reuse, and AI-powered reports.

**Independent Test**: Follow the CI guide to add Gaze to a GitHub Actions workflow on a test repository — threshold violations fail the build.

### Implementation for User Story 3

- [x] T012 [P] [US3] Create `docs/guides/ci-integration.md` — GitHub Actions workflow YAML (copy-pasteable), coverage profile generation (`-coverprofile=coverage.out`), `--coverprofile` flag to avoid double test runs, threshold enforcement (`--max-crapload`, `--max-gaze-crapload`, `--min-contract-coverage`), step summary integration (FR-016, SC-005). **Source**: `.github/workflows/test.yml`, README.md §CI Integration, spec-018, spec-020
- [x] T013 [P] [US3] Create `docs/guides/ai-reports.md` — adapter setup for each AI backend (Claude, Gemini, Ollama, OpenCode), model configuration (`--model`), `--ai` flag usage, GitHub Step Summary integration, secret management for CI (FR-017). **Source**: `internal/aireport/adapter.go`, `internal/aireport/adapter_claude.go`, `internal/aireport/adapter_gemini.go`, `internal/aireport/adapter_ollama.go`, `internal/aireport/adapter_opencode.go`, spec-018
- [x] T014 [P] [US3] Create `docs/guides/opencode-integration.md` — `gaze init` command, scaffolded agent/command files, how to use the `/gaze` command in OpenCode, what the gaze-reporter agent does (FR-018). **Source**: `internal/scaffold/scaffold.go`, `internal/scaffold/assets/`, README.md §OpenCode Integration, spec-005
- [x] T015 [P] [US3] Create `docs/guides/improving-scores.md` — organized by fix strategy: `decompose` (reduce complexity), `add_tests` (zero coverage), `add_assertions` (Q3 — has line coverage, lacks contract assertions), `decompose_and_test` (both needed). Include concrete before/after examples for each strategy (FR-019). **Source**: `internal/crap/crap.go` (`FixStrategy` constants), `internal/crap/analyze.go` (`assignFixStrategy`), spec-009

**Checkpoint**: All 4 guide pages complete. A developer can set up CI integration and improve their scores. Verify: CI guide produces working workflow (SC-005).

---

## Phase 6: User Story 4 — Developer Looks Up CLI Details (Priority: P2)

**Goal**: A user can look up any CLI command's flags, defaults, output format, and behavior edge cases on a dedicated reference page.

**Independent Test**: Select any command, read its reference page, verify every flag from `gaze <cmd> --help` appears with accurate descriptions.

### Implementation for User Story 4

- [x] T016 [P] [US4] Create `docs/reference/cli/analyze.md` — all flags from `gaze analyze --help` with type, default, description, `.gaze.yaml` interaction (FR-012). **Source**: `cmd/gaze/main.go` (Cobra command definition for `analyzeCmd`), `gaze analyze --help` output
- [x] T017 [P] [US4] Create `docs/reference/cli/crap.md` — all flags from `gaze crap --help` (FR-012). **Source**: `cmd/gaze/main.go` (`crapCmd`), `gaze crap --help` output
- [x] T018 [P] [US4] Create `docs/reference/cli/quality.md` — all flags from `gaze quality --help` (FR-012). **Source**: `cmd/gaze/main.go` (`qualityCmd`), `gaze quality --help` output
- [x] T019 [P] [US4] Create `docs/reference/cli/report.md` — all flags from `gaze report --help`, including `--ai`, `--model`, `--coverprofile`, threshold flags (FR-012). **Source**: `cmd/gaze/main.go` (`reportCmd`), `gaze report --help` output
- [x] T020 [P] [US4] Create `docs/reference/cli/self-check.md` — all flags from `gaze self-check --help` (FR-012). **Source**: `cmd/gaze/main.go` (`selfCheckCmd`), `gaze self-check --help` output
- [x] T021 [P] [US4] Create `docs/reference/cli/docscan.md` — all flags from `gaze docscan --help` (FR-012). **Source**: `cmd/gaze/main.go` (`docscanCmd`), `gaze docscan --help` output
- [x] T022 [P] [US4] Create `docs/reference/cli/schema.md` — all flags from `gaze schema --help` (FR-012). **Source**: `cmd/gaze/main.go` (`schemaCmd`), `gaze schema --help` output
- [x] T023 [P] [US4] Create `docs/reference/cli/init.md` — all flags from `gaze init --help` (FR-012). **Source**: `cmd/gaze/main.go` (`initCmd`), `gaze init --help` output
- [x] T024 [US4] Create `docs/reference/configuration.md` — all `.gaze.yaml` keys with types, defaults, validation rules, and interaction with CLI flags (FR-013). **Source**: `internal/config/config.go`, any existing `.gaze.yaml` examples in README or specs
- [x] T025 [US4] Create `docs/reference/json-schemas.md` — schema references and annotated example output for each command that supports `--format=json` (analyze, crap, quality) (FR-014). **Source**: `internal/report/schema.go` (embedded JSON Schema), actual JSON output from running commands

**Checkpoint**: All 11 reference pages complete. Every flag from every `--help` is documented (SC-002). Verify: diff each CLI reference page against `gaze <cmd> --help` output.

---

## Phase 7: User Story 5 — Contributor Understands the Architecture (Priority: P3)

**Goal**: A new contributor can understand the package structure, data flow, coding conventions, and extension points well enough to make a contribution.

**Independent Test**: A new contributor reads the architecture docs and successfully adds a stub side effect type with tests, following the documented patterns.

### Implementation for User Story 5

- [x] T026 [P] [US5] Create `docs/architecture/overview.md` — package dependency graph, data flow from CLI entry through analysis/classification/scoring to output, role of each `internal/` package (FR-020, SC-006). **Source**: README.md §Architecture, AGENTS.md, package structure (`internal/analysis/`, `internal/classify/`, `internal/crap/`, `internal/quality/`, `internal/report/`, `internal/aireport/`, `internal/scaffold/`, `internal/loader/`, `internal/config/`, `internal/docscan/`, `internal/taxonomy/`)
- [x] T027 [P] [US5] Create `docs/architecture/contributing.md` — dev environment setup, build/test commands (`go build`, `go test -race -count=1 -short ./...`, `golangci-lint run`), coding conventions (naming, error handling, import grouping, no global state), testing patterns (standard library only, no testify), spec-first workflow (FR-021, SC-006). **Source**: AGENTS.md §Build & Test Commands, §Coding Conventions, §Testing Conventions, §Git & Workflow, `.specify/` directory structure
- [x] T028 [P] [US5] Create `docs/architecture/extending.md` — how to add new side effect types (modify `internal/taxonomy/types.go`, add detection in `internal/analysis/`), new classification signals (add analyzer in `internal/classify/`), new output formats (modify `internal/report/`), new AI adapters (implement `AIAdapter` interface in `internal/aireport/`) (FR-022). **Source**: `internal/taxonomy/types.go`, `internal/analysis/`, `internal/classify/interface.go`, `internal/report/`, `internal/aireport/adapter.go`

**Checkpoint**: All 3 architecture pages complete. A contributor can set up, build, test, and extend Gaze (SC-006).

---

## Phase 8: User Story 6 — Language Porter Understands the Contracts (Priority: P3)

**Goal**: A team building "Gaze for Python" (or Rust, TypeScript, etc.) can enumerate all behavioral contracts without referencing Go source code.

**Independent Test**: A developer reads only the porting section and lists all behavioral contracts a port must satisfy, with no ambiguity about required vs optional.

### Implementation for User Story 6

- [x] T029 [P] [US6] Create `docs/porting/contracts.md` — all language-agnostic behavioral contracts: effect taxonomy (37 types, 5 tiers), classification rules (5 signals, confidence formula, tier boosts, label thresholds), scoring formulas (CRAP, GazeCRAP, quadrants). Clearly separate from Go-specific implementation details (FR-023, SC-007). **Source**: `internal/taxonomy/types.go`, `internal/crap/crap.go`, `internal/classify/score.go`
- [x] T030 [P] [US6] Create `docs/porting/requirements.md` — capability checklist: effect detection (required), classification (required), CRAP scoring (required), quality assessment (optional), AI reports (optional), docscan (optional). Each requirement labeled required or optional (FR-024). **Source**: all spec artifacts (spec-001 through spec-022), feature capability matrix derived from plan.md
- [x] T031 [P] [US6] Create `docs/porting/taxonomy-reference.md` — canonical list of all 37 effect types with tier assignments, detection status (Implemented/Defined), and scoring formulas in a format suitable for mechanical extraction (tables, not prose) (FR-025, SC-007). **Source**: `internal/taxonomy/types.go`, `internal/crap/crap.go` (formulas)

**Checkpoint**: All 3 porting pages complete. A porter can enumerate all contracts without Go source (SC-007).

---

## Phase 9: Polish — README Refactoring & Cross-References

**Purpose**: Refactor README to link into `docs/` for detailed content (FR-026), update `docs/index.md` with final cross-references, and verify all internal links resolve.

- [x] T032 [US1-US6] Refactor `README.md` — replace extended explanations (effect taxonomy tables, formula derivations, detailed flag tables) with links into `docs/`. Preserve README as a concise project overview with installation, quick examples, and links to docs/ for depth. No duplication between README and docs/ (FR-026, SC-008). **Source**: current README.md, all `docs/` pages created in T002–T031
- [x] T033 [US1-US6] Update `docs/index.md` — finalize table of contents with all actual page titles and verified relative links. Ensure every page created in T002–T031 is listed. Verify all cross-reference links across all docs resolve correctly (FR-002, SC-009). **Source**: all `docs/` pages
- [x] T034 [US1-US6] Cross-reference audit — verify: (a) every concept doc that mentions a CLI command links to the reference page, not inline flag tables (FR-027); (b) every domain term links to glossary on first use (R6); (c) P3/P4 effects are marked "Defined — detection not yet implemented" wherever they appear (FR-028); (d) all relative links resolve (SC-009)

**Checkpoint**: All documentation complete. README links to docs/ (SC-008). All links resolve (SC-009). No duplicated content.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — glossary and index BLOCK all user stories
- **Phases 3–4 (US1, US2 — P1)**: Depend on Phase 2. Can run in parallel with each other.
- **Phases 5–6 (US3, US4 — P2)**: Depend on Phase 2. Can run in parallel with each other and with Phases 3–4. Guides in Phase 5 benefit from concept pages (Phase 3) being available for cross-linking but are not strictly blocked.
- **Phases 7–8 (US5, US6 — P3)**: Depend on Phase 2. Can run in parallel with each other. Architecture pages benefit from all prior sections for cross-linking but are not strictly blocked.
- **Phase 9 (Polish)**: Depends on ALL prior phases — README refactoring and cross-reference audit require all docs to exist.

### Within Each User Story

- Tasks marked [P] within a phase can run in parallel (they write independent files)
- Tasks without [P] have implicit sequential dependencies (e.g., T011 quickstart benefits from T009 installation and T010 concepts being done first)
- T024 (configuration.md) and T025 (json-schemas.md) are sequential within Phase 6 because they may reference each other

### Parallel Opportunities

- **Phase 3**: All 5 concept pages (T004–T008) can run in parallel — each covers an independent topic
- **Phase 4**: T009 and T010 can run in parallel; T011 follows (references both)
- **Phase 5**: All 4 guide pages (T012–T015) can run in parallel — each covers an independent goal
- **Phase 6**: All 8 CLI reference pages (T016–T023) can run in parallel — each covers one subcommand
- **Phase 7**: All 3 architecture pages (T026–T028) can run in parallel
- **Phase 8**: All 3 porting pages (T029–T031) can run in parallel
- **Cross-phase**: Phases 3–8 can all run in parallel after Phase 2 completes (with best results when concept pages land first for cross-linking)

---

## Summary

| Metric | Count |
|--------|-------|
| **Total tasks** | 34 |
| **Phase 1 (Setup)** | 1 |
| **Phase 2 (Foundational)** | 2 |
| **Phase 3 (US1 — Concepts)** | 5 |
| **Phase 4 (US2 — Getting Started)** | 3 |
| **Phase 5 (US3 — Guides)** | 4 |
| **Phase 6 (US4 — CLI Reference)** | 10 |
| **Phase 7 (US5 — Architecture)** | 3 |
| **Phase 8 (US6 — Porting)** | 3 |
| **Phase 9 (Polish)** | 3 |
| **Parallelizable tasks [P]** | 28 of 34 (82%) |
| **Max parallel width** | 8 (Phase 6 CLI reference pages) |

### Tasks per User Story

| Story | Tasks | Priority |
|-------|-------|----------|
| US1 — Core Concepts | 5 (T004–T008) | P1 |
| US2 — Getting Started | 3 (T009–T011) | P1 |
| US3 — CI & Guides | 4 (T012–T015) | P2 |
| US4 — CLI Reference | 10 (T016–T025) | P2 |
| US5 — Architecture | 3 (T026–T028) | P3 |
| US6 — Porting | 3 (T029–T031) | P3 |
| Cross-cutting | 6 (T001–T003, T032–T034) | — |

<!-- spec-review: passed -->
