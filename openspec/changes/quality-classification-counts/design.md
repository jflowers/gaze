## Context

External analyzers already classify side effects: `ExternalSideEffectAnalyzer` calls `classify_signals` (spec 246) and attaches a `taxonomy.Classification` to each cached `taxonomy.SideEffect`. But `adapter.BuildQualityFromMappings` — the function that converts `test_mapping` protocol output into quality reports — drops that information on the floor. The returned `taxonomy.PackageSummary` has no classification summary, so `gaze quality --analyzer` and `gaze report` cannot surface the distribution of contractual/incidental/ambiguous effects.

Separately, the `gaze-reporter` agent's `/gaze` Full Mode derives its Classification Summary from `gaze analyze --classify`, which is Go-only (design D12: `analyze` does not register `--analyzer`/`--language`). External-analyzer projects therefore get a degraded classification section even though the classification data already exists in the `quality --analyzer` JSON.

This change closes both gaps with minimal surface area: one new taxonomy field, one tally helper, prompt updates, and a schema declaration.

## Goals / Non-Goals

### Goals

- Surface a module-wide `classification_counts` summary on the quality package summary for external analyzers.
- Tally classifications correctly (dedup by effect ID, ignore unclassified, nil when nothing classified).
- Teach `gaze-reporter` to detect the analyzer/language and source Full Mode classification from `classification_counts` for external projects.
- Keep `QualitySchema` in sync with the new field.

### Non-Goals

- Changing how `ComputeContractCoverage` or contract coverage is computed (classification already influences coverage; this only surfaces counts).
- Emitting `classification_counts` for Go-native quality (Go reports classification via `analyze --classify`; the field stays nil/absent).
- Adding analyzer detection to the `gaze analyze` command itself (out of scope; design D12 stands).
- Any change to the external analyzer protocol (no new methods, no protocol version bump).

## Decisions

### D1: `classification_counts` as a pointer `omitempty` field on `PackageSummary`

`ClassificationCounts` is added as `*ClassificationCounts` with `json:"classification_counts,omitempty"`. Go-native quality never populates it, so the field is absent from Go-native JSON.

**Rationale**: The field is only meaningful for external analyzers. A nil pointer + `omitempty` keeps Go-native output byte-for-byte unchanged, avoiding churn in existing golden outputs and the Go-native schema tests. This mirrors the existing convention where external-only signals (e.g. the `classification,omitempty` field on `SideEffect`) are optional.

### D2: Tally in `adapter.BuildQualityFromMappings`, not in `buildQualitySummary`

The count is computed from `results []taxonomy.AnalysisResult` (which carries the `classification` data), not from the already-reduced `[]taxonomy.QualityReport`. The assignment happens in `BuildQualityFromMappings` after `buildQualitySummary` returns.

**Rationale**: `buildQualitySummary` only receives `[]taxonomy.QualityReport`, which does not carry per-effect classification. The classification data lives on `results`, already a parameter of `BuildQualityFromMappings`. Computing there avoids widening `buildQualitySummary`'s signature or duplicating the tally in callers.

### D3: Dedup by side-effect ID via a `seen` map

`countClassifications` maintains a `seen map[string]bool` keyed on `SideEffect.ID`, incrementing a label bucket only the first time an ID is seen.

**Rationale**: The merged `results` may contain the same effect more than once (e.g. the analyzer returns the same effect across batches, or `mergeClassifications` fans out a label to multiple matching effects). Counting by unique ID prevents the distribution from inflating. Unclassified effects (`Classification == nil`) are skipped before the dedup check, and a `classified` bool tracks whether any effect was classified so the helper can return nil for the all-unclassified case.

### D4: Analyzer/language detection is best-effort and layered

The `gaze-reporter` prompt adds an "Analyzer & Language Detection" section: read `.gaze.yaml` `analyzers:` map first (language = map key, command = its `command`), fall back to a project-marker heuristic, and default to the Go-native path when unresolved. Detection produces the `--language`/`--analyzer` flags; actual plugin resolution is delegated to gaze's existing three-tier discovery (CLI flag → `.gaze.yaml` → `gaze-analyzer-<language>` on `$PATH`).

