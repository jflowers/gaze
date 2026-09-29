# Delta Spec: Doc-Signal Derivation

## ADDED Requirements

### Requirement: Derived signal carrier type

Doc-signal derivation SHALL produce a keyed carrier type, defined in `internal/adapter/docsignal/`, that binds a classification signal to the `(package, function, side_effect_type)` tuple it applies to:

```go
// DerivedSignal binds a taxonomy.Signal to the (package, function,
// side_effect_type) tuple it applies to.
type DerivedSignal struct {
    Package        string
    Function       string
    SideEffectType string
    Signal         taxonomy.Signal
}
```

The carrier type is internal to `docsignal`; `taxonomy.Signal` is NOT modified. At the `mergeClassifications` boundary, each `DerivedSignal` is reduced to its `Signal` field and grouped by its `(Package, Function, SideEffectType)` tuple alongside protocol-derived signals.

#### Scenario: DerivedSignal carries the tuple key

- **GIVEN** a derived signal for `ContainerMutation` declared contractual in package `mypackage`, function `process_items`
- **WHEN** the signal is produced
- **THEN** it MUST be a `DerivedSignal` with `Package: "mypackage"`, `Function: "process_items"`, `SideEffectType: "ContainerMutation"`, and a non-nil `Signal` carrying the source/weight/reasoning

---

### Requirement: Annotation grammar parsing

A doc-signal extractor MUST parse Markdown documentation files for a deterministic annotation grammar that declares side effects as contractual or incidental, keyed by the analyzer's own effect-type name.

The annotation grammar SHALL use the following syntax:

```markdown
<!-- gaze:contractual ContainerMutation -->
<!-- gaze:incidental CallbackInvocation -->
```

Alternatively, a fenced block variant SHALL be supported:

```markdown
> [!gaze-contract]
> - `ContainerMutation` — contractual (public API mutation)
> - `CallbackInvocation` — incidental (internal wiring)
```

The grammar is **type-level only**: an annotation names an effect-type and a `contractual`/`incidental` label, but no package or function. A type-level annotation is a **wildcard** that fans out to every `(package, function)` tuple in the analyzed effect list whose effect set contains the annotated type. Function-level precision is supplied by the sidecar (see "Sidecar file override").

The token grammar SHALL be:

- HTML comment form: `<!--` `gaze:` `contractual|incidental` `TypeName` `-->`, matched case-insensitively on the `gaze:` and label tokens, with exactly one type name (a non-whitespace token) between the label and the closing `-->`.
- Fenced block form: within a `> [!gaze-contract]` block, each subsequent `> - \`TypeName\` — label` line is a declaration; the human comment after the em-dash is ignored.

Comments in the fenced block are for human readers only; the extractor MUST key on the type name and the `contractual`/`incidental` label token.

The extractor MUST use analyzer-emitted type names verbatim — no translation or aliasing is performed. A reference to `ContainerMutation` in a doc MUST match only side effects whose `SideEffectType` is exactly `"ContainerMutation"`.

#### Scenario: Grammar annotations in design docs

- **GIVEN** a design doc `docs/design/containers.md` contains:
  ```markdown
  <!-- gaze:contractual ContainerMutation -->
  <!-- gaze:incidental CallbackInvocation -->
  ```
- **AND** the analyzed effect list contains `(mypackage, process_items, ContainerMutation)` and `(mypackage, process_items, CallbackInvocation)`
- **WHEN** the doc-signal extractor processes the document
- **THEN** it MUST produce two `DerivedSignal`s, one per tuple:
  - `{Package: "mypackage", Function: "process_items", SideEffectType: "ContainerMutation", Signal: {Source: "architecture_doc", SourceFile: "docs/design/containers.md", Weight: +25, Reasoning: "design doc declares ContainerMutation as contractual"}}`
  - `{Package: "mypackage", Function: "process_items", SideEffectType: "CallbackInvocation", Signal: {Source: "architecture_doc", SourceFile: "docs/design/containers.md", Weight: -25, Reasoning: "design doc declares CallbackInvocation as incidental"}}`

#### Scenario: Fenced block annotations

- **GIVEN** a README contains:
  ```markdown
  > [!gaze-contract]
  > - `ContainerMutation` — contractual (public API mutation)
  > - `CallbackInvocation` — incidental (internal wiring)
  ```
