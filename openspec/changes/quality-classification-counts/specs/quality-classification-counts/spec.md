## ADDED Requirements

### Requirement: classification_counts on PackageSummary

`taxonomy.PackageSummary` MUST carry an optional `ClassificationCounts` field serialized as `classification_counts` with `omitempty`. The field MUST be a pointer so Go-native quality output — which reports classification via `gaze analyze --classify` rather than this field — omits it entirely.

`ClassificationCounts` MUST be a struct with three integer fields: `Contractual`, `Incidental`, and `Ambiguous`, serialized as `contractual`, `incidental`, and `ambiguous` respectively.

#### Scenario: External analyzer emits classification_counts
- **GIVEN** a `PackageSummary` produced by `BuildQualityFromMappings` from an external analyzer that returned `classify_signals` labels
- **WHEN** the summary is serialized to JSON
- **THEN** it MUST include a `classification_counts` object with `contractual`, `incidental`, and `ambiguous` integer counts

#### Scenario: Go-native summary omits classification_counts
- **GIVEN** a `PackageSummary` produced by the Go-native quality path (no external analyzer)
- **WHEN** the summary is serialized to JSON
- **THEN** the `classification_counts` key MUST be absent (nil pointer + `omitempty`)

### Requirement: Count classifications from merged results

A `countClassifications` helper MUST tally side effects by classification label across `[]taxonomy.AnalysisResult`, de-duplicating by side-effect ID so a merged effect is counted once. Effects with a nil `Classification` MUST NOT be counted in any bucket. The helper MUST return nil when no effect was classified.

#### Scenario: Dedupe by ID
- **GIVEN** analysis results whose merged side effects contain two entries with the same ID and a `contractual` classification
- **WHEN** `countClassifications` runs
- **THEN** the `contractual` count MUST be 1, not 2

#### Scenario: Unclassified effects are ignored
- **GIVEN** analysis results containing both classified and unclassified (nil `Classification`) side effects
- **WHEN** `countClassifications` runs
- **THEN** only the classified effects MUST be tallied; unclassified effects MUST NOT affect any bucket

#### Scenario: Nothing classified returns nil
- **GIVEN** analysis results where every side effect has a nil `Classification`, or no results at all
- **WHEN** `countClassifications` runs
- **THEN** it MUST return nil

### Requirement: BuildQualityFromMappings populates classification_counts

`adapter.BuildQualityFromMappings` MUST set `summary.ClassificationCounts` from the provided `results` using `countClassifications` before returning the summary.

#### Scenario: Classification counts flow into the summary
- **GIVEN** `BuildQualityFromMappings` is called with mappings and analysis results whose classifications are 2 contractual, 1 incidental, 0 ambiguous
- **WHEN** the returned `*taxonomy.PackageSummary` is inspected
- **THEN** `ClassificationCounts` MUST be non-nil with `Contractual=2`, `Incidental=1`, `Ambiguous=0`

### Requirement: QualitySchema declares classification_counts

The `QualitySchema` `$defs.PackageSummary` definition MUST declare a `classification_counts` property of type object with integer properties `contractual`, `incidental`, and `ambiguous`. The property MUST NOT be added to the `required` array, so output that omits it (Go-native) continues to validate.

#### Scenario: Schema validates classification_counts presence
- **GIVEN** quality JSON whose `quality_summary` includes a `classification_counts` object
- **WHEN** the JSON is validated against `QualitySchema`
- **THEN** validation MUST succeed

#### Scenario: Schema validates classification_counts omission
- **GIVEN** quality JSON whose `quality_summary` omits `classification_counts`
- **WHEN** the JSON is validated against `QualitySchema`
- **THEN** validation MUST succeed (field not required)

## MODIFIED Requirements

### Requirement: gaze-reporter analyzer/language detection

The `gaze-reporter` agent MUST determine, before running any gaze command, whether the project uses an external analyzer, by reading the `.gaze.yaml` `analyzers:` map first and falling back to a project-marker heuristic (`pyproject.toml`/`setup.py`/`*.py` ⇒ `python`, `go.mod` ⇒ `go`, `package.json` + `tsconfig.json` ⇒ `typescript`, `Cargo.toml` ⇒ `rust`). For external languages it MUST pass `--language <lang>` (and `--analyzer <command>` when `.gaze.yaml` names a non-convention command) to `crap`, `quality`, and `docscan`. For `go` or unresolved projects it MUST use the Go-native path.

Previously: the agent ran `crap`, `quality`, and `docscan` without `--analyzer`/`--language` flags.

#### Scenario: .gaze.yaml names an analyzer
- **GIVEN** a `.gaze.yaml` with `analyzers: {python: {command: snake-eyes, args: ["--stdio"]}}`
- **WHEN** the agent prepares its gaze commands
- **THEN** it MUST pass `--language python --analyzer snake-eyes` to `crap`, `quality`, and `docscan`

#### Scenario: Heuristic detects a Go project
- **GIVEN** a project with a `go.mod` and no `.gaze.yaml` `analyzers:` map
- **WHEN** the agent prepares its gaze commands
- **THEN** it MUST use the Go-native path (no `--analyzer`/`--language` flags)

### Requirement: Full Mode classification sourcing

In Full Mode, the `gaze-reporter` agent MUST source the classification distribution from the `classification_counts` field of the `quality --analyzer` JSON for external-analyzer projects, and from `gaze analyze --classify` for Go projects. For external projects it MUST NOT run `analyze --classify`.

Previously: Full Mode always ran `analyze --classify` for the classification baseline.

#### Scenario: External project classification
- **GIVEN** an external-analyzer project
- **WHEN** the agent runs Full Mode
- **THEN** it MUST skip `analyze --classify`
- **AND** it MUST build the Classification Summary table from `classification_counts` in the `quality --analyzer` JSON

#### Scenario: Go project classification
- **GIVEN** a Go project
- **WHEN** the agent runs Full Mode
- **THEN** it MUST run `analyze --classify` and build the Classification Summary from that output

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
