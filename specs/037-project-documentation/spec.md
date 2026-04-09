# Feature Specification: Project Documentation

**Feature Branch**: `037-project-documentation`  
**Created**: 2026-04-08  
**Status**: Approved  
**Input**: User description: "Create detailed project documentation in docs/ for power users, developers, and future language porters using Diataxis framework"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Power User Learns Core Concepts (Priority: P1)

A power user has installed Gaze and run `gaze analyze` on their project. They see side effects classified as "contractual" and "ambiguous" with confidence scores, but don't understand what these labels mean, how the classification signals work, or what levers they have to influence the results. They navigate to `docs/concepts/` and find clear explanations of the side effect taxonomy, the classification model, and the scoring system — enough to understand their output and make informed decisions about thresholds and configuration.

**Why this priority**: Without conceptual documentation, Gaze's output is opaque. Users who don't understand the model can't act on the results, which directly undermines the core principle of Actionable Output.

**Independent Test**: Can be fully tested by having a new user read only the concepts section and then correctly interpret a `gaze crap` output — identifying which functions are dangerous, why, and what fix strategy applies.

**Acceptance Scenarios**:

1. **Given** a user who has never used Gaze, **When** they read `docs/concepts/side-effects.md`, **Then** they can identify which tier (P0-P4) a side effect belongs to and explain why tiers exist.
2. **Given** a user who sees "ambiguous" classifications in output, **When** they read `docs/concepts/classification.md`, **Then** they can explain the five signal analyzers and how confidence thresholds determine the label.
3. **Given** a user with CRAPload functions in Q4, **When** they read `docs/concepts/scoring.md`, **Then** they can explain what CRAP vs GazeCRAP measures and what each quadrant means.
4. **Given** a user who wants better contract coverage, **When** they read `docs/concepts/quality.md`, **Then** they understand assertion mapping, over-specification, and how tests are paired with targets.

---

### User Story 2 - New User Gets Started Quickly (Priority: P1)

A developer hears about Gaze at a conference and wants to try it on their Go project. They visit the docs and find a guided quickstart that takes them from install through their first meaningful analysis in under 10 minutes — showing them `analyze`, `crap`, and `quality` in sequence on their own code with enough context to understand the output without reading the full conceptual docs.

**Why this priority**: First impressions determine adoption. A clear getting-started path is tied for P1 because it gates all other documentation usage.

**Independent Test**: Can be fully tested by timing a developer who follows only the quickstart on a fresh Go project and verifying they complete all steps and can describe what Gaze told them.

**Acceptance Scenarios**:

1. **Given** a developer with Go installed, **When** they follow `docs/getting-started/quickstart.md`, **Then** they complete install, analyze, crap, and quality commands within 10 minutes.
2. **Given** a developer who has never used Gaze, **When** they read `docs/getting-started/concepts.md`, **Then** they can articulate why line coverage alone is insufficient and what contract coverage adds.
3. **Given** a user on macOS or Linux, **When** they follow `docs/getting-started/installation.md`, **Then** they have a working `gaze` binary on their PATH via their preferred method (Homebrew, go install, or source build).

---

### User Story 3 - Developer Configures CI Quality Gates (Priority: P2)

A team lead wants to integrate Gaze into their CI pipeline to block PRs that introduce dangerous functions. They need a step-by-step guide covering coverage profile generation, threshold configuration, AI report setup, and GitHub Actions integration — with copy-pasteable YAML snippets that work out of the box.

**Why this priority**: CI integration is the primary enterprise use case and the path to team-wide adoption. It ranks P2 because it depends on understanding the core concepts (P1).

**Independent Test**: Can be fully tested by following the CI guide to add Gaze to a GitHub Actions workflow on a test repository and verifying that threshold violations fail the build.

**Acceptance Scenarios**:

1. **Given** a GitHub Actions workflow, **When** a developer follows `docs/guides/ci-integration.md`, **Then** they have a working CI step that runs `gaze crap` with threshold enforcement.
2. **Given** a CI pipeline that generates coverage profiles, **When** a developer follows the guide, **Then** they use `--coverprofile` to avoid double test runs.
3. **Given** a team that wants AI-powered reports, **When** they follow `docs/guides/ai-reports.md`, **Then** they have `gaze report --ai=opencode` producing markdown summaries in GitHub Step Summary.