- **WHEN** the doc-signal extractor processes the document
- **THEN** it MUST produce the same two `DerivedSignal`s as the inline comment grammar for the same types, except `Signal.Source` is `"readme"` (the document basename is `README*`) with `Weight` ±15

#### Scenario: Conflicting annotations across documents

- **GIVEN** `pkg/mypackage/README.md` (PriorityOther) declares `ContainerMutation` as `contractual`
- **AND** `README.md` (PriorityModuleRoot) declares `ContainerMutation` as `incidental`
- **WHEN** the doc-signal extractor processes both documents
- **THEN** the PriorityModuleRoot document's signal MUST take precedence
- **AND** the resulting signal MUST be `{Signal: {Source: "readme", SourceFile: "README.md", Weight: -15, Reasoning: "readme declares ContainerMutation as incidental"}}`

#### Scenario: No annotations in document

- **GIVEN** a Markdown document with no `<!-- gaze:` comments and no `[!gaze-contract]` fenced blocks
- **WHEN** the doc-signal extractor processes the document
- **THEN** it MUST return an empty signal list (no error)

#### Scenario: Malformed annotations

- **GIVEN** a document containing `<!-- gaze:contractual -->` (missing effect type) or `<!-- gaze:invalidLabel ContainerMutation -->` (unknown label)
- **WHEN** the doc-signal extractor processes the document
- **THEN** the malformed annotation MUST be silently skipped
- **AND** remaining valid annotations in the same document MUST still be extracted

---

### Requirement: Sidecar file override

A sidecar file at `.uf/gaze/contracts.yaml` (or `.uf/gaze/contracts.json`) SHALL allow explicit `(package, function, side_effect_type) → label` declarations that override or extend grammar-derived signals. The sidecar format MUST be language-agnostic and project-specific.

The `function` key SHALL be the analyzer's verbatim reported function name (the value stored as `taxonomy.FunctionTarget.Function`), NOT the receiver-qualified name. Method receiver qualification is out of scope for this change.

The YAML format SHALL be:

```yaml
version: 1
contracts:
  - package: "mypackage"
    function: "process_items"
    side_effect_type: "ContainerMutation"
    label: contractual
    reasoning: "Design doc marks this as public API mutation"
  - package: "mypackage"
    function: "process_items"
    side_effect_type: "CallbackInvocation"
    label: incidental
    reasoning: "Internal wiring callback, not part of public contract"
```

JSON format SHALL be structurally equivalent.

Each sidecar entry MUST be semantically validated: an entry with an unknown `label` (anything other than `contractual`/`incidental`), an empty `package` or `function`, or a `side_effect_type` that is not a `taxonomy.IsKnownType` value MUST be skipped with a warning written to stderr. Remaining valid entries MUST still be loaded.

Sidecar signals MUST take precedence over grammar-derived signals. When both a grammar annotation (fanned out to a tuple) and a sidecar entry exist for the same `(package, function, side_effect_type)` tuple, the sidecar value MUST be used and the grammar-derived signal MUST be discarded for that tuple.

The sidecar file is optional. When no sidecar file exists at the expected path, the derivation step MUST proceed with grammar-derived signals only (no error).

#### Scenario: Sidecar overrides grammar annotation

- **GIVEN** a grammar annotation declares `ContainerMutation` as `contractual` with Weight +25
- **AND** `.uf/gaze/contracts.yaml` declares the same `(mypackage, process_items, ContainerMutation)` tuple as `incidental`
- **WHEN** the doc-signal extractor merges signals
- **THEN** the sidecar signal MUST replace the grammar-derived signal
- **AND** the resulting signal MUST be `{Signal: {Source: "sidecar", SourceFile: ".uf/gaze/contracts.yaml", Weight: -30, Reasoning: "sidecar declares ContainerMutation as incidental"}}`

#### Scenario: Sidecar extends grammar

- **GIVEN** a grammar annotation declares `ContainerMutation` as `contractual`
- **AND** `.uf/gaze/contracts.yaml` declares `StreamOutput` as `incidental` (not mentioned in any grammar annotation)
- **WHEN** the doc-signal extractor merges signals
- **THEN** both signals MUST be present in the result
- **AND** no signal conflict resolution is needed (they cover different tuples)

#### Scenario: Missing sidecar file

- **GIVEN** no `.uf/gaze/contracts.yaml` or `.uf/gaze/contracts.json` exists
- **WHEN** the doc-signal extractor runs the derivation step
- **THEN** no error MUST be returned
- **AND** grammar-derived signals MUST be returned unchanged

