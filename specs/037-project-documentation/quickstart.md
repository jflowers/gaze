# Quickstart: Adding Documentation Pages

**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md) | **Date**: 2026-04-08

This guide explains how a contributor adds a new documentation page to the correct section of `docs/`.

## Step 1: Identify the Section

Determine which Diataxis category your content belongs to:

| If you're writing... | Put it in... | Example |
|---------------------|-------------|---------|
| A step-by-step tutorial for beginners | `docs/getting-started/` | "First Analysis Walkthrough" |
| An explanation of how something works | `docs/concepts/` | "How Classification Signals Work" |
| A lookup table of flags, options, or terms | `docs/reference/` | "gaze report CLI Reference" |
| A task-oriented recipe for a specific goal | `docs/guides/` | "Setting Up Gaze in GitLab CI" |
| Internal architecture or contributor info | `docs/architecture/` | "Adding a New Output Format" |
| Language-agnostic behavioral contracts | `docs/porting/` | "Scoring Formula Contracts" |

**Rule of thumb**: Tutorials teach, explanations illuminate, references list, guides solve. If your content doesn't fit, it probably belongs in concepts (explanation).

## Step 2: Create the File

Create a new `.md` file in the appropriate directory:

```bash
# Example: adding a new guide
touch docs/guides/gitlab-ci.md
```

**Naming conventions**:
- Use kebab-case: `ci-integration.md`, not `CI_Integration.md`
- Be descriptive: `improving-scores.md`, not `scores.md`
- CLI reference pages match the subcommand: `docs/reference/cli/analyze.md`

## Step 3: Write the Content

Every documentation page follows this structure:

```markdown
# Page Title

Brief introduction (1-2 sentences) explaining what this page covers and who it's for.

## Section Heading

Content organized by logical sections...

## See Also

- [Related Page](../section/page.md) — brief description
```

**Rules**:
1. **Start with H1** (`#`): One H1 per page, matching the page's topic.
2. **Self-contained**: The page must be readable without any build step — plain Markdown on GitHub.
3. **Link, don't duplicate**: If another page covers a topic, link to it. Never copy flag tables or formula definitions.
4. **First use of domain terms**: Link to the glossary — `[CRAPload](../reference/glossary.md#crapload)`.
5. **Mark unimplemented features**: If referencing P3/P4 effects, add: *"Defined — detection not yet implemented."*
6. **Use relative links**: `[scoring](../concepts/scoring.md)`, never absolute GitHub URLs.

## Step 4: Identify Primary Sources

Before writing content, identify the authoritative source for your page's facts:

- **CLI flags** → Run `gaze <cmd> --help` and transcribe
- **Effect types** → Read `internal/taxonomy/types.go`
- **Formulas** → Read `internal/crap/crap.go`
- **Classification signals** → Read `internal/classify/score.go`
- **Config options** → Read `internal/config/config.go`
- **Spec rationale** → Read `specs/NNN-feature/spec.md`

**Never write from memory.** Always verify against the source code.

## Step 5: Update Cross-References

After creating your page:

1. **Add to `docs/index.md`**: Add your page to the table of contents in the appropriate section.
2. **Link from related pages**: If your page is referenced by existing docs, add links from those pages.
3. **Link to glossary**: Ensure any domain terms used are linked on first use.

## Step 6: Verify

Before submitting:

- [ ] File is in the correct `docs/` subdirectory
- [ ] Content is derived from authoritative sources (source code, `--help`, specs)
- [ ] No flag tables or formulas are duplicated from other pages
- [ ] Domain terms link to the glossary on first use
- [ ] All cross-reference links use relative paths and resolve correctly
- [ ] Markdown renders correctly on GitHub (check with preview)
- [ ] Page is listed in `docs/index.md`
- [ ] Unimplemented features are clearly marked

## Example: Adding a CLI Reference Page

```bash
# 1. Create the file
touch docs/reference/cli/new-command.md

# 2. Get the authoritative flag list
gaze new-command --help

# 3. Write the page with flag table derived from --help output

# 4. Add to docs/index.md under Reference > CLI
# 5. Link from any guide that mentions this command
# 6. Verify rendering on GitHub
```
