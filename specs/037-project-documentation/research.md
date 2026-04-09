# Research: Project Documentation

**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md) | **Date**: 2026-04-08

## R1: Documentation Structure — Diataxis Mapping

**Question**: How should Gaze's documentation map to the Diataxis framework, and what additional sections are needed beyond the four canonical categories?

**Decision**: Use six top-level sections, four of which map directly to Diataxis and two that extend it for project-specific needs:

| Diataxis Category | Gaze Section | Directory | Purpose |
|-------------------|-------------|-----------|---------|
| Tutorials | Getting Started | `getting-started/` | Guided learning for new users |
| Explanations | Concepts | `concepts/` | Deep understanding of the model |
| Reference | Reference | `reference/` | Lookup tables for CLI, config, schemas |
| How-To Guides | Guides | `guides/` | Task-oriented recipes for specific goals |
| *(extension)* | Architecture | `architecture/` | Contributor-facing internals |
| *(extension)* | Porting | `porting/` | Language-agnostic behavioral contracts |

**Rationale**: Architecture and porting don't fit cleanly into Diataxis categories. Architecture is partly reference (package structure) and partly explanation (data flow), but its audience (contributors) is distinct enough to warrant separation. Porting is a unique category — it extracts behavioral contracts for reimplementation in other languages, which is neither a tutorial, explanation, reference, nor how-to.

**Alternatives Considered**:
- Merge architecture into reference → Rejected: contributor docs have a different audience and reading pattern than CLI reference lookups.
- Merge porting into architecture → Rejected: porters need language-agnostic contracts without Go implementation details. Architecture docs necessarily include Go-specific patterns.
- Use a single flat `docs/` directory → Rejected: 25+ files need categorical organization for discoverability.

## R2: Content Sourcing Strategy

**Question**: Where does the authoritative content for each documentation section come from, and how do we prevent drift between docs and implementation?

**Decision**: Every documentation page has a declared "primary source" in the plan's Content Source Mapping table. During implementation, content MUST be extracted from these sources — never written from memory or general knowledge. The primary sources are:

1. **Source code** (`internal/*/`): Authoritative for types, formulas, constants, and behavior. Read the actual Go code, not summaries.
2. **CLI `--help` output**: Authoritative for flag names, types, defaults, and descriptions. Run `gaze <cmd> --help` and transcribe.
3. **Spec artifacts** (`specs/*/spec.md`): Authoritative for design rationale and feature scope.
4. **README.md**: Authoritative for current user-facing descriptions (to be refactored, not duplicated).
5. **`.gaze.yaml`** and `internal/config/config.go`: Authoritative for configuration options.

**Drift Prevention**: The plan's Content Source Mapping table serves as a traceability matrix. During review, each doc page can be verified against its declared sources. SC-002 (flags match `--help`), SC-003 (effect types match `types.go`), and SC-004 (formulas match `crap.go`) are concrete drift-detection criteria.

## R3: Cross-Referencing Strategy

**Question**: How should documentation pages reference each other, and how do we handle the boundary between README and docs/?

**Decision**: Three rules govern cross-references:

1. **Relative Markdown links**: All cross-references use relative paths (e.g., `[scoring](../concepts/scoring.md)`). No absolute URLs to the GitHub repository — this ensures docs work in any Markdown viewer, not just GitHub.

2. **Link, don't duplicate**: When a concept doc mentions a CLI command, it links to the reference page (e.g., "See [gaze crap reference](../reference/cli/crap.md) for all flags"). Flag tables are NEVER duplicated outside reference pages (FR-027).

3. **README as gateway**: README stays concise — a project overview with installation, quick examples, and links into `docs/` for depth. Extended explanations (effect taxonomy tables, formula derivations, detailed flag tables) move to `docs/` and README links to them. The README retains enough context to be useful standalone but defers to `docs/` for anything longer than a paragraph (FR-026).

