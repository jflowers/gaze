# Implementation Plan: Project Documentation

**Branch**: `037-project-documentation` | **Date**: 2026-04-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/037-project-documentation/spec.md`

## Summary

Create comprehensive project documentation in `docs/` organized by the Diataxis framework (tutorials, explanations, reference, how-to guides) plus architecture and porting sections. This is a **documentation-only change** — no production Go code, no test code, no new packages or types. The deliverables are ~25 Markdown files covering getting-started, concepts, reference, guides, architecture, and porting topics, plus a README refactoring to link into `docs/` instead of duplicating extended explanations.

## Technical Context

**Language/Version**: N/A — Markdown files only (no Go code changes)
**Primary Dependencies**: None — plain Markdown, no static site generator, no build step
**Storage**: Filesystem only — `docs/` directory at repository root
**Testing**: Manual verification — documentation accuracy checked against source code, CLI `--help` output, and existing spec artifacts
**Target Platform**: GitHub (rendered Markdown), any Markdown viewer, text editors
**Project Type**: Single project — documentation lives alongside source in the same repository
**Performance Goals**: N/A — static files
**Constraints**: GitHub-flavored Markdown; no external tooling required to read; all files self-contained
**Scale/Scope**: ~25 documentation pages across 6 sections, plus README refactoring

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Accuracy | **PASS** | All documentation content MUST be derived from authoritative sources: source code (`internal/taxonomy/types.go`, `internal/crap/crap.go`, `internal/classify/score.go`, etc.), CLI `--help` output, existing spec artifacts, and the README. Formulas, effect types, tier assignments, and flag tables will be verified against the implementation. No fabricated features. FR-028 explicitly requires marking unimplemented P3/P4 detection as "defined, detection not yet implemented." |
| II. Minimal Assumptions | **PASS** | FR-003 requires plain Markdown with no build step — readable on GitHub, in any viewer, or in a text editor. No static site generator, no custom tooling, no preprocessing. The only assumption is GitHub-flavored Markdown support (tables, fenced code blocks), which is standard. |
| III. Actionable Output | **PASS** | Every documentation page guides users toward concrete actions: the quickstart produces real output (SC-001), the CI guide produces a working workflow (SC-005), the scoring docs include worked examples (SC-004), and the porting guide enumerates all behavioral contracts (SC-007). FR-019 organizes improvement guidance by fix strategy with before/after examples. |
| IV. Testability | **PASS** | Documentation accuracy is verifiable by manual inspection: flag tables can be diffed against `--help` output (SC-002), effect type lists can be diffed against `taxonomy/types.go` (SC-003), formulas can be verified against `crap.go` (SC-004). No programmatic test code is needed because the deliverables are documentation files, not executable code. The spec's success criteria (SC-001 through SC-010) define concrete, measurable verification procedures. |

**Gate Result**: ✅ PASS — All four principles satisfied. Proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/037-project-documentation/
├── plan.md              # This file
├── research.md          # Phase 0 output — doc structure decisions
├── data-model.md        # Phase 1 output — documentation page taxonomy
├── quickstart.md        # Phase 1 output — contributor guide for adding docs
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
docs/
├── index.md                        # Navigation hub (FR-002)
├── getting-started/
│   ├── installation.md             # FR-004
│   ├── quickstart.md               # FR-005
│   └── concepts.md                 # FR-006
├── concepts/
│   ├── side-effects.md             # FR-007
│   ├── classification.md           # FR-008
│   ├── scoring.md                  # FR-009
│   ├── quality.md                  # FR-010
│   └── analysis-pipeline.md        # FR-011
├── reference/
│   ├── cli/
│   │   ├── analyze.md              # FR-012
│   │   ├── crap.md                 # FR-012
│   │   ├── quality.md              # FR-012
│   │   ├── report.md               # FR-012
│   │   ├── self-check.md           # FR-012
│   │   ├── docscan.md              # FR-012
│   │   ├── schema.md               # FR-012
│   │   └── init.md                 # FR-012
│   ├── configuration.md            # FR-013
│   ├── json-schemas.md             # FR-014
│   └── glossary.md                 # FR-015
├── guides/
│   ├── ci-integration.md           # FR-016
│   ├── ai-reports.md               # FR-017
│   ├── opencode-integration.md     # FR-018
│   └── improving-scores.md         # FR-019
├── architecture/
│   ├── overview.md                 # FR-020
│   ├── contributing.md             # FR-021
│   └── extending.md                # FR-022
└── porting/
    ├── contracts.md                # FR-023
    ├── requirements.md             # FR-024
    └── taxonomy-reference.md       # FR-025

README.md                           # FR-026 — refactored to link into docs/
```

**Structure Decision**: Flat Diataxis-based directory structure under `docs/` with one subdirectory per section. CLI reference pages get their own `cli/` subdirectory under `reference/` because there are 8 subcommands. All other sections are flat. This matches the spec's FR requirements 1:1 — every FR maps to exactly one file.

## Content Source Mapping

Each documentation page has authoritative sources that MUST be consulted during implementation:

| Doc Page | Primary Sources |
|----------|----------------|
| `getting-started/installation.md` | README.md §Installation, `.goreleaser.yaml`, `go.mod` |
| `getting-started/quickstart.md` | README.md §Commands, actual CLI output |
| `getting-started/concepts.md` | README.md introduction, spec-001, spec-003 |
| `concepts/side-effects.md` | `internal/taxonomy/types.go` (all 37 types), spec-001 |
| `concepts/classification.md` | `internal/classify/score.go`, `classify.go`, `interface.go`, `visibility.go`, `callers.go`, `naming.go`, `godoc.go`, spec-002 |
| `concepts/scoring.md` | `internal/crap/crap.go` (Formula, ClassifyQuadrant, FixStrategy), spec-004 |
| `concepts/quality.md` | `internal/quality/quality.go`, `assertion.go`, `mapping.go`, `pairing.go`, spec-003 |
| `concepts/analysis-pipeline.md` | `internal/analysis/*.go`, `internal/loader/`, spec-001 |
| `reference/cli/*.md` | `cmd/gaze/main.go` (Cobra command definitions), `gaze <cmd> --help` output |
| `reference/configuration.md` | `internal/config/config.go`, `.gaze.yaml` |
| `reference/json-schemas.md` | `internal/report/schema.go`, actual JSON output |
| `reference/glossary.md` | All concept docs, README, spec artifacts |
| `guides/ci-integration.md` | `.github/workflows/test.yml`, README §CI Integration, spec-018 |
| `guides/ai-reports.md` | `internal/aireport/`, README §`gaze report`, spec-018 |
| `guides/opencode-integration.md` | `internal/scaffold/`, README §OpenCode Integration, spec-005 |
| `guides/improving-scores.md` | `internal/crap/crap.go` (FixStrategy), spec-009 |
| `architecture/overview.md` | README §Architecture, AGENTS.md, package structure |
| `architecture/contributing.md` | AGENTS.md, `.specify/`, spec workflow |
| `architecture/extending.md` | `internal/analysis/`, `internal/classify/`, `internal/aireport/` |
| `porting/contracts.md` | `internal/taxonomy/types.go`, `internal/crap/crap.go`, `internal/classify/score.go` |
| `porting/requirements.md` | All spec artifacts, feature capability matrix |
| `porting/taxonomy-reference.md` | `internal/taxonomy/types.go`, `internal/crap/crap.go` |

## Complexity Tracking

No complexity violations. This is a documentation-only spec with no production code, no new abstractions, and no architectural changes.

## Coverage Strategy

N/A — No production code or test code is introduced by this spec. Documentation accuracy is verified by the success criteria (SC-001 through SC-010), which define manual verification procedures:

- SC-002: Diff CLI reference pages against `gaze <cmd> --help` output
- SC-003: Diff effect type list against `internal/taxonomy/types.go`
- SC-004: Verify formulas against `internal/crap/crap.go`
- SC-009: Validate Markdown rendering on GitHub

## Implementation Phases

### Phase 1: Foundation (Getting Started + Index)
Create `docs/` directory structure, `docs/index.md` navigation hub, and all getting-started pages. These are the entry points for new users and must be complete before other sections can cross-reference them.

**Files**: `docs/index.md`, `docs/getting-started/installation.md`, `docs/getting-started/quickstart.md`, `docs/getting-started/concepts.md`

### Phase 2: Concepts (Explanations)
Create all concept pages. These are the knowledge foundation that reference, guide, and porting pages link into.

**Files**: `docs/concepts/side-effects.md`, `docs/concepts/classification.md`, `docs/concepts/scoring.md`, `docs/concepts/quality.md`, `docs/concepts/analysis-pipeline.md`

### Phase 3: Reference
Create CLI reference pages, configuration reference, JSON schema docs, and glossary. These depend on concepts being defined so they can link back.

**Files**: `docs/reference/cli/*.md` (8 files), `docs/reference/configuration.md`, `docs/reference/json-schemas.md`, `docs/reference/glossary.md`

### Phase 4: Guides (How-To)
Create practical guides. These depend on both concepts and reference being available for cross-linking.

**Files**: `docs/guides/ci-integration.md`, `docs/guides/ai-reports.md`, `docs/guides/opencode-integration.md`, `docs/guides/improving-scores.md`

### Phase 5: Architecture + Porting
Create contributor and porter documentation. These are lower priority (P3) and depend on all other sections being complete for accurate cross-references.

**Files**: `docs/architecture/overview.md`, `docs/architecture/contributing.md`, `docs/architecture/extending.md`, `docs/porting/contracts.md`, `docs/porting/requirements.md`, `docs/porting/taxonomy-reference.md`

### Phase 6: README Refactoring + Final Cross-References
Refactor README to link into `docs/` for detailed content. Update `docs/index.md` with final cross-references. Verify all internal links resolve.

**Files**: `README.md` (modified), `docs/index.md` (updated)

## Post-Design Constitution Re-Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Accuracy | **PASS** | Content source mapping table ensures every doc page is derived from authoritative source code and spec artifacts. No content is fabricated. |
| II. Minimal Assumptions | **PASS** | Plain Markdown, no build step, no tooling dependencies. GitHub-flavored Markdown is the only assumption. |
| III. Actionable Output | **PASS** | Quickstart produces real CLI output. CI guide produces working YAML. Scoring docs include worked examples. Improvement guide organized by fix strategy with before/after. |
| IV. Testability | **PASS** | Success criteria SC-001–SC-010 define concrete verification procedures. No programmatic tests needed for documentation files. |

**Gate Result**: ✅ PASS — Post-design re-check confirms all principles satisfied.