---

### User Story 4 - Developer Looks Up CLI Details (Priority: P2)

A user knows which Gaze command they want but needs to check the exact flags, output format, or behavior edge cases. They navigate to a per-command reference page and find every flag documented with its type, default, description, and relationship to `.gaze.yaml` configuration — plus example invocations and sample output.

**Why this priority**: Reference documentation is the most-visited doc type after initial onboarding. It ranks P2 because users need conceptual grounding first.

**Independent Test**: Can be fully tested by selecting any command, reading its reference page, and verifying that every flag listed in `gaze <cmd> --help` appears in the docs with accurate descriptions.

**Acceptance Scenarios**:

1. **Given** any Gaze subcommand, **When** a user reads its reference page under `docs/reference/cli/`, **Then** every flag from `--help` is documented with type, default, and description.
2. **Given** a user who needs JSON output, **When** they read `docs/reference/json-schemas.md`, **Then** they find the schema and example output for each command that supports `--format=json`.
3. **Given** a user who wants to customize classification, **When** they read `docs/reference/configuration.md`, **Then** they can write a `.gaze.yaml` file with all available options and understand their effects.

---

### User Story 5 - Contributor Understands the Architecture (Priority: P3)

A developer wants to contribute to Gaze — fixing a bug, adding a new side effect type, or building a new classification signal. They need an architecture overview that explains the package structure, data flow from CLI entry through analysis to output, and the extension points. They also need coding conventions, testing patterns, and the spec-first workflow so their contribution matches project standards.

**Why this priority**: Contributor documentation enables community growth but serves a smaller audience than user-facing docs.

**Independent Test**: Can be fully tested by having a new contributor read the architecture docs and then successfully add a stub side effect type with tests, following the documented patterns.

**Acceptance Scenarios**:

1. **Given** a new contributor, **When** they read `docs/architecture/overview.md`, **Then** they can draw the data flow from CLI invocation through analysis, classification, and scoring to output.
2. **Given** a developer who wants to add a side effect type, **When** they read `docs/architecture/extending.md`, **Then** they know which files to modify and what tests to write.
3. **Given** a first-time contributor, **When** they read `docs/architecture/contributing.md`, **Then** they can set up a dev environment, run tests, and create a spec-compliant PR.

---

### User Story 6 - Language Porter Understands the Contracts (Priority: P3)

A team wants to build "Gaze for Python" (or Rust, TypeScript, etc.). They need to understand which aspects of Gaze are language-agnostic contracts that a port must honor (effect taxonomy, tier definitions, classification rules, scoring formulas, output schemas) versus Go-specific implementation details (AST visitors, SSA builders). The porting guide gives them a complete list of behavioral requirements without prescribing how to implement them.

**Why this priority**: Language portability is a future goal. The v1 porting section defines contracts and requirements but defers conformance test suites to a future spec.

**Independent Test**: Can be fully tested by having a developer read only the porting section and then listing all behavioral contracts a port must satisfy, with no ambiguity about what is required vs optional.

**Acceptance Scenarios**:

1. **Given** a developer building Gaze-for-Python, **When** they read `docs/porting/contracts.md`, **Then** they can enumerate all 37 side effect types, their tier assignments, and the classification confidence formula.
2. **Given** a language porter, **When** they read `docs/porting/requirements.md`, **Then** they have a checklist of capabilities (effect detection, classification, CRAP scoring, quality assessment) with clear required/optional labels.
3. **Given** a porter implementing scoring, **When** they read `docs/porting/taxonomy-reference.md`, **Then** they can reproduce the CRAP, GazeCRAP, and quadrant formulas exactly.

---

### Edge Cases

- What happens when a concept doc references a CLI command — does it link to the reference page or inline the usage? (Link to reference; never duplicate flag tables.)
- How does documentation handle features that are defined in taxonomy but not yet detected (P3/P4 effects)? (Document them as "defined, detection not yet implemented" with their tier and type.)
- What happens when README content overlaps with docs/ content? (README stays concise and links into docs/ for depth. No duplication of flag tables, formulas, or extended explanations.)
- How are docs versioned across Gaze releases? (Out of scope for v1. Docs track the `main` branch. Versioned docs are deferred to a future spec.)

