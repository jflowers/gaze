# Data Model: Project Documentation

**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md) | **Date**: 2026-04-08

This spec introduces no Go types, packages, or code entities. The "data model" describes the documentation page taxonomy — the organizational structure of Markdown files that constitute the deliverable.

## Documentation Page Taxonomy

### Entity: Documentation Page

A single Markdown file covering one topic. Every page in `docs/` conforms to this structure:

| Property | Type | Description |
|----------|------|-------------|
| **Title** | H1 heading | The page's primary heading (first line, `# Title`) |
| **Audience** | Implicit | Determined by section: power user (concepts), new user (getting-started), developer (guides, reference), contributor (architecture), porter (porting) |
| **Category** | Directory | One of: `getting-started`, `concepts`, `reference`, `guides`, `architecture`, `porting` |
| **Diataxis Type** | Mapped | Tutorial (getting-started), Explanation (concepts), Reference (reference), How-To (guides), N/A (architecture, porting) |
| **Primary Sources** | Metadata | Listed in plan.md Content Source Mapping — the authoritative code/spec files this page derives from |
| **Cross-References** | Relative links | Links to other docs pages using relative Markdown paths |
| **Format** | Constraint | GitHub-flavored Markdown, self-contained, no build step (FR-003) |

### Entity: Navigation Index

The `docs/index.md` file that maps audiences to entry points:

| Property | Type | Description |
|----------|------|-------------|
| **Audience Routing** | Section per audience | Sections for: "New to Gaze?", "Understanding the Model", "CLI Reference", "Integrating with CI", "Contributing", "Porting to Other Languages" |
| **Table of Contents** | Nested list | Complete listing of all pages organized by section |
| **Entry Points** | Links | Direct links to the recommended starting page for each audience |

### Entity: Behavioral Contract (Porting Section)

A language-agnostic rule documented in the porting section. Not a code entity — a documentation entity that describes what any Gaze implementation must honor:

| Property | Type | Description |
|----------|------|-------------|
| **Contract ID** | Implicit | Derived from the rule's domain (e.g., "EFFECT-001: ReturnValue detection") |
| **Domain** | Category | One of: Effect Taxonomy, Classification Rules, Scoring Formulas, Output Schemas |
| **Requirement Level** | Label | `Required` or `Optional` per FR-024 |
| **Specification** | Prose + formula | The behavioral rule in language-agnostic terms |
| **Go Reference** | Link | Where this contract is implemented in the Go codebase (for verification, not for porting) |

## Section Taxonomy

### getting-started/ (Tutorials — Diataxis)

**Audience**: New users who have never used Gaze.
**Reading Pattern**: Sequential — installation → concepts → quickstart.
**Content Type**: Step-by-step guided instructions with expected output.

| File | FR | Content Summary |
|------|-----|----------------|
| `installation.md` | FR-004 | Homebrew, `go install`, build-from-source, platform notes |
| `concepts.md` | FR-006 | Why line coverage is insufficient, what contract coverage adds |
| `quickstart.md` | FR-005 | Guided walkthrough: analyze → crap → quality on user's own project |

### concepts/ (Explanations — Diataxis)

**Audience**: Power users who want to understand the model.
**Reading Pattern**: Topic-based — users jump to the concept they need.
**Content Type**: Explanatory prose with diagrams, tables, and worked examples.

| File | FR | Content Summary |
|------|-----|----------------|
| `side-effects.md` | FR-007 | All 37 effect types, 5 tiers, definitions, detection status |
| `classification.md` | FR-008 | 5 signal analyzers, confidence scoring, tier boosts, 3 labels |
| `scoring.md` | FR-009 | CRAP formula, GazeCRAP formula, 4 quadrants, fix strategies, CRAPload |
| `quality.md` | FR-010 | Test-target pairing, assertion detection, 4 mapping passes, contract coverage |
| `analysis-pipeline.md` | FR-011 | AST + SSA dual analysis, phase composition, data flow |

