## Why

External analyzers (e.g. `snake-eyes` for Python) already return side-effect classification labels via the `classify_signals` protocol method, and `ExternalSideEffectAnalyzer` attaches them to each `taxonomy.SideEffect` (issue #246). But that classification is never surfaced to users: `gaze quality --analyzer` reports contract coverage and gap hints, yet the module-wide distribution of *contractual / incidental / ambiguous* effects is invisible in both the JSON and the `/gaze` Full Mode report.

Meanwhile the `/gaze` Full Mode report derives its Classification Summary exclusively from Go-native `gaze analyze --classify`. That command does not register `--analyzer`/`--language` (design D12), so for external-analyzer projects the report falls back to a degraded classification section — wrong or absent counts — even though the data already exists in the `quality --analyzer` output.

This change (1) surfaces a `classification_counts` summary on the quality package summary for external analyzers, and (2) teaches the `gaze-reporter` agent to detect the analyzer/language and source the classification from `classification_counts` for external projects (retaining `analyze --classify` for Go). The JSON schema that validates quality output is updated to declare the new field.

Closes #278.

## What Changes

- Add `ClassificationCounts` (`Contractual`/`Incidental`/`Ambiguous` integers) to `taxonomy.PackageSummary` as an `omitempty` pointer field, so Go-native quality output (which reports classification via `analyze --classify`) does not emit the field.
- Populate `ClassificationCounts` in `adapter.BuildQualityFromMappings` by tallying merged side-effect classifications, de-duplicated by effect ID.
- Extend the `gaze-reporter` agent prompt and the `/gaze` command prompt with analyzer/language detection (`.gaze.yaml` `analyzers:` map first, then a project-marker heuristic) and thread `--analyzer`/`--language` through `crap`, `quality`, and `docscan`. Full Mode sources its Classification Summary from `classification_counts` for external projects.
- Declare `classification_counts` in the `QualitySchema` `PackageSummary` definition so machine-validated quality output reflects the new field.
- Make `gaze crap` and `gaze quality` honor `--language` alone (not just `--analyzer`) for external-analyzer dispatch, matching `gaze docscan`, so the agent's layered detection (which supplies `--language` when `.gaze.yaml` is absent) actually triggers the external path.
- Synchronize the third embedded `gaze-reporter.md` copy (`internal/aireport/assets/agents/gaze-reporter.md`, served by `gaze report --ai=opencode`), and document the new field in `docs/reference/json-schemas.md`.

## Capabilities

### New Capabilities

- `classification_counts surfacing`: `gaze quality --analyzer` JSON includes a module-wide `classification_counts` summary (contractual/incidental/ambiguous) derived from the external analyzer's `classify_signals` labels.

### Modified Capabilities

- `taxonomy.PackageSummary`: gains an optional `classification_counts` field.
- `adapter.BuildQualityFromMappings`: populates `classification_counts` from merged analysis results.
- `gaze-reporter` agent: detects analyzer/language and, in Full Mode, sources classification from `classification_counts` for external projects instead of `analyze --classify`.
- `/gaze` command: documents analyzer/language detection in its instructions.
- `QualitySchema`: `PackageSummary` definition declares `classification_counts`.

### Removed Capabilities

- None.

## Impact

### Files Modified

- `internal/taxonomy/types.go` — add `ClassificationCounts` struct and `ClassificationCounts *ClassificationCounts` field on `PackageSummary`
- `internal/adapter/quality.go` — add `countClassifications` helper; populate `summary.ClassificationCounts` in `BuildQualityFromMappings`
- `internal/adapter/quality_internal_test.go` — tests for `countClassifications` and `BuildQualityFromMappings` classification counts
- `internal/scaffold/assets/agents/gaze-reporter.md` — analyzer/language detection section; `--analyzer`/`--language` on crap/quality/docscan; external classification sourcing in Full Mode
- `internal/scaffold/assets/commands/gaze.md` — frontmatter description and analyzer/language detection note
- `.opencode/agents/gaze-reporter.md` — dogfood copy (byte-identical to scaffold asset, embed.FS sync invariant #262/#263)
- `.opencode/commands/gaze.md` — dogfood copy (byte-identical to scaffold asset)
- `internal/aireport/assets/agents/gaze-reporter.md` — embedded copy served by `gaze report --ai=opencode` (byte-identical to scaffold asset; `TestEmbeddedPromptMatchesScaffold` invariant)
- `cmd/gaze/main.go` — `runCrap`/`runQuality` dispatch honor `--language` (matching `runDocscan`); `initExternalSession` error message for language-only resolution
- `cmd/gaze/external_analyzer_test.go` — regression test proving `--language`-only triggers the external path
- `internal/quality/quality_test.go` — assert Go-native JSON omits `classification_counts`
- `internal/report/schema.go` — declare `classification_counts` in `QualitySchema` `$defs.PackageSummary`
- `internal/report/report_test.go` — schema validation test for `classification_counts` (presence and omission)
- `docs/reference/json-schemas.md` — document `classification_counts` in the `PackageSummary` field table

### Downstream Benefits

- External-analyzer users get an accurate classification distribution in `gaze quality` and `/gaze` Full Mode, matching what Go-native users already receive from `analyze --classify`.
- Quality JSON remains schema-validated; the schema now reflects the field the code emits.

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md` v1.3.0).

### I. Accuracy

**Assessment**: PASS

Classification counts are derived directly from the external analyzer's `classify_signals` labels already attached to side effects, and de-duplicated by effect ID to avoid double-counting merged effects. The `/gaze` Full Mode report stops reporting a degraded classification section for external projects and instead reflects the real labels, eliminating a known source of inaccurate output.

### II. Minimal Assumptions

**Assessment**: PASS

Analyzer/language detection is best-effort and layered: `.gaze.yaml` `analyzers:` map first, then a project-marker heuristic, falling back to the Go-native path when unresolved. `classification_counts` is an `omitempty` pointer field, so Go-native output is unchanged. No new mandatory dependency, annotation, or restructuring is introduced; analyzer discovery is delegated to gaze's existing three-tier mechanism.

### III. Actionable Output

**Assessment**: PASS

`classification_counts` is machine-parseable JSON (contractual/incidental/ambiguous integers) and is declared in the quality JSON schema. The counts feed the `/gaze` Full Mode Classification Summary table, guiding users to the same actionable remediation (reduce ambiguous rate, address incidental effects) that Go-native classification already enables.

### IV. Testability

**Assessment**: PASS

`countClassifications` is a pure data transformation tested with synthetic inputs (dedup by ID, nil/unclassified handling, label counting). `BuildQualityFromMappings` classification counts are asserted in its existing table-driven test. The schema change is covered by a validation test that exercises both presence and omission of `classification_counts`. All new code is testable in isolation; no external services required.
<!-- scaffolded by uf vdev -->