## Requirements *(mandatory)*

### Functional Requirements

#### Documentation Structure

- **FR-001**: Documentation MUST be organized in a `docs/` directory at the repository root using the Diataxis framework: tutorials (getting-started), explanations (concepts), reference (reference), and how-to guides (guides).
- **FR-002**: Documentation MUST include a `docs/index.md` that serves as a navigation hub, routing readers to the appropriate section based on their role (power user, developer, contributor, language porter).
- **FR-003**: Each documentation file MUST be self-contained plain Markdown with no build step required — readable directly on GitHub, in any Markdown viewer, or in a text editor.

#### Getting Started (Tutorials)

- **FR-004**: Documentation MUST include an installation guide covering Homebrew, `go install`, and build-from-source methods with platform-specific notes.
- **FR-005**: Documentation MUST include a quickstart tutorial that walks a user from first install through `analyze`, `crap`, and `quality` on their own project in a single guided flow.
- **FR-006**: Documentation MUST include a conceptual introduction explaining why line coverage is insufficient and what contract-level analysis adds — establishing the mental model before tool usage.

#### Concepts (Explanations)

- **FR-007**: Documentation MUST explain the side effect taxonomy: all five tiers (P0-P4), their definitions, and all 37 effect types with descriptions and examples.
- **FR-008**: Documentation MUST explain the classification model: the five signal analyzers (interface, visibility, caller, naming, godoc), confidence scoring, tier-based boosts, and the three classification labels with their threshold semantics.
- **FR-009**: Documentation MUST explain the scoring system: CRAP formula, GazeCRAP formula, the four quadrants, fix strategies, CRAPload, and GazeCRAPload — with worked examples.
- **FR-010**: Documentation MUST explain the quality assessment model: test-target pairing, assertion detection, the four mechanical mapping passes (direct identity, indirect root, helper bridge, inline call), contract coverage, and over-specification.
- **FR-011**: Documentation MUST explain the analysis pipeline: how AST and SSA analysis work together, what each phase detects, and how they compose into the full analysis result.

#### Reference

- **FR-012**: Documentation MUST include one reference page per CLI subcommand (8 total: analyze, crap, quality, report, self-check, docscan, schema, init), each listing every flag with its type, default value, description, and interaction with `.gaze.yaml` configuration.
- **FR-013**: Documentation MUST include a configuration reference for `.gaze.yaml` documenting all available keys, their types, defaults, and validation rules.
- **FR-014**: Documentation MUST include JSON output documentation with schema references and annotated example output for each command that supports `--format=json`.
- **FR-015**: Documentation MUST include a glossary defining all domain-specific terms (CRAPload, GazeCRAP, quadrant names, contract coverage, over-specification, assertion mapping, fix strategy labels, tier names, classification labels, signal names).

#### Guides (How-To)

- **FR-016**: Documentation MUST include a CI integration guide with copy-pasteable workflow YAML for GitHub Actions, covering coverage profile generation, threshold enforcement, and AI report output.
- **FR-017**: Documentation MUST include an AI report guide covering adapter setup for each supported AI backend (Claude, Gemini, Ollama, OpenCode), model configuration, and Step Summary integration.
- **FR-018**: Documentation MUST include an OpenCode integration guide covering `gaze init`, the scaffolded agent/command files, and how to use the `/gaze` command.
- **FR-019**: Documentation MUST include a practical score improvement guide organized by fix strategy (decompose, add_tests, add_assertions, decompose_and_test) with concrete before/after examples.

#### Architecture (Developer)

- **FR-020**: Documentation MUST include an architecture overview showing the package dependency graph, data flow from CLI entry through analysis to output, and the role of each `internal/` package.
- **FR-021**: Documentation MUST include a contributing guide covering dev environment setup, build/test commands, coding conventions, testing patterns, and the spec-first development workflow.
- **FR-022**: Documentation MUST include an extension guide explaining how to add new side effect types, new classification signals, new output formats, and new AI adapters.