### reference/ (Reference — Diataxis)

**Audience**: Users who know what they want and need to look up details.
**Reading Pattern**: Lookup — users search for a specific flag, config key, or term.
**Content Type**: Tables, structured lists, complete enumerations.

| File | FR | Content Summary |
|------|-----|----------------|
| `cli/analyze.md` | FR-012 | Every flag with type, default, description, config interaction |
| `cli/crap.md` | FR-012 | (same pattern for each of 8 subcommands) |
| `cli/quality.md` | FR-012 | |
| `cli/report.md` | FR-012 | |
| `cli/self-check.md` | FR-012 | |
| `cli/docscan.md` | FR-012 | |
| `cli/schema.md` | FR-012 | |
| `cli/init.md` | FR-012 | |
| `configuration.md` | FR-013 | All `.gaze.yaml` keys, types, defaults, validation |
| `json-schemas.md` | FR-014 | Schema references, annotated example output per command |
| `glossary.md` | FR-015 | All domain terms, canonical definitions |

### guides/ (How-To — Diataxis)

**Audience**: Developers with a specific goal (CI setup, score improvement).
**Reading Pattern**: Task-oriented — follow steps to achieve a concrete outcome.
**Content Type**: Step-by-step instructions with copy-pasteable code/YAML.

| File | FR | Content Summary |
|------|-----|----------------|
| `ci-integration.md` | FR-016 | GitHub Actions YAML, coverage profiles, thresholds |
| `ai-reports.md` | FR-017 | Claude/Gemini/Ollama/OpenCode adapter setup, model config |
| `opencode-integration.md` | FR-018 | `gaze init`, scaffolded files, `/gaze` command usage |
| `improving-scores.md` | FR-019 | Fix strategies with before/after examples |

### architecture/ (Extension — not Diataxis)

**Audience**: Contributors who want to modify or extend Gaze.
**Reading Pattern**: Onboarding — read overview, then drill into specific areas.
**Content Type**: Diagrams, package descriptions, coding patterns, workflow guides.

| File | FR | Content Summary |
|------|-----|----------------|
| `overview.md` | FR-020 | Package dependency graph, data flow, package roles |
| `contributing.md` | FR-021 | Dev setup, build/test, conventions, spec-first workflow |
| `extending.md` | FR-022 | Adding effect types, signals, output formats, AI adapters |

### porting/ (Extension — not Diataxis)

**Audience**: Teams building Gaze for other languages.
**Reading Pattern**: Contract extraction — enumerate all behavioral requirements.
**Content Type**: Structured tables, formulas, checklists with required/optional labels.

| File | FR | Content Summary |
|------|-----|----------------|
| `contracts.md` | FR-023 | All language-agnostic behavioral contracts |
| `requirements.md` | FR-024 | Capability checklist with required/optional labels |
| `taxonomy-reference.md` | FR-025 | Canonical effect types, tiers, formulas for mechanical extraction |

## Cross-Reference Rules

1. **Concept → Reference**: Concept pages link to CLI reference pages when mentioning commands. Never duplicate flag tables.
2. **Guide → Concept + Reference**: Guides link to concepts for background and reference for details.
3. **Architecture → Concept**: Architecture pages link to concepts for domain model explanations.
4. **Porting → Concept**: Porting pages link to concepts for context but are self-contained for behavioral contracts.
5. **Getting Started → Concept**: Quickstart links to concepts for deeper understanding after the tutorial.
6. **Glossary ← All**: All pages link to glossary on first use of domain terms.
7. **README → docs/**: README links into docs/ for all extended content. No duplication.

## No Code Entities

This spec introduces:
- ❌ No new Go packages
- ❌ No new Go types or interfaces
- ❌ No new functions or methods
- ❌ No new test files
- ❌ No changes to existing Go code
- ❌ No changes to existing test code
- ❌ No new dependencies

The only file modifications are:
- ✅ New Markdown files in `docs/`
- ✅ Modified `README.md` (refactored to link into `docs/`)