#### Scenario: Invalid sidecar file

- **GIVEN** `.uf/gaze/contracts.yaml` exists but contains malformed YAML
- **WHEN** the doc-signal extractor loads the sidecar
- **THEN** a warning MUST be written to stderr
- **AND** the derivation step MUST proceed with grammar-derived signals only (graceful degradation)

#### Scenario: Semantically invalid sidecar entry

- **GIVEN** `.uf/gaze/contracts.yaml` contains one entry with `label: "unknown"` and one valid entry
- **WHEN** the doc-signal extractor loads the sidecar
- **THEN** a warning MUST be written to stderr for the invalid entry
- **AND** the invalid entry MUST be skipped
- **AND** the valid entry MUST still be loaded

---

### Requirement: Doc-signal derivation orchestration

The `DeriveSignals` function in `internal/adapter/docsignal/` MUST orchestrate the full derivation pipeline: parse grammar annotations from `docscan.DocumentFile` values, fan them out to per-function tuples using the analyzed effect list, load and merge sidecar overrides, and return `[]DerivedSignal`.

```go
func DeriveSignals(
    docs []docscan.DocumentFile,
    cached []taxonomy.AnalysisResult,
    moduleRoot string,
) []DerivedSignal
```

The `cached` parameter is the analyzer's full result set; it supplies the `(package, function)` tuples and effect types that type-level grammar annotations fan out to. Sidecar entries produce full tuples directly and do not require a matching entry in `cached`.

The derivation step MUST be scoped to external analyzers only. Go-native projects use `internal/classify/godoc.go` and are not affected by this change.

#### Scenario: Full derivation pipeline with grammar and sidecar

- **GIVEN** a `docscan.DocumentFile` list with 3 Markdown docs, two containing grammar annotations
- **AND** an analyzed result set containing the annotated effect types
- **AND** a `.uf/gaze/contracts.yaml` with one override and one new entry
- **WHEN** `DeriveSignals(docs, cached, moduleRoot)` is called
- **THEN** grammar annotations from both annotated docs MUST be extracted and fanned out to per-function tuples
- **AND** sidecar entries MUST be merged with precedence over grammar
- **AND** the result MUST contain the union of all unique `(package, function, side_effect_type)` tuples

#### Scenario: Derivation with no docs

- **GIVEN** an empty `docscan.DocumentFile` list (no Markdown files in the project)
- **WHEN** `DeriveSignals(docs, cached, moduleRoot)` is called
- **THEN** an empty signal list MUST be returned (no error)

#### Scenario: Derivation with no annotated docs

- **GIVEN** a `docscan.DocumentFile` list where no document contains grammar annotations
- **AND** no sidecar file exists
- **WHEN** `DeriveSignals(docs, cached, moduleRoot)` is called
- **THEN** an empty signal list MUST be returned (no error)

---

### Requirement: Adapter integration

The `ExternalSideEffectAnalyzer` MUST run the doc-signal derivation step unconditionally after `analyze`/`analyze_stream` populates the cached results, regardless of whether the analyzer supports `classify_signals`. Protocol `classify_signals` is fetched only when the capability is present. The doc-derived signals MUST be passed to `mergeClassifications` alongside any `classify_signals` protocol signals.

Doc-derived signals and protocol-derived signals SHALL be combined additively — both signal sets are passed to `classify.ComputeScore`, which sums weights. There is no conflict resolution between doc signals and protocol signals (the scorer handles weight summation naturally).

#### Scenario: Doc signals supplement protocol signals

- **GIVEN** an external analyzer supports `classify_signals: true` and returns a type-annotation signal for `(mypackage, process_items, ContainerMutation)` with Weight +30
- **AND** docscan discovers a design doc declaring `ContainerMutation` as `contractual` (Weight +25)
- **WHEN** the adapter processes classification
- **THEN** both signals MUST be passed to `mergeClassifications`
- **AND** `classify.ComputeScore` MUST receive both signals for the same `(mypackage, process_items, ContainerMutation)` tuple
- **AND** the combined confidence score MUST reflect the sum of both weights

#### Scenario: Doc signals when analyzer lacks classify_signals

- **GIVEN** an external analyzer does not support `classify_signals` (capability flag is `false` or absent)
- **AND** docscan discovers annotated design docs
- **WHEN** the adapter processes classification
- **THEN** the doc-signal derivation MUST still run
- **AND** only doc-derived signals MUST be passed to `mergeClassifications`
- **AND** `classify.ComputeScore` MUST still produce classification labels based on doc signals alone