#### Porting (Future Languages)

- **FR-023**: Documentation MUST include a contracts document that extracts all language-agnostic behavioral contracts: the effect taxonomy, tier definitions, classification rules, and scoring formulas — clearly separated from Go-specific implementation details.
- **FR-024**: Documentation MUST include a requirements document listing all capabilities a Gaze port must implement, with each requirement labeled as required or optional.
- **FR-025**: Documentation MUST include a taxonomy reference with the canonical list of effect types, tier assignments, and scoring formulas in a format suitable for mechanical extraction.

#### Cross-Cutting

- **FR-026**: The README MUST be updated to link into `docs/` for detailed content, removing any duplicated extended explanations while preserving its role as a concise project overview.
- **FR-027**: Every concept doc that references a CLI command MUST link to the corresponding reference page rather than duplicating flag tables.
- **FR-028**: Documentation MUST clearly mark features that are defined but not yet implemented (P3/P4 effect types) as "defined, detection not yet implemented."

### Key Entities

- **Documentation Page**: A single Markdown file covering one topic. Has a title, audience indicator, and belongs to one Diataxis category (tutorial, explanation, reference, how-to).
- **Navigation Index**: The `docs/index.md` file that maps audiences to entry points and provides a complete table of contents.
- **Behavioral Contract**: A language-agnostic rule that any Gaze implementation must honor — an effect type, a scoring formula, a classification rule, or an output schema requirement. Lives in the porting section.

### Dependencies

- Spec 001 (side-effect-detection): Source for effect type definitions and detection logic documentation.
- Spec 002 (contract-classification): Source for classification signal system documentation.
- Spec 003 (test-quality-metrics): Source for quality assessment model documentation.
- Spec 004 (composite-metrics): Source for CRAP/GazeCRAP scoring formulas.
- Spec 018 (ci-report): Source for AI report pipeline and adapter documentation.
- Spec 020 (report-coverprofile): Source for `--coverprofile` flag documentation.
- Existing README.md: Source content to be refactored (not duplicated) into docs/.
- Existing GoDoc comments: Authoritative source for exported API descriptions — docs reference these, not duplicate them.

### Assumptions

- No static site generator is required for v1. Plain Markdown files readable on GitHub are sufficient.
- Documentation tracks the `main` branch. Multi-version documentation is out of scope.
- The porting section defines contracts and requirements but does not include a conformance test suite — that is a future spec.
- Documentation follows GitHub-flavored Markdown conventions (tables, fenced code blocks, task lists).
- The audience for porting docs has strong software engineering skills and can translate behavioral contracts into their target language without hand-holding.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can complete the quickstart tutorial and produce meaningful output from `gaze analyze`, `gaze crap`, and `gaze quality` on their own Go project within 10 minutes of starting the tutorial.
- **SC-002**: Every flag listed in `gaze <subcommand> --help` for all 8 subcommands has a corresponding entry in the relevant `docs/reference/cli/` page with type, default, and description.
- **SC-003**: All 37 defined side effect types appear in `docs/concepts/side-effects.md` with their tier assignment, description, and detection status (implemented vs defined-only).
- **SC-004**: The CRAP formula, GazeCRAP formula, quadrant boundaries, and fix strategy rules in `docs/concepts/scoring.md` produce identical results to the implementation when applied to the same inputs.
- **SC-005**: The CI integration guide produces a working GitHub Actions workflow that enforces CRAPload thresholds when followed step-by-step on a test repository.
- **SC-006**: A developer who reads only the architecture and contributing docs can set up a dev environment, run all tests, and identify the correct package to modify for a given change type.
- **SC-007**: A language porter who reads only the porting section can enumerate all behavioral contracts (effect types, tier rules, scoring formulas, classification rules) without referencing Go source code.
- **SC-008**: The README links to docs/ for all topics that have dedicated documentation pages, with no extended explanations duplicated between README and docs/.
- **SC-009**: All documentation files are valid Markdown that renders correctly on GitHub without any build step, tooling, or preprocessing.
- **SC-010**: The glossary covers all domain-specific terms used across the documentation, with each term defined in one canonical location and referenced (not redefined) elsewhere.
