# gaze init

Scaffold OpenCode agent and command files into the current directory. After running this command, you can use the `/gaze` command in OpenCode to generate AI-powered quality reports.

## Synopsis

```
gaze init [flags]
```

## Arguments

This command takes no positional arguments. It always operates on the current working directory.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | `bool` | `false` | Overwrite existing files. Without this flag, existing user-owned files are skipped. |

## Scaffolded Files

The `init` command creates 9 files with the following directory structure:

```
.opencode/
├── agents/
│   ├── gaze-reporter.md          # AI agent prompt for quality reporting (user-owned)
│   ├── gaze-test-generator.md    # Test generation agent
│   └── divisor-gaze-testing.md   # Test-quality reviewer persona for review councils
├── commands/
│   ├── gaze.md                   # /gaze command definition (user-owned)
│   ├── gaze-fix.md               # /gaze fix command definition
│   └── speckit.testreview.md     # /speckit.testreview command definition
├── dcp.jsonc                     # DCP context-preservation configuration
└── references/
    ├── example-report.md         # Canonical example output for the reporter
    └── doc-scoring-model.md      # Document-enhanced classification scoring model
```

### File Ownership Model

Gaze uses a mixed ownership model for scaffolded files:

- **User-owned files** (`agents/gaze-reporter.md`, `commands/gaze.md`): Created once, never overwritten on subsequent `gaze init` runs (unless `--force` is used). Users can customize these files.
- **Tool-owned files** (everything under `references/`, plus `agents/gaze-test-generator.md`, `agents/divisor-gaze-testing.md`, `commands/gaze-fix.md`, `commands/speckit.testreview.md`, and `dcp.jsonc`): Updated automatically on `gaze init` if the content has changed. These files are maintained by Gaze and should not be manually edited.

### Divisor Review-Config Registration

When run in a project that already uses unbound-force tooling (a `.uf/` directory exists), `gaze init` also registers the `divisor-gaze-testing` persona in the existing review configuration: it appends the reviewer entry to `.uf/reviewer-capabilities.yaml` and swaps the generic `divisor-testing` advisor for `divisor-gaze-testing` in the `code`, `spec`, and `feedback` advisor lists of `.uf/review-matrix.yaml` (the `triage` list is untouched). The merge is idempotent, comment-preserving, and never creates `.uf` files — projects without a `.uf/` directory get a notice instead.

`gaze init` also removes a legacy `.opencode/agents/reviewer-testing.md` when its content carries the gaze version marker (`<!-- scaffolded by gaze ... -->`) — the file was superseded by `agents/divisor-gaze-testing.md`. Marker-less (user-customized) copies are left untouched.

## Configuration Interaction

This command does not read `.gaze.yaml`.

## Examples

### Initialize OpenCode integration

```bash
cd /path/to/your/go/project
gaze init
```

```text
Notice: no .uf directory found; skipping Divisor review-config registration.
Gaze never creates .uf files; unbound-force projects register the persona automatically.

Gaze OpenCode integration initialized:
  created: .opencode/agents/divisor-gaze-testing.md
  created: .opencode/agents/gaze-reporter.md
  created: .opencode/agents/gaze-test-generator.md
  created: .opencode/commands/gaze-fix.md
  created: .opencode/commands/gaze.md
  created: .opencode/commands/speckit.testreview.md
  created: .opencode/dcp.jsonc
  created: .opencode/references/doc-scoring-model.md
  created: .opencode/references/example-report.md

Run /gaze for quality reports and /speckit.testreview for testability analysis.
```

### Re-initialize (update tool-owned files)

```bash
gaze init
```

On subsequent runs, user-owned files are skipped and tool-owned files are updated if their content has changed. In projects with a `.uf/` directory, the persona registration is idempotent (already-registered files are not rewritten), and a gaze-marked legacy `.opencode/agents/reviewer-testing.md` orphan is removed with a `removed:` summary line.

### Force overwrite all files

```bash
gaze init --force
```

Overwrites all files, including user-customized agent prompts and commands.

## See Also

- [OpenCode Integration](../../guides/opencode-integration.md) — how to use `/gaze` after initialization
- [`gaze report`](report.md) — the CLI equivalent of the `/gaze` OpenCode command
