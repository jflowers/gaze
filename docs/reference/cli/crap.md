# gaze crap

Compute [CRAP](../../concepts/scoring.md) (Change Risk Anti-Patterns) scores by combining cyclomatic complexity with test coverage. Reports per-function CRAP scores and the project's [CRAPload](../glossary.md) (count of functions above the threshold).

When contract coverage data is available (via the integrated quality pipeline), also computes [GazeCRAP](../glossary.md) scores and [quadrant](../../concepts/scoring.md) classifications.

## Synopsis

```
gaze crap [packages...] [flags]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `packages` | Yes (1+) | One or more Go package patterns (e.g., `./...`, `./internal/crap`, `./cmd/...`) |

At least one package pattern is required. Use `./...` to analyze the entire module.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | `string` | `text` | Output format: `text` or `json` |
| `--coverprofile` | `string` | `""` (generate via `go test`) | Path to a pre-generated Go coverage profile. When omitted, Gaze runs `go test -coverprofile` automatically. |
| `--crap-threshold` | `float64` | `15` | CRAP score threshold for flagging functions. Functions at or above this score are counted in the CRAPload. |
| `--gaze-crap-threshold` | `float64` | `15` | GazeCRAP score threshold. Used when contract coverage is available to compute GazeCRAPload. |
| `--max-crapload` | `int` | `0` (no limit) | CI gate: fail with non-zero exit code if CRAPload exceeds this value. |
| `--max-gaze-crapload` | `int` | `0` (no limit) | CI gate: fail with non-zero exit code if GazeCRAPload exceeds this value. |
| `--ai-mapper` | `string` | `""` | AI backend for assertion mapping fallback: `claude`, `gemini`, `ollama`, or `opencode`. When set, unmapped assertions are sent to the AI for semantic matching. |
| `--ai-mapper-model` | `string` | `""` | Model name for the AI mapper. Required when `--ai-mapper=ollama`. |

## Configuration Interaction

The `gaze crap` command does not directly read `.gaze.yaml`. However, the integrated quality pipeline (which provides contract coverage for GazeCRAP) uses classification thresholds from the config file internally.

The `--coverprofile` flag is the primary configuration point — providing a pre-generated profile avoids running `go test` again (useful in CI where tests have already run).

## Examples

### Basic CRAP analysis

```bash
gaze crap ./...
```

```
CRAP Scores
═══════════════════════════════════════════════════════════════

  Function                    Complexity  Coverage  CRAP
  ─────────────────────────── ──────────  ────────  ────
  runAnalyze                  8           85.0%     10.2
  loadConfig                  12          92.0%     13.8
  ...

Summary
  Total functions: 142
  Average complexity: 4.2
  Average line coverage: 78.5%
  CRAPload: 5 (threshold: 15.0)
```

### CI quality gate with thresholds

```bash
gaze crap ./... --max-crapload=10 --max-gaze-crapload=5
```

Exits with code 1 if CRAPload exceeds 10 or GazeCRAPload exceeds 5. Prints a CI summary line to stderr:

```
CRAPload: 5/10 (PASS) | GazeCRAPload: 3/5 (PASS)
```

### Using a pre-generated coverage profile

```bash
# Generate coverage during test run
go test -race -count=1 -coverprofile=coverage.out ./...

# Analyze without re-running tests
gaze crap ./... --coverprofile=coverage.out
```

### JSON output

```bash
gaze crap ./... --format=json | jq '.summary.crapload'
```

See [JSON Schemas](../json-schemas.md) for the full output structure.

## See Also

- [Scoring](../../concepts/scoring.md) — CRAP formula, GazeCRAP, quadrants, and fix strategies
- [CI Integration](../../guides/ci-integration.md) — setting up quality gates in GitHub Actions
- [Improving Scores](../../guides/improving-scores.md) — remediation strategies by fix type
- [JSON Schemas](../json-schemas.md) — schema reference for `--format=json` output
- [`gaze quality`](quality.md) — contract coverage analysis (feeds into GazeCRAP)
- [`gaze report`](report.md) — AI-formatted quality report combining all analyses