**Alternatives Considered**:
- Absolute GitHub URLs → Rejected: breaks offline/local viewing and couples docs to GitHub hosting.
- Inline flag tables in concept docs → Rejected: creates duplication that drifts. FR-027 explicitly prohibits this.
- Remove README content entirely → Rejected: README is the first thing visitors see on GitHub. It must be self-sufficient for a quick overview.

## R4: Unimplemented Feature Documentation

**Question**: How should docs handle P3/P4 effect types that are defined in the taxonomy but not yet detected?

**Decision**: Per FR-028 and the spec's edge case resolution:

- P3/P4 effect types appear in `docs/concepts/side-effects.md` with their tier, type name, and description.
- Each unimplemented type is marked with a clear label: **"Defined — detection not yet implemented"**.
- The porting guide (`docs/porting/taxonomy-reference.md`) includes ALL 37 types with a `Status` column: `Implemented` or `Defined`.
- No documentation claims that Gaze detects P3/P4 effects. The distinction between "defined in taxonomy" and "detected by analysis" is explicit.

**Rationale**: Porters need the complete taxonomy to build a conformant implementation. Users need to know what Gaze can and cannot detect today. Omitting defined-but-unimplemented types would create a false impression of completeness; including them without status labels would create a false impression of capability.

## R5: Documentation Versioning

**Question**: How are docs versioned across Gaze releases?

**Decision**: Per the spec's edge case resolution: **out of scope for v1**. Documentation tracks the `main` branch. There is no multi-version documentation system, no version selector, and no archived doc snapshots.

**Rationale**: Gaze is pre-1.0. The documentation structure and content are still evolving. Adding versioning infrastructure now would be premature optimization. A future spec can introduce versioned docs (e.g., via a static site generator with version support) when the project reaches stability.

## R6: Glossary as Single Source of Truth

**Question**: How do we ensure domain-specific terms are defined consistently across 25+ documentation pages?

**Decision**: `docs/reference/glossary.md` is the canonical definition location for all domain terms (FR-015, SC-010). Other pages that use these terms:

1. **First use in a page**: Link to the glossary entry (e.g., "[CRAPload](../reference/glossary.md#crapload)").
2. **Subsequent uses**: Use the term without linking (readers can refer back to the glossary).
3. **Never redefine**: No page other than the glossary provides a definition. Pages may provide *context* (e.g., "CRAPload, which measures the count of high-risk functions, is the primary CI gate metric") but not a *definition*.

**Terms to include** (derived from spec FR-015):
CRAPload, GazeCRAPload, GazeCRAP, CRAP, quadrant names (Q1 Safe, Q2 Complex But Tested, Q3 Simple But Underspecified, Q4 Dangerous), contract coverage, over-specification, assertion mapping, fix strategy labels (decompose, add_tests, add_assertions, decompose_and_test), tier names (P0–P4), classification labels (contractual, incidental, ambiguous), signal names (interface, visibility, caller, naming, godoc), side effect, behavioral contract, Diataxis.

## R7: File Count and Scope Validation

**Question**: Is the planned file count (~25 Markdown files + README modification) appropriate for the scope?

**Decision**: Yes. The file count maps 1:1 to functional requirements:

| Section | Files | FRs Covered |
|---------|-------|-------------|
| Root | 1 (`index.md`) | FR-002 |
| Getting Started | 3 | FR-004, FR-005, FR-006 |
| Concepts | 5 | FR-007, FR-008, FR-009, FR-010, FR-011 |
| Reference | 11 (8 CLI + 3) | FR-012, FR-013, FR-014, FR-015 |
| Guides | 4 | FR-016, FR-017, FR-018, FR-019 |
| Architecture | 3 | FR-020, FR-021, FR-022 |
| Porting | 3 | FR-023, FR-024, FR-025 |
| README | 1 (modified) | FR-026 |
| **Total** | **31** | **28 FRs** |

The reference/cli/ section accounts for the higher count (8 subcommands × 1 page each). No file is gratuitous — each maps to a specific FR.