#### Scenario: No doc signals available

- **GIVEN** an external analyzer is running against a project with no Markdown docs
- **WHEN** the adapter processes classification
- **THEN** `mergeClassifications` MUST receive an empty or nil doc-signal list
- **AND** classification MUST proceed identically to current behavior (no change)

---

### Requirement: Signal weight conventions

Doc-derived signals SHALL carry the following weight conventions, determined **by document filename** (not by priority tier):

| Source | Contractual weight | Incidental weight |
|---|---|---|
| `architecture_doc` (any non-README document) | +25 | -25 |
| `readme` (document basename `README*`, case-insensitive) | +15 | -15 |
| `sidecar` | +30 | -30 |

`Signal.Source` SHALL be the bare source token (`"architecture_doc"`, `"readme"`, or `"sidecar"`); the document path SHALL be carried in `Signal.SourceFile`. This matches the existing `taxonomy.Signal` convention (source type vs. source file).

The `sidecar` weight (+30/-30) SHALL be higher than any grammar-derived weight to reflect the explicit, intentional nature of sidecar declarations.

> **Note**: a single grammar-derived signal on a P2-P4 effect (base confidence 50) yields 75 — still `ambiguous` at the default contractual threshold of 80. Only sidecar (±30) or multiple grammar signals cross that threshold for P2-P4 effects. This is a known calibration limitation; the weights are configurable indirectly via the existing `.gaze.yaml` classification thresholds.

#### Scenario: Architecture doc weight for non-README doc

- **GIVEN** a document at `mypackage/DESIGN.md` (basename `DESIGN.md`, not `README*`)
- **WHEN** the document contains a `<!-- gaze:contractual ContainerMutation -->` annotation
- **THEN** the signal MUST have `Source: "architecture_doc"` and Weight +25

#### Scenario: README weight for root-level doc

- **GIVEN** a document at `README.md` (basename `README.md`)
- **WHEN** the document contains a `<!-- gaze:incidental StreamOutput -->` annotation
- **THEN** the signal MUST have `Source: "readme"` and Weight -15

#### Scenario: Sidecar weight

- **GIVEN** `.uf/gaze/contracts.yaml` declares `ContainerMutation` as `contractual`
- **WHEN** the sidecar signal is produced
- **THEN** the signal MUST have `Source: "sidecar"` and Weight +30

---

## MODIFIED Requirements

### Requirement: mergeClassifications signature

The `mergeClassifications` function in `internal/adapter/classify.go` SHALL accept an additional parameter for doc-derived signals. The existing parameters are preserved unchanged.

Previous signature:

```go
mergeClassifications(cached []taxonomy.AnalysisResult, signals []protocol.ClassifySignalData, cfg *config.GazeConfig)
```

New signature:

```go
mergeClassifications(
    cached []taxonomy.AnalysisResult,
    signals []protocol.ClassifySignalData,
    docSignals []docsignal.DerivedSignal,
    cfg *config.GazeConfig,
)
```

Doc signals SHALL be grouped by `(Package, Function, SideEffectType)` alongside protocol signals (reducing each `DerivedSignal` to its `Signal` field), and both groups SHALL be passed to `classify.ComputeScore`. The function MUST NOT distinguish between signal sources during scoring — all signals for a given tuple are passed together. The existing empty-signal early return (`if len(signals) == 0 { return }`) MUST be extended to also account for `docSignals` (e.g., `if len(signals) == 0 && len(docSignals) == 0 { return }`), so that doc-signal-only classification proceeds when the analyzer does not advertise `classify_signals`.

#### Scenario: Merged signals passed to scorer

- **GIVEN** `signals` contains `{Package: "pkg", Function: "Foo", SideEffectType: "ContainerMutation", Source: "type_annotation", Weight: 30}`
- **AND** `docSignals` contains `{Package: "pkg", Function: "Foo", SideEffectType: "ContainerMutation", Signal: {Source: "architecture_doc", SourceFile: "DESIGN.md", Weight: 25}}` for the same tuple
- **WHEN** `mergeClassifications` is called
- **THEN** both signals MUST be passed to `classify.ComputeScore` in the same `[]taxonomy.Signal` slice
- **AND** the scorer MUST sum weights: 30 + 25 = 55

---

## REMOVED Requirements

None.