**Rationale**: The agent must not reimplement analyzer discovery. Gaze already owns three-tier discovery; the prompt only needs to supply the flags that trigger it. Layering `.gaze.yaml` before the heuristic honors explicit user intent while keeping the no-config case working.

### D5: Full Mode classification sourced from `classification_counts` for external projects

Full Mode for external projects skips `analyze --classify` and builds the Classification Summary table from `classification_counts` in the `quality --analyzer` JSON; Go projects retain `analyze --classify`.

**Rationale**: `analyze --classify` is Go-only and would fail or return empty for external projects. `classification_counts` already contains the merged `classify_signals` labels, which is exactly what the table needs. Go projects keep the existing (correct) baseline.

### D6: Schema mirrors the field, not required

`classification_counts` is added to `QualitySchema.$defs.PackageSummary.properties` as an inline object (type object, integer properties `contractual`/`incidental`/`ambiguous`) but NOT to the `required` array.

**Rationale**: The project convention is that every `PackageSummary` field is mirrored into `QualitySchema`. Because the field is `omitempty` (absent for Go-native), it cannot be `required`. This matches how `ssa_degraded`, `skipped_tests`, and `skipped_test_names` are declared (present in properties, absent from required).

### D7: `crap` and `quality` honor `--language` alone for external dispatch

`runCrap` and `runQuality` switch their dispatch condition from `p.analyzerFlag != ""` to `p.analyzerFlag != "" || p.languageFlag != ""`, matching `runDocscan`. `initExternalSession` gains a language-specific error message when resolution fails without an `--analyzer` flag.

**Rationale**: The agent's layered detection supplies only `--language` when `.gaze.yaml` is absent, relying on gaze's tier-3 PATH-convention discovery (`gaze-analyzer-<language>`). Without this change `crap`/`quality` silently fall through to the Go-native path and never produce `classification_counts`, defeating the change's headline scenario for heuristic-detected projects.

## Coverage Strategy

Unit tests for `countClassifications` in `quality_internal_test.go` (dedup by ID, unclassified ignored, nil-when-unclassified) plus `BuildQualityFromMappings` assertions on the returned `ClassificationCounts` (2/1/0). A schema validation test in `report_test.go` exercises `classification_counts` both present and omitted against `QualitySchema`. `TestWriteJSON_Structure` asserts Go-native JSON omits `classification_counts`. `TestCrapWithLanguageOnly_TriggersExternalPath` proves `--language`-only dispatch reaches the external path (not the Go-native fallthrough). No e2e tests needed — the tally is pure data transformation and the schema test uses the existing `santhosh-tekuri/jsonschema` harness.

## Risks / Trade-offs

### R1: Nil-vs-zero ambiguity for analyzers with zero classified effects

An analyzer that returns only unclassified effects yields a nil `classification_counts` (field absent), which is indistinguishable from "no classification support." This is acceptable: both cases render "classification unavailable" in the report, and the `reason`/`assertion` fields already disambiguate degradation when needed.

### R2: Heuristic mis-detection in the agent prompt

The project-marker heuristic may mis-detect a polyglot repo. This is mitigated by `.gaze.yaml` `analyzers:` taking precedence and by the fallback to Go-native when unresolved. Mis-detection only affects report sourcing, never analysis correctness.

### R3: Dogfood drift (three copies)

Three `gaze-reporter.md` copies must stay byte-identical: the scaffold asset (`internal/scaffold/assets/agents/gaze-reporter.md`), the dogfood copy (`.opencode/agents/gaze-reporter.md`), and the embedded copy served by `gaze report --ai=opencode` (`internal/aireport/assets/agents/gaze-reporter.md`, `//go:embed` as `defaultPromptRaw`). `TestEmbeddedPromptMatchesScaffold` enforces the scaffold↔embedded pair, and `TestEmbeddedAssetsMatchSource` enforces the scaffold↔dogfood pair. This change updates all three identically.

### R4: Schema permissiveness

`QualitySchema` has no `additionalProperties: false`, so the missing `classification_counts` declaration would not break tests; adding it is a documentation-correctness improvement, not a bug fix. It is included here to satisfy the project convention that types.go fields are mirrored into the schema.
<!-- scaffolded by uf vdev -->
